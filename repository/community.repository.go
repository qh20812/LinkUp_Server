package repository

import (
	"context"
	"errors"
	"fmt"
	"linkup/dto"
	"linkup/models"
	"linkup/utils"
	"sort"
	"time"

	"gorm.io/gorm"
)

type CommunityRepository struct {
	db *gorm.DB
}

func NewCommunityRepository(db *gorm.DB) *CommunityRepository {
	return &CommunityRepository{db: db}
}

// CreateCommunityWithDefaultGroupChat tạo community, group_member, user_roles, chat và participant trong 1 transaction.
func (r *CommunityRepository) CreateCommunityWithDefaultGroupChat(
	ctx context.Context,
	community *models.Community,
	member *models.GroupMember,
	userRoles []models.UserRole,
	chat *models.Chat,
	participants []models.ChatParticipant,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(community).Error; err != nil {
			return fmt.Errorf("lỗi khi lưu thông tin cộng đồng: %w", err)
		}
		if err := tx.Create(member).Error; err != nil {
			return fmt.Errorf("lỗi khi lưu thông tin thành viên: %w", err)
		}
		for i := range userRoles {
			userRoles[i].ID = utils.GenerateUUID()
			userRoles[i].AssignedAt = time.Now().UTC()
			if err := tx.Create(&userRoles[i]).Error; err != nil {
				return fmt.Errorf("lỗi khi gán role cho người tạo: %w", err)
			}
		}
		if err := tx.Create(chat).Error; err != nil {
			return fmt.Errorf("lỗi khi tạo group chat mặc định: %w", err)
		}
		for i := range participants {
			if err := tx.Create(&participants[i]).Error; err != nil {
				return fmt.Errorf("lỗi khi thêm participant vào group chat: %w", err)
			}
		}
		return nil
	})
}

// FindDefaultGroupChatByCommunity tìm group chat mặc định của community.
func (r *CommunityRepository) FindDefaultGroupChatByCommunity(ctx context.Context, communityID string) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.WithContext(ctx).
		Where("community_id = ? AND type = ?", communityID, models.ChatTypeGroup).
		First(&chat).Error
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

// AddCommunityMemberAndGroupChat thêm member vào community và group chat trong 1 transaction.
func (r *CommunityRepository) AddCommunityMemberAndGroupChat(ctx context.Context, communityID, userID, chatID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()

		member := models.NewGroupMember(communityID, userID)
		member.ID = utils.GenerateUUID()
		member.JoinedAt = now
		member.Points = 0
		if err := tx.Create(&member).Error; err != nil {
			return fmt.Errorf("thêm thành viên thất bại: %w", err)
		}

		var groupMemberRole models.Role
		if err := tx.Where("name = ?", models.RoleGroupMember).First(&groupMemberRole).Error; err != nil {
			return fmt.Errorf("không tìm thấy role GROUP_MEMBER: %w", err)
		}

		userRole := models.NewScopedUserRole(userID, groupMemberRole.ID, communityID, models.ScopeTypeCommunity)
		userRole.ID = utils.GenerateUUID()
		userRole.AssignedAt = now
		if err := tx.Create(&userRole).Error; err != nil {
			return fmt.Errorf("gán role thành viên thất bại: %w", err)
		}

		participant := models.NewChatParticipant(chatID, userID, models.ChatRoleMember)
		participant.ID = utils.GenerateUUID()
		participant.JoinedAt = now
		if err := tx.Create(&participant).Error; err != nil {
			return fmt.Errorf("thêm participant vào group chat thất bại: %w", err)
		}

		return nil
	})
}

// CreateWithRoles tạo community, group_member, và user_roles trong 1 transaction.
func (r *CommunityRepository) CreateWithRoles(ctx context.Context, community *models.Community, member *models.GroupMember, userRoles []models.UserRole) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(community).Error; err != nil {
			return fmt.Errorf("lỗi khi lưu thông tin cộng đồng: %w", err)
		}
		if err := tx.Create(member).Error; err != nil {
			return fmt.Errorf("lỗi khi lưu thông tin thành viên: %w", err)
		}
		for i := range userRoles {
			userRoles[i].ID = utils.GenerateUUID()
			userRoles[i].AssignedAt = time.Now().UTC()
			if err := tx.Create(&userRoles[i]).Error; err != nil {
				return fmt.Errorf("lỗi khi gán role cho người tạo: %w", err)
			}
		}
		return nil
	})
}

