package repository

import (
	"context"
	"errors"
	"fmt"

	"linkup/dto"
	"linkup/models"

	"gorm.io/gorm"
)

type FriendRepository struct {
	db *gorm.DB
}

func NewFriendRepository(db *gorm.DB) *FriendRepository {
	return &FriendRepository{db: db}
}

func (r *FriendRepository) IsAcceptedFriend(ctx context.Context, userA, userB string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("friends").
		Where("((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)) AND status = ?", userA, userB, userB, userA, models.FriendStatusAccepted).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check friend status: %w", err)
	}
	return count > 0, nil
}
func (r *FriendRepository) FindBySenderAndReceiver(ctx context.Context, senderID, receiverID string) (*models.Friend, error) {
	var friend models.Friend
	err := r.db.WithContext(ctx).
		Where("sender_id = ? AND receiver_id = ?", senderID, receiverID).
		First(&friend).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find friend request: %w", err)
	}
	return &friend, nil
}

func (r *FriendRepository) Create(ctx context.Context, friend *models.Friend) error {
	tx := r.db.WithContext(ctx).Create(friend)
	if tx.Error != nil {
		return fmt.Errorf("create friend request: %w", tx.Error)
	}
	return nil
}

func (r *FriendRepository) FindByID(ctx context.Context, id string) (*models.Friend, error) {
	var friend models.Friend
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&friend).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find friend request by id: %w", err)
	}
	return &friend, nil
}

func (r *FriendRepository) FindByReceiverID(ctx context.Context, receiverID string, status models.FriendStatus) ([]models.Friend, error) {
	var friends []models.Friend
	q := r.db.WithContext(ctx).Where("receiver_id = ?", receiverID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("created_at DESC").Find(&friends).Error
	if err != nil {
		return nil, fmt.Errorf("find friend requests by receiver: %w", err)
	}
	return friends, nil
}

func (r *FriendRepository) FindBySenderID(ctx context.Context, senderID string, status models.FriendStatus) ([]models.Friend, error) {
	var friends []models.Friend
	q := r.db.WithContext(ctx).Where("sender_id = ?", senderID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("created_at DESC").Find(&friends).Error
	if err != nil {
		return nil, fmt.Errorf("find friend requests by sender: %w", err)
	}
	return friends, nil
}

func (r *FriendRepository) UpdateStatus(ctx context.Context, id string, status models.FriendStatus) error {
	tx := r.db.WithContext(ctx).Model(&models.Friend{}).Where("id = ?", id).Update("status", status)
	if tx.Error != nil {
		return fmt.Errorf("update friend request status: %w", tx.Error)
	}
	return nil
}

func (r *FriendRepository) Delete(ctx context.Context, id string) error {
	tx := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Friend{})
	if tx.Error != nil {
		return fmt.Errorf("delete friend request: %w", tx.Error)
	}
	return nil
}

func (r *FriendRepository) FindAcceptedFriends(ctx context.Context, userID string, page, pageSize int) ([]dto.FriendItem, int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Table("friends").
		Where("(sender_id = ? OR receiver_id = ?) AND status = ?", userID, userID, models.FriendStatusAccepted).
		Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("count friends: %w", err)
	}
	if total == 0 {
		return []dto.FriendItem{}, 0, nil
	}

	offset := (page - 1) * pageSize
	type friendRow struct {
		UserID      string
		DisplayName string
		AvatarURI   string
	}
	query := `
		SELECT
			(CASE WHEN f.sender_id = ? THEN f.receiver_id ELSE f.sender_id END) AS user_id,
			COALESCE(p.display_name, '') AS display_name,
			COALESCE(p.avatar_uri, '') AS avatar_uri
		FROM friends f
		LEFT JOIN profiles p
			ON p.user_id = (CASE WHEN f.sender_id = ? THEN f.receiver_id ELSE f.sender_id END)
		WHERE (f.sender_id = ? OR f.receiver_id = ?) AND f.status = ?
		ORDER BY f.created_at DESC
		LIMIT ? OFFSET ?`
	args := []interface{}{
		userID, userID,
		userID, userID, string(models.FriendStatusAccepted),
		pageSize, offset,
	}
	var rows []friendRow
	err = r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("find accepted friends: %w", err)
	}

	items := make([]dto.FriendItem, len(rows))
	for i, row := range rows {
		items[i] = dto.FriendItem{
			UserID:      row.UserID,
			DisplayName: row.DisplayName,
			AvatarURI:   row.AvatarURI,
			Status:      string(models.FriendStatusAccepted),
		}
	}
	return items, total, nil
}

