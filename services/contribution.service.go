package services

import (
	"context"
	"errors"
	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	ErrChallengeNotFound            = errors.New("challenge không tồn tại")
	ErrChallengeInactive            = errors.New("challenge không còn hoạt động")
	ErrChallengeNotStarted          = errors.New("challenge chưa bắt đầu")
	ErrChallengeEnded               = errors.New("challenge đã kết thúc")
	ErrChallengeAlreadyJoined       = errors.New("bạn đã tham gia challenge này")
	ErrChallengeParticipantLimitHit = errors.New("challenge đã đủ số lượng người tham gia")
)

type ContributionService struct {
	contributionRepo   *repository.ContributionRepository
	communityRepo      *repository.CommunityRepository
	profileRepo        *repository.ProfileRepository
	postRepo           *repository.PostRepository
	notificationService *NotificationService
	validation         *validations.ContributionValidation
	groupRole          *utils.GroupRoleChecker
}

func NewContributionService(contributionRepo *repository.ContributionRepository, communityRepo *repository.CommunityRepository, profileRepo *repository.ProfileRepository, postRepo *repository.PostRepository, notificationService *NotificationService, validation *validations.ContributionValidation) *ContributionService {
	return &ContributionService{
		contributionRepo:   contributionRepo,
		communityRepo:      communityRepo,
		profileRepo:        profileRepo,
		postRepo:           postRepo,
		notificationService: notificationService,
		validation:         validation,
		groupRole:          utils.NewGroupRoleChecker(communityRepo.GetUserRole),
	}
}

func (s *ContributionService) GetPolicy(ctx context.Context, communityID string) (*models.CommunityPolicy, error) {
	if _, err := s.communityRepo.FindByID(ctx, communityID); err != nil {
		return nil, validations.ErrCommunityNotFound
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

	existing, err := s.contributionRepo.GetPolicy(ctx, communityID)
	if err == nil {
		policy.ID = existing.ID
		policy.CreatedAt = existing.CreatedAt
	}
	if policy.ID == "" {
		policy.ID = utils.GenerateUUID()
	}
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = now
	}

	return s.contributionRepo.UpsertPolicy(ctx, policy)
}

