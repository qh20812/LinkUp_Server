package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"linkup/config"
	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
)

type AuthService struct {
	authRepo           *repository.AuthRepository
	profileRepo        *repository.ProfileRepository
	banRepo            *repository.BanRepository
	adminSettingsRepo  *repository.AdminSettingsRepository
	emailVerifyService *EmailVerificationService
	googleVerifier     GoogleIDTokenVerifier
	env                config.Env
}

func NewAuthService(authRepo *repository.AuthRepository, profileRepo *repository.ProfileRepository, banRepo *repository.BanRepository, adminSettingsRepo *repository.AdminSettingsRepository, env config.Env) *AuthService {
	return &AuthService{
		authRepo:          authRepo,
		profileRepo:       profileRepo,
		banRepo:           banRepo,
		adminSettingsRepo: adminSettingsRepo,
		env:               env,
	}
}

func (s *AuthService) SetEmailVerificationService(svc *EmailVerificationService) {
	s.emailVerifyService = svc
}


func (s *AuthService) SetSessionRepository(sessionRepo *repository.UserSessionRepository) {
	s.sessionRepo = sessionRepo
}

func (s *AuthService) isMaintenanceMode(ctx context.Context) (bool, error) {
	if err != nil {
		return false, err
	}
	return cfg != nil && cfg.Value == "true", nil
}

func (s *AuthService) getMinPasswordLength(ctx context.Context) int {
	cfg, err := s.adminSettingsRepo.GetByKey(ctx, "password_min_length")
	if err != nil || cfg == nil {
		return 8
	}
	n, err := strconv.Atoi(cfg.Value)
	if err != nil || n < 8 {
		return 8
	}
	if n > 50 {
		return 50
	}
	return n
}

func (s *AuthService) getJWTExpiryMinutes(ctx context.Context) int {
	cfg, err := s.adminSettingsRepo.GetByKey(ctx, "jwt_expiry_minutes")
	if err != nil || cfg == nil {
		return clamp(s.env.JWTExpiresIn, 1, 60)
	}
	n, err := strconv.Atoi(cfg.Value)
	if err != nil || n < 1 {
		return clamp(s.env.JWTExpiresIn, 1, 60)
	}
	if n > 60 {
		return 60
	}
	return n
}

func (s *AuthService) getRefreshTokenExpiryDays(ctx context.Context) int {
	cfg, err := s.adminSettingsRepo.GetByKey(ctx, "refresh_token_expiry_days")
	if err != nil || cfg == nil {
		return 7
	}
	n, err := strconv.Atoi(cfg.Value)
	if err != nil || n < 1 {
		return 7
	}
	if n > 30 {
		return 30
	}
	return n
}

func (s *AuthService) getMaxLoginAttempts(ctx context.Context) int {
	cfg, err := s.adminSettingsRepo.GetByKey(ctx, "max_login_attempts")
	if err != nil || cfg == nil {
		return 5
	}
	n, err := strconv.Atoi(cfg.Value)
	if err != nil || n < 1 {
		return 5
	}
	if n > 10 {
		return 10
	}
	return n
}

func (s *AuthService) getDefaultUserRole(ctx context.Context) models.RoleName {
	cfg, err := s.adminSettingsRepo.GetByKey(ctx, "default_user_role")
	if err != nil || cfg == nil {
		return models.RoleUser
	}
	switch strings.ToLower(strings.TrimSpace(cfg.Value)) {
	case "admin":
		return models.RoleAdmin
	default:
		return models.RoleUser
	}
}

