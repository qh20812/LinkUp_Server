package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"linkup/config"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/utils"
)

func AuthMiddleware(env config.Env, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeMissingAuthorization))
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeInvalidToken))
			c.Abort()
			return
		}

		token, err := utils.ParseToken(env.JWTSecret, parts[1])
		if err != nil || !token.Valid {
			errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeInvalidToken))
			c.Abort()
			return
		}

		claims := token.Claims.(*utils.TokenClaims)
		if claims.TokenType != "access" {
			errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeInvalidToken))
			c.Abort()
			return
		}

		if db != nil {
			var user models.User
			if err := db.WithContext(c.Request.Context()).Select("status, token_version").Where("id = ?", claims.UserID).First(&user).Error; err != nil {
				errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeUserNotFound))
				c.Abort()
				return
			}

			if !user.IsActive() {
				if user.IsBanned() {
					errorsapp.RespondError(c, http.StatusForbidden, errorsapp.New(errorsapp.ErrCodeAccountBanned))
				} else {
					errorsapp.RespondError(c, http.StatusForbidden, errorsapp.New(errorsapp.ErrCodeAccountInactive))
				}
				c.Abort()
				return
			}

			if claims.TokenVersion != user.TokenVersion {
				code := errorsapp.ErrCodeSessionExpired
				var sysConfig models.SystemConfig
				if err := db.WithContext(c.Request.Context()).Where("`key` = ?", "maintenance_mode").First(&sysConfig).Error; err == nil && sysConfig.Value == "true" {
					code = errorsapp.ErrCodeMaintenanceMode
				}
				errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(code))
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
					errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeSessionExpired))
					c.Abort()
					return
				}
				if session.RevokedAt != nil || time.Now().UTC().After(session.ExpiresAt) {
					errorsapp.RespondError(c, http.StatusUnauthorized, errorsapp.New(errorsapp.ErrCodeSessionExpired))
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

// OptionalAuth tries to parse the JWT token and set userID/email if present,
// but does NOT abort if the token is missing or invalid.
func OptionalAuth(env config.Env, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token, err := utils.ParseToken(env.JWTSecret, parts[1])
		if err != nil || !token.Valid {
			c.Next()
			return
		}

		claims := token.Claims.(*utils.TokenClaims)
		if claims.TokenType != "access" {
			c.Next()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("sessionID", claims.ID)
		c.Next()
	}
}
