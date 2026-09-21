package services

import (
	"bytes"
	"context"
	"testing"
	"time"

	"linkup/dto"
	"linkup/models"
	"linkup/repository"
)

// ─── B.1: key_version bump ───────────────────────────────────────

// TestUpsertUserKeyBumpsVersionOnChange — đăng ký cùng public key giữ nguyên
// version; đổi public key (identity/thiết bị mới) thì key_version tăng 1.
func TestUpsertUserKeyBumpsVersionOnChange(t *testing.T) {
	db, e2eRepo := connectAndMigrateE2E(t)
	ctx := context.Background()

	svc := NewE2EService(e2eRepo, repository.NewChatRepository(db))

	start := time.Now().UTC()
	if err := svc.RegisterUserKey(ctx, "user-a", "pub-alpha"); err != nil {
		t.Fatalf("register v1: %v", err)
	}
	key, err := e2eRepo.GetUserKey(ctx, "user-a")
	if err != nil {
		t.Fatalf("get key v1: %v", err)
	}
	if key == nil || key.KeyVersion != 1 {
		t.Fatalf("key after first register: key=%+v (want version 1)", key)
	}
	if key.PublicKey != "pub-alpha" {
		t.Fatalf("public key mismatch: %s", key.PublicKey)
	}

	// Cùng public key → không bump.
	if err := svc.RegisterUserKey(ctx, "user-a", "pub-alpha"); err != nil {
		t.Fatalf("re-register same key: %v", err)
	}
	key, _ = e2eRepo.GetUserKey(ctx, "user-a")
	if key.KeyVersion != 1 {
		t.Fatalf("same key bumped version to %d (want 1)", key.KeyVersion)
	}

	// Đổi public key → bump lên 2.
	if err := svc.RegisterUserKey(ctx, "user-a", "pub-beta"); err != nil {
		t.Fatalf("register v2: %v", err)
	}
	key, _ = e2eRepo.GetUserKey(ctx, "user-a")
	if key.KeyVersion != 2 {
		t.Fatalf("changed key kept version %d (want 2)", key.KeyVersion)
	}
	if key.PublicKey != "pub-beta" {
		t.Fatalf("public key after change: %s (want pub-beta)", key.PublicKey)
	}
	if !key.UpdatedAt.After(start) {
		t.Fatalf("updated_at not refreshed: created=%v updated=%v", key.CreatedAt, key.UpdatedAt)
	}
}

// ─── B.1: re-key khi đối phương đổi identity ─────────────────────

