package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func GenerateTokenPair(secret string, userID string, email string, accessTTL, refreshTTL time.Duration) (string, string, error) {
	accessToken, err := generateToken(secret, userID, email, "access", accessTTL)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := generateToken(secret, userID, email, "refresh", refreshTTL)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func GenerateToken(secret string, userID string, email string, tokenType string, ttl time.Duration) (string, error) {
	return generateToken(secret, userID, email, tokenType, ttl)
}

func ParseToken(secret string, tokenString string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
}

func generateToken(secret string, userID string, email string, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := TokenClaims{
		UserID:    userID,
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   email,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
