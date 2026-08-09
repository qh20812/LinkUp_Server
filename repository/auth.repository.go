package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"linkup/dto"
	"linkup/models"
	"linkup/utils"

	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("không tìm thấy người dùng")

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	tx := r.db.WithContext(ctx).Create(user)
	if tx.Error != nil {
		return nil, fmt.Errorf("insert user: %w", tx.Error)
	}
	return user, nil
}

func (r *AuthRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &user, nil
}

func (r *AuthRepository) IsUsernameTaken(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check username: %w", err)
	}
	return count > 0, nil
}

func (r *AuthRepository) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
	tx := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("password_hash", hashedPassword)
	if tx.Error != nil {
		return fmt.Errorf("update password: %w", tx.Error)
	}
	return nil
}

func (r *AuthRepository) FindByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &user, nil
}

func (r *AuthRepository) FindByIDs(ctx context.Context, userIDs []string) ([]models.User, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	var users []models.User
	if err := r.db.WithContext(ctx).Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("find users by ids: %w", err)
	}
	return users, nil
}

// SavePasswordHistory lưu lịch sử mật khẩu của người dùng vào cơ sở dữ liệu.
func (r *AuthRepository) SavePasswordHistory(ctx context.Context, userID string, hashedPassword string) error {
	history := &models.PasswordHistory{
		ID:           utils.GenerateUUID(),
		UserID:       userID,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
	}
	tx := r.db.WithContext(ctx).Create(history)
	if tx.Error != nil {
		return fmt.Errorf("save password history: %w", tx.Error)
	}
	return nil
}

// HasRole kiểm tra xem người dùng có vai trò cụ thể hay không (platform role, scope=NULL).
func (r *AuthRepository) HasRole(ctx context.Context, userID string, roleName models.RoleName) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("user_roles").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.name = ? AND user_roles.scope_id IS NULL", userID, roleName).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check role: %w", err)
	}
	return count > 0, nil
}

// HasScopedRole kiểm tra user có role trong scope (community/chat) cụ thể không.
func (r *AuthRepository) HasScopedRole(ctx context.Context, userID string, roleName models.RoleName, scopeID string, scopeType models.ScopeType) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("user_roles").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.name = ? AND user_roles.scope_id = ? AND user_roles.scope_type = ?",
			userID, roleName, scopeID, scopeType).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check scoped role: %w", err)
	}
	return count > 0, nil
}

// GetPasswordHistoryByUserID lấy danh sách lịch sử mật khẩu của người dùng theo userID.
func (r *AuthRepository) GetPasswordHistoryByUserID(ctx context.Context, userID string) ([]models.PasswordHistory, error) {
	var histories []models.PasswordHistory
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&histories).Error
	if err != nil {
		return nil, fmt.Errorf("get password history: %w", err)
	}
	return histories, nil
}

// AssignUserRole gán một role cho user trong user_roles.
// scopeID/scopeType = nil cho platform role, có giá trị cho scoped role (community/chat).
func (r *AuthRepository) AssignUserRole(ctx context.Context, userID string, roleName models.RoleName, scopeID *string, scopeType *models.ScopeType) error {
	var role models.Role
	if err := r.db.WithContext(ctx).Where("name = ?", roleName).First(&role).Error; err != nil {
		return fmt.Errorf("find role %s: %w", roleName, err)
	}

	userRole := models.NewUserRole(userID, role.ID)
	userRole.ID = utils.GenerateUUID()
	userRole.ScopeID = scopeID
	userRole.ScopeType = scopeType
	userRole.AssignedAt = time.Now().UTC()

	if err := r.db.WithContext(ctx).Create(&userRole).Error; err != nil {
		return fmt.Errorf("assign role %s: %w", roleName, err)
	}
	return nil
}

func (r *AuthRepository) GetUserRole(ctx context.Context, userID string) (string, error) {
	var roleName string
	err := r.db.WithContext(ctx).
		Table("user_roles").
		Select("roles.name").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND user_roles.scope_id IS NULL", userID).
		Order("CASE roles.name WHEN 'SUPER_ADMIN' THEN 0 WHEN 'ADMIN' THEN 1 WHEN 'PARTNER' THEN 2 ELSE 3 END").
		Limit(1).
		Scan(&roleName).Error
	if err != nil {
		return "", fmt.Errorf("get user role: %w", err)
	}
	if roleName == "" {
		return "USER", nil
	}
	return roleName, nil
}

