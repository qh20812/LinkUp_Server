package services

import (
	"context"
	"errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"time"
)

type GroupChatService struct {
	groupRepo *repository.GroupChatRepository
}

func NewGroupChatService(groupRepo *repository.GroupChatRepository) *GroupChatService {
	return &GroupChatService{groupRepo: groupRepo}
}

func (s *GroupChatService) CreateGroup(ctx context.Context, userID string, name, avatarURI string, memberIDs []string) (*models.Chat, error) {
	if len(name) < 3 || len(name) > 50 {
		return nil, errors.New("tên nhóm chat phải từ 3 đến 50 ký tự")
	}

	// 🌟 Sử dụng hàm NewChat có sẵn của bạn với Type là group
	group := models.NewChat(models.ChatTypeGroup, name, avatarURI)
	group.ID = utils.GenerateUUID()
	group.CreatedAt = time.Now().UTC()

	// Khởi tạo Admin (Người tạo nhóm)
	adminPart := models.NewChatParticipant(group.ID, userID, models.ChatRoleAdmin)
	adminPart.ID = utils.GenerateUUID()
	adminPart.JoinedAt = time.Now().UTC()

	participants := []models.ChatParticipant{adminPart}

	// Vòng lặp thêm các thành viên được chọn ban đầu
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

func (s *GroupChatService) AddMember(ctx context.Context, chatID, requesterID, newMemberID string) error {
	isRequesterMember, err := s.groupRepo.IsUserMember(ctx, chatID, requesterID)
	if err != nil {
		return err
	}
	if !isRequesterMember {
		return errors.New("bạn không phải thành viên của nhóm này nên không có quyền mời người khác")
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