// AssignUserRole gán một role cho user trong user_roles (scoped role).
func (r *CommunityRepository) AssignUserRole(ctx context.Context, userID string, roleName models.RoleName, scopeID string, scopeType models.ScopeType) error {
	var role models.Role
	if err := r.db.WithContext(ctx).Where("name = ?", roleName).First(&role).Error; err != nil {
		return fmt.Errorf("find role %s: %w", roleName, err)
	}

	userRole := models.NewScopedUserRole(userID, role.ID, scopeID, scopeType)
	userRole.ID = utils.GenerateUUID()
	userRole.AssignedAt = time.Now().UTC()

	if err := r.db.WithContext(ctx).Create(&userRole).Error; err != nil {
		return fmt.Errorf("assign role %s: %w", roleName, err)
	}
	return nil
}

// FindRoleByName tìm role theo tên, trả lỗi nếu không tìm thấy.
func (r *CommunityRepository) FindRoleByName(ctx context.Context, roleName models.RoleName, out *models.Role) error {
	return r.db.WithContext(ctx).Where("name = ?", roleName).First(out).Error
}

func (r *CommunityRepository) FindByID(ctx context.Context, id string) (*models.Community, error) {
	var community models.Community
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&community).Error
	if err != nil {
		return nil, err
	}
	return &community, nil
}

func (r *CommunityRepository) FindByName(ctx context.Context, name string) (*models.Community, error) {
	var community models.Community
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&community).Error
	if err != nil {
		return nil, err
	}
	return &community, nil
}

func (r *CommunityRepository) IsNameTaken(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Community{}).
		Where("name = ?", name).
		Count(&count).Error
	return count > 0, err
}

// UpdateBackground cập nhật background_uri của community.
func (r *CommunityRepository) UpdateBackground(ctx context.Context, communityID, backgroundURI string) error {
	return r.db.WithContext(ctx).
		Model(&models.Community{}).
		Where("id = ?", communityID).
		Update("background_uri", backgroundURI).Error
}

// IsUserAdmin kiểm tra user có phải admin của community không (dựa trên user_roles).
func (r *CommunityRepository) IsUserAdmin(ctx context.Context, communityID, userID string) (bool, error) {
	var count int64
	scopeType := models.ScopeTypeCommunity
	err := r.db.WithContext(ctx).
		Table("user_roles").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND user_roles.scope_id = ? AND user_roles.scope_type = ?",
			userID, communityID, scopeType).
		Where("roles.name IN (?, ?)", models.RoleGroupAdmin, models.RoleCommunityAdmin).
		Count(&count).Error
	return count > 0, err
}

// IsUserMember kiểm tra user có phải member của community không (dựa trên user_roles).
func (r *CommunityRepository) IsUserMember(ctx context.Context, communityID, userID string) (bool, error) {
	var count int64
	scopeType := models.ScopeTypeCommunity
	err := r.db.WithContext(ctx).
		Table("user_roles").
		Where("user_roles.user_id = ? AND user_roles.scope_id = ? AND user_roles.scope_type = ?",
			userID, communityID, scopeType).
		Count(&count).Error
	return count > 0, err
}

// GetUserRole lấy role của user trong community từ user_roles.
func (r *CommunityRepository) GetUserRole(ctx context.Context, communityID, userID string) (models.GroupRole, error) {
	type userRoleWithName struct {
		models.UserRole
		RoleName models.RoleName `gorm:"column:name"`
	}
	var result userRoleWithName
	scopeType := models.ScopeTypeCommunity
	err := r.db.WithContext(ctx).
		Table("user_roles").
		Select("user_roles.*, roles.name").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND user_roles.scope_id = ? AND user_roles.scope_type = ?",
			userID, communityID, scopeType).
		Where("roles.name IN (?, ?, ?, ?)",
			models.RoleGroupAdmin, models.RoleGroupMod, models.RoleGroupMember, models.RoleCommunityAdmin).
		Order("CASE roles.name WHEN 'COMMUNITY_ADMIN' THEN 1 WHEN 'GROUP_ADMIN' THEN 2 WHEN 'GROUP_MOD' THEN 3 ELSE 4 END").
		First(&result).Error
	if err != nil {
		return "", err
	}
	return mapRoleNameToGroupRole(result.RoleName), nil
}

