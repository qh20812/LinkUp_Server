package services

import (
	"context"
	"errors"
	"linkup/dto"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ContributionService struct {
	contributionRepo    *repository.ContributionRepository
	communityRepo       *repository.CommunityRepository
	profileRepo         *repository.ProfileRepository
	notificationService *NotificationService
	validation          *validations.ContributionValidation
	groupRole           *utils.GroupRoleChecker
}

func NewContributionService(contributionRepo *repository.ContributionRepository, communityRepo *repository.CommunityRepository, profileRepo *repository.ProfileRepository, notificationService *NotificationService, validation *validations.ContributionValidation) *ContributionService {
	return &ContributionService{
		contributionRepo:    contributionRepo,
		communityRepo:       communityRepo,
		profileRepo:         profileRepo,
		notificationService: notificationService,
		validation:          validation,
		groupRole:           utils.NewGroupRoleChecker(communityRepo.GetUserRole),
	}
}

func (s *ContributionService) GetPolicy(ctx context.Context, communityID, userID string) (*models.CommunityPolicy, error) {
	if _, err := s.communityRepo.FindByID(ctx, communityID); err != nil {
		return nil, validations.ErrCommunityNotFound
	}

	if err := s.groupRole.RequireRole(ctx, communityID, userID, models.GroupRoleMember); err != nil {
		return nil, validations.ErrNotCommunityMember
	}

	policy, err := s.contributionRepo.GetPolicy(ctx, communityID)
	if err == nil {
		return policy, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	policy = s.defaultPolicy(communityID)
	if err := s.contributionRepo.UpsertPolicy(ctx, policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func (s *ContributionService) RequireMember(ctx context.Context, communityID, userID string) error {
	if _, err := s.communityRepo.FindByID(ctx, communityID); err != nil {
		return validations.ErrCommunityNotFound
	}
	if err := s.groupRole.RequireRole(ctx, communityID, userID, models.GroupRoleMember); err != nil {
		return validations.ErrNotCommunityMember
	}
	return nil
}

func (s *ContributionService) UpdatePolicy(ctx context.Context, adminID, communityID string, input dto.UpdatePolicyInput) error {
	if _, err := s.communityRepo.FindByID(ctx, communityID); err != nil {
		return validations.ErrCommunityNotFound
	}
	if err := s.groupRole.RequireRole(ctx, communityID, adminID, models.GroupRoleAdmin); err != nil {
		return validations.ErrNotCommunityAdmin
	}
	if err := s.validation.ValidateUpdatePolicyInput(input); err != nil {
		return err
	}

	now := time.Now().UTC()
	policy := &models.CommunityPolicy{
		CommunityID:                 communityID,
		PostWeight:                  input.PostWeight,
		CommentWeight:               input.CommentWeight,
		ReactionWeight:              input.ReactionWeight,
		EventWeight:                 input.EventWeight,
		TopContributorThreshold:     input.TopContributorThreshold,
		ModeratorPromotionThreshold: input.ModeratorPromotionThreshold,
		AutoPromoteEnabled:          input.AutoPromoteEnabled,
		BadgeEnabled:                input.BadgeEnabled,
		UpdatedAt:                   &now,
	}

	return s.contributionRepo.UpsertPolicy(ctx, policy)
}

func (s *ContributionService) RecalculateScore(ctx context.Context, communityID, userID string) (*models.MemberContribution, error) {
	if _, err := s.communityRepo.FindByID(ctx, communityID); err != nil {
		return nil, validations.ErrCommunityNotFound
	}

	policy, err := s.ensurePolicy(ctx, communityID)
	if err != nil {
		return nil, err
	}

	contribution, err := s.ensureContribution(ctx, communityID, userID)
	if err != nil {
		return nil, err
	}

	return s.recalculateAndPersistContribution(ctx, contribution, policy)
}

func (s *ContributionService) GetLeaderboard(ctx context.Context, communityID string, page, pageSize int) ([]dto.LeaderboardItem, error) {
	if _, err := s.communityRepo.FindByID(ctx, communityID); err != nil {
		return nil, validations.ErrCommunityNotFound
	}

	page, pageSize = s.validation.NormalizePagination(page, pageSize)
	offset := (page - 1) * pageSize
	items, err := s.contributionRepo.GetLeaderboard(ctx, communityID, offset, pageSize)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *ContributionService) GetCommunityMembers(ctx context.Context, communityID string, page, pageSize int) ([]dto.CommunityMemberItem, error) {
	if _, err := s.communityRepo.FindByID(ctx, communityID); err != nil {
		return nil, validations.ErrCommunityNotFound
	}

	page, pageSize = s.validation.NormalizePagination(page, pageSize)
	offset := (page - 1) * pageSize
	return s.contributionRepo.GetCommunityMembers(ctx, communityID, offset, pageSize)
}

func (s *ContributionService) GetContribution(ctx context.Context, communityID, userID string) (*models.MemberContribution, error) {
	if _, err := s.communityRepo.FindByID(ctx, communityID); err != nil {
		return nil, validations.ErrCommunityNotFound
	}

	return s.ensureContribution(ctx, communityID, userID)
}

func (s *ContributionService) GetContributionResponse(ctx context.Context, communityID, userID string) (*dto.ContributionResponse, error) {
	contribution, err := s.GetContribution(ctx, communityID, userID)
	if err != nil {
		return nil, err
	}

	policy, err := s.ensurePolicy(ctx, communityID)
	if err != nil {
		return nil, err
	}

	updated, err := s.recalculateAndPersistContribution(ctx, contribution, policy)
	if err != nil {
		return nil, err
	}

	displayName := ""
	avatarURI := ""
	if s.profileRepo != nil {
		if profile, err := s.profileRepo.FindByUserID(ctx, userID); err == nil && profile != nil {
			displayName = profile.DisplayName
			avatarURI = profile.AvatarURI
		}
	}

	return &dto.ContributionResponse{
		UserID:              userID,
		DisplayName:         displayName,
		AvatarURI:           avatarURI,
		ValidPosts:          updated.ValidPosts,
		QualityComments:     updated.QualityComments,
		PositiveReactions:   updated.PositiveReactions,
		EventParticipations: updated.EventParticipations,
		ContributionScore:   updated.ContributionScore,
		BadgeType:           updated.BadgeType,
		PromotedToMod:       updated.PromotedToMod,
	}, nil
}

// Trạng thái hiển thị của challenge, tính từ status trong DB + mốc thời gian.
const (
	ChallengeStateUpcoming  = "upcoming"
	ChallengeStateActive    = "active"
	ChallengeStateEnded     = "ended"
	ChallengeStateCancelled = "cancelled"
)

// MaxChallengePostsPerDay là trần số bài được tính điểm mỗi ngày cho mỗi người trong
// một challenge — chặn farm điểm bằng cách spam hashtag.
const MaxChallengePostsPerDay = 3

// challengeAwardDate trả mốc ngày (UTC) dùng cho trần điểm theo ngày.
func challengeAwardDate(now time.Time) time.Time {
	return now.UTC().Truncate(24 * time.Hour)
}

// challengeState suy ra trạng thái thực tế: status='active' nhưng đã qua end_date
// vẫn phải hiển thị là đã kết thúc (DB không tự cập nhật status).
func challengeState(challenge *models.CommunityChallenge, now time.Time) string {
	switch challenge.Status {
	case models.ChallengeStatusCancelled:
		return ChallengeStateCancelled
	case models.ChallengeStatusCompleted:
		return ChallengeStateEnded
	}
	if now.Before(challenge.StartDate) {
		return ChallengeStateUpcoming
	}
	if now.After(challenge.EndDate) {
		return ChallengeStateEnded
	}
	return ChallengeStateActive
}

// challengeStatePriority dùng để sắp xếp: đang diễn ra → sắp diễn ra → đã kết thúc → đã huỷ.
func challengeStatePriority(state string) int {
	switch state {
	case ChallengeStateActive:
		return 0
	case ChallengeStateUpcoming:
		return 1
	case ChallengeStateEnded:
		return 2
	default:
		return 3
	}
}

func (s *ContributionService) buildChallengeResponse(ctx context.Context, challenge *models.CommunityChallenge, participants int, joined, canManage bool, dailyAwarded int) dto.ChallengeResponse {
	creatorName := ""
	if s.profileRepo != nil {
		if profile, err := s.profileRepo.FindByUserID(ctx, challenge.CreatorID); err == nil && profile != nil {
			creatorName = profile.DisplayName
		}
	}
	return dto.ChallengeResponse{
		ID:                challenge.ID,
		CommunityID:       challenge.CommunityID,
		CreatorID:         challenge.CreatorID,
		CreatorName:       creatorName,
		Title:             challenge.Title,
		Description:       challenge.Description,
		Hashtag:           challenge.Hashtag,
		PointsPerPost:     challenge.PointsPerPost,
		StartDate:         challenge.StartDate,
		EndDate:           challenge.EndDate,
		MaxParticipants:   challenge.MaxParticipants,
		Status:            string(challenge.Status),
		State:             challengeState(challenge, time.Now().UTC()),
		ParticipantsCount: participants,
		Joined:            joined,
		CanManage:         canManage,
		DailyPostLimit:    MaxChallengePostsPerDay,
		DailyPostsAwarded: dailyAwarded,
		CreatedAt:         challenge.CreatedAt,
	}
}

// challengeDailyAwarded trả số bài đã được tính điểm hôm nay của viewer.
// Lỗi ở đây chỉ ảnh hưởng hiển thị nên trả 0 thay vì chặn cả request.
func (s *ContributionService) challengeDailyAwarded(ctx context.Context, challengeID, viewerID string) int {
	if viewerID == "" {
		return 0
	}
	counts, err := s.contributionRepo.GetChallengeDailyAwardCounts(
		ctx, viewerID, []string{challengeID}, challengeAwardDate(time.Now().UTC()),
	)
	if err != nil {
		return 0
	}
	return counts[challengeID]
}

// GetChallengeResponses trả danh sách challenge kèm state/joined/can_manage.
// viewerID rỗng nghĩa là khách chưa đăng nhập.
func (s *ContributionService) GetChallengeResponses(ctx context.Context, communityID, viewerID string, includeEnded bool) ([]dto.ChallengeResponse, error) {
	if _, err := s.communityRepo.FindByID(ctx, communityID); err != nil {
		return nil, validations.ErrCommunityNotFound
	}

	challenges, err := s.contributionRepo.GetChallengesByCommunity(ctx, communityID, includeEnded)
	if err != nil {
		return nil, err
	}
	if len(challenges) == 0 {
		return []dto.ChallengeResponse{}, nil
	}

	ids := make([]string, 0, len(challenges))
	for _, challenge := range challenges {
		ids = append(ids, challenge.ID)
	}

	counts, err := s.contributionRepo.CountChallengeParticipantsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	joined := map[string]bool{}
	canManage := false
	dailyAwarded := map[string]int{}
	if viewerID != "" {
		joined, err = s.contributionRepo.GetJoinedChallengeIDs(ctx, viewerID, ids)
		if err != nil {
			return nil, err
		}
		dailyAwarded, err = s.contributionRepo.GetChallengeDailyAwardCounts(
			ctx, viewerID, ids, challengeAwardDate(time.Now().UTC()),
		)
		if err != nil {
			return nil, err
		}
		if isAdmin, roleErr := s.groupRole.IsAdmin(ctx, communityID, viewerID); roleErr == nil {
			canManage = isAdmin
		}
	}

	responses := make([]dto.ChallengeResponse, 0, len(challenges))
	for i := range challenges {
		challenge := &challenges[i]
		responses = append(responses, s.buildChallengeResponse(
			ctx, challenge, counts[challenge.ID], joined[challenge.ID], canManage, dailyAwarded[challenge.ID],
		))
	}

	sort.SliceStable(responses, func(i, j int) bool {
		pi, pj := challengeStatePriority(responses[i].State), challengeStatePriority(responses[j].State)
		if pi != pj {
			return pi < pj
		}
		return responses[i].StartDate.Before(responses[j].StartDate)
	})

	return responses, nil
}

// GetChallengeDetail trả chi tiết một challenge trong community.
func (s *ContributionService) GetChallengeDetail(ctx context.Context, communityID, challengeID, viewerID string) (*dto.ChallengeResponse, error) {
	if _, err := s.communityRepo.FindByID(ctx, communityID); err != nil {
		return nil, validations.ErrCommunityNotFound
	}

	challenge, err := s.contributionRepo.GetChallengeByID(ctx, challengeID)
	if err != nil || challenge.CommunityID != communityID {
		return nil, errorsapp.New(errorsapp.ErrCodeContribChallengeNotFound)
	}

	participants, err := s.contributionRepo.CountChallengeParticipants(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	joined, canManage := false, false
	if viewerID != "" {
		if _, err := s.contributionRepo.FindChallengeParticipant(ctx, challengeID, viewerID); err == nil {
			joined = true
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if isAdmin, roleErr := s.groupRole.IsAdmin(ctx, communityID, viewerID); roleErr == nil {
			canManage = isAdmin
		}
	}

	response := s.buildChallengeResponse(ctx, challenge, participants, joined, canManage,
		s.challengeDailyAwarded(ctx, challengeID, viewerID))
	return &response, nil
}

// UpdateChallenge cho phép admin sửa nội dung, thời gian, phần thưởng và đổi trạng thái
// (active → completed/cancelled, hoặc mở lại).
func (s *ContributionService) UpdateChallenge(ctx context.Context, adminID, communityID, challengeID string, input dto.UpdateChallengeInput) (*dto.ChallengeResponse, error) {
	if _, err := s.communityRepo.FindByID(ctx, communityID); err != nil {
		return nil, validations.ErrCommunityNotFound
	}
	if err := s.groupRole.RequireRole(ctx, communityID, adminID, models.GroupRoleAdmin); err != nil {
		return nil, validations.ErrNotCommunityAdmin
	}

	challenge, err := s.contributionRepo.GetChallengeByID(ctx, challengeID)
	if err != nil || challenge.CommunityID != communityID {
		return nil, errorsapp.New(errorsapp.ErrCodeContribChallengeNotFound)
	}

	merged := dto.CreateChallengeInput{
		Title:           challenge.Title,
		Description:     challenge.Description,
		Hashtag:         challenge.Hashtag,
		PointsPerPost:   challenge.PointsPerPost,
		StartDate:       challenge.StartDate.Format(time.RFC3339),
		EndDate:         challenge.EndDate.Format(time.RFC3339),
		MaxParticipants: challenge.MaxParticipants,
	}
	if input.Title != nil {
		merged.Title = *input.Title
	}
	if input.Description != nil {
		merged.Description = *input.Description
	}
	if input.Hashtag != nil {
		merged.Hashtag = *input.Hashtag
	}
	if input.PointsPerPost != nil {
		merged.PointsPerPost = *input.PointsPerPost
	}
	if input.StartDate != nil {
		merged.StartDate = *input.StartDate
	}
	if input.EndDate != nil {
		merged.EndDate = *input.EndDate
	}
	if input.MaxParticipants != nil {
		if *input.MaxParticipants <= 0 {
			merged.MaxParticipants = nil
		} else {
			limit := *input.MaxParticipants
			merged.MaxParticipants = &limit
		}
	}

	start, end, err := s.validation.ValidateUpdateChallenge(merged)
	if err != nil {
		return nil, err
	}

	status := challenge.Status
	if input.Status != nil {
		normalized := strings.ToLower(strings.TrimSpace(*input.Status))
		if err := s.validation.ValidateChallengeStatus(normalized); err != nil {
			return nil, err
		}
		status = models.ChallengeStatus(normalized)
	}

	challenge.Title = strings.TrimSpace(merged.Title)
	challenge.Description = strings.TrimSpace(merged.Description)
	challenge.Hashtag = strings.TrimSpace(merged.Hashtag)
	challenge.PointsPerPost = merged.PointsPerPost
	challenge.StartDate = start
	challenge.EndDate = end
	challenge.MaxParticipants = merged.MaxParticipants
	challenge.Status = status

	if err := s.contributionRepo.UpdateChallenge(ctx, challenge); err != nil {
		return nil, err
	}

	participants, err := s.contributionRepo.CountChallengeParticipants(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	joined := false
	if _, err := s.contributionRepo.FindChallengeParticipant(ctx, challengeID, adminID); err == nil {
		joined = true
	}

	response := s.buildChallengeResponse(ctx, challenge, participants, joined, true,
		s.challengeDailyAwarded(ctx, challengeID, adminID))
	return &response, nil
}

func (s *ContributionService) GetChallengeParticipants(ctx context.Context, challengeID string) ([]dto.ChallengeParticipantItem, error) {
	if _, err := s.contributionRepo.GetChallengeByID(ctx, challengeID); err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodeContribChallengeNotFound)
	}
	return s.contributionRepo.GetChallengeParticipants(ctx, challengeID)
}

func (s *ContributionService) CreateChallenge(ctx context.Context, adminID, communityID string, input dto.CreateChallengeInput) error {
	if _, err := s.communityRepo.FindByID(ctx, communityID); err != nil {
		return validations.ErrCommunityNotFound
	}
	if err := s.groupRole.RequireRole(ctx, communityID, adminID, models.GroupRoleAdmin); err != nil {
		return validations.ErrNotCommunityAdmin
	}

	startDate, endDate, err := s.validation.ValidateCreateChallenge(input)
	if err != nil {
		return err
	}

	challenge := &models.CommunityChallenge{
		CommunityID:     communityID,
		CreatorID:       adminID,
		Title:           strings.TrimSpace(input.Title),
		Description:     strings.TrimSpace(input.Description),
		Hashtag:         strings.TrimSpace(input.Hashtag),
		PointsPerPost:   input.PointsPerPost,
		StartDate:       startDate,
		EndDate:         endDate,
		MaxParticipants: input.MaxParticipants,
		Status:          models.ChallengeStatusActive,
		CreatedAt:       time.Now().UTC(),
	}

	return s.contributionRepo.CreateChallenge(ctx, challenge)
}

func (s *ContributionService) JoinChallenge(ctx context.Context, userID, challengeID string) error {
	challenge, err := s.contributionRepo.GetChallengeByID(ctx, challengeID)
	if err != nil {
		return errorsapp.New(errorsapp.ErrCodeContribChallengeNotFound)
	}
	if challenge.Status != models.ChallengeStatusActive {
		return errorsapp.New(errorsapp.ErrCodeContribChallengeInactive)
	}

	if err := s.groupRole.RequireRole(ctx, challenge.CommunityID, userID, models.GroupRoleMember); err != nil {
		return validations.ErrNotCommunityMember
	}

	now := time.Now().UTC()
	if now.Before(challenge.StartDate) {
		return errorsapp.New(errorsapp.ErrCodeContribChallengeNotStarted)
	}
	if now.After(challenge.EndDate) {
		return errorsapp.New(errorsapp.ErrCodeContribChallengeEnded)
	}

	err = s.contributionRepo.JoinChallengeAtomic(ctx, &models.ChallengeParticipant{
		ChallengeID: challengeID,
		UserID:      userID,
		JoinedAt:    now,
	}, challenge.MaxParticipants)
	if err != nil {
		if errors.Is(err, repository.ErrRepoChallengeAlreadyJoined) {
			return errorsapp.New(errorsapp.ErrCodeContribAlreadyJoined)
		}
		if errors.Is(err, repository.ErrRepoChallengeParticipantLimitHit) {
			return errorsapp.New(errorsapp.ErrCodeContribParticipantLimitHit)
		}
		return err
	}
	return nil
}

func (s *ContributionService) GetActiveChallenges(ctx context.Context, communityID string) ([]models.CommunityChallenge, error) {
	if _, err := s.communityRepo.FindByID(ctx, communityID); err != nil {
		return nil, validations.ErrCommunityNotFound
	}
	return s.contributionRepo.GetActiveChallenges(ctx, communityID)
}

func (s *ContributionService) ProcessChallengePost(ctx context.Context, communityID, userID, content string) error {
	if strings.TrimSpace(content) == "" {
		return nil
	}

	hashtags := extractHashtags(content)
	if len(hashtags) == 0 {
		return nil
	}

	// Only find challenges in this community (prevents cross-community leak).
	matchedChallenges := make(map[string]models.CommunityChallenge)
	for _, hashtag := range hashtags {
		challenges, err := s.contributionRepo.FindActiveChallengesByHashtag(ctx, communityID, hashtag)
		if err != nil {
			return err
		}
		for _, challenge := range challenges {
			matchedChallenges[challenge.ID] = challenge
		}
	}

	if len(matchedChallenges) == 0 {
		return nil
	}

	// Silently skip non-members (safety guard — upstream already enforces this).
	if err := s.groupRole.RequireRole(ctx, communityID, userID, models.GroupRoleMember); err != nil {
		return nil
	}

	communityDeltas := make(map[string]*models.MemberContribution)
	now := time.Now().UTC()

	for _, challenge := range matchedChallenges {
		if now.Before(challenge.StartDate) || now.After(challenge.EndDate) {
			continue
		}

		// Chỉ người đã chủ động join mới được tính điểm. Người chưa join vẫn đăng bài
		// bình thường, chỉ không được cộng vào challenge (tránh chiếm suất của người khác).
		if _, err := s.contributionRepo.FindChallengeParticipant(ctx, challenge.ID, userID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}

		// Trần điểm/ngày: chỉ N bài đầu tiên trong ngày mỗi challenge được tính;
		// các bài sau bị bỏ qua hoàn toàn để chặn farm điểm bằng spam hashtag.
		awarded, err := s.contributionRepo.IncrementChallengeDailyAward(
			ctx, challenge.ID, userID, challengeAwardDate(now), MaxChallengePostsPerDay, challenge.PointsPerPost,
		)
		if err != nil {
			return err
		}
		if !awarded {
			continue
		}

		if err := s.contributionRepo.IncrementChallengeParticipant(ctx, challenge.ID, userID, challenge.PointsPerPost); err != nil {
			return err
		}

		if _, exists := communityDeltas[challenge.CommunityID]; !exists {
			communityDeltas[challenge.CommunityID] = &models.MemberContribution{CommunityID: challenge.CommunityID, UserID: userID}
		}
		communityDeltas[challenge.CommunityID].EventParticipations++
	}

	for deltaCommunityID, delta := range communityDeltas {
		for i := 0; i < delta.EventParticipations; i++ {
			if _, err := s.ensureContribution(ctx, deltaCommunityID, userID); err != nil {
				return err
			}
			if err := s.contributionRepo.AtomicIncrement(ctx, deltaCommunityID, userID, "event_participations"); err != nil {
				return err
			}
		}

		contribution, err := s.contributionRepo.GetContribution(ctx, deltaCommunityID, userID)
		if err != nil {
			return err
		}
		policy, err := s.ensurePolicy(ctx, deltaCommunityID)
		if err != nil {
			return err
		}
		if err := s.recalculateAndPersistScoreOnly(ctx, contribution, policy); err != nil {
			return err
		}
	}

	return nil
}

func (s *ContributionService) IncrementValidPosts(ctx context.Context, communityID, userID string) error {
	if _, err := s.ensureContribution(ctx, communityID, userID); err != nil {
		return err
	}
	if err := s.contributionRepo.AtomicIncrement(ctx, communityID, userID, "valid_posts"); err != nil {
		return err
	}
	contribution, err := s.contributionRepo.GetContribution(ctx, communityID, userID)
	if err != nil {
		return err
	}
	policy, err := s.ensurePolicy(ctx, communityID)
	if err != nil {
		return err
	}
	return s.recalculateAndPersistScoreOnly(ctx, contribution, policy)
}

func (s *ContributionService) IncrementQualityComments(ctx context.Context, communityID, userID string) error {
	if _, err := s.ensureContribution(ctx, communityID, userID); err != nil {
		return err
	}
	if err := s.contributionRepo.AtomicIncrement(ctx, communityID, userID, "quality_comments"); err != nil {
		return err
	}
	contribution, err := s.contributionRepo.GetContribution(ctx, communityID, userID)
	if err != nil {
		return err
	}
	policy, err := s.ensurePolicy(ctx, communityID)
	if err != nil {
		return err
	}
	return s.recalculateAndPersistScoreOnly(ctx, contribution, policy)
}

func (s *ContributionService) IncrementPositiveReactions(ctx context.Context, communityID, userID string) error {
	if _, err := s.ensureContribution(ctx, communityID, userID); err != nil {
		return err
	}
	if err := s.contributionRepo.AtomicIncrement(ctx, communityID, userID, "positive_reactions"); err != nil {
		return err
	}
	contribution, err := s.contributionRepo.GetContribution(ctx, communityID, userID)
	if err != nil {
		return err
	}
	policy, err := s.ensurePolicy(ctx, communityID)
	if err != nil {
		return err
	}
	return s.recalculateAndPersistScoreOnly(ctx, contribution, policy)
}

func (s *ContributionService) IncrementEventParticipations(ctx context.Context, communityID, userID string) error {
	if _, err := s.ensureContribution(ctx, communityID, userID); err != nil {
		return err
	}
	if err := s.contributionRepo.AtomicIncrement(ctx, communityID, userID, "event_participations"); err != nil {
		return err
	}
	contribution, err := s.contributionRepo.GetContribution(ctx, communityID, userID)
	if err != nil {
		return err
	}
	policy, err := s.ensurePolicy(ctx, communityID)
	if err != nil {
		return err
	}
	return s.recalculateAndPersistScoreOnly(ctx, contribution, policy)
}

func (s *ContributionService) DecrementQualityComments(ctx context.Context, communityID, userID string) error {
	if _, err := s.ensureContribution(ctx, communityID, userID); err != nil {
		return err
	}
	if err := s.contributionRepo.AtomicDecrement(ctx, communityID, userID, "quality_comments"); err != nil {
		return err
	}
	contribution, err := s.contributionRepo.GetContribution(ctx, communityID, userID)
	if err != nil {
		return err
	}
	policy, err := s.ensurePolicy(ctx, communityID)
	if err != nil {
		return err
	}
	return s.recalculateAndPersistScoreOnly(ctx, contribution, policy)
}

func (s *ContributionService) DecrementPositiveReactions(ctx context.Context, communityID, userID string) error {
	if _, err := s.ensureContribution(ctx, communityID, userID); err != nil {
		return err
	}
	if err := s.contributionRepo.AtomicDecrement(ctx, communityID, userID, "positive_reactions"); err != nil {
		return err
	}
	contribution, err := s.contributionRepo.GetContribution(ctx, communityID, userID)
	if err != nil {
		return err
	}
	policy, err := s.ensurePolicy(ctx, communityID)
	if err != nil {
		return err
	}
	return s.recalculateAndPersistScoreOnly(ctx, contribution, policy)
}

func (s *ContributionService) checkAndAssignBadge(contribution *models.MemberContribution, policy *models.CommunityPolicy) {
	if !policy.BadgeEnabled {
		contribution.BadgeType = nil
		return
	}
	if contribution.ContributionScore >= policy.TopContributorThreshold {
		badge := "Top Contributor"
		contribution.BadgeType = &badge
		return
	}
	contribution.BadgeType = nil
}

func (s *ContributionService) checkAndPromoteToMod(ctx context.Context, contribution *models.MemberContribution, policy *models.CommunityPolicy) error {
	if !policy.AutoPromoteEnabled {
		return nil
	}
	if contribution.PromotedToMod {
		return nil
	}
	if contribution.ContributionScore >= policy.ModeratorPromotionThreshold {
		if s.communityRepo == nil {
			contribution.PromotedToMod = true
			return nil
		}
		isMod, err := s.groupRole.IsModOrAbove(ctx, contribution.CommunityID, contribution.UserID)
		if err != nil {
			return err
		}
		if isMod {
			return nil
		}

		// Update role first (idempotent — harmless if another goroutine already did it).
		if err := s.communityRepo.UpdateUserRole(ctx, contribution.CommunityID, contribution.UserID, models.RoleGroupMod); err != nil {
			return err
		}

		// Atomically claim the promotion flag — only one goroutine wins.
		claimed, err := s.contributionRepo.TryClaimPromotion(ctx, contribution.CommunityID, contribution.UserID)
		if err != nil {
			return err
		}
		if !claimed {
			return nil
		}

		s.notificationService.Create(ctx, contribution.UserID, nil, models.NotificationTypeCommunityRoleChanged, "bạn đã được thăng chức lên Moderator nhờ điểm đóng góp", nil, nil, nil)
		contribution.PromotedToMod = true
	}
	return nil
}

func (s *ContributionService) recalculateAndPersistContribution(ctx context.Context, contribution *models.MemberContribution, policy *models.CommunityPolicy) (*models.MemberContribution, error) {
	contribution.ContributionScore = s.calculateContributionScore(contribution, policy)
	s.checkAndAssignBadge(contribution, policy)
	if err := s.checkAndPromoteToMod(ctx, contribution, policy); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	contribution.LastCalculatedAt = now
	contribution.UpdatedAt = &now

	if err := s.contributionRepo.UpsertContribution(ctx, contribution); err != nil {
		return nil, err
	}
	return contribution, nil
}

func (s *ContributionService) recalculateAndPersistScoreOnly(ctx context.Context, contribution *models.MemberContribution, policy *models.CommunityPolicy) error {
	contribution.ContributionScore = s.calculateContributionScore(contribution, policy)
	s.checkAndAssignBadge(contribution, policy)
	if err := s.checkAndPromoteToMod(ctx, contribution, policy); err != nil {
		return err
	}
	now := time.Now().UTC()
	contribution.LastCalculatedAt = now
	contribution.UpdatedAt = &now
	return s.contributionRepo.UpdateContributionCalculatedFields(ctx, contribution)
}

func (s *ContributionService) calculateContributionScore(contribution *models.MemberContribution, policy *models.CommunityPolicy) int {
	return (contribution.ValidPosts * policy.PostWeight) +
		(contribution.QualityComments * policy.CommentWeight) +
		(contribution.PositiveReactions * policy.ReactionWeight) +
		(contribution.EventParticipations * policy.EventWeight)
}

func (s *ContributionService) ensurePolicy(ctx context.Context, communityID string) (*models.CommunityPolicy, error) {
	policy, err := s.contributionRepo.GetPolicy(ctx, communityID)
	if err == nil {
		return policy, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	policy = s.defaultPolicy(communityID)
	if err := s.contributionRepo.UpsertPolicy(ctx, policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func (s *ContributionService) ensureContribution(ctx context.Context, communityID, userID string) (*models.MemberContribution, error) {
	contribution, err := s.contributionRepo.GetContribution(ctx, communityID, userID)
	if err == nil {
		return contribution, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	contribution = &models.MemberContribution{
		CommunityID:         communityID,
		UserID:              userID,
		ValidPosts:          0,
		QualityComments:     0,
		PositiveReactions:   0,
		EventParticipations: 0,
		ContributionScore:   0,
		BadgeType:           nil,
		PromotedToMod:       false,
		LastCalculatedAt:    time.Now().UTC(),
		CreatedAt:           time.Now().UTC(),
	}
	if err := s.contributionRepo.UpsertContribution(ctx, contribution); err != nil {
		return nil, err
	}
	return contribution, nil
}

func (s *ContributionService) defaultPolicy(communityID string) *models.CommunityPolicy {
	now := time.Now().UTC()
	return &models.CommunityPolicy{
		CommunityID:                 communityID,
		PostWeight:                  10,
		CommentWeight:               5,
		ReactionWeight:              2,
		EventWeight:                 20,
		TopContributorThreshold:     2500,
		ModeratorPromotionThreshold: 5000,
		AutoPromoteEnabled:          true,
		BadgeEnabled:                true,
		CreatedAt:                   now,
	}
}

func extractHashtags(content string) []string {
	tokens := strings.Fields(content)
	result := make([]string, 0, len(tokens))
	seen := map[string]bool{}

	for _, token := range tokens {
		token = strings.Trim(token, ",.!?;:)]}\"'")
		if !strings.HasPrefix(token, "#") || len(token) < 2 {
			continue
		}
		key := strings.ToLower(token)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, token)
	}

	return result
}