func (s *AuthService) Register(ctx context.Context, input dto.RegisterInput) (dto.AuthResponse, error) {
	cfg, err := s.adminSettingsRepo.GetByKey(ctx, "allow_registration")
	if err != nil {
		return dto.AuthResponse{}, err
	}
	if cfg != nil && cfg.Value == "false" {
		return dto.AuthResponse{}, errors.New("đăng ký đã bị tắt bởi quản trị viên")
	}

	maint, err := s.isMaintenanceMode(ctx)
	if err != nil {
		return dto.AuthResponse{}, err
	}
	if maint {
		return dto.AuthResponse{}, errors.New("hệ thống đang bảo trì")
	}

	if len(input.Password) < s.getMinPasswordLength(ctx) {
		return dto.AuthResponse{}, fmt.Errorf("mật khẩu phải có ít nhất %d ký tự", s.getMinPasswordLength(ctx))
	}
	if len(input.Password) > 50 {
		return dto.AuthResponse{}, errors.New("mật khẩu không được vượt quá 50 ký tự")
	}

	email := normalizeEmail(input.Email)

	if _, err := s.authRepo.FindByEmail(ctx, email); err == nil {
		return dto.AuthResponse{}, errors.New("email đã tồn tại")
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return dto.AuthResponse{}, err
	}

	username, err := utils.GenerateUsername(email, func(u string) (bool, error) {
		return s.authRepo.IsUsernameTaken(ctx, u)
	})
	if err != nil {
		return dto.AuthResponse{}, err
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	createdUser, err := s.authRepo.Create(ctx, &models.User{
		ID:                utils.GenerateUUID(),
		Username:          username,
		Email:             email,
		PasswordHash:      hashedPassword,
		Status:            models.UserStatusActive,
		StorageQuotaBytes: models.DefaultStorageQuotaBytes,
		StorageUsedBytes:  0,
	})
	if err != nil {
		return dto.AuthResponse{}, err
	}

	if err := s.authRepo.SavePasswordHistory(ctx, createdUser.ID, hashedPassword); err != nil {
		return dto.AuthResponse{}, err
	}

	if _, err := s.profileRepo.Create(ctx, &models.Profile{
		ID:          utils.GenerateUUID(),
		UserID:      createdUser.ID,
		DisplayName: input.DisplayName,
	}); err != nil {
		return dto.AuthResponse{}, err
	}

	if err := s.authRepo.AssignUserRole(ctx, createdUser.ID, s.getDefaultUserRole(ctx), nil, nil); err != nil {
		return dto.AuthResponse{}, err
	}

	// Check if email verification is required
	reqVerify := false
	if cfg, err := s.adminSettingsRepo.GetByKey(ctx, "require_email_verify"); err == nil && cfg != nil && cfg.Value == "true" {
		reqVerify = true
	}

	if reqVerify && s.emailVerifyService != nil {
		if err := s.emailVerifyService.SendVerificationEmail(ctx, createdUser.ID, createdUser.Email, createdUser.Username); err != nil {
			return dto.AuthResponse{}, err
		}
		return buildAuthResponse(*createdUser, "", "", s.accessTTL(ctx), s.refreshTTL(ctx), true), nil
	}

	accessToken, refreshToken, err := s.generateTokens(ctx, createdUser, "USER", "")
	if err != nil {
		return dto.AuthResponse{}, err
	}

	return buildAuthResponse(*createdUser, accessToken, refreshToken, s.accessTTL(ctx), s.refreshTTL(ctx), false), nil
}

func (s *AuthService) GoogleLogin(ctx context.Context, idToken string) (dto.AuthResponse, error) {
	if s.googleVerifier == nil {
		return dto.AuthResponse{}, errors.New("đăng nhập google chưa được cấu hình")
	}

	claims, err := s.googleVerifier.Verify(ctx, idToken)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	email := normalizeEmail(claims.Email)
	user, err := s.authRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return dto.AuthResponse{}, err
	}
	if user == nil {
		user, err = s.authRepo.FindByGoogleID(ctx, claims.GoogleID)
		if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
			return dto.AuthResponse{}, err
		}
	}

	if user == nil {
		user, err = s.createUserFromGoogle(ctx, claims, email)
		if err != nil {
			return dto.AuthResponse{}, err
		}
	} else {
		if user.GoogleID == nil {
			if err := s.authRepo.LinkGoogleAccount(ctx, user.ID, claims.GoogleID, time.Now().UTC()); err != nil {
				return dto.AuthResponse{}, err
			}
			user.GoogleID = &claims.GoogleID
		} else if *user.GoogleID != claims.GoogleID {
			return dto.AuthResponse{}, errors.New("tài khoản google không khớp với tài khoản hiện tại")
		}

		if err := s.ensureBanStatus(ctx, user); err != nil {
			return dto.AuthResponse{}, err
		}
		if !user.IsActive() {
			return dto.AuthResponse{}, errors.New("tài khoản không hoạt động")
		}
	}

	maint, err := s.isMaintenanceMode(ctx)
	if err != nil {
		return dto.AuthResponse{}, err
	}
	role, err := s.authRepo.GetUserRole(ctx, user.ID)
	if err != nil {
		return dto.AuthResponse{}, err
	}
	if maint && role != string(models.RoleSuperAdmin) && role != string(models.RoleAdmin) {
		return dto.AuthResponse{}, errors.New("hệ thống đang bảo trì")
	}

	accessToken, refreshToken, err := s.generateTokens(ctx, user, role)
	if err != nil {
		return dto.AuthResponse{}, err
	}
	return buildAuthResponse(*user, accessToken, refreshToken, s.accessTTL(ctx), s.refreshTTL(ctx), false), nil
}

