package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"linkup/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type E2ERepository struct {
	db *gorm.DB
}

func NewE2ERepository(db *gorm.DB) *E2ERepository {
	return &E2ERepository{db: db}
}

// UpsertUserKey registers the user's public key. Nếu public key đổi so với lần
// đăng ký trước (identity/thiết bị mới) thì tăng key_version để đối phương phát
// hiện và re-key; giữ nguyên version nếu khóa không thay đổi.
func (r *E2ERepository) UpsertUserKey(ctx context.Context, key *models.UserE2EKey) error {
	existing, err := r.GetUserKey(ctx, key.UserID)
	if err != nil {
		return fmt.Errorf("get user e2e key before upsert: %w", err)
	}
	now := time.Now().UTC()

	// Chưa từng đăng ký → tạo mới với version 1.
	if existing == nil {
		if key.KeyVersion == 0 {
			key.KeyVersion = 1
		}
		key.CreatedAt = now
		key.UpdatedAt = now
		if err := r.db.WithContext(ctx).Create(key).Error; err != nil {
			return fmt.Errorf("insert user e2e key: %w", err)
		}
		return nil
	}

	updates := map[string]any{"updated_at": now}
	if existing.PublicKey != key.PublicKey {
		updates["public_key"] = key.PublicKey
		updates["key_version"] = existing.KeyVersion + 1
	}
	if err := r.db.WithContext(ctx).
		Model(&models.UserE2EKey{}).
		Where("user_id = ?", key.UserID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("update user e2e key: %w", err)
	}
	return nil
}

func (r *E2ERepository) GetUserKey(ctx context.Context, userID string) (*models.UserE2EKey, error) {
	var key models.UserE2EKey
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user e2e key: %w", err)
	}
	return &key, nil
}

// UpsertChatKey persists a single wrapped chat key entry. First-writer-wins.
func (r *E2ERepository) UpsertChatKey(ctx context.Context, key *models.ChatE2EKey) error {
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chat_id"}, {Name: "user_id"}},
		DoNothing: true,
	}).Create(key).Error
	if err != nil {
		return fmt.Errorf("insert chat e2e key: %w", err)
	}
	return nil
}

// lockChatRows serializes chat-key setup/read on a per-chat basis by taking a
// FOR UPDATE lock on the chats row. Hai client cùng tạo khóa cho một chat sẽ
// được xử lý tuần tự: người ghi trước thắng cả 2 row (cả hai participant), người
// sau không chèn được gì và tự reconcile về khóa của người thắng. chatIDs phải
// được sắp xếp để tránh deadlock khi hai request đi theo thứ tự ngược nhau.
func lockChatRows(ctx context.Context, tx *gorm.DB, chatIDs []string) error {
	for _, chatID := range chatIDs {
		var chat struct{ ID string }
		if err := tx.WithContext(ctx).Raw(
			"SELECT id FROM chats WHERE id = ? FOR UPDATE", chatID,
		).Scan(&chat).Error; err != nil {
			return fmt.Errorf("lock chat row %s: %w", chatID, err)
		}
	}
	return nil
}

func uniqueSortedChatIDs(keys []models.ChatE2EKey) []string {
	seen := make(map[string]struct{}, len(keys))
	ids := make([]string, 0, len(keys))
	for _, k := range keys {
		if _, ok := seen[k.ChatID]; ok {
			continue
		}
		seen[k.ChatID] = struct{}{}
		ids = append(ids, k.ChatID)
	}
	sort.Strings(ids)
	return ids
}

// UpsertChatKeys persists a batch of wrapped chat key entries. First-writer-wins:
// một khi (chat_id, user_id) đã tồn tại thì không ghi đè. Row-lock (FOR UPDATE)
// trên từng chat khiến batch insert được serialize — người ghi trước thắng cả 2
// row của một chat, đảm bảo hai client tạo khóa đồng thời quy về cùng một khóa.
func (r *E2ERepository) UpsertChatKeys(ctx context.Context, keys []models.ChatE2EKey) error {
	if len(keys) == 0 {
		return nil
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockChatRows(ctx, tx, uniqueSortedChatIDs(keys)); err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "chat_id"}, {Name: "user_id"}},
			DoNothing: true,
		}).Create(&keys).Error
	})
	if err != nil {
		return fmt.Errorf("insert chat e2e keys: %w", err)
	}
	return nil
}

// UpsertChatKeyForce ghi đè wrapped key của một row (chat_id, user_id) — dùng
// cho re-key khi đối phương đổi identity. Controller/service đảm bảo chỉ row
// của chính user gọi bị ghi; không đụng row của người khác.
func (r *E2ERepository) UpsertChatKeyForce(ctx context.Context, key *models.ChatE2EKey) error {
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chat_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"wrapped_key", "nonce"}),
	}).Create(key).Error
	if err != nil {
		return fmt.Errorf("upsert chat e2e key force: %w", err)
	}
	return nil
}

// GetChatKey reads the caller's wrapped copy of the chat's symmetric key. Đọc
// trong cùng transaction + row-lock để luôn lấy snapshot nhất quán (không đọc
// được "nửa cặp" khi transaction setup của bên kia đang dang dở).
func (r *E2ERepository) GetChatKey(ctx context.Context, chatID, userID string) (*models.ChatE2EKey, error) {
	var key models.ChatE2EKey
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockChatRows(ctx, tx, []string{chatID}); err != nil {
			return err
		}
		return tx.Where("chat_id = ? AND user_id = ?", chatID, userID).First(&key).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get chat e2e key: %w", err)
	}
	return &key, nil
}

// UpsertRecovery lưu (hoặc thay thế) backup khôi phục khóa chat. Mỗi lần bật
// mới / đổi PIN đều ghi đè toàn bộ salt+blob+check; bộ đếm attempts về 0.
func (r *E2ERepository) UpsertRecovery(ctx context.Context, rec *models.UserE2ERecovery) error {
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"salt", "blob", "pin_check", "recovery_check", "attempts", "locked_until", "updated_at"}),
	}).Create(rec).Error
	if err != nil {
		return fmt.Errorf("upsert user e2e recovery: %w", err)
	}
	return nil
}

// GetRecovery đọc backup khôi phục của user. Trả nil nếu chưa từng bật.
func (r *E2ERepository) GetRecovery(ctx context.Context, userID string) (*models.UserE2ERecovery, error) {
	var rec models.UserE2ERecovery
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user e2e recovery: %w", err)
	}
	return &rec, nil
}

// DeleteRecovery xóa backup khôi phục (tắt tính năng khôi phục).
func (r *E2ERepository) DeleteRecovery(ctx context.Context, userID string) error {
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.UserE2ERecovery{}).Error
	if err != nil {
		return fmt.Errorf("delete user e2e recovery: %w", err)
	}
	return nil
}

// InvalidateRecoveryAttempt tăng bộ đếm lần thử sai; khi gán lockedUntil thì
// các lần thử sau trong khoảng đó đều bị từ chối (rate-limit tăng dần).
func (r *E2ERepository) InvalidateRecoveryAttempt(ctx context.Context, userID string, lockedUntil *time.Time) error {
	updates := map[string]any{
		"attempts":   gorm.Expr("attempts + 1"),
		"updated_at": time.Now().UTC(),
	}
	if lockedUntil != nil {
		updates["locked_until"] = lockedUntil
	}
	err := r.db.WithContext(ctx).
		Model(&models.UserE2ERecovery{}).
		Where("user_id = ?", userID).
		Updates(updates).Error
	if err != nil {
		return fmt.Errorf("invalidate recovery attempt: %w", err)
	}
	return nil
}
