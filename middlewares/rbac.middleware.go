package middlewares

import (
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
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
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied - Platform role not found"})
			c.Abort()
			return
		}

		userRole := models.RoleName(roleNameStr)

		// Kiểm tra xem role của user có nằm trong danh sách được phép không
		isAllowed := false
		for _, role := range allowedRoles {
			if userRole == role {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this resource"})
			c.Abort()
			return
		}

		c.Set("userRole", string(userRole))

		c.Next()
	}
}

// RequireContributionLevel checks that the authenticated user's contribution
// score in the community (identified by :communityID URL param) meets or
// exceeds the given threshold. Requires AuthMiddleware to run first.
func RequireContributionLevel(db *gorm.DB, threshold int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		userID := userIDVal.(string)

		communityID := c.Param("communityID")
		if communityID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing community ID"})
			c.Abort()
			return
		}

		var score int
		err := db.Table("member_contributions").
			Select("COALESCE(contribution_score, 0)").
			Where("community_id = ? AND user_id = ?", communityID, userID).
			Scan(&score).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check contribution level"})
			c.Abort()
			return
		}

		if score < threshold {
			c.JSON(http.StatusForbidden, gin.H{"error": "Điểm đóng góp chưa đủ để thực hiện hành động này"})
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

		// Nếu là ADMIN hoặc SUPER_ADMIN thì được quyền đi tiếp (Bypass)
		if userRole == models.RoleSuperAdmin || userRole == models.RoleAdmin {
			c.Next()
			return
		}

		// Nếu là PARTNER, bắt đầu kiểm tra chủ sở hữu của bản ghi quảng cáo
		if userRole == models.RolePartner {
			adID := c.Param("id")
			if adID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Missing ad ID"})
				c.Abort()
				return
			}

			ad, err := adService.GetAdByID(adID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Advertisement not found"})
				c.Abort()
				return
			}

			// Chống xem lén / sửa lén dữ liệu
			if ad.PartnerID != userID {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. You do not own this advertisement"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