// TestRekeyChatOverwritesSelfRowOnly — RekeyChat chỉ ghi đè row (chat_id,
// user_id) của CHÍNH caller, không đụng row của người khác, và chỉ cho phép
// thành viên chat.
func TestRekeyChatOverwritesSelfRowOnly(t *testing.T) {
	db, e2eRepo := connectAndMigrateE2E(t)
	ctx := context.Background()

	svc := NewE2EService(e2eRepo, repository.NewChatRepository(db))

	const chatID = "chat-rekey-1"
	seedE2EChat(t, db, chatID)
	seedE2EParticipant(t, db, chatID, "user-a")
	seedE2EParticipant(t, db, chatID, "user-b")

	sharedAbefore := bytes.Repeat([]byte{0xA1}, 32) // shared secret cũ (A cũ × B cũ)
	sharedBbefore := bytes.Repeat([]byte{0xB2}, 32) // B bọc row của chính B
	sharedCafter := bytes.Repeat([]byte{0xC3}, 32)  // shared secret mới (A × B mới)
	chatKey := bytes.Repeat([]byte{0x42}, 32)

	// Row ban đầu: A bọc bằng shared cũ, B bọc bằng shared B.
	wrapAold, nAold := aesWrap(t, sharedAbefore, chatKey)
	wrapB, nB := aesWrap(t, sharedBbefore, chatKey)
	if err := e2eRepo.UpsertChatKey(ctx, &models.ChatE2EKey{
		ChatID: chatID, UserID: "user-a", WrappedKey: wrapAold, Nonce: nAold,
	}); err != nil {
		t.Fatalf("seed key A: %v", err)
	}
	if err := e2eRepo.UpsertChatKey(ctx, &models.ChatE2EKey{
		ChatID: chatID, UserID: "user-b", WrappedKey: wrapB, Nonce: nB,
	}); err != nil {
		t.Fatalf("seed key B: %v", err)
	}

	// A re-wrap khóa chuẩn bằng shared secret mới và ghi đè row của mình.
	wrapAnew, nAnew := aesWrap(t, sharedCafter, chatKey)
	if err := svc.RekeyChat(ctx, "user-a", chatID, nAnew, wrapAnew); err != nil {
		t.Fatalf("rekey A: %v", err)
	}

	rowA, err := e2eRepo.GetChatKey(ctx, chatID, "user-a")
	if err != nil {
		t.Fatalf("get key A after rekey: %v", err)
	}
	if rowA == nil {
		t.Fatal("row A missing after rekey")
	}
	gotA := unwrapEither(t, sharedAbefore, sharedCafter, rowA.WrappedKey, rowA.Nonce)
	if !bytes.Equal(gotA, chatKey) {
		t.Fatalf("row A does not hold the canonical chat key: %x", gotA)
	}

	// Row B không bị đụng.
	rowB, err := e2eRepo.GetChatKey(ctx, chatID, "user-b")
	if err != nil {
		t.Fatalf("get key B: %v", err)
	}
	if rowB == nil {
		t.Fatal("row B missing after rekey of A")
	}
	if rowB.WrappedKey != wrapB || rowB.Nonce != nB {
		t.Fatalf("row B changed by A's rekey: %+v", rowB)
	}
	gotB := unwrapEither(t, sharedBbefore, sharedAbefore, rowB.WrappedKey, rowB.Nonce)
	if !bytes.Equal(gotB, chatKey) {
		t.Fatalf("row B does not hold the canonical chat key: %x", gotB)
	}

	// Người không phải thành viên không được rekey.
	const chat2 = "chat-rekey-2"
	seedE2EChat(t, db, chat2)
	seedE2EParticipant(t, db, chat2, "user-owner")
	if err := svc.RekeyChat(ctx, "user-outside", chat2, "n", "w"); err == nil {
		t.Fatal("rekey by non-participant should fail")
	}
}

// ─── B.3: put / meta / delete recovery ───────────────────────────

func TestPutGetMetaDeleteRecovery(t *testing.T) {
	db, e2eRepo := connectAndMigrateE2E(t)
	ctx := context.Background()

	svc := NewE2EService(e2eRepo, repository.NewChatRepository(db))

	// Chưa từng bật → has_blob=false.
	meta, err := svc.GetRecoveryMeta(ctx, "user-rec")
	if err != nil {
		t.Fatalf("get meta (empty): %v", err)
	}
	if meta == nil || meta.HasBlob {
		t.Fatalf("empty recovery meta: %+v", meta)
	}

	input := dto.PutRecoveryRequest{
		Salt:          "salt-b64",
		Blob:          "blob-b64",
		PinCheck:      "pin-check",
		RecoveryCheck: "recovery-check",
	}
	if err := svc.PutRecovery(ctx, "user-rec", input); err != nil {
		t.Fatalf("put recovery: %v", err)
	}

	meta, err = svc.GetRecoveryMeta(ctx, "user-rec")
	if err != nil {
		t.Fatalf("get meta: %v", err)
	}
	if meta == nil || !meta.HasBlob || meta.Salt != input.Salt || meta.UpdatedAt == nil {
		t.Fatalf("recovery meta after put: %+v", meta)
	}

	if err := svc.DeleteRecovery(ctx, "user-rec"); err != nil {
		t.Fatalf("delete recovery: %v", err)
	}
	meta, _ = svc.GetRecoveryMeta(ctx, "user-rec")
	if meta == nil || meta.HasBlob {
		t.Fatalf("meta after delete: %+v (want has_blob=false)", meta)
	}
}

