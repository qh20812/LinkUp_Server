package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"linkup/config"
	"linkup/models"
	"linkup/utils"
)

func AuthMiddleware(env config.Env, db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "thiếu header authorization"})
            c.Abort()
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "header authorization không hợp lệ"})
            c.Abort()
            return
        }

        token, err := utils.ParseToken(env.JWTSecret, parts[1])
        if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token không hợp lệ"})
            c.Abort()
            return
        }

        claims := token.Claims.(*utils.TokenClaims)
        if claims.TokenType != "access" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "token không hợp lệ"})
            c.Abort()
            return
        }

	if db != nil {
		var user models.User
		if err := db.WithContext(c.Request.Context()).Select("status, token_version").Where("id = ?", claims.UserID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "người dùng không tồn tại"})
			c.Abort()
			return
		}

		if !user.IsActive() {
			if user.IsBanned() {
				c.JSON(http.StatusForbidden, gin.H{"error": "tài khoản đang bị ban"})
			} else {
				c.JSON(http.StatusForbidden, gin.H{"error": "tài khoản chưa được kích hoạt"})
			}
			c.Abort()
			return
		}

		if claims.TokenVersion != user.TokenVersion {
			errMsg := "phiên đăng nhập đã hết hạn"
			var sysConfig models.SystemConfig
			if err := db.WithContext(c.Request.Context()).Where("`key` = ?", "maintenance_mode").First(&sysConfig).Error; err == nil && sysConfig.Value == "true" {
				errMsg = "hệ thống đang bảo trì"
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": errMsg})
			c.Abort()
			return
		}

		// Per-session revocation: tokens bound to a session carry the session id
		// (jti). Legacy tokens (no jti) skip this check but still expire via the
		// token_version check above.
		if claims.ID != "" {
			var session models.UserSession
			if err := db.WithContext(c.Request.Context()).
				Select("revoked_at, expires_at").
				Where("id = ? AND user_id = ?", claims.ID, claims.UserID).
				First(&session).Error; err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "phiên đăng nhập đã hết hạn"})
				c.Abort()
				return
			}
			if session.RevokedAt != nil || time.Now().UTC().After(session.ExpiresAt) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "phiên đăng nhập đã hết hạn"})
				c.Abort()
				return
			}
		}
	}

        c.Set("userID", claims.UserID)
        c.Set("email", claims.Email)
        c.Set("sessionID", claims.ID)
        c.Next()
    }
}