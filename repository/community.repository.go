package repository

import (
	"context"
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
