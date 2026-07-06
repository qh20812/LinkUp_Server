package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/ws"
)

type VoiceCallService struct {
	callRepo   *repository.CallRepository
	friendRepo *repository.FriendRepository
	hub        *ws.Hub
}

func NewVoiceCallService(callRepo *repository.CallRepository, friendRepo *repository.FriendRepository, hub *ws.Hub) *VoiceCallService {
	return &VoiceCallService{
		callRepo:   callRepo,
		friendRepo: friendRepo,
		hub:        hub,
	}
}

func (s *VoiceCallService) InitiateCall(ctx context.Context, callerID string, payload dto.CallInitiatePayload) (*models.Call, error) {
	if callerID == payload.CalleeID {
		return nil, errors.New("không thể gọi cho chính mình")
	}

	callType := models.ParseCallType(payload.CallType)

	activeAsCaller, err := s.callRepo.FindActiveByUserID(ctx, callerID)
	if err != nil {
		return nil, fmt.Errorf("kiểm tra trạng thái: %w", err)
	}
	if activeAsCaller != nil {
		return nil, errors.New("bạn đang có cuộc gọi khác")
	}

	isFriend, err := s.friendRepo.IsAcceptedFriend(ctx, callerID, payload.CalleeID)
	if err != nil {
		return nil, fmt.Errorf("kiểm tra bạn bè: %w", err)
	}
	if !isFriend {
		return nil, errors.New("chỉ có thể gọi cho bạn bè")
	}

	now := time.Now().UTC()
	call := &models.Call{
		ID:                 utils.GenerateUUID(),
		CallerID:           callerID,
		CalleeID:           payload.CalleeID,
		CallType:           callType,
		IsGroup:            false,
		Status:             models.CallStatusCalling,
		StartedAt:          nil,
		EndedAt:            nil,
		Duration:           0,
		VideoEnabledCaller: false,
		VideoEnabledCallee: false,
		CreatedAt:          now,
	}

	isBusy, err := s.callRepo.CreateIfNotBusy(ctx, call)
	if err != nil {
		return nil, fmt.Errorf("tạo cuộc gọi: %w", err)
	}
	if isBusy {
		s.hub.SendToUser(callerID, ws.OutgoingMessage{
			Type: "call:busy",
			Data: dto.CallBusyPayload{CalleeID: payload.CalleeID},
		})
		return nil, nil
	}

	if s.hub.IsUserOnline(payload.CalleeID) {
		s.hub.SendToUser(payload.CalleeID, ws.OutgoingMessage{
			Type: "call:incoming",
			Data: dto.CallIncomingPayload{
				CallID:    call.ID,
				CallerID:  callerID,
				CallType:  string(callType),
				Timestamp: now.UnixMilli(),
			},
		})
	}

	s.hub.SendToUser(callerID, ws.OutgoingMessage{
		Type: "call:status",
		Data: dto.CallStatusPayload{
			CallID:             call.ID,
			Status:             string(models.CallStatusCalling),
			CallerID:           callerID,
			CalleeID:           payload.CalleeID,
			CallType:           string(callType),
			VideoEnabledCaller: false,
			VideoEnabledCallee: false,
		},
	})

	return call, nil
}

func (s *VoiceCallService) AcceptCall(ctx context.Context, userID string, callID string) error {
	call, err := s.callRepo.FindByID(ctx, callID)
	if err != nil {
		return fmt.Errorf("tìm cuộc gọi: %w", err)
	}
	if call == nil {
		return errors.New("cuộc gọi không tồn tại")
	}
	if userID != call.CalleeID {
		return errors.New("chỉ người nhận mới có thể chấp nhận cuộc gọi")
	}
	if call.Status != models.CallStatusCalling && call.Status != models.CallStatusRinging {
		return errors.New("cuộc gọi không ở trạng thái chờ")
	}

	now := time.Now().UTC()
	if err := s.callRepo.UpdateStatus(ctx, callID, models.CallStatusConnected, &now, nil, 0); err != nil {
		return fmt.Errorf("cập nhật trạng thái: %w", err)
	}

	payload := dto.CallStatusPayload{
		CallID:             call.ID,
		Status:             string(models.CallStatusConnected),
		CallerID:           call.CallerID,
		CalleeID:           call.CalleeID,
		CallType:           string(call.CallType),
		VideoEnabledCaller: call.VideoEnabledCaller,
		VideoEnabledCallee: call.VideoEnabledCallee,
		StartedAt:          ptr(now.UnixMilli()),
	}

	s.hub.SendToUsers([]string{call.CallerID, call.CalleeID}, ws.OutgoingMessage{
		Type: "call:status",
		Data: payload,
	})

	return nil
}

