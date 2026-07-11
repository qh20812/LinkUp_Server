package ws

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"linkup/config"
	"linkup/utils"
)

// Phase 4 fix: Configurable CheckOrigin via WS_ALLOWED_ORIGINS env var.
// - Empty or "*": allows all origins (default, suitable for dev)
// - Comma-separated list: only allows origins matching the list
// This mitigates Cross-Site WebSocket Hijacking (CSWSH) in production.
var allowedOrigins = parseAllowedOrigins(os.Getenv("WS_ALLOWED_ORIGINS"))

func parseAllowedOrigins(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "*" {
		return nil // nil means "allow all"
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			origins = append(origins, strings.ToLower(p))
		}
	}
	return origins
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		if allowedOrigins == nil {
			return true // wildcard: allow all
		}
		origin := r.Header.Get("Origin")
		if origin == "" {
			return false
		}
		originLower := strings.ToLower(origin)
		for _, o := range allowedOrigins {
			if originLower == o {
				return true
			}
		}
		return false
	},
}

func ServeWS(hub *Hub, env config.Env) gin.HandlerFunc {
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
			log.Printf("ws upgrade: %v", err)
			return
		}

		client := NewClient(context.Background(), conn, hub, nil, nil, claims.UserID)
		hub.register <- client

		go client.WritePump()
		go client.ReadPump()
	}
}