// TestUnlockRecoveryRoundTripAndReset — mở khóa đúng trả blob; nhập sai tăng
// bộ đếm; mở khóa đúng reset bộ đếm về 0.
func TestUnlockRecoveryRoundTripAndReset(t *testing.T) {
	db, e2eRepo := connectAndMigrateE2E(t)
	ctx := context.Background()

	svc := NewE2EService(e2eRepo, repository.NewChatRepository(db))

	const userID = "user-rec"
	input := dto.PutRecoveryRequest{
		Salt:          "salt-b64",
		Blob:          "blob-b64",
		PinCheck:      "pin-check",
		RecoveryCheck: "recovery-check",
	}
	if err := svc.PutRecovery(ctx, userID, input); err != nil {
		t.Fatalf("put recovery: %v", err)
	}

	// Sai 3 lần (dưới ngưỡng lock).
	for i := 0; i < 3; i++ {
		if _, err := svc.UnlockRecovery(ctx, userID, "wrong"); err == nil {
			t.Fatalf("wrong attempt %d should fail", i+1)
		}
	}
	rec, err := e2eRepo.GetRecovery(ctx, userID)
	if err != nil {
		t.Fatalf("get recovery after wrong attempts: %v", err)
	}
	if rec.Attempts != 3 {
		t.Fatalf("attempts after 3 wrong: %d (want 3)", rec.Attempts)
	}
	if rec.LockedUntil != nil {
		t.Fatalf("should not be locked at attempts=%d", rec.Attempts)
	}

	// Mở khóa bằng recovery check → trả blob, reset attempts.
	blob, err := svc.UnlockRecovery(ctx, userID, input.RecoveryCheck)
	if err != nil {
		t.Fatalf("unlock with recovery check: %v", err)
	}
	if blob != input.Blob {
		t.Fatalf("unlock returned blob %q (want %q)", blob, input.Blob)
	}
	rec, _ = e2eRepo.GetRecovery(ctx, userID)
	if rec.Attempts != 0 || rec.LockedUntil != nil {
		t.Fatalf("attempts/lock not reset after success: attempts=%d locked=%v", rec.Attempts, rec.LockedUntil)
	}
}

// TestUnlockRecoveryRateLimit — quá recoveryMaxAttempts lần sai liên tiếp →
// khóa tạm thời với locked_until; trong lúc khóa, kể cả check đúng cũng bị từ
// chối; hết hạn → mở khóa thành công và reset.
func TestUnlockRecoveryRateLimit(t *testing.T) {
	db, e2eRepo := connectAndMigrateE2E(t)
	ctx := context.Background()

	svc := NewE2EService(e2eRepo, repository.NewChatRepository(db))

	const userID = "user-lock"
	input := dto.PutRecoveryRequest{
		Salt:          "salt-b64",
		Blob:          "blob-b64",
		PinCheck:      "pin-check",
		RecoveryCheck: "recovery-check",
	}
	if err := svc.PutRecovery(ctx, userID, input); err != nil {
		t.Fatalf("put recovery: %v", err)
	}

	// Sai đủ recoveryMaxAttempts lần liên tiếp → lần cuối đặt khóa.
	const max = recoveryMaxAttempts // = 5
	before := time.Now().UTC()
	for i := 0; i < max; i++ {
		if _, err := svc.UnlockRecovery(ctx, userID, "wrong"); err == nil {
			t.Fatalf("wrong attempt %d should fail", i+1)
		}
	}
	rec, err := e2eRepo.GetRecovery(ctx, userID)
	if err != nil {
		t.Fatalf("get recovery after lock: %v", err)
	}
	if rec.Attempts != max {
		t.Fatalf("attempts after %d wrong: %d (want %d)", max, rec.Attempts, max)
	}
	if rec.LockedUntil == nil {
		t.Fatal("locked_until not set after hitting the attempt limit")
	}
	if !rec.LockedUntil.After(before) {
		t.Fatalf("locked_until %v is not in the future (now %v)", rec.LockedUntil, before)
	}

	// Đang khóa → check ĐÚNG vẫn bị từ chối.
	if _, err := svc.UnlockRecovery(ctx, userID, input.PinCheck); err == nil {
		t.Fatal("correct pin should still be rejected while locked")
	}

	// Hết hạn khóa (chỉnh backdated) → mở khóa thành công + reset.
	past := time.Now().UTC().Add(-time.Hour)
	if err := db.Model(&models.UserE2ERecovery{}).
		Where("user_id = ?", userID).
		Update("locked_until", past).Error; err != nil {
		t.Fatalf("backdate locked_until: %v", err)
	}
	blob, err := svc.UnlockRecovery(ctx, userID, input.PinCheck)
	if err != nil {
		t.Fatalf("unlock after lock expiry: %v", err)
	}
	if blob != input.Blob {
		t.Fatalf("unlock after expiry returned %q (want %q)", blob, input.Blob)
	}
	rec, _ = e2eRepo.GetRecovery(ctx, userID)
	if rec.Attempts != 0 || rec.LockedUntil != nil {
		t.Fatalf("attempts/lock not reset after unlock: attempts=%d locked=%v", rec.Attempts, rec.LockedUntil)
	}
}