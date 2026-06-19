package routes

import (
	"fmt"
	"net/http"
	"strings"

	"linkup/config"
	"linkup/controllers"

	"github.com/golang-jwt/jwt/v5" // Hoặc thư viện jwt bạn đang dùng trong dự án
	"github.com/gin-gonic/gin"
)

// Bổ sung thêm tham số env config.Env để lấy JWTSecret
func RegisterPostRoutes(router *gin.Engine, ctrl *controllers.PostController, env config.Env) {
	postGroup := router.Group("/posts")
	
	// Áp dụng Tháp canh bảo mật (AuthMiddleware) riêng cho cổng POST tạo bài viết
	postGroup.POST("", AuthMiddleware(env), ctrl.CreatePost)
	
	// Cổng GET xem bài viết công khai thì không cần ép đăng nhập
	postGroup.GET("", ctrl.GetPosts)
}

// Hàm Middleware trung gian để bóc tách và kiểm tra Token từ Postman gửi lên
func AuthMiddleware(env config.Env) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Thiếu header Authorization"})
			c.Abort()
			return
		}

		// Định dạng chuẩn: "Bearer <token>" -> Cần cắt chữ "Bearer " ra
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Định dạng Authorization phải là Bearer <token>"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Giải mã (Parse) token bằng Secret Key của dự án
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

		// Lấy thông tin user_id nằm trong phần lõi (Claims) của Token ra
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// 🌟 QUAN TRỌNG: Gán userId vào context với key là "userId" khớp với Controller
			// Tùy theo hàm login bạn gán key gì khi tạo token (id, user_id, sub, ...), bạn map cho đúng nhé.
			if userID, exists := claims["user_id"]; exists {
				c.Set("userId", userID) 
			} else if sub, exists := claims["sub"]; exists {
				c.Set("userId", sub)
			}
		}

		c.Next() // Token hợp lệ, cho phép request đi tiếp vào Controller
	}
}