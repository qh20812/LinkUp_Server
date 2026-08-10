package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	TokenType    string `json:"token_type"`
	TokenVersion int    `json:"token_version"`
	jwt.RegisteredClaims
}

func GenerateTokenPair(secret string, userID string, email string, role string, tokenVersion int, accessTTL, refreshTTL time.Duration) (string, string, error) {
	return generateTokenPair(secret, userID, email, role, tokenVersion, "", accessTTL, refreshTTL)
}

// GenerateTokenPairWithSession issues an access/refresh pair bound to a user
// session. The session ID is embedded as the JWT `jti` claim, which the auth
// middleware uses to enforce per-session revocation (list/revoke devices).
func GenerateTokenPairWithSession(secret string, userID string, email string, role string, tokenVersion int, sessionID string, accessTTL, refreshTTL time.Duration) (string, string, error) {
	return generateTokenPair(secret, userID, email, role, tokenVersion, sessionID, accessTTL, refreshTTL)
}

func generateTokenPair(secret string, userID string, email string, role string, tokenVersion int, sessionID string, accessTTL, refreshTTL time.Duration) (string, string, error) {
	accessToken, err := generateToken(secret, userID, email, role, tokenVersion, "access", sessionID, accessTTL)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := generateToken(secret, userID, email, role, tokenVersion, "refresh", sessionID, refreshTTL)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func GenerateToken(secret string, userID string, email string, role string, tokenVersion int, tokenType string, ttl time.Duration) (string, error) {
	return generateToken(secret, userID, email, role, tokenVersion, tokenType, "", ttl)
}

func ParseToken(secret string, tokenString string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
}

func generateToken(secret string, userID string, email string, role string, tokenVersion int, tokenType string, sessionID string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := TokenClaims{
		UserID:       userID,
		Email:        email,
		Role:         role,
		TokenType:    tokenType,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        sessionID,
			Subject:   email,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