func (r *FriendRepository) FindPair(ctx context.Context, userA, userB string) (*models.Friend, error) {
	var friend models.Friend
	err := r.db.WithContext(ctx).
		Where("((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?))", userA, userB, userB, userA).
		First(&friend).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find friend pair: %w", err)
	}
	return &friend, nil
}

func (r *FriendRepository) DeletePair(ctx context.Context, userA, userB string) error {
	tx := r.db.WithContext(ctx).
		Where("((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)) AND status = ?", userA, userB, userB, userA, models.FriendStatusAccepted).
		Delete(&models.Friend{})
	if tx.Error != nil {
		return fmt.Errorf("delete friend pair: %w", tx.Error)
	}
	return nil
}

func (r *FriendRepository) GetFriendSuggestions(ctx context.Context, userID string, page, pageSize int) ([]dto.FriendSuggestionItem, int64, error) {
	eligibleWhere := `u.id != ? AND u.status = ?
		AND NOT EXISTS (
			SELECT 1 FROM friends f
			WHERE f.status = ?
			  AND ((f.sender_id = ? AND f.receiver_id = u.id) OR (f.sender_id = u.id AND f.receiver_id = ?))
		)
		AND NOT EXISTS (
			SELECT 1 FROM friends f
			WHERE f.status = ?
			  AND ((f.sender_id = ? AND f.receiver_id = u.id) OR (f.sender_id = u.id AND f.receiver_id = ?))
		)
		AND NOT EXISTS (SELECT 1 FROM blocks b WHERE b.user_id = ? AND b.blocked_user_id = u.id)
		AND NOT EXISTS (SELECT 1 FROM blocks b WHERE b.user_id = u.id AND b.blocked_user_id = ?)
		AND NOT EXISTS (SELECT 1 FROM user_roles ur
			JOIN roles r ON r.id = ur.role_id
			WHERE ur.user_id = u.id AND r.name IN (?, ?))`

	eligibleArgs := func() []interface{} {
		return []interface{}{
			userID, models.UserStatusActive,
			models.FriendStatusAccepted, userID, userID,
			models.FriendStatusPending, userID, userID,
			userID, userID,
			models.RoleSuperAdmin, models.RoleAdmin,
		}
	}

	var total int64
	if err := r.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM users u WHERE `+eligibleWhere, eligibleArgs()...).Scan(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count friend suggestions: %w", err)
	}
	if total == 0 {
		return []dto.FriendSuggestionItem{}, 0, nil
	}

	offset := (page - 1) * pageSize

	type suggestionRow struct {
		ID          string
		DisplayName string
		AvatarURI   string
		MutualCount int
	}

	query := `SELECT u.id,
		COALESCE(p.display_name, '') AS display_name,
		COALESCE(p.avatar_uri, '') AS avatar_uri,
		COUNT(DISTINCT CASE WHEN uf.id IS NOT NULL THEN m.id END) AS mutual_count
	FROM users u
	LEFT JOIN profiles p ON p.user_id = u.id
	LEFT JOIN friends myf ON myf.status = ?
		AND (myf.sender_id = ? OR myf.receiver_id = ?)
	LEFT JOIN users m ON m.id = CASE WHEN myf.sender_id = ? THEN myf.receiver_id ELSE myf.sender_id END
	LEFT JOIN friends uf ON uf.status = ?
		AND ((uf.sender_id = u.id AND uf.receiver_id = m.id) OR (uf.sender_id = m.id AND uf.receiver_id = u.id))
	WHERE ` + eligibleWhere + `
	GROUP BY u.id, p.display_name, p.avatar_uri
	ORDER BY mutual_count DESC, u.id
	LIMIT ? OFFSET ?`

	perArgs := []interface{}{models.FriendStatusAccepted, userID, userID, userID, models.FriendStatusAccepted}
	perArgs = append(perArgs, eligibleArgs()...)
	perArgs = append(perArgs, pageSize, offset)

	var rows []suggestionRow
	if err := r.db.WithContext(ctx).Raw(query, perArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("get friend suggestions: %w", err)
	}
	if len(rows) == 0 {
		return []dto.FriendSuggestionItem{}, 0, nil
	}

	items := make([]dto.FriendSuggestionItem, len(rows))
	for i, row := range rows {
		items[i] = dto.FriendSuggestionItem{
			UserID:      row.ID,
			DisplayName: row.DisplayName,
			AvatarURI:   row.AvatarURI,
			MutualCount: row.MutualCount,
		}
	}
	return items, total, nil
}