func (s *VoiceCallService) RejectCall(ctx context.Context, userID string, callID string) error {
	call, err := s.callRepo.FindByID(ctx, callID)
	if err != nil {
		return fmt.Errorf("tìm cuộc gọi: %w", err)
	}
	if call == nil {
		return errors.New("cuộc gọi không tồn tại")
	}
	if userID != call.CalleeID {
		return errors.New("chỉ người nhận mới có thể từ chối cuộc gọi")
	}
	if call.Status != models.CallStatusCalling && call.Status != models.CallStatusRinging {
		return errors.New("cuộc gọi không ở trạng thái chờ")
	}

	now := time.Now().UTC()
	if err := s.callRepo.UpdateStatus(ctx, callID, models.CallStatusRejected, nil, &now, 0); err != nil {
		return fmt.Errorf("cập nhật trạng thái: %w", err)
	}

	payload := dto.CallStatusPayload{
		CallID:             call.ID,
		Status:             string(models.CallStatusRejected),
		CallerID:           call.CallerID,
		CalleeID:           call.CalleeID,
		CallType:           string(call.CallType),
		VideoEnabledCaller: call.VideoEnabledCaller,
		VideoEnabledCallee: call.VideoEnabledCallee,
		EndedAt:            ptr(now.UnixMilli()),
	}

	s.hub.SendToUsers([]string{call.CallerID, call.CalleeID}, ws.OutgoingMessage{
		Type: "call:status",
		Data: payload,
	})

	return nil
}

func (s *VoiceCallService) EndCall(ctx context.Context, userID string, callID string) error {
	call, err := s.callRepo.FindByID(ctx, callID)
	if err != nil {
		return fmt.Errorf("tìm cuộc gọi: %w", err)
	}
	if call == nil {
		return errors.New("cuộc gọi không tồn tại")
	}
	if userID != call.CallerID && userID != call.CalleeID {
		return errors.New("không phải người tham gia cuộc gọi")
	}
	if call.Status == models.CallStatusEnded || call.Status == models.CallStatusRejected || call.Status == models.CallStatusMissed {
		return errors.New("cuộc gọi đã kết thúc")
	}

	now := time.Now().UTC()
	var duration int
	if call.StartedAt != nil {
		duration = int(now.Sub(*call.StartedAt).Seconds())
	}

	newStatus := models.CallStatusEnded
	if call.Status == models.CallStatusCalling || call.Status == models.CallStatusRinging {
		newStatus = models.CallStatusMissed
	}

	if err := s.callRepo.UpdateStatus(ctx, callID, newStatus, nil, &now, duration); err != nil {
		return fmt.Errorf("cập nhật trạng thái: %w", err)
	}

	payload := dto.CallStatusPayload{
		CallID:             call.ID,
		Status:             string(newStatus),
		CallerID:           call.CallerID,
		CalleeID:           call.CalleeID,
		CallType:           string(call.CallType),
		VideoEnabledCaller: call.VideoEnabledCaller,
		VideoEnabledCallee: call.VideoEnabledCallee,
		EndedAt:            ptr(now.UnixMilli()),
		Duration:           duration,
	}

	s.hub.SendToUsers([]string{call.CallerID, call.CalleeID}, ws.OutgoingMessage{
		Type: "call:status",
		Data: payload,
	})

	return nil
}

func (s *VoiceCallService) GetCallHistory(ctx context.Context, userID string, limit, offset int) ([]models.Call, int64, error) {
	calls, err := s.callRepo.GetHistory(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("lấy lịch sử cuộc gọi: %w", err)
	}
	total, err := s.callRepo.CountHistory(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("đếm lịch sử cuộc gọi: %w", err)
	}
	return calls, total, nil
}