func (s *ContributionService) RecalculateScore(ctx context.Context, communityID, userID string) (*models.MemberContribution, error) {
	if _, err := s.communityRepo.FindByID(ctx, communityID); err != nil {
		return nil, validations.ErrCommunityNotFound
	}

	policy, err := s.GetPolicy(ctx, communityID)
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

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	limit := page * pageSize
	items, err := s.contributionRepo.GetLeaderboard(ctx, communityID, limit)
	if err != nil {
		return nil, err
	}

	start := (page - 1) * pageSize
	if start >= len(items) {
		return []dto.LeaderboardItem{}, nil
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	result := append([]dto.LeaderboardItem(nil), items[start:end]...)
	s.sortLeaderboard(result)
	return result, nil
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

	isModerator, err := s.groupRole.IsModOrAbove(ctx, communityID, userID)
	if err != nil {
		isModerator = false
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
		IsModerator:         isModerator,
		PromotedToMod:       updated.PromotedToMod,
	}, nil
}

func (s *ContributionService) GetActiveChallengeResponses(ctx context.Context, communityID string) ([]dto.ChallengeResponse, error) {
	challenges, err := s.GetActiveChallenges(ctx, communityID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ChallengeResponse, 0, len(challenges))
	for _, challenge := range challenges {
		participantsCount, err := s.contributionRepo.CountChallengeParticipants(ctx, challenge.ID)
		if err != nil {
			return nil, err
		}
		responses = append(responses, dto.ChallengeResponse{
			ID:                challenge.ID,
			Title:             challenge.Title,
			Description:       challenge.Description,
			Hashtag:           challenge.Hashtag,
			PointsPerPost:     challenge.PointsPerPost,
			StartDate:         challenge.StartDate,
			EndDate:           challenge.EndDate,
			Status:            string(challenge.Status),
			ParticipantsCount: participantsCount,
		})
	}

	return responses, nil
}

func (s *ContributionService) GetChallengeParticipants(ctx context.Context, challengeID string) ([]dto.ChallengeParticipantItem, error) {
	return s.contributionRepo.GetChallengeParticipants(ctx, challengeID)
}

func (s *ContributionService) CreateChallenge(ctx context.Context, adminID, communityID string, input dto.CreateChallengeInput) error {
	if _, err := s.communityRepo.FindByID(ctx, communityID); err != nil {
		return validations.ErrCommunityNotFound
	}
	if err := s.groupRole.RequireRole(ctx, communityID, adminID, models.GroupRoleAdmin); err != nil {
		return validations.ErrNotCommunityAdmin
	}
	if err := s.validation.ValidateCreateChallenge(input); err != nil {
		return err
	}

	startDate, err := time.Parse(time.RFC3339, strings.TrimSpace(input.StartDate))
	if err != nil {
		return validations.ErrDateFormatInvalid
	}
	endDate, err := time.Parse(time.RFC3339, strings.TrimSpace(input.EndDate))
	if err != nil {
		return validations.ErrDateFormatInvalid
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
		return ErrChallengeNotFound
	}
	if challenge.Status != models.ChallengeStatusActive {
		return ErrChallengeInactive
	}

	now := time.Now().UTC()
	if now.Before(challenge.StartDate) {
		return ErrChallengeNotStarted
	}
	if now.After(challenge.EndDate) {
		return ErrChallengeEnded
	}

	participants, err := s.contributionRepo.GetChallengeParticipants(ctx, challengeID)
	if err != nil {
		return err
	}
	for _, participant := range participants {
		if participant.UserID == userID {
			return ErrChallengeAlreadyJoined
		}
	}
	if challenge.MaxParticipants != nil && len(participants) >= *challenge.MaxParticipants {
		return ErrChallengeParticipantLimitHit
	}

	return s.contributionRepo.JoinChallenge(ctx, &models.ChallengeParticipant{
		ChallengeID: challengeID,
		UserID:      userID,
		JoinedAt:    now,
	})
}

func (s *ContributionService) GetActiveChallenges(ctx context.Context, communityID string) ([]models.CommunityChallenge, error) {
	if _, err := s.communityRepo.FindByID(ctx, communityID); err != nil {
		return nil, validations.ErrCommunityNotFound
	}
	return s.contributionRepo.GetActiveChallenges(ctx, communityID)
}

func (s *ContributionService) ProcessChallengePost(ctx context.Context, postID, userID, content string) error {
	if strings.TrimSpace(content) == "" {
		return nil
	}

	if _, err := s.postRepo.FindByID(ctx, postID); err != nil {
		return errors.New("bài viết không tồn tại")
	}

	hashtags := extractHashtags(content)
	if len(hashtags) == 0 {
		return nil
	}

	matchedChallenges := make(map[string]models.CommunityChallenge)
	for _, hashtag := range hashtags {
		challenges, err := s.contributionRepo.FindActiveChallengesByHashtag(ctx, hashtag)
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

	communityDeltas := make(map[string]*models.MemberContribution)
	now := time.Now().UTC()

	for _, challenge := range matchedChallenges {
		if now.Before(challenge.StartDate) || now.After(challenge.EndDate) {
			continue
		}

		participant, err := s.contributionRepo.FindChallengeParticipant(ctx, challenge.ID, userID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			participant = nil
		}
		if participant == nil {
			if err := s.contributionRepo.JoinChallenge(ctx, &models.ChallengeParticipant{
				ChallengeID: challenge.ID,
				UserID:      userID,
				JoinedAt:    now,
			}); err != nil {
				return err
			}
			participant = &models.ChallengeParticipant{ChallengeID: challenge.ID, UserID: userID, JoinedAt: now}
		}

		participant.PostsCount++
		participant.TotalPointsEarned += challenge.PointsPerPost
		if err := s.contributionRepo.UpdateChallengeParticipant(ctx, challenge.ID, userID, participant.PostsCount, participant.TotalPointsEarned); err != nil {
			return err
		}

		if _, exists := communityDeltas[challenge.CommunityID]; !exists {
			communityDeltas[challenge.CommunityID] = &models.MemberContribution{CommunityID: challenge.CommunityID, UserID: userID}
		}
		communityDeltas[challenge.CommunityID].EventParticipations++
	}

	for communityID, delta := range communityDeltas {
		contribution, err := s.ensureContribution(ctx, communityID, userID)
		if err != nil {
			return err
		}
		contribution.EventParticipations += delta.EventParticipations
		policy, err := s.GetPolicy(ctx, communityID)
		if err != nil {
			return err
		}
		if _, err := s.recalculateAndPersistContribution(ctx, contribution, policy); err != nil {
			return err
		}
	}

	return nil
}

func (s *ContributionService) IncrementValidPosts(ctx context.Context, communityID, userID string) error {
	contribution, err := s.ensureContribution(ctx, communityID, userID)
	if err != nil {
		return err
	}
	contribution.ValidPosts++
	policy, err := s.GetPolicy(ctx, communityID)
	if err != nil {
		return err
	}
	_, err = s.recalculateAndPersistContribution(ctx, contribution, policy)
	return err
}

func (s *ContributionService) IncrementQualityComments(ctx context.Context, communityID, userID string) error {
	contribution, err := s.ensureContribution(ctx, communityID, userID)
	if err != nil {
		return err
	}
	contribution.QualityComments++
	policy, err := s.GetPolicy(ctx, communityID)
	if err != nil {
		return err
	}
	_, err = s.recalculateAndPersistContribution(ctx, contribution, policy)
	return err
}

func (s *ContributionService) IncrementPositiveReactions(ctx context.Context, communityID, userID string) error {
	contribution, err := s.ensureContribution(ctx, communityID, userID)
	if err != nil {
		return err
	}
	contribution.PositiveReactions++
	policy, err := s.GetPolicy(ctx, communityID)
	if err != nil {
		return err
	}
	_, err = s.recalculateAndPersistContribution(ctx, contribution, policy)
	return err
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
		if err := s.communityRepo.UpdateUserRole(ctx, contribution.CommunityID, contribution.UserID, models.RoleGroupMod); err != nil {
			return err
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

func (s *ContributionService) sortLeaderboard(items []dto.LeaderboardItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].ContributionScore == items[j].ContributionScore {
			return items[i].UserID < items[j].UserID
		}
		return items[i].ContributionScore > items[j].ContributionScore
	})
}