// createUserFromGoogle tạo tài khoản mới từ thông tin Google đã xác thực.
func (s *AuthService) createUserFromGoogle(ctx context.Context, claims *GoogleClaims, email string) (*models.User, error) {
	cfg, err := s.adminSettingsRepo.GetByKey(ctx, "allow_registration")
	if err != nil {
		return nil, err
	}
	if cfg != nil && cfg.Value == "false" {
		return nil, errors.New("đăng ký đã bị tắt bởi quản trị viên")
	}

	username, err := utils.GenerateUsername(email, func(u string) (bool, error) {
		return s.authRepo.IsUsernameTaken(ctx, u)
	})
	if err != nil {
		return nil, err
	}

	displayName := strings.TrimSpace(claims.Name)
	if displayName == "" {
		displayName = strings.Split(email, "@")[0]
	}

	googleID := claims.GoogleID
	now := time.Now().UTC()
	createdUser, err := s.authRepo.Create(ctx, &models.User{
		ID:                utils.GenerateUUID(),
		Username:          username,
		Email:             email,
		PasswordHash:      "", // không có mật khẩu -> không dùng được Login truyền thống
		Status:            models.UserStatusActive,
		StorageQuotaBytes: models.DefaultStorageQuotaBytes,
		EmailVerifiedAt:   &now,
		GoogleID:          &googleID,
	})
	if err != nil {
		return nil, err
	}

	if _, err := s.profileRepo.Create(ctx, &models.Profile{
		ID:          utils.GenerateUUID(),
		UserID:      createdUser.ID,
		DisplayName: displayName,
		AvatarURI:   claims.Picture,
	}); err != nil {
		return nil, err
	}

	if err := s.authRepo.AssignUserRole(ctx, createdUser.ID, s.getDefaultUserRole(ctx), nil, nil); err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (s *AuthService) Login(ctx context.Context, input dto.LoginInput) (dto.AuthResponse, error) {
	email := normalizeEmail(input.Email)
	user, err := s.authRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return dto.AuthResponse{}, errors.New("email hoặc mật khẩu không hợp lệ")
		}
		return dto.AuthResponse{}, err
	}

	if err := s.ensureBanStatus(ctx, user); err != nil {
		return dto.AuthResponse{}, err
	}

	// Self-deactivated accounts may log in again (which reactivates them).
	// Admin suspensions (status=suspended without self_deactivated_at) stay blocked.
	if !user.IsActive() && !(user.Status == models.UserStatusSuspended && user.SelfDeactivatedAt != nil) {
		return dto.AuthResponse{}, fmt.Errorf("tài khoản chưa được kích hoạt")
	}

	role, err := s.authRepo.GetUserRole(ctx, user.ID)
	if err != nil {
		return dto.AuthResponse{}, err
	}
	isPrivileged := role == string(models.RoleSuperAdmin) || role == string(models.RoleAdmin)

	if !isPrivileged && user.IsLocked() {
		return dto.AuthResponse{}, errors.New("tài khoản tạm thời bị khóa do nhập sai nhiều lần, vui lòng thử lại sau 15 phút")
	}

	if err := utils.ComparePassword(user.PasswordHash, input.Password); err != nil {
		if !isPrivileged {
			maxAttempts := s.getMaxLoginAttempts(ctx)
			if err := s.authRepo.IncrementLoginAttempts(ctx, user.ID, maxAttempts); err != nil {
				return dto.AuthResponse{}, err
			}
			remaining := maxAttempts - user.LoginAttempts - 1
			if remaining > 0 {
				return dto.AuthResponse{}, fmt.Errorf("email hoặc mật khẩu không hợp lệ. Bạn còn %d lần thử", remaining)
			}
			return dto.AuthResponse{}, errors.New("email hoặc mật khẩu không hợp lệ. Tài khoản của bạn đã bị khóa tạm thời")
		}
		return dto.AuthResponse{}, errors.New("email hoặc mật khẩu không hợp lệ")
	}

	if err := s.authRepo.ResetLoginAttempts(ctx, user.ID); err != nil {
		return dto.AuthResponse{}, err
	}

	// Restore a self-deactivated account after a successful login.
	if user.Status == models.UserStatusSuspended && user.SelfDeactivatedAt != nil {
		if err := s.authRepo.Reactivate(ctx, user.ID); err != nil {
			return dto.AuthResponse{}, err
		}
		user.Status = models.UserStatusActive
		user.SelfDeactivatedAt = nil
	}

	maint, err := s.isMaintenanceMode(ctx)
	if err != nil {
		return dto.AuthResponse{}, err
	}
	if maint && role != string(models.RoleSuperAdmin) && role != string(models.RoleAdmin) {
		return dto.AuthResponse{}, errors.New("hệ thống đang bảo trì")
	}

	// Check email verification if required
	if cfg, err := s.adminSettingsRepo.GetByKey(ctx, "require_email_verify"); err == nil && cfg != nil && cfg.Value == "true" {
		if !user.IsEmailVerified() {
			return dto.AuthResponse{}, errors.New("vui lòng xác thực email trước khi đăng nhập")
		}
	}

	sessionID, err := s.createSession(ctx, user, deviceName, ipAddress, userAgent)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	accessToken, refreshToken, err := s.generateTokens(ctx, user, role, sessionID)
	if err != nil {
		return dto.AuthResponse{}, err
	}

	return buildAuthResponse(*user, accessToken, refreshToken, s.accessTTL(ctx), s.refreshTTL(ctx), false), nil
}

