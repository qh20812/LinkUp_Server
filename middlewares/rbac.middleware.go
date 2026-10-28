package middlewares

import (
	"linkup/errors"
	"linkup/models"
	"linkup/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// platformRoles đọc toàn bộ role phạm vi nền tảng (scope_id IS NULL) của user.
// Một user có thể có nhiều role cùng lúc — seed cấp cả PARTNER và USER cho tài
// khoản đối tác. Trước đây chỉ lấy một dòng (Scan vào string) nên kết quả phụ
// thuộc thứ tự MySQL trả về: có lúc nhận USER nên tài khoản đối tác bị 403 dù
// có PARTNER. Vì là biến nên test có thể thay bằng dữ liệu giả.
var platformRoles = func(db *gorm.DB, userID string) ([]models.RoleName, error) {
	var names []string
	err := db.Table("roles").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Where("user_roles.scope_id IS NULL").
		Pluck("roles.name", &names).Error
	if err != nil {
		return nil, err
	}

	roles := make([]models.RoleName, 0, len(names))
	for _, name := range names {
		roles = append(roles, models.RoleName(name))
	}
	return roles, nil
}

// roleRank xếp hạng quyền. Khi user có nhiều role, role mạnh nhất được chọn để
// ghi vào context (vd. ADMIN + PARTNER → ADMIN, để admin quản lý được ad của
// mọi partner thay vì bị coi là partner sở hữu).
var roleRank = map[models.RoleName]int{
	models.RoleSuperAdmin: 4,
	models.RoleAdmin:      3,
	models.RolePartner:    2,
	models.RoleUser:       1,
}

func containsRole(roles []models.RoleName, target models.RoleName) bool {
	for _, role := range roles {
		if role == target {
			return true
		}
	}
	return false
}

// effectiveRole chọn role mạnh nhất trong số role user có mà route cho phép.
func effectiveRole(userRoles, allowedRoles []models.RoleName) (models.RoleName, bool) {
	var best models.RoleName
	found := false
	for _, role := range userRoles {
		if !containsRole(allowedRoles, role) {
			continue
		}
		if !found || roleRank[role] > roleRank[best] {
			best = role
			found = true
		}
	}
	return best, found
}

func RequireRoles(db *gorm.DB, allowedRoles ...models.RoleName) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("userID")
		if !exists {
			errors.RespondError(c, http.StatusUnauthorized, errors.New(errors.ErrCodeRbacAuthRequired))
			c.Abort()
			return
		}
		userID := userIDVal.(string)

		userRoles, err := platformRoles(db, userID)
		if err != nil || len(userRoles) == 0 {
			errors.RespondError(c, http.StatusForbidden, errors.New(errors.ErrCodeRbacRoleNotFound))
			c.Abort()
			return
		}

		userRole, isAllowed := effectiveRole(userRoles, allowedRoles)
		if !isAllowed {
			errors.RespondError(c, http.StatusForbidden, errors.New(errors.ErrCodeRbacPermissionDenied))
			c.Abort()
			return
		}

		c.Set("userRole", string(userRole))

		c.Next()
	}
}

// RequireContributionLevel checks that the authenticated user's contribution
// score in the community (identified by :communityID URL param) meets or
// exceeds the given threshold. Requires AuthMiddleware(env, db) to run first.
func RequireContributionLevel(db *gorm.DB, threshold int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("userID")
		if !exists {
			errors.RespondError(c, http.StatusUnauthorized, errors.New(errors.ErrCodeRbacAuthRequired))
			c.Abort()
			return
		}
		userID := userIDVal.(string)

		communityID := c.Param("communityID")
		if communityID == "" {
			errors.RespondError(c, http.StatusBadRequest, errors.New(errors.ErrCodeRbacMissingCommunityID))
			c.Abort()
			return
		}

		var score int
		err := db.Table("member_contributions").
			Select("COALESCE(contribution_score, 0)").
			Where("community_id = ? AND user_id = ?", communityID, userID).
			Scan(&score).Error
		if err != nil {
			errors.RespondError(c, http.StatusInternalServerError, errors.New(errors.ErrCodeRbacContributionCheckFailed))
			c.Abort()
			return
		}

		if score < threshold {
			errors.RespondError(c, http.StatusForbidden, errors.New(errors.ErrCodeRbacContributionInsufficient))
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckAdOwnership kiểm tra tính chính chủ đối với vai trò PARTNER
func CheckAdOwnership(db *gorm.DB, adService services.AdService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, _ := c.Get("userID")
		userID, _ := userIDVal.(string)

		// RequireRoles luôn chạy trước và đã tra role; chỉ khi thiếu mới tự tra
		// lại để middleware này dùng độc lập vẫn đúng.
		var userRoles []models.RoleName
		if roleVal, exists := c.Get("userRole"); exists {
			if roleStr, ok := roleVal.(string); ok && roleStr != "" {
				userRoles = []models.RoleName{models.RoleName(roleStr)}
			}
		}
		if len(userRoles) == 0 {
			roles, err := platformRoles(db, userID)
			if err != nil || len(roles) == 0 {
				errors.RespondError(c, http.StatusForbidden, errors.New(errors.ErrCodeRbacRoleNotFound))
				c.Abort()
				return
			}
			userRoles = roles
		}

		if containsRole(userRoles, models.RoleSuperAdmin) || containsRole(userRoles, models.RoleAdmin) {
			c.Next()
			return
		}

		if containsRole(userRoles, models.RolePartner) {
			adID := c.Param("id")
			if adID == "" {
				errors.RespondError(c, http.StatusBadRequest, errors.New(errors.ErrCodeRbacMissingAdID))
				c.Abort()
				return
			}

			ad, err := adService.GetAdByID(adID)
			if err != nil {
				errors.RespondError(c, http.StatusNotFound, errors.New(errors.ErrCodeRbacAdNotFound))
				c.Abort()
				return
			}

			if ad.PartnerID != userID {
				errors.RespondError(c, http.StatusForbidden, errors.New(errors.ErrCodeRbacAdAccessDenied))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