func (s *VoiceCallService) ToggleMute(ctx context.Context, userID string, callID string, muted bool) error {
	call, err := s.callRepo.FindByID(ctx, callID)
	if err != nil {
		return fmt.Errorf("tìm cuộc gọi: %w", err)
	}
	if call == nil {
		return errors.New("cuộc gọi không tồn tại")
	}
	if call.Status != models.CallStatusConnected {
		return errors.New("cuộc gọi không ở trạng thái kết nối")
	}

	switch userID {
	case call.CallerID:
		if err := s.callRepo.UpdateMuted(ctx, callID, &muted, nil); err != nil {
			return fmt.Errorf("cập nhật mute: %w", err)
		}
	case call.CalleeID:
		if err := s.callRepo.UpdateMuted(ctx, callID, nil, &muted); err != nil {
			return fmt.Errorf("cập nhật mute: %w", err)
		}
	default:
		return errors.New("không phải người tham gia cuộc gọi")
	}

	s.hub.SendToUsers([]string{call.CallerID, call.CalleeID}, ws.OutgoingMessage{
		Type: "call:mute",
		Data: map[string]interface{}{
			"call_id": callID,
			"user_id": userID,
			"muted":   muted,
		},
	})

	return nil
}

func (s *VoiceCallService) ToggleVideo(ctx context.Context, userID string, callID string, videoEnabled bool) error {
	call, err := s.callRepo.FindByID(ctx, callID)
	if err != nil {
		return fmt.Errorf("tìm cuộc gọi: %w", err)
	}
	if call == nil {
		return errors.New("cuộc gọi không tồn tại")
	}
	if call.CallType != models.CallTypeVideo {
		return errors.New("cuộc gọi không phải video call")
	}
	if call.Status != models.CallStatusConnected {
		return errors.New("cuộc gọi không ở trạng thái kết nối")
	}

	switch userID {
	case call.CallerID:
		if err := s.callRepo.UpdateVideoEnabled(ctx, callID, &videoEnabled, nil); err != nil {
			return fmt.Errorf("cập nhật video: %w", err)
		}
	case call.CalleeID:
		if err := s.callRepo.UpdateVideoEnabled(ctx, callID, nil, &videoEnabled); err != nil {
			return fmt.Errorf("cập nhật video: %w", err)
		}
	default:
		return errors.New("không phải người tham gia cuộc gọi")
	}

	s.hub.SendToUsers([]string{call.CallerID, call.CalleeID}, ws.OutgoingMessage{
		Type: "call:video",
		Data: map[string]interface{}{
			"call_id":       callID,
			"user_id":       userID,
			"video_enabled": videoEnabled,
		},
	})

	return nil
}

func (s *VoiceCallService) GetCallDetail(ctx context.Context, userID string, callID string) (*models.Call, error) {
	call, err := s.callRepo.FindByID(ctx, callID)
	if err != nil {
		return nil, fmt.Errorf("tìm cuộc gọi: %w", err)
	}
	if call == nil {
		return nil, errors.New("cuộc gọi không tồn tại")
	}
	if userID != call.CallerID && userID != call.CalleeID {
		return nil, errors.New("không phải người tham gia cuộc gọi")
	}
	return call, nil
}

func (s *VoiceCallService) HandleSignal(ctx context.Context, senderID string, callID string, signal json.RawMessage) error {
	call, err := s.callRepo.FindByID(ctx, callID)
	if err != nil {
		return fmt.Errorf("tìm cuộc gọi: %w", err)
	}
	if call == nil {
		return errors.New("cuộc gọi không tồn tại")
	}
	if senderID != call.CallerID && senderID != call.CalleeID {
		return errors.New("không phải người tham gia cuộc gọi")
	}

	receiverID := call.CalleeID
	if senderID == call.CalleeID {
		receiverID = call.CallerID
	}

	s.hub.SendToUser(receiverID, ws.OutgoingMessage{
		Type: "call:signal",
		Data: dto.CallSignalPayload{
			CallID:   callID,
			SenderID: senderID,
			Signal:   signal,
		},
	})

	return nil
}

func ptr[T any](v T) *T {
	return &v
}
