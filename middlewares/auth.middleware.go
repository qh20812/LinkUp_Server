package middlewares

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "linkup/config"
    "linkup/utils"
)

func AuthMiddleware(env config.Env) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
            c.Abort()
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
            c.Abort()
            return
        }

        token, err := utils.ParseToken(env.JWTSecret, parts[1])
        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }

        claims := token.Claims.(*utils.TokenClaims)
        c.Set("userID", claims.UserID)
        c.Set("email", claims.Email)
        c.Next()
    }
}