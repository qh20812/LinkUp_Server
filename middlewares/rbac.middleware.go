package middlewares

import (
	"linkup/errors"
	"linkup/models"
	"linkup/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RequireRoles(db *gorm.DB, allowedRoles ...models.RoleName) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("userID")
		if !exists {
			errors.RespondError(c, http.StatusUnauthorized, errors.New(errors.ErrCodeRbacAuthRequired))
			c.Abort()
			return
		}
		userID := userIDVal.(string)

		var roleNameStr string
		err := db.Table("roles").
			Select("roles.name").
			Joins("JOIN user_roles ON user_roles.role_id = roles.id").
			Where("user_roles.user_id = ?", userID).
			Where("user_roles.scope_id IS NULL").
			Scan(&roleNameStr).Error

		if err != nil || roleNameStr == "" {
			errors.RespondError(c, http.StatusForbidden, errors.New(errors.ErrCodeRbacRoleNotFound))
			c.Abort()
			return
		}

		userRole := models.RoleName(roleNameStr)

		isAllowed := false
		for _, role := range allowedRoles {
			if userRole == role {
				isAllowed = true
				break
			}
		}

		if isAllowed == false {
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
		userID := userIDVal.(string)
		roleVal, exists := c.Get("userRole")
		var userRole models.RoleName

		if !exists {
			var roleNameStr string
			db.Table("roles").
				Select("roles.name").
				Joins("JOIN user_roles ON user_roles.role_id = roles.id").
				Where("user_roles.user_id = ?", userID).
				Where("user_roles.scope_id IS NULL").
				Scan(&roleNameStr)
			userRole = models.RoleName(roleNameStr)
		} else {
			userRole = models.RoleName(roleVal.(string))
		}

		if userRole == models.RoleSuperAdmin || userRole == models.RoleAdmin {
			c.Next()
			return
		}

		if userRole == models.RolePartner {
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
