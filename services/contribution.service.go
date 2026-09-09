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
	"strings"
	"time"

	"gorm.io/gorm"
)

type ContributionService struct {
	contributionRepo   *repository.ContributionRepository
	communityRepo      *repository.CommunityRepository
	profileRepo        *repository.ProfileRepository
	notificationService *NotificationService
	validation         *validations.ContributionValidation
	groupRole          *utils.GroupRoleChecker
}

func NewContributionService(contributionRepo *repository.ContributionRepository, communityRepo *repository.CommunityRepository, profileRepo *repository.ProfileRepository, notificationService *NotificationService, validation *validations.ContributionValidation) *ContributionService {
	return &ContributionService{
		contributionRepo:   contributionRepo,
		communityRepo:      communityRepo,
		profileRepo:        profileRepo,
		notificationService: notificationService,
		validation:         validation,
		groupRole:          utils.NewGroupRoleChecker(communityRepo.GetUserRole),
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
			MaxParticipants:   challenge.MaxParticipants,
			Status:            string(challenge.Status),
			ParticipantsCount: participantsCount,
		})
	}

	return responses, nil
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

		participant, err := s.contributionRepo.FindChallengeParticipant(ctx, challenge.ID, userID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			participant = nil
		}
		if participant == nil {
			if err := s.contributionRepo.JoinChallengeAtomic(ctx, &models.ChallengeParticipant{
				ChallengeID: challenge.ID,
				UserID:      userID,
				JoinedAt:    now,
			}, challenge.MaxParticipants); err != nil {
				if errors.Is(err, repository.ErrRepoChallengeAlreadyJoined) {
					participant, err = s.contributionRepo.FindChallengeParticipant(ctx, challenge.ID, userID)
					if err != nil {
						return err
					}
				} else if errors.Is(err, repository.ErrRepoChallengeParticipantLimitHit) {
					continue
				} else {
					return err
				}
			}
			if participant == nil {
				participant = &models.ChallengeParticipant{ChallengeID: challenge.ID, UserID: userID, JoinedAt: now}
			}
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
		for i := 0; i < delta.EventParticipations; i++ {
			if _, err := s.ensureContribution(ctx, communityID, userID); err != nil {
				return err
			}
			if err := s.contributionRepo.AtomicIncrement(ctx, communityID, userID, "event_participations"); err != nil {
				return err
			}
		}

		contribution, err := s.contributionRepo.GetContribution(ctx, communityID, userID)
		if err != nil {
			return err
		}
		policy, err := s.ensurePolicy(ctx, communityID)
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
