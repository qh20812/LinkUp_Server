package groupws

import (
	"context"
	"log"
	"net/http"

	"linkup/config"
	"linkup/services"
	"linkup/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func ServeGroupWS(
	hub *Hub,
	messageService *services.GroupMessageService,
	groupService *services.GroupChatService,
	env config.Env,
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