// CreateJoinRequest tạo một yêu cầu tham gia community.
func (r *CommunityRepository) CreateJoinRequest(ctx context.Context, req *models.CommunityJoinRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

// DeleteNonPendingJoinRequests xóa các join request cũ (rejected/approved) để cho phép gửi lại.
func (r *CommunityRepository) DeleteNonPendingJoinRequests(ctx context.Context, communityID, userID string) error {
	return r.db.WithContext(ctx).
		Where("community_id = ? AND user_id = ? AND status != ?", communityID, userID, models.JoinRequestStatusPending).
		Delete(&models.CommunityJoinRequest{}).Error
}

// FindPendingJoinRequestByUserAndCommunity tìm yêu cầu pending của user trong community.
func (r *CommunityRepository) FindPendingJoinRequestByUserAndCommunity(ctx context.Context, communityID, userID string) (*models.CommunityJoinRequest, error) {
	var req models.CommunityJoinRequest
	err := r.db.WithContext(ctx).
		Where("community_id = ? AND user_id = ? AND status = ?", communityID, userID, models.JoinRequestStatusPending).
		First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// FindJoinRequestByID tìm join request theo ID.
func (r *CommunityRepository) FindJoinRequestByID(ctx context.Context, requestID string) (*models.CommunityJoinRequest, error) {
	var req models.CommunityJoinRequest
	err := r.db.WithContext(ctx).Where("id = ?", requestID).First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// FindPendingJoinRequestsByCommunity lấy danh sách yêu cầu pending của community.
func (r *CommunityRepository) FindPendingJoinRequestsByCommunity(ctx context.Context, communityID string) ([]models.CommunityJoinRequest, error) {
	var requests []models.CommunityJoinRequest
	err := r.db.WithContext(ctx).
		Where("community_id = ? AND status = ?", communityID, models.JoinRequestStatusPending).
		Order("created_at ASC").
		Find(&requests).Error
	return requests, err
}

// ApproveJoinRequest transaction: cập nhật status = approved + tạo group_member + gán GROUP_MEMBER role.
// Nếu chatID != nil, thêm user vào chat_participants của group chat.
func (r *CommunityRepository) ApproveJoinRequest(ctx context.Context, requestID string, chatID *string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var req models.CommunityJoinRequest
		if err := tx.Where("id = ? AND status = ?", requestID, models.JoinRequestStatusPending).First(&req).Error; err != nil {
			return fmt.Errorf("yêu cầu tham gia không tồn tại hoặc đã được xử lý: %w", err)
		}

		now := time.Now().UTC()
		req.Status = models.JoinRequestStatusApproved
		req.RespondedAt = &now
		if err := tx.Save(&req).Error; err != nil {
			return fmt.Errorf("cập nhật trạng thái yêu cầu thất bại: %w", err)
		}

		member := models.NewGroupMember(req.CommunityID, req.UserID)
		member.ID = utils.GenerateUUID()
		member.JoinedAt = now
		member.Points = 0
		if err := tx.Create(&member).Error; err != nil {
			return fmt.Errorf("thêm thành viên thất bại: %w", err)
		}

		var groupMemberRole models.Role
		if err := tx.Where("name = ?", models.RoleGroupMember).First(&groupMemberRole).Error; err != nil {
			return fmt.Errorf("không tìm thấy role GROUP_MEMBER: %w", err)
		}

		userRole := models.NewScopedUserRole(req.UserID, groupMemberRole.ID, req.CommunityID, models.ScopeTypeCommunity)
		userRole.ID = utils.GenerateUUID()
		userRole.AssignedAt = now
		if err := tx.Create(&userRole).Error; err != nil {
			return fmt.Errorf("gán role thành viên thất bại: %w", err)
		}

		if chatID != nil {
			participant := models.NewChatParticipant(*chatID, req.UserID, models.ChatRoleMember)
			participant.ID = utils.GenerateUUID()
			participant.JoinedAt = now
			if err := tx.Create(&participant).Error; err != nil {
				return fmt.Errorf("thêm participant vào group chat thất bại: %w", err)
			}
		}

		return nil
	})
}

// RejectJoinRequest cập nhật status = rejected.
func (r *CommunityRepository) RejectJoinRequest(ctx context.Context, requestID string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.CommunityJoinRequest{}).
		Where("id = ? AND status = ?", requestID, models.JoinRequestStatusPending).
		Updates(map[string]interface{}{
			"status":      models.JoinRequestStatusRejected,
			"responded_at": now,
		}).Error
}

