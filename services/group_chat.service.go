package services

import (
	"context"
	"errors"
	"fmt"
	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"time"
	"unicode/utf8"
)

type GroupChatService struct {
	groupRepo *repository.GroupChatRepository
	chatRepo  *repository.ChatRepository // Inject thêm để hỗ trợ một số kiểm tra chéo nếu cần
}

func NewGroupChatService(groupRepo *repository.GroupChatRepository, chatRepo *repository.ChatRepository) *GroupChatService {
	return &GroupChatService{
		groupRepo: groupRepo,
		chatRepo:  chatRepo,
	}
}

func (s *GroupChatService) CreateGroup(ctx context.Context, userID string, name, avatarURI string, memberIDs []string) (*models.Chat, error) {
	if len(name) < 3 || len(name) > 50 {
		return nil, errors.New("tên nhóm chat phải từ 3 đến 50 ký tự")
	}

	group := models.NewChat(models.ChatTypeGroup, name, avatarURI)
	group.ID = utils.GenerateUUID()
	group.CreatedAt = time.Now().UTC()

	// Khởi tạo khóa mã hóa cho Group Chat
	encKey, err := utils.GenerateEncryptionKey()
	if err != nil {
		return nil, fmt.Errorf("không thể khởi tạo khóa bảo mật cho nhóm: %w", err)
	}
	group.EncryptionKey = encKey

	// 🔥 LƯU Ý: Đảm bảo models.ChatRoleAdmin dưới đây trả về đúng chuỗi "CHAT_ADMIN" của bạn
	adminPart := models.NewChatParticipant(group.ID, userID, models.ChatRoleAdmin)
	adminPart.ID = utils.GenerateUUID()
	adminPart.JoinedAt = time.Now().UTC()

	participants := []models.ChatParticipant{adminPart}

	for _, memberID := range memberIDs {
		if memberID == userID || memberID == "" {
			continue
		}

		memberPart := models.NewChatParticipant(group.ID, memberID, models.ChatRoleMember)
		memberPart.ID = utils.GenerateUUID()
		memberPart.JoinedAt = time.Now().UTC()

		participants = append(participants, memberPart)
	}

	if err := s.groupRepo.CreateGroup(ctx, &group, participants); err != nil {
		return nil, err
	}
	return &group, nil
}

// 1. CHỨC NĂNG: RỜI KHỎI NHÓM CHAT
func (s *GroupChatService) LeaveGroup(ctx context.Context, chatID, userID string) error {
	isMember, err := s.groupRepo.IsUserMember(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("bạn không phải là thành viên của nhóm này")
	}

	return s.groupRepo.LeaveGroup(ctx, chatID, userID)
}

// 2. CHỨC NĂNG: BAN THÀNH VIÊN (CHẶN QUAY LẠI)
func (s *GroupChatService) BanMember(ctx context.Context, chatID, adminID, targetUserID string) error {
	if adminID == targetUserID {
		return errors.New("bạn không thể tự chặn chính mình")
	}

	// Kiểm tra quyền CHAT_ADMIN
	isAdmin, err := s.groupRepo.IsUserAdmin(ctx, chatID, adminID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("chỉ quản trị viên (CHAT_ADMIN) mới có quyền chặn thành viên")
	}

	// Kiểm tra xem mục tiêu đã bị ban trước đó chưa
	isBanned, err := s.groupRepo.IsUserBanned(ctx, chatID, targetUserID)
	if err != nil {
		return err
	}
	if isBanned {
		return errors.New("người dùng này đã bị chặn từ trước")
	}

	banData := &models.GroupChatBan{
		ID:        utils.GenerateUUID(),
		ChatID:    chatID,
		UserID:    targetUserID,
		BannedBy:  adminID,
		CreatedAt: time.Now().UTC(),
	}

	return s.groupRepo.BanUser(ctx, banData)
}

// 3. CẬP NHẬT CHỨC NĂNG: THÊM THÀNH VIÊN (TÍCH HỢP KIỂM TRA BAN REJOIN)
func (s *GroupChatService) AddMember(ctx context.Context, chatID, requesterID, newMemberID string) error {
	isRequesterMember, err := s.groupRepo.IsUserMember(ctx, chatID, requesterID)
	if err != nil {
		return err
	}
	if !isRequesterMember {
		return errors.New("bạn không phải thành viên của nhóm này nên không có quyền mời người khác")
	}

	// [BAN REJOIN SECURITY CHECK]: Ngăn chặn người bị Ban quay trở lại nhóm
	isBanned, err := s.groupRepo.IsUserBanned(ctx, chatID, newMemberID)
	if err != nil {
		return err
	}
	if isBanned {
		return errors.New("người dùng này đã bị chặn bởi Admin và không thể tham gia lại nhóm")
	}

	isTargetMember, err := s.groupRepo.IsUserMember(ctx, chatID, newMemberID)
	if err != nil {
		return err
	}
	if isTargetMember {
		return errors.New("người dùng này đã là thành viên của nhóm từ trước")
	}

	participant := models.NewChatParticipant(chatID, newMemberID, models.ChatRoleMember)
	participant.ID = utils.GenerateUUID()
	participant.JoinedAt = time.Now().UTC()

	return s.groupRepo.AddMember(ctx, &participant)
}

