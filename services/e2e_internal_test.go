package services

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"os"
	"sync"
	"testing"
	"time"

	"linkup/models"
	"linkup/repository"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ─── Integration infrastructure ──────────────────────────────────

func connectAndMigrateE2E(t *testing.T) (*gorm.DB, *repository.E2ERepository) {
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		t.Skip("TEST_DSN not set; skipping DB-dependent test")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Chat{},
		&models.ChatParticipant{},
		&models.ChatE2EKey{},
		&models.UserE2EKey{},
		&models.UserE2ERecovery{},
	); err != nil {
		t.Fatalf("auto-migrate: %v", err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM user_e2e_recovery")
		db.Exec("DELETE FROM user_e2e_keys")
		db.Exec("DELETE FROM chat_e2e_keys")
		db.Exec("DELETE FROM chat_participants")
		db.Exec("DELETE FROM chats")
	})
	return db, repository.NewE2ERepository(db)
}

func seedE2EChat(t *testing.T, db *gorm.DB, chatID string) {
	chat := models.Chat{
		ID:        chatID,
		Type:      models.ChatTypeDirect,
		Name:      "e2e-test",
		Status:    models.ChatStatusActive,
		CreatedAt: time.Now().UTC(),
	}
	if err := db.Create(&chat).Error; err != nil {
		t.Fatalf("seed chat: %v", err)
	}
}

func seedE2EParticipant(t *testing.T, db *gorm.DB, chatID, userID string) {
	t.Helper()
	p := models.ChatParticipant{
		ID:       "cp-" + chatID + "-" + userID,
		ChatID:   chatID,
		UserID:   userID,
		Role:     models.ChatRoleMember,
		JoinedAt: time.Now().UTC(),
	}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("seed participant: %v", err)
	}
}

// aesWrap mô phỏng khâu wrap khóa chat: mã hóa chatKey bằng "shared secret"
// (32 byte) theo AES-256-GCM, như client thật với shared secret ECDH/HKDF.
func aesWrap(t *testing.T, sharedSecret, chatKey []byte) (wrapped, nonce string) {
	t.Helper()
	block, err := aes.NewCipher(sharedSecret)
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("new gcm: %v", err)
	}
	iv := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(iv); err != nil {
		t.Fatalf("read iv: %v", err)
	}
	ct := gcm.Seal(nil, iv, chatKey, nil)
	return hex.EncodeToString(ct), hex.EncodeToString(iv)
}

func aesUnwrap(t *testing.T, sharedSecret []byte, wrapped, nonce string) ([]byte, error) {
	t.Helper()
	block, err := aes.NewCipher(sharedSecret)
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("new gcm: %v", err)
	}
	iv, err := hex.DecodeString(nonce)
	if err != nil {
		t.Fatalf("decode nonce: %v", err)
	}
	ct, err := hex.DecodeString(wrapped)
	if err != nil {
		t.Fatalf("decode wrapped: %v", err)
	}
	return gcm.Open(nil, iv, ct, nil)
}

// unwrapEither thử giải mã một row bằng cả hai shared secret có thể; chỉ trả về
// đúng khóa của writer đã thắng cuộc đua (AES-GCM fail nếu khóa sai).
func unwrapEither(t *testing.T, sharedA, sharedB []byte, wrapped, nonce string) []byte {
	t.Helper()
	var found []byte
	for _, shared := range [][]byte{sharedA, sharedB} {
		plain, err := aesUnwrap(t, shared, wrapped, nonce)
		if err != nil {
			continue
		}
		if found != nil {
			t.Fatalf("row unwraps under cả hai shared secret — test setup lỗi")
		}
		found = plain
	}
	if found == nil {
		t.Fatalf("row không unwrap được dưới bất kỳ shared secret nào")
	}
	return found
}

// TestUpsertChatKeysConcurrentConsistency chứng minh race khóa chat đã được
// serialize: hai client tạo khóa đồng thời (khóa KA, KB khác nhau) cho cùng một
// chat phải kết thúc với hai row (mỗi participant một row) mang CÙNG một khóa.
// Nếu không có row-lock FOR UPDATE, hai insert xem kẽ nhau sẽ để lại khóa phân
// kỳ → test này fail.
func TestUpsertChatKeysConcurrentConsistency(t *testing.T) {
	db, e2eRepo := connectAndMigrateE2E(t)
	seedE2EChat(t, db, "chat-race-1")

	const (
		userA = "user-a"
		userB = "user-b"
	)
	sharedA := bytes.Repeat([]byte{0xA1}, 32)
	sharedB := bytes.Repeat([]byte{0xB2}, 32)
	keyKA := bytes.Repeat([]byte{0x11}, 32)
	keyKB := bytes.Repeat([]byte{0x22}, 32)

	wrapA1, nA1 := aesWrap(t, sharedA, keyKA)
	wrapA2, nA2 := aesWrap(t, sharedA, keyKA)
	wrapB1, nB1 := aesWrap(t, sharedB, keyKB)
	wrapB2, nB2 := aesWrap(t, sharedB, keyKB)

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e2eRepo.UpsertChatKeys(ctx, []models.ChatE2EKey{
			{ChatID: "chat-race-1", UserID: userA, WrappedKey: wrapA1, Nonce: nA1},
			{ChatID: "chat-race-1", UserID: userB, WrappedKey: wrapA2, Nonce: nA2},
		}); err != nil {
			t.Errorf("writer A: %v", err)
		}
	}()
	go func() {
		defer wg.Done()
		<-start
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e2eRepo.UpsertChatKeys(ctx, []models.ChatE2EKey{
			{ChatID: "chat-race-1", UserID: userA, WrappedKey: wrapB1, Nonce: nB1},
			{ChatID: "chat-race-1", UserID: userB, WrappedKey: wrapB2, Nonce: nB2},
		}); err != nil {
			t.Errorf("writer B: %v", err)
		}
	}()
	close(start)
	wg.Wait()

	ctx := context.Background()
	rowA, err := e2eRepo.GetChatKey(ctx, "chat-race-1", userA)
	if err != nil {
		t.Fatalf("get key A: %v", err)
	}
	rowB, err := e2eRepo.GetChatKey(ctx, "chat-race-1", userB)
	if err != nil {
		t.Fatalf("get key B: %v", err)
	}
	if rowA == nil || rowB == nil {
		t.Fatalf("thiếu row khóa: A=%v B=%v", rowA, rowB)
	}

	keyA := unwrapEither(t, sharedA, sharedB, rowA.WrappedKey, rowA.Nonce)
	keyB := unwrapEither(t, sharedA, sharedB, rowB.WrappedKey, rowB.Nonce)
	if !bytes.Equal(keyA, keyB) {
		t.Fatalf("hai participant giữ 2 khóa phân kỳ: %x != %x", keyA, keyB)
	}
}