// UpdateUserRole cập nhật role của user trong community (update in-place).
func (r *CommunityRepository) UpdateUserRole(ctx context.Context, communityID, userID string, newRoleName models.RoleName) error {
	var newRole models.Role
	if err := r.db.WithContext(ctx).Where("name = ?", newRoleName).First(&newRole).Error; err != nil {
		return fmt.Errorf("không tìm thấy role %s: %w", newRoleName, err)
	}

	scopeType := models.ScopeTypeCommunity
	result := r.db.WithContext(ctx).
		Model(&models.UserRole{}).
		Where("user_id = ? AND scope_id = ? AND scope_type = ?", userID, communityID, scopeType).
		Update("role_id", newRole.ID)
	if result.Error != nil {
		return fmt.Errorf("cập nhật role thất bại: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("không tìm thấy role cần cập nhật")
	}
	return nil
}

// FindMembersByCommunity lấy danh sách thành viên của community kèm profile và role.
func (r *CommunityRepository) FindMembersByCommunity(ctx context.Context, communityID string) ([]dto.CommunityMemberItem, error) {
	var groupMembers []models.GroupMember
	if err := r.db.WithContext(ctx).
		Where("community_id = ?", communityID).
		Find(&groupMembers).Error; err != nil {
		return nil, fmt.Errorf("lấy danh sách thành viên thất bại: %w", err)
	}
	if len(groupMembers) == 0 {
		return []dto.CommunityMemberItem{}, nil
	}

	type userRoleRow struct {
		UserID   string `gorm:"column:user_id"`
		RoleName string `gorm:"column:role_name"`
	}
	var roleRows []userRoleRow
	scopeType := models.ScopeTypeCommunity
	if err := r.db.WithContext(ctx).
		Table("user_roles").
		Select("user_roles.user_id, roles.name AS role_name").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.scope_id = ? AND user_roles.scope_type = ?", communityID, scopeType).
		Find(&roleRows).Error; err != nil {
		return nil, fmt.Errorf("lấy vai trò thành viên thất bại: %w", err)
	}

	rolePriority := map[string]int{
		"COMMUNITY_ADMIN": 1,
		"GROUP_ADMIN":     2,
		"GROUP_MOD":       3,
		"GROUP_MEMBER":    4,
	}
	bestRole := make(map[string]string)
	for _, row := range roleRows {
		if current, exists := bestRole[row.UserID]; !exists || rolePriority[row.RoleName] < rolePriority[current] {
			bestRole[row.UserID] = row.RoleName
		}
	}

	var profiles []models.Profile
	userIDs := make([]string, 0, len(groupMembers))
	for _, gm := range groupMembers {
		userIDs = append(userIDs, gm.UserID)
	}
	if err := r.db.WithContext(ctx).
		Where("user_id IN ?", userIDs).
		Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("lấy hồ sơ thành viên thất bại: %w", err)
	}
	profileMap := make(map[string]*models.Profile, len(profiles))
	for i := range profiles {
		profileMap[profiles[i].UserID] = &profiles[i]
	}

	members := make([]dto.CommunityMemberItem, 0, len(groupMembers))
	for _, gm := range groupMembers {
		item := dto.CommunityMemberItem{
			UserID:   gm.UserID,
			Role:     bestRole[gm.UserID],
			JoinedAt: gm.JoinedAt,
		}
		if p, ok := profileMap[gm.UserID]; ok {
			item.DisplayName = p.DisplayName
			item.AvatarURI = p.AvatarURI
		}
		members = append(members, item)
	}

	sort.Slice(members, func(i, j int) bool {
		return rolePriority[members[i].Role] < rolePriority[members[j].Role]
	})

	return members, nil
}

