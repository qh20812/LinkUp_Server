package utils

import (
	"crypto/rand"
	"math/big"
)

const inviteCodeCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func GenerateInviteCode() (string, error) {
	code := make([]byte, 6)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(inviteCodeCharset))))
		if err != nil {
			return "", err
		}
		code[i] = inviteCodeCharset[n.Int64()]
	}
	return string(code), nil
}