func (s *GroupChatService) GetSettings(ctx context.Context, chatID, userID string) (*dto.GroupChatSettingsResponse, error) {
	isMember, err := s.groupRepo.IsUserMember(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("bạn không phải là thành viên của nhóm")
	}

	groupSettings, err := s.groupRepo.GetSettings(ctx, chatID)
	if err != nil {
		return nil, err
	}

	memberSettings, err := s.groupRepo.GetMemberSettings(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}

	return &dto.GroupChatSettingsResponse{
		ChatID:               chatID,
		NotificationsEnabled: memberSettings.NotificationsEnabled,
		AllowMemberAdd:       groupSettings.AllowMemberAdd,
	}, nil
}

func (s *GroupChatService) UpdateSettings(ctx context.Context, chatID, requestID string, input *dto.GroupChatSettingsDTO) (*dto.GroupChatSettingsResponse, error) {
	isMember, err := s.groupRepo.IsUserMember(ctx, chatID, requestID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("bạn không phải là thành viên của nhóm")
	}

	isAdmin, err := s.groupRepo.IsUserAdmin(ctx, chatID, requestID)
	if err != nil {
		return nil, err
	}

	if input.NotificationsEnabled != nil {
		memberSettings, err := s.groupRepo.GetMemberSettings(ctx, chatID, requestID)
		if err != nil {
			return nil, err
		}
		memberSettings.NotificationsEnabled = *input.NotificationsEnabled
		memberSettings.UpdatedAt = time.Now().UTC()

		if err := s.groupRepo.UpsertMemberSettings(ctx, memberSettings); err != nil {
			return nil, err
		}
	}

	if input.Name != nil {
		if !isAdmin {
			return nil, errors.New("chỉ admin mới có quyền đổi tên nhóm")
		}
		if utf8.RuneCountInString(*input.Name) < 3 || utf8.RuneCountInString(*input.Name) > 50 {
			return nil, errors.New("tên nhóm phải từ 3 đến 50 ký tự")
		}
	}

	if input.AvatarURI != nil && !isAdmin {
		return nil, errors.New("chỉ admin mới có quyền đổi avatar nhóm")
	}

	if input.AllowMemberAdd != nil && !isAdmin {
		return nil, errors.New("chỉ admin mới có quyền cấu hình quyền thêm thành viên")
	}

	if input.AllowMemberAdd != nil || input.Name != nil || input.AvatarURI != nil {
		if !isAdmin {
			return nil, errors.New("chỉ admin mới có quyền cập nhật thiết lập toàn nhóm")
		}

		groupSettings, err := s.groupRepo.GetSettings(ctx, chatID)
		if err != nil {
			return nil, err
		}

		if input.AllowMemberAdd != nil {
			groupSettings.AllowMemberAdd = *input.AllowMemberAdd
		}

		if err := s.groupRepo.UpsertSettings(ctx, groupSettings); err != nil {
			return nil, err
		}

		if input.Name != nil || input.AvatarURI != nil {
			chat, err := s.chatRepo.FindChatByID(ctx, chatID)
			if err != nil {
				return nil, err
			}
			if input.Name != nil {
				chat.Name = *input.Name
			}
			if input.AvatarURI != nil {
				chat.AvatarURI = *input.AvatarURI
			}
			if err := s.chatRepo.UpdateChat(ctx, chat); err != nil {
				return nil, err
			}
		}
	}

	// trả về trạng thái hiện tại cho requester
	memberSettings, err := s.groupRepo.GetMemberSettings(ctx, chatID, requestID)
	if err != nil {
		return nil, err
	}
	groupSettings, err := s.groupRepo.GetSettings(ctx, chatID)
	if err != nil {
		return nil, err
	}

	return &dto.GroupChatSettingsResponse{
		ChatID:               chatID,
		NotificationsEnabled: memberSettings.NotificationsEnabled,
		AllowMemberAdd:       groupSettings.AllowMemberAdd,
	}, nil
}

func (s *GroupChatService) TransferAdmin(ctx context.Context, chatID, requestID, targetUserID string) error {
	if requestID == targetUserID {
		return errors.New("không thể chuyển quyền cho chính mình")
	}

	isAdmin, err := s.groupRepo.IsUserAdmin(ctx, chatID, requestID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("chỉ admin mới có thể chuyển quyền quản trị")
	}

	isTargetMember, err := s.groupRepo.IsUserMember(ctx, chatID, targetUserID)
	if err != nil {
		return err
	}
	if !isTargetMember {
		return errors.New("người nhận phải là thành viên của nhóm")
	}

	settings, err := s.groupRepo.GetSettings(ctx, chatID)
	if err != nil {
		return err
	}
	if settings.LastAdminTransferAt != nil {
		nextAllowed := settings.LastAdminTransferAt.AddDate(0, 1, 0) // +1 month
		if time.Now().UTC().Before(nextAllowed) {
			return fmt.Errorf("quyền admin chỉ có thể chuyển 1 lần mỗi tháng; lần chuyển trước: %s", settings.LastAdminTransferAt.UTC().Format(time.RFC3339))
		}
	}

	return s.groupRepo.TransferAdmin(ctx, chatID, requestID, targetUserID, time.Now().UTC())
}
