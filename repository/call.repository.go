package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"linkup/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrCallBusy    = errors.New("người dùng đang bận")
	ErrCallNotFound = errors.New("cuộc gọi không tồn tại")
)

type CallRepository struct {
	db *gorm.DB
}

func NewCallRepository(db *gorm.DB) *CallRepository {
	return &CallRepository{db: db}
}

func (r *CallRepository) Create(ctx context.Context, call *models.Call) error {
	tx := r.db.WithContext(ctx).Create(call)
	if tx.Error != nil {
		return fmt.Errorf("create call: %w", tx.Error)
	}
	return nil
}

func (r *CallRepository) CreateIfNotBusy(ctx context.Context, call *models.Call) (bool, error) {
	var isBusy bool
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Raw(
			`SELECT COUNT(*) FROM calls
			 WHERE (caller_id = ? OR callee_id = ?)
			   AND status IN (?, ?, ?)
			 FOR UPDATE`,
			call.CalleeID, call.CalleeID,
			models.CallStatusCalling,
			models.CallStatusRinging,
			models.CallStatusConnected,
		).Scan(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			isBusy = true
			return nil
		}
		return tx.Create(call).Error
	})
	if err != nil {
		return false, fmt.Errorf("create call: %w", err)
	}
	return isBusy, nil
}

func (r *CallRepository) FindByID(ctx context.Context, id string) (*models.Call, error) {
	var call models.Call
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&call).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find call by id: %w", err)
	}
	return &call, nil
}

func (r *CallRepository) FindActiveByUserID(ctx context.Context, userID string) (*models.Call, error) {
	return r.FindActiveByUser(ctx, userID)
}

func (r *CallRepository) FindActiveByUser(ctx context.Context, userID string) (*models.Call, error) {
	var call models.Call
	err := r.db.WithContext(ctx).
		Where("(caller_id = ? OR callee_id = ?) AND status IN ?", userID, userID, []models.CallStatus{
			models.CallStatusCalling,
			models.CallStatusRinging,
			models.CallStatusConnected,
		}).
		Order("created_at DESC").
		First(&call).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find active call by user: %w", err)
	}
	return &call, nil
}

func (r *CallRepository) FindActiveBetween(ctx context.Context, userA, userB string) (*models.Call, error) {
	var call models.Call
	err := r.db.WithContext(ctx).
		Where("((caller_id = ? AND callee_id = ?) OR (caller_id = ? AND callee_id = ?)) AND status IN ?",
			userA, userB, userB, userA,
			[]models.CallStatus{
				models.CallStatusCalling,
				models.CallStatusRinging,
				models.CallStatusConnected,
			}).
		Order("created_at DESC").
		First(&call).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find active call between users: %w", err)
	}
	return &call, nil
}

// UpdateStatus updates the status and optional timestamp/duration fields of a call.
// Returns ErrCallNotFound if no call with the given ID exists.
func (r *CallRepository) UpdateStatus(ctx context.Context, id string, status models.CallStatus, startedAt, endedAt *time.Time, duration int) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if startedAt != nil {
		updates["started_at"] = *startedAt
	}
	if endedAt != nil {
		updates["ended_at"] = *endedAt
	}
	if duration > 0 {
		updates["duration"] = duration
	}
	tx := r.db.WithContext(ctx).Model(&models.Call{}).Where("id = ?", id).Updates(updates)
	if tx.Error != nil {
		return fmt.Errorf("update call status: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return ErrCallNotFound
	}
	return nil
}

// AcceptCallAtomic atomically transitions a call from calling/ringing to connected.
// This eliminates the TOCTOU race between FindByID + status check + UpdateStatus:
// the UPDATE itself performs the status check via WHERE clause, so if a concurrent
// request changes the status between our read and write, RowsAffected == 0.
// Returns the updated call on success, or an error describing the failure.
func (r *CallRepository) AcceptCallAtomic(ctx context.Context, callID, userID string) (*models.Call, error) {
	var call *models.Call
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Atomic conditional UPDATE: only transitions if current status is
		// calling or ringing AND the user is the callee.
		result := tx.Model(&models.Call{}).
			Where("id = ? AND callee_id = ? AND status IN (?, ?)",
				callID, userID,
				models.CallStatusCalling,
				models.CallStatusRinging,
			).
			Updates(map[string]interface{}{
				"status":     models.CallStatusConnected,
				"started_at": time.Now().UTC(),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			// Could be "not found" or "wrong status" — query to determine which.
			var findErr error
			call, findErr = r.findByIdTx(tx, callID)
			if findErr != nil {
				return findErr
			}
			if call == nil {
				return ErrCallNotFound
			}
			return errors.New("cuộc gọi không ở trạng thái chờ")
		}
		// Fetch the updated call to return caller/callee IDs for WS events.
		var findErr error
		call, findErr = r.findByIdTx(tx, callID)
		return findErr
	})
	if err != nil {
		return nil, err
	}
	return call, nil
}