// IsUserCreator kiểm tra user có phải người tạo community không.
func (r *CommunityRepository) IsUserCreator(ctx context.Context, communityID, userID string) (bool, error) {
	var community models.Community
	err := r.db.WithContext(ctx).Where("id = ?", communityID).First(&community).Error
	if err != nil {
		return false, err
	}
	return community.CreatorID == userID, nil
}

// RemoveMember xóa UserRole và GroupMember của user trong community (transaction).
func (r *CommunityRepository) RemoveMember(ctx context.Context, communityID, userID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		scopeType := models.ScopeTypeCommunity
		if err := tx.Where("user_id = ? AND scope_id = ? AND scope_type = ?", userID, communityID, scopeType).
			Delete(&models.UserRole{}).Error; err != nil {
			return fmt.Errorf("xóa vai trò thất bại: %w", err)
		}
		if err := tx.Where("community_id = ? AND user_id = ?", communityID, userID).
			Delete(&models.GroupMember{}).Error; err != nil {
			return fmt.Errorf("xóa thành viên thất bại: %w", err)
		}
		return nil
	})
}

// FindCommunityAdmins lấy danh sách user IDs có quyền admin (COMMUNITY_ADMIN hoặc GROUP_ADMIN) trong community.
func (r *CommunityRepository) FindCommunityAdmins(ctx context.Context, communityID string) ([]string, error) {
	var userIDs []string
	scopeType := models.ScopeTypeCommunity
	err := r.db.WithContext(ctx).
		Table("user_roles").
		Select("DISTINCT user_roles.user_id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.scope_id = ? AND user_roles.scope_type = ?", communityID, scopeType).
		Where("roles.name IN (?, ?)", models.RoleCommunityAdmin, models.RoleGroupAdmin).
		Pluck("user_roles.user_id", &userIDs).Error
	if err != nil {
		return nil, fmt.Errorf("lấy danh sách admin thất bại: %w", err)
	}
	return userIDs, nil
}

// FindCommunityMemberIDs lấy danh sách tất cả user IDs là thành viên của community.
func (r *CommunityRepository) FindCommunityMemberIDs(ctx context.Context, communityID string) ([]string, error) {
	var userIDs []string
	err := r.db.WithContext(ctx).
		Table("group_members").
		Select("user_id").
		Where("community_id = ?", communityID).
		Pluck("user_id", &userIDs).Error
	if err != nil {
		return nil, fmt.Errorf("lấy danh sách thành viên thất bại: %w", err)
	}
	return userIDs, nil
}

// FindByCreator lấy danh sách cộng đồng do một người dùng tạo.
func (r *CommunityRepository) FindByCreator(ctx context.Context, creatorID string) ([]models.Community, error) {
	var communities []models.Community
	err := r.db.WithContext(ctx).Where("creator_id = ?", creatorID).Find(&communities).Error
	if err != nil {
		return nil, fmt.Errorf("tìm cộng đồng theo người tạo thất bại: %w", err)
	}
	return communities, nil
}

// FindOldestMember tìm thành viên tham gia sớm nhất (trừ userID chỉ định).
func (r *CommunityRepository) FindOldestMember(ctx context.Context, communityID, excludeUserID string) (*models.GroupMember, error) {
	var member models.GroupMember
	err := r.db.WithContext(ctx).
		Where("community_id = ? AND user_id <> ?", communityID, excludeUserID).
		Order("joined_at ASC").
		First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("tìm thành viên cũ nhất thất bại: %w", err)
	}
	return &member, nil
}

// FindHighestContributionMember tìm thành viên có điểm đóng góp cao nhất (trừ userID chỉ định).
func (r *CommunityRepository) FindHighestContributionMember(ctx context.Context, communityID, excludeUserID string) (*models.MemberContribution, error) {
	var mc models.MemberContribution
	err := r.db.WithContext(ctx).
		Where("community_id = ? AND user_id <> ?", communityID, excludeUserID).
		Order("contribution_score DESC").
		First(&mc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("tìm thành viên có điểm đóng góp cao nhất thất bại: %w", err)
	}
	return &mc, nil
}

