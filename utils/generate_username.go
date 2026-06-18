package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

var usernameRe = regexp.MustCompile(`[^a-zA-Z0-9_.]`)

type UsernameTakenFunc func(username string) (bool, error)

func GenerateUsername(email string, isTaken UsernameTakenFunc) (string, error) {
	base := extractBase(email)
	base = sanitizeBase(base)

	if len(base) > 20 {
		base = base[:20]
	}
	if len(base) < 3 {
		base = base + "user"
	}
	if len(base) > 18 {
		base = base[:18]
	}

	username := base
	taken, err := isTaken(username)
	if err != nil {
		return "", fmt.Errorf("check username: %w", err)
	}
	if !taken {
		return username, nil
	}

	for attempts := 0; attempts < 10; attempts++ {
		suffix, err := randomSuffix(4)
		if err != nil {
			return "", err
		}
		username = base + suffix
		taken, err := isTaken(username)
		if err != nil {
			return "", fmt.Errorf("check username: %w", err)
		}
		if !taken {
			return username, nil
		}
	}

	return "", fmt.Errorf("could not generate unique username after 10 attempts")
}

func extractBase(email string) string {
	at := strings.LastIndex(email, "@")
	if at == -1 {
		return email
	}
	return email[:at]
}

func sanitizeBase(s string) string {
	s = usernameRe.ReplaceAllString(s, "")
	return strings.ToLower(s)
}

func randomSuffix(length int) (string, error) {
	const digits = "0123456789"
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", fmt.Errorf("random digit: %w", err)
		}
		result[i] = digits[n.Int64()]
	}
	return string(result), nil
}