// createSession records a login session and returns its ID, which is embedded
// in the issued tokens as the `jti` claim. Expired rows are cleaned up lazily.
func (s *AuthService) createSession(ctx context.Context, user *models.User, deviceName, ipAddress, userAgent string) (string, error) {
	if s.sessionRepo == nil {
		return "", nil
	}

	_ = s.sessionRepo.CleanupExpired(ctx)

	now := time.Now().UTC()
	session := &models.UserSession{
		ID:           utils.GenerateUUID(),
		UserID:       user.ID,
		DeviceName:   deviceName,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		CreatedAt:    now,
		ExpiresAt:    now.Add(s.refreshTTL(ctx)),
		LastActiveAt: now,
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return "", err
	}
	return session.ID, nil
}

func (s *AuthService) generateTokens(ctx context.Context, user *models.User, role string, sessionID string) (string, string, error) {
	if sessionID == "" {
		return utils.GenerateTokenPair(s.env.JWTSecret, user.ID, user.Email, role, user.TokenVersion, s.accessTTL(ctx), s.refreshTTL(ctx))
	}
	return utils.GenerateTokenPairWithSession(s.env.JWTSecret, user.ID, user.Email, role, user.TokenVersion, sessionID, s.accessTTL(ctx), s.refreshTTL(ctx))
}

func (s *AuthService) Logout(ctx context.Context, userID string) error {
	if s.sessionRepo != nil {
		_ = s.sessionRepo.RevokeAllByUserID(ctx, userID)
	}
	return s.authRepo.IncrementTokenVersion(ctx, userID)
}

func (s *AuthService) accessTTL(ctx context.Context) time.Duration {
	min := s.getJWTExpiryMinutes(ctx)
	if min <= 0 {
		return 15 * time.Minute
	}
	return time.Duration(min) * time.Minute
}

func (s *AuthService) refreshTTL(ctx context.Context) time.Duration {
	days := s.getRefreshTokenExpiryDays(ctx)
	return time.Duration(days) * 24 * time.Hour
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func buildAuthResponse(user models.User, accessToken, refreshToken string, accessTTL, refreshTTL time.Duration, verifyEmail bool) dto.AuthResponse {
	return dto.AuthResponse{
		User: dto.AuthUserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Status:    string(user.Status),
			CreatedAt: user.CreatedAt,
		},
		Tokens: dto.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(accessTTL.Seconds()),
			RefreshTTLIn: int64(refreshTTL.Seconds()),
		},
		Storage:     dto.StorageInfo{
			QuotaBytes: user.StorageQuotaBytes,
			UsedBytes:  user.StorageUsedBytes,
			AvailBytes: user.AvailableStorageBytes(),
		},
		VerifyEmail: verifyEmail,
	}
}