// ── Invite code ─────────────────────────────────────────────────────────────

func (r *CommunityRepository) CreateInviteCode(ctx context.Context, code *models.CommunityInviteCode) error {
	return r.db.WithContext(ctx).Create(code).Error
}

func (r *CommunityRepository) FindInviteCodeByCode(ctx context.Context, code string) (*models.CommunityInviteCode, error) {
	var inviteCode models.CommunityInviteCode
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&inviteCode).Error
	if err != nil {
		return nil, err
	}
	return &inviteCode, nil
}

func (r *CommunityRepository) FindInviteCodeByID(ctx context.Context, id string) (*models.CommunityInviteCode, error) {
	var inviteCode models.CommunityInviteCode
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&inviteCode).Error
	if err != nil {
		return nil, err
	}
	return &inviteCode, nil
}

func (r *CommunityRepository) ListInviteCodesByCommunity(ctx context.Context, communityID string) ([]models.CommunityInviteCode, error) {
	var codes []models.CommunityInviteCode
	err := r.db.WithContext(ctx).
		Where("community_id = ?", communityID).
		Order("created_at DESC").
		Find(&codes).Error
	return codes, err
}

func (r *CommunityRepository) DeactivateInviteCode(ctx context.Context, codeID string) error {
	return r.db.WithContext(ctx).
		Model(&models.CommunityInviteCode{}).
		Where("id = ?", codeID).
		Update("is_active", false).Error
}

func (r *CommunityRepository) IncrementInviteCodeUsedCount(ctx context.Context, tx *gorm.DB, codeID string) error {
	db := r.db.WithContext(ctx)
	if tx != nil {
		db = tx
	}
	return db.
		Model(&models.CommunityInviteCode{}).
		Where("id = ?", codeID).
		UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error
}

// ── Direct invitation ───────────────────────────────────────────────────────

func (r *CommunityRepository) CreateInvitation(ctx context.Context, inv *models.CommunityInvitation) error {
	return r.db.WithContext(ctx).Create(inv).Error
}

func (r *CommunityRepository) FindInvitationByID(ctx context.Context, id string) (*models.CommunityInvitation, error) {
	var inv models.CommunityInvitation
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *CommunityRepository) FindPendingInvitation(ctx context.Context, communityID, inviteeID string) (*models.CommunityInvitation, error) {
	var inv models.CommunityInvitation
	err := r.db.WithContext(ctx).
		Where("community_id = ? AND invitee_id = ? AND status = ?", communityID, inviteeID, models.InvitationStatusPending).
		First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *CommunityRepository) UpdateInvitationStatus(ctx context.Context, tx *gorm.DB, id string, status models.InvitationStatus) error {
	now := time.Now().UTC()
	db := r.db.WithContext(ctx)
	if tx != nil {
		db = tx
	}
	return db.
		Model(&models.CommunityInvitation{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       status,
			"responded_at": now,
		}).Error
}

type invitationWithCommunity struct {
	models.CommunityInvitation
	CommunityName string `gorm:"column:name"`
}

func (r *CommunityRepository) ListPendingInvitationsByInvitee(ctx context.Context, inviteeID string) ([]invitationWithCommunity, error) {
	var invites []invitationWithCommunity
	err := r.db.WithContext(ctx).
		Table("community_invitations").
		Select("community_invitations.*, communities.name").
		Joins("JOIN communities ON communities.id = community_invitations.community_id").
		Where("community_invitations.invitee_id = ? AND community_invitations.status = ?", inviteeID, models.InvitationStatusPending).
		Order("community_invitations.created_at DESC").
		Find(&invites).Error
	return invites, err
}

