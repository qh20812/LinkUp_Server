package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// Tạo key 256-bit cho AES-256
func GenerateEncryptionKey() (string, error) {
	key := make([]byte, 32) //256 bits
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// Mã hóa nội dung tin nhắn bằng AES-256-GCM
func EncryptMessage(plaintext, keyString string) (string, error) {
	// Decode key
	key, err := base64.StdEncoding.DecodeString(keyString)
	if err != nil {
		return "", fmt.Errorf("decode key failed: %w", err)
	}

	//Tạo AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher failed: %w", err)
	}

	// Tạo GCM (Galois/Counter Mode) để có authenticated encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm failed: %w", err)
	}

	// Tạo nonce ngẫu nhiên (12 bytes)
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return "", fmt.Errorf("generate nonce failed: %w", err)
	}

	// Mã hóa
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Giải mã nội dung tin nhắn
func DecryptMessage(encryptedBase64, keyString string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(keyString)
	if err != nil {
		return "", fmt.Errorf("decode key failed: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", fmt.Errorf("decode ciphertext failed: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm failed: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, encryptedText := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, encryptedText, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt failed: %w", err)
	}

	return string(plaintext), nil
}