package middlewares

import (
    "net/http"
    "strings"

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
            if err := db.WithContext(c.Request.Context()).Select("status").Where("id = ?", claims.UserID).First(&user).Error; err != nil {
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
        }

        c.Set("userID", claims.UserID)
        c.Set("email", claims.Email)
        c.Next()
    }
}