// TransferCommunityOwnership chuyển quyền sở hữu cộng đồng.
// Cập nhật creator_id, gán/quản lý user_roles và cập nhật participant role trong group chat mặc định.
func (r *CommunityRepository) TransferCommunityOwnership(ctx context.Context, communityID, oldCreatorID, newCreatorID string, keepAdmin bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()

		if err := tx.Model(&models.Community{}).Where("id = ?", communityID).Update("creator_id", newCreatorID).Error; err != nil {
			return fmt.Errorf("cập nhật creator_id thất bại: %w", err)
		}

		var communityAdminRole, groupAdminRole, groupMemberRole models.Role
		if err := tx.Where("name = ?", models.RoleCommunityAdmin).First(&communityAdminRole).Error; err != nil {
			return fmt.Errorf("không tìm thấy role COMMUNITY_ADMIN: %w", err)
		}
		if err := tx.Where("name = ?", models.RoleGroupAdmin).First(&groupAdminRole).Error; err != nil {
			return fmt.Errorf("không tìm thấy role GROUP_ADMIN: %w", err)
		}
		if err := tx.Where("name = ?", models.RoleGroupMember).First(&groupMemberRole).Error; err != nil {
			return fmt.Errorf("không tìm thấy role GROUP_MEMBER: %w", err)
		}

		scopeType := models.ScopeTypeCommunity
		adminRoleIDs := []string{communityAdminRole.ID, groupAdminRole.ID}

		if keepAdmin {
			if err := tx.Where("user_id = ? AND scope_id = ? AND scope_type = ? AND role_id IN ?",
				oldCreatorID, communityID, scopeType, adminRoleIDs).
				Delete(&models.UserRole{}).Error; err != nil {
				return fmt.Errorf("xóa admin roles cũ thất bại: %w", err)
			}
		} else {
			if err := tx.Where("user_id = ? AND scope_id = ? AND scope_type = ? AND role_id IN ?",
				oldCreatorID, communityID, scopeType, adminRoleIDs).
				Delete(&models.UserRole{}).Error; err != nil {
				return fmt.Errorf("xóa admin roles cũ thất bại: %w", err)
			}
			var count int64
			tx.Model(&models.UserRole{}).Where("user_id = ? AND scope_id = ? AND scope_type = ? AND role_id = ?",
				oldCreatorID, communityID, scopeType, groupMemberRole.ID).Count(&count)
			if count == 0 {
				memberRole := models.NewScopedUserRole(oldCreatorID, groupMemberRole.ID, communityID, scopeType)
				memberRole.ID = utils.GenerateUUID()
				memberRole.AssignedAt = now
				if err := tx.Create(&memberRole).Error; err != nil {
					return fmt.Errorf("gán member role cho chủ cũ thất bại: %w", err)
				}
			}
		}

		if err := tx.Where("user_id = ? AND scope_id = ? AND scope_type = ? AND role_id = ?",
			newCreatorID, communityID, scopeType, groupMemberRole.ID).
			Delete(&models.UserRole{}).Error; err != nil {
			return fmt.Errorf("xóa member role của chủ mới thất bại: %w", err)
		}

		for _, roleID := range adminRoleIDs {
			userRole := models.NewScopedUserRole(newCreatorID, roleID, communityID, scopeType)
			userRole.ID = utils.GenerateUUID()
			userRole.AssignedAt = now
			if err := tx.Create(&userRole).Error; err != nil {
				return fmt.Errorf("gán admin role cho chủ mới thất bại: %w", err)
			}
		}

		var chat models.Chat
		err := tx.Where("community_id = ? AND type = ?", communityID, models.ChatTypeGroup).First(&chat).Error
		if err == nil {
			if err := tx.Model(&models.ChatParticipant{}).
				Where("chat_id = ? AND user_id = ?", chat.ID, newCreatorID).
				Update("role", models.ChatRoleAdmin).Error; err != nil {
				return fmt.Errorf("cập nhật participant role cho chủ mới thất bại: %w", err)
			}
		}

		return nil
	})
}

// ── Admin: Community management ─────────────────────────────────────────────

