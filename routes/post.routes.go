package routes

import (
	"fmt"
	"net/http"
	"strings"

	"linkup/config"
	"linkup/controllers"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RegisterPostRoutes(router *gin.Engine, ctrl *controllers.PostController, env config.Env) {
	postGroup := router.Group("/posts")
	{
		postGroup.POST("", AuthMiddleware(env), ctrl.CreatePost)
		postGroup.GET("", ctrl.GetPosts)
		postGroup.GET("/:id", ctrl.ViewPostDetail)
		postGroup.POST("/:id/react", AuthMiddleware(env), ctrl.ReactPost)
		postGroup.POST("/:id/comments", AuthMiddleware(env), ctrl.CreateComment)
	}
}

func AuthMiddleware(env config.Env) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Thiếu header Authorization"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Định dạng Authorization phải là Bearer <token>"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("phương thức ký không hợp lệ: %v", token.Header["alg"])
			}
			return []byte(env.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không hợp lệ hoặc đã hết hạn"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if userID, exists := claims["user_id"]; exists {
				c.Set("userId", userID)
			} else if sub, exists := claims["sub"]; exists {
				c.Set("userId", sub)
			}
		}

		c.Next()
	}
}