func (s *AuthService) ChangePassword(ctx context.Context, userID string, input dto.ChangePasswordInput) error {
	if err := validations.NewAuthValidation().ValidatePassword(input.NewPassword); err != nil {
		return err
	}

	if len(input.NewPassword) < s.getMinPasswordLength(ctx) {
		return fmt.Errorf("mật khẩu phải có ít nhất %d ký tự", s.getMinPasswordLength(ctx))
	}
	if len(input.NewPassword) > 50 {
		return errors.New("mật khẩu không được vượt quá 50 ký tự")
	}

	user, err := s.authRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return errors.New("không tìm thấy người dùng")
		}
		return err
	}

	if err := utils.ComparePassword(user.PasswordHash, input.OldPassword); err != nil {
		return errors.New("mật khẩu hiện tại không đúng")
	}

	if input.OldPassword == input.NewPassword {
		return validations.ErrPasswordSameAsOld
	}

	histories, err := s.authRepo.GetPasswordHistoryByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, history := range histories {
		if err := utils.ComparePassword(history.PasswordHash, input.NewPassword); err == nil {
			return errors.New("không thể sử dụng lại mật khẩu trước đó")
		}
	}

	hashedPassword, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return err
	}

	if err := s.authRepo.UpdatePasswordWithHistory(ctx, userID, user.PasswordHash, hashedPassword); err != nil {
		return err
	}

	// A password change invalidates every existing token/session so that a
	// compromised account is fully locked out elsewhere.
	if s.sessionRepo != nil {
		_ = s.sessionRepo.RevokeAllByUserID(ctx, userID)
	}
	return s.authRepo.IncrementTokenVersion(ctx, userID)
}

func (s *AuthService) RefreshToken(ctx context.Context, input dto.RefreshTokenInput, deviceName, ipAddress, userAgent string) (dto.TokenResponse, error) {
	token, err := utils.ParseToken(s.env.JWTSecret, input.RefreshToken)
	if err != nil || !token.Valid {
		return dto.TokenResponse{}, errors.New("refresh token không hợp lệ hoặc đã hết hạn")
	}

	claims, ok := token.Claims.(*utils.TokenClaims)
	if !ok {
		return dto.TokenResponse{}, errors.New("refresh token không hợp lệ")
	}

	if claims.TokenType != "refresh" {
		return dto.TokenResponse{}, errors.New("token không phải refresh token")
	}

	user, err := s.authRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return dto.TokenResponse{}, errors.New("người dùng không tồn tại")
		}
		return dto.TokenResponse{}, err
	}

	if err := s.ensureBanStatus(ctx, user); err != nil {
		return dto.TokenResponse{}, err
	}

	if !user.IsActive() {
		return dto.TokenResponse{}, errors.New("tài khoản chưa được kích hoạt")
	}

	// Reject refresh tokens issued before the latest token_version bump (e.g.
	// after logout or password change), mirroring the middleware for access tokens.
	if claims.TokenVersion != user.TokenVersion {
		return dto.TokenResponse{}, errors.New("phiên đăng nhập đã hết hạn")
	}

	role, err := s.authRepo.GetUserRole(ctx, user.ID)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	// Rotate: revoke the session that the old refresh token belonged to.
	if s.sessionRepo != nil && claims.ID != "" {
		_, _ = s.sessionRepo.Revoke(ctx, claims.ID, claims.UserID)
	}

	sessionID, err := s.createSession(ctx, user, deviceName, ipAddress, userAgent)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	accessToken, refreshToken, err := s.generateTokens(ctx, user, role, sessionID)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	return dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTTL(ctx).Seconds()),
		RefreshTTLIn: int64(s.refreshTTL(ctx).Seconds()),
	}, nil
}

func (s *AuthService) ensureBanStatus(ctx context.Context, user *models.User) error {
	if user.Status != models.UserStatusBanned {
		return nil
	}

	ban, err := s.banRepo.GetLatestBanByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrBanNotFound) {
			return fmt.Errorf("tài khoản đang bị ban")
		}
		return err
	}

	if ban.ExpiresAt != nil && ban.ExpiresAt.Before(time.Now().UTC()) {
		if err := s.authRepo.UpdateUserStatus(ctx, user.ID, models.UserStatusActive); err != nil {
			return err
		}
		user.Status = models.UserStatusActive
		return nil
	}

	if ban.ExpiresAt != nil {
		return fmt.Errorf("tài khoản đang bị ban đến %s", ban.ExpiresAt.Format("2006-01-02 15:04:05"))
	}

	return errors.New("tài khoản bị ban vô thời hạn")
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