// RejectCallAtomic atomically transitions a call from calling/ringing to rejected.
// Same TOCTOU protection as AcceptCallAtomic.
func (r *CallRepository) RejectCallAtomic(ctx context.Context, callID, userID string) (*models.Call, error) {
	var call *models.Call
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.Call{}).
			Where("id = ? AND callee_id = ? AND status IN (?, ?)",
				callID, userID,
				models.CallStatusCalling,
				models.CallStatusRinging,
			).
			Updates(map[string]interface{}{
				"status":   models.CallStatusRejected,
				"ended_at": time.Now().UTC(),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			var findErr error
			call, findErr = r.findByIdTx(tx, callID)
			if findErr != nil {
				return findErr
			}
			if call == nil {
				return ErrCallNotFound
			}
			return errors.New("cuộc gọi không ở trạng thái chờ")
		}
		var findErr error
		call, findErr = r.findByIdTx(tx, callID)
		return findErr
	})
	if err != nil {
		return nil, err
	}
	return call, nil
}

// findByIdTx queries a call within an existing transaction.
func (r *CallRepository) findByIdTx(tx *gorm.DB, id string) (*models.Call, error) {
	var call models.Call
	err := tx.Where("id = ?", id).First(&call).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &call, nil
}

// UpdateMuted updates mute state for one or both participants.
// Returns ErrCallNotFound if no call with the given ID exists.
func (r *CallRepository) UpdateMuted(ctx context.Context, id string, mutedCaller, mutedCallee *bool) error {
	updates := map[string]interface{}{}
	if mutedCaller != nil {
		updates["muted_caller"] = *mutedCaller
	}
	if mutedCallee != nil {
		updates["muted_callee"] = *mutedCallee
	}
	if len(updates) == 0 {
		return nil
	}
	tx := r.db.WithContext(ctx).Model(&models.Call{}).Where("id = ?", id).Updates(updates)
	if tx.Error != nil {
		return fmt.Errorf("update call mute: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return ErrCallNotFound
	}
	return nil
}

// UpdateVideoEnabled updates video state for one or both participants.
// Returns ErrCallNotFound if no call with the given ID exists.
func (r *CallRepository) UpdateVideoEnabled(ctx context.Context, id string, videoCaller, videoCallee *bool) error {
	updates := map[string]interface{}{}
	if videoCaller != nil {
		updates["video_enabled_caller"] = *videoCaller
	}
	if videoCallee != nil {
		updates["video_enabled_callee"] = *videoCallee
	}
	if len(updates) == 0 {
		return nil
	}
	tx := r.db.WithContext(ctx).Model(&models.Call{}).Where("id = ?", id).Updates(updates)
	if tx.Error != nil {
		return fmt.Errorf("update call video: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return ErrCallNotFound
	}
	return nil
}

func (r *CallRepository) GetHistory(ctx context.Context, userID string, limit, offset int) ([]models.Call, error) {
	var calls []models.Call
	err := r.db.WithContext(ctx).
		Where("caller_id = ? OR callee_id = ?", userID, userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&calls).Error
	if err != nil {
		return nil, fmt.Errorf("get call history: %w", err)
	}
	return calls, nil
}

func (r *CallRepository) CountHistory(ctx context.Context, userID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&models.Call{}).
		Where("caller_id = ? OR callee_id = ?", userID, userID).
		Count(&total).Error
	if err != nil {
		return 0, fmt.Errorf("count call history: %w", err)
	}
	return total, nil
}

// CallHistoryFilter holds optional filters for querying call history.
type CallHistoryFilter struct {
	CallType *string // "voice" | "video"
	Status   *string // "missed" | "ended" | "rejected"
	Sort     string  // "created_at" | "duration"
	Order    string  // "asc" | "desc"
	Limit    int
	Offset   int
}

// allowedSortColumns whitelist prevents SQL injection in ORDER BY.
var allowedSortColumns = map[string]bool{
	"created_at": true,
	"duration":   true,
	"call_type":  true,
	"status":     true,
}

// applyFilter builds a GORM query with user involvement, hidden-call exclusion,
// optional type/status filters, sort, and pagination.
func (r *CallRepository) applyFilter(ctx context.Context, userID string, f CallHistoryFilter) *gorm.DB {
	// LEFT JOIN call_hidden to exclude hidden calls — more efficient than NOT IN subquery.
	tx := r.db.WithContext(ctx).
		Model(&models.Call{}).
		Joins("LEFT JOIN call_hidden ON calls.id = call_hidden.call_id AND call_hidden.user_id = ?", userID).
		Where("calls.caller_id = ? OR calls.callee_id = ?", userID, userID).
		Where("call_hidden.call_id IS NULL")

	if f.CallType != nil {
		tx = tx.Where("calls.call_type = ?", *f.CallType)
	}
	if f.Status != nil {
		tx = tx.Where("calls.status = ?", *f.Status)
	}

	// Whitelist sort column to prevent SQL injection.
	sortCol := "calls.created_at"
	if f.Sort != "" && allowedSortColumns[f.Sort] {
		sortCol = "calls." + f.Sort
	}

	// Whitelist order direction to prevent SQL injection.
	orderClause := clause.OrderByColumn{
		Column: clause.Column{Name: sortCol},
		Desc:   strings.ToUpper(f.Order) != "ASC",
	}
	tx = tx.Clauses(clause.OrderBy{Columns: []clause.OrderByColumn{orderClause}})

	if f.Limit > 0 {
		tx = tx.Limit(f.Limit)
	}
	if f.Offset > 0 {
		tx = tx.Offset(f.Offset)
	}
	return tx
}

// GetHistoryFiltered returns paginated call history with optional filters,
// excluding calls the user has hidden.
func (r *CallRepository) GetHistoryFiltered(ctx context.Context, userID string, f CallHistoryFilter) ([]models.Call, error) {
	var calls []models.Call
	err := r.applyFilter(ctx, userID, f).Find(&calls).Error
	if err != nil {
		return nil, fmt.Errorf("get call history filtered: %w", err)
	}
	return calls, nil
}

// CountHistoryFiltered returns total count matching the same filters.
func (r *CallRepository) CountHistoryFiltered(ctx context.Context, userID string, f CallHistoryFilter) (int64, error) {
	var total int64
	err := r.applyFilter(ctx, userID, f).Count(&total).Error
	if err != nil {
		return 0, fmt.Errorf("count call history filtered: %w", err)
	}
	return total, nil
}

// CountMissedSince returns the number of missed calls for the user
// that occurred at or after the given timestamp.
// Uses >= (not >) to avoid a race where a missed call created at the same
// millisecond as MarkMissedRead would be skipped.
func (r *CallRepository) CountMissedSince(ctx context.Context, userID string, since *time.Time) (int64, error) {
	tx := r.db.WithContext(ctx).
		Model(&models.Call{}).
		Joins("LEFT JOIN call_hidden ON calls.id = call_hidden.call_id AND call_hidden.user_id = ?", userID).
		Where("calls.callee_id = ? AND calls.status = ?", userID, models.CallStatusMissed).
		Where("call_hidden.call_id IS NULL")
	if since != nil {
		tx = tx.Where("calls.created_at >= ?", *since)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count missed calls: %w", err)
	}
	return total, nil
}

// FindActiveByParticipant returns the most recent call for the user that is
// still in progress (calling/ringing/connected). Used to clean up calls
// abandoned when a participant's connection drops.
func (r *CallRepository) FindActiveByParticipant(ctx context.Context, userID string) (*models.Call, error) {
	var call models.Call
	err := r.db.WithContext(ctx).
		Where("(caller_id = ? OR callee_id = ?) AND status IN (?, ?, ?)", userID, userID,
			models.CallStatusCalling, models.CallStatusRinging, models.CallStatusConnected).
		Order("created_at DESC").
		First(&call).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find active call: %w", err)
	}
	return &call, nil
}

// MarkMissedRead sets last_read_missed_at on the user's profile to now,
// so subsequent CountMissedSince only counts calls after this moment.
func (r *CallRepository) MarkMissedRead(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	tx := r.db.WithContext(ctx).
		Model(&models.Profile{}).
		Where("user_id = ?", userID).
		Update("last_read_missed_at", now)
	if tx.Error != nil {
		return fmt.Errorf("mark missed read: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("profile not found for user %s", userID)
	}
	return nil
}

// HideCall inserts a call_hidden row so the call no longer appears
// in that user's history. The call remains visible to the other party.
func (r *CallRepository) HideCall(ctx context.Context, userID, callID string) error {
	hidden := models.CallHidden{
		CallID: callID,
		UserID: userID,
	}
	tx := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&hidden)
	if tx.Error != nil {
		return fmt.Errorf("hide call: %w", tx.Error)
	}
	return nil
}