func (r *AuthRepository) CountUsers(ctx context.Context, keyword string, status string) (int64, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&models.User{}).
		Joins("LEFT JOIN profiles ON profiles.user_id = users.id").
		Where("users.id NOT IN (SELECT user_roles.user_id FROM user_roles JOIN roles ON roles.id = user_roles.role_id WHERE roles.name IN ('SUPER_ADMIN', 'ADMIN') AND user_roles.scope_id IS NULL)")

	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("users.username LIKE ? OR users.email LIKE ? OR profiles.display_name LIKE ?", like, like, like)
	}

	if status != "" {
		q = q.Where("users.status = ?", status)
	}

	if err := q.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}

func (r *AuthRepository) ListUsers(ctx context.Context, keyword string, status string, limit, offset int) ([]dto.AdminUserListItem, error) {
	var results []dto.AdminUserListItem
	q := r.db.WithContext(ctx).
		Model(&models.User{}).
		Select("users.id, users.username, users.email, users.status, COALESCE(profiles.display_name, '') AS display_name, COALESCE(profiles.avatar_uri, '') AS avatar_uri, users.created_at, users.updated_at").
		Joins("LEFT JOIN profiles ON profiles.user_id = users.id").
		Where("users.id NOT IN (SELECT user_roles.user_id FROM user_roles JOIN roles ON roles.id = user_roles.role_id WHERE roles.name IN ('SUPER_ADMIN', 'ADMIN') AND user_roles.scope_id IS NULL)").
		Order("users.created_at DESC").
		Limit(limit).
		Offset(offset)

	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("users.username LIKE ? OR users.email LIKE ? OR profiles.display_name LIKE ?", like, like, like)
	}

	if status != "" {
		q = q.Where("users.status = ?", status)
	}

	if err := q.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return results, nil
}

func (r *AuthRepository) UpdateUserStatus(ctx context.Context, userID string, status models.UserStatus) error {
	tx := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("status", status)
	if tx.Error != nil {
		return fmt.Errorf("update user status: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *AuthRepository) IncrementAllTokenVersions(ctx context.Context) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE users
		SET token_version = token_version + 1
		WHERE id NOT IN (
			SELECT ur.user_id FROM user_roles ur
			JOIN roles r ON r.id = ur.role_id
			WHERE r.name IN ('SUPER_ADMIN', 'ADMIN')
			AND ur.scope_id IS NULL
		)
	`).Error
}

func (r *AuthRepository) IncrementTokenVersion(ctx context.Context, userID string) error {
	tx := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).UpdateColumn("token_version", gorm.Expr("token_version + 1"))
	if tx.Error != nil {
		return fmt.Errorf("increment token version: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *AuthRepository) IncrementLoginAttempts(ctx context.Context, userID string, maxAttempts int) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE users
		SET login_attempts = login_attempts + 1,
		    locked_until = CASE
			WHEN login_attempts >= ? THEN DATE_ADD(NOW(), INTERVAL 15 MINUTE)
			ELSE locked_until
		    END
		WHERE id = ?
	`, maxAttempts, userID).Error
}

func (r *AuthRepository) ResetLoginAttempts(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"login_attempts": 0,
		"locked_until":   nil,
	}).Error
}

func (r *AuthRepository) UpdateEmailVerifiedAtTx(ctx context.Context, tx *gorm.DB, userID string, verifiedAt time.Time) error {
	result := tx.Model(&models.User{}).Where("id = ?", userID).Update("email_verified_at", verifiedAt)
	if result.Error != nil {
		return fmt.Errorf("update email_verified_at: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *AuthRepository) FindByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("google_id = ?", googleID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by google id: %w", err)
	}
	return &user, nil
}

func (r *AuthRepository) LinkGoogleAccount(ctx context.Context, userID string, googleID string, verifiedAt time.Time) error {
	res := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ? AND google_id IS NULL", userID).
		Update("google_id", googleID)
	if res.Error != nil {
		return fmt.Errorf("link google account: %w", res.Error)
	}
	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ? AND email_verified_at IS NULL", userID).
		Update("email_verified_at", verifiedAt).Error; err != nil {
		return fmt.Errorf("mark google email verified: %w", err)
	}
	return nil
}