func (r *CommunityRepository) ListCommunitiesAdmin(ctx context.Context, keyword, status, privacy string, pageSize, offset int) ([]dto.AdminCommunityListItem, error) {
	query := r.db.WithContext(ctx).
		Table("communities").
		Select(`communities.id, communities.name, communities.creator_id, communities.privacy,
			communities.status, communities.created_at,
			COALESCE((SELECT COUNT(*) FROM group_members WHERE community_id = communities.id), 0) AS member_count,
			COALESCE((SELECT display_name FROM profiles WHERE user_id = communities.creator_id), '') AS creator_name`)

	if keyword != "" {
		query = query.Where("communities.name LIKE ?", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("communities.status = ?", status)
	}
	if privacy != "" {
		query = query.Where("communities.privacy = ?", privacy)
	}

	type communityRow struct {
		ID          string    `gorm:"column:id"`
		Name        string    `gorm:"column:name"`
		CreatorID   string    `gorm:"column:creator_id"`
		CreatorName string    `gorm:"column:creator_name"`
		MemberCount int       `gorm:"column:member_count"`
		Privacy     string    `gorm:"column:privacy"`
		Status      string    `gorm:"column:status"`
		CreatedAt   time.Time `gorm:"column:created_at"`
	}
	var rows []communityRow
	if err := query.Order("communities.created_at DESC").Offset(offset).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list communities admin: %w", err)
	}

	items := make([]dto.AdminCommunityListItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, dto.AdminCommunityListItem{
			ID:          r.ID,
			Name:        r.Name,
			CreatorID:   r.CreatorID,
			CreatorName: r.CreatorName,
			MemberCount: r.MemberCount,
			Privacy:     r.Privacy,
			Status:      r.Status,
			CreatedAt:   r.CreatedAt,
		})
	}
	return items, nil
}

func (r *CommunityRepository) CountCommunitiesAdmin(ctx context.Context, keyword, status, privacy string) (int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Community{})
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if privacy != "" {
		query = query.Where("privacy = ?", privacy)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count communities admin: %w", err)
	}
	return total, nil
}

func (r *CommunityRepository) UpdateStatus(ctx context.Context, communityID string, status models.CommunityStatus) error {
	return r.db.WithContext(ctx).Model(&models.Community{}).Where("id = ?", communityID).Update("status", status).Error
}

func (r *CommunityRepository) FindCommunityMembersWithProfiles(ctx context.Context, communityID string) ([]dto.AdminCommunityMember, error) {
	type memberRow struct {
		UserID      string    `gorm:"column:user_id"`
		DisplayName string    `gorm:"column:display_name"`
		AvatarURI   string    `gorm:"column:avatar_uri"`
		Role        string    `gorm:"column:role_name"`
		JoinedAt    time.Time `gorm:"column:joined_at"`
	}
	var rows []memberRow
	err := r.db.WithContext(ctx).
		Table("group_members").
		Select(`DISTINCT group_members.user_id, group_members.joined_at,
			COALESCE(profiles.display_name, '') AS display_name,
			COALESCE(profiles.avatar_uri, '') AS avatar_uri,
			COALESCE((SELECT roles.name FROM user_roles
				JOIN roles ON roles.id = user_roles.role_id
				WHERE user_roles.user_id = group_members.user_id
				AND user_roles.scope_id = ? AND user_roles.scope_type = ?
				ORDER BY CASE roles.name
					WHEN 'COMMUNITY_ADMIN' THEN 1
					WHEN 'GROUP_ADMIN' THEN 2
					WHEN 'GROUP_MOD' THEN 3
					ELSE 4 END
				LIMIT 1), 'GROUP_MEMBER') AS role_name`, communityID, models.ScopeTypeCommunity).
		Joins("LEFT JOIN profiles ON profiles.user_id = group_members.user_id").
		Where("group_members.community_id = ?", communityID).
		Order("group_members.joined_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("find community members with profiles: %w", err)
	}

	items := make([]dto.AdminCommunityMember, 0, len(rows))
	for _, r := range rows {
		items = append(items, dto.AdminCommunityMember{
			UserID:      r.UserID,
			DisplayName: r.DisplayName,
			AvatarURI:   r.AvatarURI,
			Role:        r.Role,
		})
	}
	return items, nil
}

// mapRoleNameToGroupRole ánh xạ role name từ roles table sang GroupRole enum.
func mapRoleNameToGroupRole(name models.RoleName) models.GroupRole {
	switch name {
	case models.RoleCommunityAdmin, models.RoleGroupAdmin:
		return models.GroupRoleAdmin
	case models.RoleGroupMod:
		return models.GroupRoleMod
	default:
		return models.GroupRoleMember
	}
}
