package services

import (
	"context"
	"errors"
	"fmt"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"time"
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
// Hàm BanMember tối ưu lại trong services/group_chat_service.go
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

	// Kiểm tra xem mục tiêu đã bị ban từ trước chưa
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

	// Gọi duy nhất hàm này là đủ (Repo tự lo việc lưu bảng Ban và xóa bảng Participant)
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
