package groupws

import (
	"context"
	"log"
	"net/http"

	"linkup/config"
	"linkup/models"
	"linkup/services"
	"linkup/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func checkUserActive(c *gin.Context, db *gorm.DB, userID string) bool {
	if db == nil {
		return true
	}
	var user models.User
	if err := db.WithContext(c.Request.Context()).Select("status").Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return false
	}
	if !user.IsActive() {
		c.JSON(http.StatusForbidden, gin.H{"error": "account is banned or suspended"})
		return false
	}
	return true
}

func ServeGroupWS(
	hub *Hub,
	messageService *services.GroupMessageService,
	groupService *services.GroupChatService,
	env config.Env,
	db *gorm.DB,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Query("token")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		token, err := utils.ParseToken(env.JWTSecret, tokenString)
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims := token.Claims.(*utils.TokenClaims)
		if claims.TokenType != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token type"})
			return
		}

		if !checkUserActive(c, db, claims.UserID) {
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("group ws upgrade: %v", err)
			return
		}

		client := NewClient(
			context.Background(),
			conn,
			hub,
			messageService,
			groupService,
			claims.UserID,
		)

		hub.RegisterClient(client)
		go client.WritePump()
		go client.ReadPump()
	}
}

func ServeGroupCallWS(
	hub *Hub,
	messageService *services.GroupMessageService,
	groupService *services.GroupChatService,
	groupChatHub *Hub,
	env config.Env,
	db *gorm.DB,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Query("token")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		token, err := utils.ParseToken(env.JWTSecret, tokenString)
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims := token.Claims.(*utils.TokenClaims)
		if claims.TokenType != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token type"})
			return
		}

		if !checkUserActive(c, db, claims.UserID) {
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("group call ws upgrade: %v", err)
			return
		}

		client := NewClientWithMode(
			context.Background(),
			conn,
			hub,
			messageService,
			groupService,
			claims.UserID,
			"call",
			groupChatHub,
		)

		hub.RegisterClient(client)
		go client.WritePump()
		go client.ReadPump()
	}
}
