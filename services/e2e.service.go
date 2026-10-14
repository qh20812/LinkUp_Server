package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"linkup/ws"
)

type E2EService struct {
	e2eRepo *repository.E2ERepository
	chatRepo *repository.ChatRepository
	// chatHub là Hub của /api/chats/ws (khác hub notification ở /api/ws).
	// Optional — nil khi chưa wire (test/service-only) thì bỏ qua notify.
	chatHub *ws.Hub
}

func NewE2EService(e2eRepo *repository.E2ERepository, chatRepo *repository.ChatRepository) *E2EService {
	return &E2EService{e2eRepo: e2eRepo, chatRepo: chatRepo}
}

// SetChatHub gắn Hub chat để phát event key-updated cho participant đang online.
func (s *E2EService) SetChatHub(h *ws.Hub) {
	s.chatHub = h
}

func (s *E2EService) RegisterUserKey(ctx context.Context, userID, publicKey string) error {
	now := time.Now().UTC()
	key := models.UserE2EKey{
		UserID:    userID,
		PublicKey: publicKey,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.e2eRepo.UpsertUserKey(ctx, &key)
}

func (s *E2EService) GetUserKey(ctx context.Context, userID string) (*models.UserE2EKey, error) {
	return s.e2eRepo.GetUserKey(ctx, userID)
}

// RekeyChat cập nhật khóa bọc của CHÍNH user khi đối phương đổi identity:
// client re-wrap khóa chuẩn (đang giữ local) bằng shared secret mới rồi ghi đè
// row (chat_id, user_id) của mình, để người ở thiết bị mới vẫn unwrap được khóa
// cũ. Không tạo khóa chat mới, không đụng row của người khác.
func (s *E2EService) RekeyChat(ctx context.Context, callerID, chatID, nonce, wrappedKey string) error {
	ok, err := s.chatRepo.IsUserParticipant(ctx, chatID, callerID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("not a participant of chat %s", chatID)
	}

	key := models.ChatE2EKey{
		ChatID:     chatID,
		UserID:     callerID,
		WrappedKey: wrappedKey,
		Nonce:      nonce,
		CreatedAt:  time.Now().UTC(),
	}
	return s.e2eRepo.UpsertChatKeyForce(ctx, &key)
}

// StoreChatKeys stores the wrapped chat key for each participant. The caller
// must be a participant of every chat being registered.
func (s *E2EService) StoreChatKeys(ctx context.Context, callerID string, inputs []dto.ChatE2EKeyInput) error {
	if len(inputs) == 0 {
		return nil
	}

	now := time.Now().UTC()
	keys := make([]models.ChatE2EKey, 0, len(inputs))
	for _, in := range inputs {
		ok, err := s.chatRepo.IsUserParticipant(ctx, in.ChatID, callerID)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("not a participant of chat %s", in.ChatID)
		}

		keys = append(keys, models.ChatE2EKey{
			ChatID:     in.ChatID,
			UserID:     in.UserID,
			WrappedKey: in.WrappedKey,
			Nonce:      in.Nonce,
			CreatedAt:  now,
		})
	}
	if err := s.e2eRepo.UpsertChatKeys(ctx, keys); err != nil {
		return err
	}
	s.notifyKeyUpdated(keys, callerID)
	return nil
}

// keyUpdateRecipients gom user cần báo theo từng chat: mọi user trong batch
// trừ caller, khử trùng lặp. Hàm thuần để unit-test không cần DB/hub.
func keyUpdateRecipients(keys []models.ChatE2EKey, callerID string) map[string][]string {
	out := make(map[string][]string)
	seen := make(map[string]map[string]bool)
	for _, k := range keys {
		if k.UserID == callerID {
			continue
		}
		if seen[k.ChatID] == nil {
			seen[k.ChatID] = make(map[string]bool)
		}
		if seen[k.ChatID][k.UserID] {
			continue
		}
		seen[k.ChatID][k.UserID] = true
		out[k.ChatID] = append(out[k.ChatID], k.UserID)
	}
	return out
}

// notifyKeyUpdated báo cho các participant còn lại (đang online) rằng khóa E2E
// của chat đã đổi để client tự adopt + giải mã lại tin đang kẹt — thay vì chờ
// user F5. Bắn trực tiếp WsEvent wire-format của chat client.
func (s *E2EService) notifyKeyUpdated(keys []models.ChatE2EKey, callerID string) {
	if s.chatHub == nil {
		return
	}
	for chatID, userIDs := range keyUpdateRecipients(keys, callerID) {
		payload, err := json.Marshal(map[string]string{"chat_id": chatID})
		if err != nil {
			continue
		}
		raw, err := json.Marshal(dto.WsEvent{Type: "chat:e2e_key_updated", Payload: payload})
		if err != nil {
			continue
		}
		for _, uid := range userIDs {
			s.chatHub.SendRawToUser(uid, raw)
		}
	}
}

