package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/ws"
)

type VoiceCallService struct {
	callRepo     *repository.CallRepository
	friendRepo   *repository.FriendRepository
	profileRepo  *repository.ProfileRepository
	notifService *NotificationService
	hub          *ws.Hub
}

func NewVoiceCallService(callRepo *repository.CallRepository, friendRepo *repository.FriendRepository, profileRepo *repository.ProfileRepository, notifService *NotificationService, hub *ws.Hub) *VoiceCallService {
	return &VoiceCallService{
		callRepo:     callRepo,
		friendRepo:   friendRepo,
		profileRepo:  profileRepo,
		notifService: notifService,
		hub:          hub,
	}
}

func (s *VoiceCallService) InitiateCall(ctx context.Context, callerID string, payload dto.CallInitiatePayload) (*models.Call, error) {
	if callerID == payload.CalleeID {
		return nil, errorsapp.New(errorsapp.ErrCodeCallSelfCall)
	}

	callType := models.ParseCallType(payload.CallType)

	isFriend, err := s.friendRepo.IsAcceptedFriend(ctx, callerID, payload.CalleeID)
	if err != nil {
		return nil, fmt.Errorf("kiểm tra bạn bè: %w", err)
	}
	if !isFriend {
		return nil, errorsapp.New(errorsapp.ErrCodeCallNotFriend)
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

// AcceptCall atomically transitions a call to connected status.
func (s *VoiceCallService) AcceptCall(ctx context.Context, userID string, callID string) error {
	// The atomic method handles: call existence check, callee ownership check,
	// status validation, and update — all in a single conditional UPDATE.
	call, err := s.callRepo.AcceptCallAtomic(ctx, callID, userID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
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

// RejectCall atomically transitions a call to rejected status.
func (s *VoiceCallService) RejectCall(ctx context.Context, userID string, callID string) error {
	call, err := s.callRepo.RejectCallAtomic(ctx, callID, userID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
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
		return errorsapp.New(errorsapp.ErrCodeCallNotFound)
	}
	if userID != call.CallerID && userID != call.CalleeID {
		return errorsapp.New(errorsapp.ErrCodeCallNotParticipant)
	}
	if call.Status == models.CallStatusEnded || call.Status == models.CallStatusRejected || call.Status == models.CallStatusMissed || call.Status == models.CallStatusCancelled {
		return errorsapp.New(errorsapp.ErrCodeCallAlreadyEnded)
	}

	now := time.Now().UTC()
	var duration int
	if call.StartedAt != nil {
		duration = int(now.Sub(*call.StartedAt).Seconds())
	}

	newStatus := models.CallStatusEnded
	if call.Status == models.CallStatusCalling || call.Status == models.CallStatusRinging {
		if userID == call.CallerID {
			newStatus = models.CallStatusCancelled
		} else {
			newStatus = models.CallStatusMissed
		}
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

	// Real-time push: notify the other party based on who ended the call.
	switch newStatus {
	case models.CallStatusMissed:
		// Callee didn't answer — notify callee of missed call.
		s.hub.SendToUser(call.CalleeID, ws.OutgoingMessage{
			Type: "call:missed",
			Data: dto.CallMissedPayload{
				CallID:    call.ID,
				CallerID:  call.CallerID,
				Timestamp: now.UnixMilli(),
			},
		})
		if s.notifService != nil {
			senderID := call.CallerID
			s.notifService.Create(ctx, call.CalleeID, &senderID, models.NotificationTypeVoiceCall, "đã gọi nhỡ cho bạn", nil, &call.CallerID, nil)
		}
	case models.CallStatusCancelled:
		// Caller cancelled before answer — notify callee of cancelled call.
		s.hub.SendToUser(call.CalleeID, ws.OutgoingMessage{
			Type: "call:cancelled",
			Data: dto.CallMissedPayload{
				CallID:    call.ID,
				CallerID:  call.CallerID,
				Timestamp: now.UnixMilli(),
			},
		})
	}

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

// GetCallHistoryFiltered returns paginated, filtered call history enriched
// with the other party's profile info (display name, avatar).
func (s *VoiceCallService) GetCallHistoryFiltered(ctx context.Context, userID string, f repository.CallHistoryFilter) ([]dto.CallHistoryItem, int64, error) {
	calls, err := s.callRepo.GetHistoryFiltered(ctx, userID, f)
	if err != nil {
		return nil, 0, fmt.Errorf("lấy lịch sử cuộc gọi: %w", err)
	}
	total, err := s.callRepo.CountHistoryFiltered(ctx, userID, f)
	if err != nil {
		return nil, 0, fmt.Errorf("đếm lịch sử cuộc gọi: %w", err)
	}
	if len(calls) == 0 {
		return []dto.CallHistoryItem{}, total, nil
	}

	// Batch-load profiles of the "other user" for each call.
	otherIDs := make([]string, 0, len(calls))
	for i := range calls {
		if calls[i].CalleeID == userID {
			otherIDs = append(otherIDs, calls[i].CallerID)
		} else {
			otherIDs = append(otherIDs, calls[i].CalleeID)
		}
	}

	profiles, err := s.profileRepo.FindByIDs(ctx, otherIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("lấy hồ sơ người dùng: %w", err)
	}
	profileMap := make(map[string]models.Profile, len(profiles))
	for _, p := range profiles {
		profileMap[p.UserID] = p
	}

	items := make([]dto.CallHistoryItem, len(calls))
	for i, c := range calls {
		otherID := c.CallerID
		if c.CallerID == userID {
			otherID = c.CalleeID
		}

		brief := dto.UserBrief{ID: otherID, DisplayName: "Người dùng"}
		if p, ok := profileMap[otherID]; ok {
			brief = dto.UserBrief{
				ID:          otherID,
				DisplayName: p.DisplayName,
				AvatarURL:   p.AvatarURI,
			}
		}

		direction := "outgoing"
		if c.CallerID != userID {
			direction = "incoming"
		}

		var startedAt, endedAt *int64
		if c.StartedAt != nil {
			startedAt = ptr(c.StartedAt.UnixMilli())
		}
		if c.EndedAt != nil {
			endedAt = ptr(c.EndedAt.UnixMilli())
		}

		items[i] = dto.CallHistoryItem{
			ID:        c.ID,
			OtherUser: brief,
			CallType:  string(c.CallType),
			Direction: direction,
			Status:    string(c.Status),
			IsMissed:  c.Status == models.CallStatusMissed,
			Duration:  c.Duration,
			StartedAt: startedAt,
			EndedAt:   endedAt,
			CreatedAt: c.CreatedAt.UnixMilli(),
		}
	}
	return items, total, nil
}

// GetMissedCallCount returns the number of unread missed calls for the user.
// Reads the user's last_read_missed_at timestamp from their profile.
func (s *VoiceCallService) GetMissedCallCount(ctx context.Context, userID string) (int64, error) {
	profile, err := s.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("lấy hồ sơ: %w", err)
	}
	if profile == nil {
		return 0, nil
	}

	var since *time.Time
	if profile.LastReadMissedAt != nil {
		since = profile.LastReadMissedAt
	}

	count, err := s.callRepo.CountMissedSince(ctx, userID, since)
	if err != nil {
		return 0, fmt.Errorf("đếm cuộc gọi nhỡ: %w", err)
	}
	return count, nil
}

// MarkMissedAsRead records the current time as the user's last-read marker
// so future GetMissedCallCount only counts calls after this moment.
func (s *VoiceCallService) MarkMissedAsRead(ctx context.Context, userID string) error {
	if err := s.callRepo.MarkMissedRead(ctx, userID); err != nil {
		return fmt.Errorf("đánh dấu đã đọc: %w", err)
	}
	return nil
}

// HideCallFromHistory removes a call from the user's history view.
// The call remains visible to the other party.
func (s *VoiceCallService) HideCallFromHistory(ctx context.Context, userID, callID string) error {
	// Verify the user is a participant of this call.
	call, err := s.callRepo.FindByID(ctx, callID)
	if err != nil {
		return fmt.Errorf("tìm cuộc gọi: %w", err)
	}
	if call == nil {
		return errorsapp.New(errorsapp.ErrCodeCallNotFound)
	}
	if userID != call.CallerID && userID != call.CalleeID {
		return errorsapp.New(errorsapp.ErrCodeCallNotParticipant)
	}

	if err := s.callRepo.HideCall(ctx, userID, callID); err != nil {
		return fmt.Errorf("ẩn cuộc gọi: %w", err)
	}
	return nil
}

func (s *VoiceCallService) ToggleMute(ctx context.Context, userID string, callID string, muted bool) error {
	call, err := s.callRepo.FindByID(ctx, callID)
	if err != nil {
		return fmt.Errorf("tìm cuộc gọi: %w", err)
	}
	if call == nil {
		return errorsapp.New(errorsapp.ErrCodeCallNotFound)
	}
	if call.Status != models.CallStatusConnected {
		return errorsapp.New(errorsapp.ErrCodeCallNotConnected)
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
		return errorsapp.New(errorsapp.ErrCodeCallNotParticipant)
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
		return errorsapp.New(errorsapp.ErrCodeCallNotFound)
	}
	if call.CallType != models.CallTypeVideo {
		return errorsapp.New(errorsapp.ErrCodeCallNotVideo)
	}
	if call.Status != models.CallStatusConnected {
		return errorsapp.New(errorsapp.ErrCodeCallNotConnected)
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
		return errorsapp.New(errorsapp.ErrCodeCallNotParticipant)
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
		return nil, errorsapp.New(errorsapp.ErrCodeCallNotFound)
	}
	if userID != call.CallerID && userID != call.CalleeID {
		return nil, errorsapp.New(errorsapp.ErrCodeCallNotParticipant)
	}
	return call, nil
}

func (s *VoiceCallService) HandleSignal(ctx context.Context, senderID string, callID string, signal json.RawMessage) error {
	call, err := s.callRepo.FindByID(ctx, callID)
	if err != nil {
		return fmt.Errorf("tìm cuộc gọi: %w", err)
	}
	if call == nil {
		return errorsapp.New(errorsapp.ErrCodeCallNotFound)
	}
	if senderID != call.CallerID && senderID != call.CalleeID {
		return errorsapp.New(errorsapp.ErrCodeCallNotParticipant)
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