func (s *E2EService) GetChatKey(ctx context.Context, userID, chatID string) (*dto.ChatE2EKeyResponse, error) {
	ok, err := s.chatRepo.IsUserParticipant(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("not a participant of chat %s", chatID)
	}

	key, err := s.e2eRepo.GetChatKey(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, nil
	}
	return &dto.ChatE2EKeyResponse{
		ChatID:     key.ChatID,
		WrappedKey: key.WrappedKey,
		Nonce:      key.Nonce,
	}, nil
}

// recoveryMaxAttempts — số lần nhập sai tối đa trước khi khóa tạm thời.
const recoveryMaxAttempts = 5

// recoveryLockBaseMinutes — thời gian khóa cơ sở (phút), nhân dần theo số lần
// đã bị khóa (rate-limit tăng dần).
const recoveryLockBaseMinutes = 10

// PutRecovery lưu (hoặc thay mới) backup khôi phục khóa chat của user. Blob,
// salt và hai hash check đều do client tạo; server chỉ lưu opaque. Bất kỳ lần
// bật/đổi PIN nào cũng reset bộ đếm thử sai.
func (s *E2EService) PutRecovery(ctx context.Context, userID string, input dto.PutRecoveryRequest) error {
	rec := &models.UserE2ERecovery{
		UserID:        userID,
		Salt:          input.Salt,
		Blob:          input.Blob,
		PinCheck:      input.PinCheck,
		RecoveryCheck: input.RecoveryCheck,
		Attempts:      0,
		LockedUntil:   nil,
		UpdatedAt:     time.Now().UTC(),
	}
	return s.e2eRepo.UpsertRecovery(ctx, rec)
}

// GetRecoveryMeta trả về trạng thái backup khôi phục. Trả has_blob=false nếu
// user chưa từng bật; salt được trả vì client cần nó để dẫn khóa giải mã sau
// khi mở khóa đúng.
func (s *E2EService) GetRecoveryMeta(ctx context.Context, userID string) (*dto.RecoveryMetaResponse, error) {
	rec, err := s.e2eRepo.GetRecovery(ctx, userID)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return &dto.RecoveryMetaResponse{HasBlob: false}, nil
	}
	updatedAt := rec.UpdatedAt
	return &dto.RecoveryMetaResponse{
		HasBlob:   true,
		Salt:      rec.Salt,
		UpdatedAt: &updatedAt,
	}, nil
}

// UnlockRecovery xác thực giá trị check (băm của PIN hoặc recovery key) để lấy
// blob backup. Sai quá recoveryMaxAttempts lần → khóa tạm thời với thời gian
// nhân dần: lần khóa thứ n bị chờ recoveryLockBaseMinutes × n phút.
func (s *E2EService) UnlockRecovery(ctx context.Context, userID, check string) (string, error) {
	rec, err := s.e2eRepo.GetRecovery(ctx, userID)
	if err != nil {
		return "", err
	}
	if rec == nil {
		return "", fmt.Errorf("chưa có dữ liệu khôi phục")
	}

	now := time.Now().UTC()
	if rec.LockedUntil != nil && now.Before(*rec.LockedUntil) {
		return "", fmt.Errorf("đã khóa do nhập sai nhiều lần, thử lại sau ít phút")
	}

	if rec.PinCheck != check && rec.RecoveryCheck != check {
		// Nhập sai → tăng bộ đếm. Cứ mỗi 5 lần sai liên tiếp thì khóa thêm một
		// khoảng tăng dần (10 × số lần đã khóa).
		var lockedUntil *time.Time
		if rec.Attempts+1 >= recoveryMaxAttempts {
			lockNumber := (rec.Attempts / recoveryMaxAttempts) + 1
			t := now.Add(time.Duration(recoveryLockBaseMinutes*lockNumber) * time.Minute)
			lockedUntil = &t
		}
		if lockErr := s.e2eRepo.InvalidateRecoveryAttempt(ctx, userID, lockedUntil); lockErr != nil {
			return "", lockErr
		}
		return "", fmt.Errorf("mã xác minh không đúng")
	}

	// Mở khóa đúng → reset bộ đếm để lần sau không bị khóa oan.
	rec.Attempts = 0
	rec.LockedUntil = nil
	rec.UpdatedAt = now
	if resetErr := s.e2eRepo.UpsertRecovery(ctx, rec); resetErr != nil {
		return "", resetErr
	}
	return rec.Blob, nil
}

// DeleteRecovery xóa backup khôi phục (user tắt tính năng khôi phục).
func (s *E2EService) DeleteRecovery(ctx context.Context, userID string) error {
	return s.e2eRepo.DeleteRecovery(ctx, userID)
}
