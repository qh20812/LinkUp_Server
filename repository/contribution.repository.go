package repository

import (
	"context"
	"errors"
	"fmt"
	"linkup/dto"
	"linkup/models"
	"linkup/utils"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrRepoChallengeAlreadyJoined       = errors.New("bạn đã tham gia challenge này")
	ErrRepoChallengeParticipantLimitHit = errors.New("challenge đã đủ số lượng người tham gia")
)

type ContributionRepository struct {
	db *gorm.DB
}

func NewContributionRepository(db *gorm.DB) *ContributionRepository {
	return &ContributionRepository{db: db}
}

// === Policy ===
func (r *ContributionRepository) GetPolicy(ctx context.Context, communityID string) (*models.CommunityPolicy, error) {
	var policy models.CommunityPolicy
	err := r.db.WithContext(ctx).
		Where("community_id = ?", communityID).
		First(&policy).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

func (r *ContributionRepository) UpsertPolicy(ctx context.Context, policy *models.CommunityPolicy) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.CommunityPolicy
		err := tx.Where("community_id = ?", policy.CommunityID).First(&existing).Error
		now := time.Now().UTC()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				if policy.ID == "" {
					policy.ID = utils.GenerateUUID()
				}
				if policy.CreatedAt.IsZero() {
					policy.CreatedAt = now
				}
				policy.UpdatedAt = nil
				return tx.Create(policy).Error
			}
			return fmt.Errorf("find policy: %w", err)
		}

		policy.ID = existing.ID
		policy.CreatedAt = existing.CreatedAt
		policy.UpdatedAt = &now
		return tx.Save(policy).Error
	})
}

// === Member Contribution ===
func (r *ContributionRepository) GetContribution(ctx context.Context, communityID, userID string) (*models.MemberContribution, error) {
	var contribution models.MemberContribution
	err := r.db.WithContext(ctx).
		Where("community_id = ? AND user_id = ?", communityID, userID).
		First(&contribution).Error
	if err != nil {
		return nil, err
	}
	return &contribution, nil
}

func (r *ContributionRepository) UpsertContribution(ctx context.Context, contribution *models.MemberContribution) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.MemberContribution
		err := tx.Where("community_id = ? AND user_id = ?", contribution.CommunityID, contribution.UserID).First(&existing).Error
		now := time.Now().UTC()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				if contribution.ID == "" {
					contribution.ID = utils.GenerateUUID()
				}
				if contribution.CreatedAt.IsZero() {
					contribution.CreatedAt = now
				}
				if contribution.LastCalculatedAt.IsZero() {
					contribution.LastCalculatedAt = now
				}
				contribution.UpdatedAt = nil
				return tx.Create(contribution).Error
			}
			return fmt.Errorf("find contribution: %w", err)
		}

		contribution.ID = existing.ID
		contribution.CreatedAt = existing.CreatedAt
		contribution.UpdatedAt = &now
		if contribution.LastCalculatedAt.IsZero() {
			contribution.LastCalculatedAt = now
		}
		return tx.Save(contribution).Error
	})
}

func (r *ContributionRepository) GetLeaderboard(ctx context.Context, communityID string, offset, limit int) ([]dto.LeaderboardItem, error) {
	type leaderboardRow struct {
		UserID            string  `gorm:"column:user_id"`
		DisplayName       string  `gorm:"column:display_name"`
		AvatarURI         string  `gorm:"column:avatar_uri"`
		ContributionScore int     `gorm:"column:contribution_score"`
		BadgeType         *string `gorm:"column:badge_type"`
	}

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	var rows []leaderboardRow
	err := r.db.WithContext(ctx).
		Table("member_contributions AS mc").
		Select(`mc.user_id,
				COALESCE(p.display_name, '') AS display_name,
				COALESCE(p.avatar_uri, '') AS avatar_uri,
				mc.contribution_score,
				mc.badge_type`).
		Joins("LEFT JOIN profiles p ON p.user_id = mc.user_id").
		Where("mc.community_id = ?", communityID).
		Order("mc.contribution_score DESC, mc.created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]dto.LeaderboardItem, 0, len(rows))
	for i, row := range rows {
		items = append(items, dto.LeaderboardItem{
			Rank:              offset + i + 1,
			UserID:            row.UserID,
			DisplayName:       row.DisplayName,
			AvatarURI:         row.AvatarURI,
			ContributionScore: row.ContributionScore,
			BadgeType:         row.BadgeType,
		})
	}
	return items, nil
}

func (r *ContributionRepository) GetCommunityMembers(ctx context.Context, communityID string, offset, limit int) ([]dto.CommunityMemberItem, error) {
	type memberRow struct {
		UserID            string    `gorm:"column:user_id"`
		DisplayName       string    `gorm:"column:display_name"`
		AvatarURI         string    `gorm:"column:avatar_uri"`
		Role              string    `gorm:"column:role"`
		JoinedAt          time.Time `gorm:"column:joined_at"`
		ContributionScore int       `gorm:"column:contribution_score"`
		BadgeType         *string   `gorm:"column:badge_type"`
	}

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	var rows []memberRow
	err := r.db.WithContext(ctx).
		Table("group_members AS gm").
		Select(`gm.user_id,
				COALESCE(p.display_name, '') AS display_name,
				COALESCE(p.avatar_uri, '') AS avatar_uri,
				COALESCE((
					SELECT r2.name FROM user_roles ur2
					JOIN roles r2 ON r2.id = ur2.role_id
					WHERE ur2.user_id = gm.user_id
					  AND ur2.scope_id = gm.community_id
					  AND ur2.scope_type = 'community'
					  AND r2.name IN ('GROUP_ADMIN', 'GROUP_MOD', 'GROUP_MEMBER')
					ORDER BY CASE r2.name
						WHEN 'GROUP_ADMIN' THEN 1
						WHEN 'GROUP_MOD' THEN 2
						WHEN 'GROUP_MEMBER' THEN 3
						ELSE 4 END
					LIMIT 1
				), '') AS role,
				gm.joined_at,
				COALESCE(mc.contribution_score, 0) AS contribution_score,
				mc.badge_type`).
		Joins("LEFT JOIN profiles p ON p.user_id = gm.user_id").
		Joins("LEFT JOIN member_contributions mc ON mc.community_id = gm.community_id AND mc.user_id = gm.user_id").
		Where("gm.community_id = ?", communityID).
		Order("mc.contribution_score DESC, gm.joined_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]dto.CommunityMemberItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.CommunityMemberItem{
			UserID:            row.UserID,
			DisplayName:       row.DisplayName,
			AvatarURI:         row.AvatarURI,
			Role:              row.Role,
			JoinedAt:          row.JoinedAt,
			ContributionScore: row.ContributionScore,
			BadgeType:         row.BadgeType,
		})
	}
	return items, nil
}

// === Challenge ===
func (r *ContributionRepository) CreateChallenge(ctx context.Context, challenge *models.CommunityChallenge) error {
	if challenge.ID == "" {
		challenge.ID = utils.GenerateUUID()
	}
	if challenge.Status == "" {
		challenge.Status = models.ChallengeStatusActive
	}
	if challenge.CreatedAt.IsZero() {
		challenge.CreatedAt = time.Now().UTC()
	}
	return r.db.WithContext(ctx).Create(challenge).Error
}

func (r *ContributionRepository) GetChallengeByID(ctx context.Context, challengeID string) (*models.CommunityChallenge, error) {
	var challenge models.CommunityChallenge
	err := r.db.WithContext(ctx).
		Where("id = ?", challengeID).
		First(&challenge).Error
	if err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (r *ContributionRepository) GetActiveChallenges(ctx context.Context, communityID string) ([]models.CommunityChallenge, error) {
	var challenges []models.CommunityChallenge
	err := r.db.WithContext(ctx).
		Where("community_id = ? AND status = ?", communityID, models.ChallengeStatusActive).
		Order("end_date ASC, created_at DESC").
		Find(&challenges).Error
	return challenges, err
}

// GetChallengesByCommunity trả challenge của community.
// includeEnded=false chỉ trả challenge còn hiệu lực (status active và chưa hết hạn).
func (r *ContributionRepository) GetChallengesByCommunity(ctx context.Context, communityID string, includeEnded bool) ([]models.CommunityChallenge, error) {
	query := r.db.WithContext(ctx).Where("community_id = ?", communityID)
	if !includeEnded {
		query = query.Where("status = ? AND end_date >= ?", models.ChallengeStatusActive, time.Now().UTC())
	}
	var challenges []models.CommunityChallenge
	err := query.Order("start_date ASC, created_at DESC").Find(&challenges).Error
	return challenges, err
}

// UpdateChallenge ghi đè toàn bộ field nghiệp vụ (kể cả giá trị zero/NULL).
func (r *ContributionRepository) UpdateChallenge(ctx context.Context, challenge *models.CommunityChallenge) error {
	var maxParticipants any
	if challenge.MaxParticipants != nil {
		maxParticipants = *challenge.MaxParticipants
	}
	return r.db.WithContext(ctx).
		Model(&models.CommunityChallenge{}).
		Where("id = ?", challenge.ID).
		Updates(map[string]any{
			"title":            challenge.Title,
			"description":      challenge.Description,
			"hashtag":          challenge.Hashtag,
			"points_per_post":  challenge.PointsPerPost,
			"start_date":       challenge.StartDate,
			"end_date":         challenge.EndDate,
			"max_participants": maxParticipants,
			"status":           challenge.Status,
		}).Error
}

// CountChallengeParticipantsByIDs đếm participant của nhiều challenge trong 1 query.
func (r *ContributionRepository) CountChallengeParticipantsByIDs(ctx context.Context, challengeIDs []string) (map[string]int, error) {
	counts := make(map[string]int, len(challengeIDs))
	if len(challengeIDs) == 0 {
		return counts, nil
	}
	type row struct {
		ChallengeID string `gorm:"column:challenge_id"`
		Total       int    `gorm:"column:total"`
	}
	var rows []row
	if err := r.db.WithContext(ctx).
		Model(&models.ChallengeParticipant{}).
		Select("challenge_id, COUNT(*) AS total").
		Where("challenge_id IN ?", challengeIDs).
		Group("challenge_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.ChallengeID] = row.Total
	}
	return counts, nil
}

// GetJoinedChallengeIDs trả về tập challenge mà user đã tham gia trong danh sách truyền vào.
func (r *ContributionRepository) GetJoinedChallengeIDs(ctx context.Context, userID string, challengeIDs []string) (map[string]bool, error) {
	joined := make(map[string]bool, len(challengeIDs))
	if len(challengeIDs) == 0 {
		return joined, nil
	}
	var ids []string
	if err := r.db.WithContext(ctx).
		Model(&models.ChallengeParticipant{}).
		Where("user_id = ? AND challenge_id IN ?", userID, challengeIDs).
		Pluck("challenge_id", &ids).Error; err != nil {
		return nil, err
	}
	for _, id := range ids {
		joined[id] = true
	}
	return joined, nil
}

// IncrementChallengeDailyAward cộng 1 bài được tính điểm cho (challenge, user, ngày).
// Trả về false khi đã chạm trần maxPosts trong ngày — bài đó không được cộng điểm,
// nhằm chặn farm điểm bằng cách spam hashtag.
// Khóa row bằng SELECT ... FOR UPDATE nên hai bài đồng thời không thể vượt trần.
func (r *ContributionRepository) IncrementChallengeDailyAward(ctx context.Context, challengeID, userID string, awardDate time.Time, maxPosts, points int) (bool, error) {
	if maxPosts <= 0 {
		return false, nil
	}

	now := time.Now().UTC()
	awarded := false

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Đảm bảo row tồn tại trước khi khóa; INSERT IGNORE tránh lỗi khi hai request
		// đồng thời tạo row cho cùng ngày.
		if err := tx.Exec(
			`INSERT IGNORE INTO challenge_daily_awards
				(id, challenge_id, user_id, award_date, posts_count, points_earned, created_at, updated_at)
			 VALUES (?, ?, ?, ?, 0, 0, ?, ?)`,
			utils.GenerateUUID(), challengeID, userID, awardDate, now, now,
		).Error; err != nil {
			return err
		}

		var award models.ChallengeDailyAward
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("challenge_id = ? AND user_id = ? AND award_date = ?", challengeID, userID, awardDate).
			First(&award).Error; err != nil {
			return err
		}

		if award.PostsCount >= maxPosts {
			return nil
		}

		if err := tx.Model(&models.ChallengeDailyAward{}).
			Where("id = ?", award.ID).
			Updates(map[string]any{
				"posts_count":   award.PostsCount + 1,
				"points_earned": award.PointsEarned + points,
				"updated_at":    now,
			}).Error; err != nil {
			return err
		}

		awarded = true
		return nil
	})
	if err != nil {
		return false, err
	}
	return awarded, nil
}

// GetChallengeDailyAwardCounts trả số bài đã được tính điểm hôm nay của user theo từng challenge.
func (r *ContributionRepository) GetChallengeDailyAwardCounts(ctx context.Context, userID string, challengeIDs []string, awardDate time.Time) (map[string]int, error) {
	counts := make(map[string]int, len(challengeIDs))
	if len(challengeIDs) == 0 {
		return counts, nil
	}

	type row struct {
		ChallengeID string `gorm:"column:challenge_id"`
		PostsCount  int    `gorm:"column:posts_count"`
	}
	var rows []row
	if err := r.db.WithContext(ctx).
		Model(&models.ChallengeDailyAward{}).
		Select("challenge_id, posts_count").
		Where("user_id = ? AND award_date = ? AND challenge_id IN ?", userID, awardDate, challengeIDs).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.ChallengeID] = row.PostsCount
	}
	return counts, nil
}

func (r *ContributionRepository) FindActiveChallengesByHashtag(ctx context.Context, communityID, hashtag string) ([]models.CommunityChallenge, error) {
	var challenges []models.CommunityChallenge
	err := r.db.WithContext(ctx).
		Where("community_id = ? AND LOWER(hashtag) = LOWER(?) AND status = ?", communityID, hashtag, models.ChallengeStatusActive).
		Order("end_date ASC, created_at DESC").
		Find(&challenges).Error
	return challenges, err
}

func (r *ContributionRepository) CountChallengeParticipants(ctx context.Context, challengeID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ChallengeParticipant{}).
		Where("challenge_id = ?", challengeID).
		Count(&count).Error
	return int(count), err
}

func (r *ContributionRepository) FindChallengeParticipant(ctx context.Context, challengeID, userID string) (*models.ChallengeParticipant, error) {
	var participant models.ChallengeParticipant
	err := r.db.WithContext(ctx).
		Where("challenge_id = ? AND user_id = ?", challengeID, userID).
		First(&participant).Error
	if err != nil {
		return nil, err
	}
	return &participant, nil
}

func (r *ContributionRepository) JoinChallenge(ctx context.Context, participant *models.ChallengeParticipant) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.ChallengeParticipant
		err := tx.Where("challenge_id = ? AND user_id = ?", participant.ChallengeID, participant.UserID).First(&existing).Error
		now := time.Now().UTC()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				if participant.ID == "" {
					participant.ID = utils.GenerateUUID()
				}
				if participant.JoinedAt.IsZero() {
					participant.JoinedAt = now
				}
				return tx.Create(participant).Error
			}
			return fmt.Errorf("find participant: %w", err)
		}

		participant.ID = existing.ID
		participant.JoinedAt = existing.JoinedAt
		return tx.Save(participant).Error
	})
}

// JoinChallengeAtomic atomically checks limit + duplicate before inserting.
// Uses FOR UPDATE on the challenge row to serialize concurrent join attempts.
func (r *ContributionRepository) JoinChallengeAtomic(ctx context.Context, participant *models.ChallengeParticipant, maxParticipants *int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Serialize: lock the challenge row so only one goroutine can proceed.
		var challenge models.CommunityChallenge
		if err := tx.Where("id = ?", participant.ChallengeID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&challenge).Error; err != nil {
			return err
		}

		// Check if already joined (race-safe: inside locked transaction).
		var existing int64
		if err := tx.Model(&models.ChallengeParticipant{}).
			Where("challenge_id = ? AND user_id = ?", participant.ChallengeID, participant.UserID).
			Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return ErrRepoChallengeAlreadyJoined
		}

		// Check participant limit (race-safe: count is accurate under lock).
		var count int64
		if err := tx.Model(&models.ChallengeParticipant{}).
			Where("challenge_id = ?", participant.ChallengeID).
			Count(&count).Error; err != nil {
			return err
		}
		if maxParticipants != nil && count >= int64(*maxParticipants) {
			return ErrRepoChallengeParticipantLimitHit
		}

		// Insert
		if participant.ID == "" {
			participant.ID = utils.GenerateUUID()
		}
		if participant.JoinedAt.IsZero() {
			participant.JoinedAt = time.Now().UTC()
		}
		return tx.Create(participant).Error
	})
}

// IncrementChallengeParticipant cộng dồn posts_count và total_points_earned ngay trong
// một câu UPDATE, tránh mất lượt cập nhật khi nhiều bài viết được xử lý song song.
func (r *ContributionRepository) IncrementChallengeParticipant(ctx context.Context, challengeID, userID string, points int) error {
	return r.db.WithContext(ctx).
		Model(&models.ChallengeParticipant{}).
		Where("challenge_id = ? AND user_id = ?", challengeID, userID).
		Updates(map[string]interface{}{
			"posts_count":         gorm.Expr("posts_count + 1"),
			"total_points_earned": gorm.Expr("total_points_earned + ?", points),
		}).Error
}

func (r *ContributionRepository) GetChallengeParticipants(ctx context.Context, challengeID string) ([]dto.ChallengeParticipantItem, error) {
	type participantRow struct {
		UserID            string    `gorm:"column:user_id"`
		DisplayName       string    `gorm:"column:display_name"`
		AvatarURI         string    `gorm:"column:avatar_uri"`
		PostsCount        int       `gorm:"column:posts_count"`
		TotalPointsEarned int       `gorm:"column:total_points_earned"`
		JoinedAt          time.Time `gorm:"column:joined_at"`
	}

	var rows []participantRow
	err := r.db.WithContext(ctx).
		Table("challenge_participants AS cp").
		Select(`cp.user_id,
			COALESCE(p.display_name, '') AS display_name,
			COALESCE(p.avatar_uri, '') AS avatar_uri,
			cp.posts_count,
			cp.total_points_earned,
			cp.joined_at`).
		Joins("LEFT JOIN profiles p ON p.user_id = cp.user_id").
		Where("cp.challenge_id = ?", challengeID).
		Order("cp.total_points_earned DESC, cp.joined_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]dto.ChallengeParticipantItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.ChallengeParticipantItem{
			UserID:            row.UserID,
			DisplayName:       row.DisplayName,
			AvatarURI:         row.AvatarURI,
			PostsCount:        row.PostsCount,
			TotalPointsEarned: row.TotalPointsEarned,
			JoinedAt:          row.JoinedAt,
		})
	}
	return items, nil
}

// === Counting queries backed by stored contribution counters ===
func (r *ContributionRepository) CountValidPostsByUser(ctx context.Context, communityID, userID string) (int, error) {
	contribution, err := r.GetContribution(ctx, communityID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return contribution.ValidPosts, nil
}

func (r *ContributionRepository) CountQualityCommentsByUser(ctx context.Context, communityID, userID string) (int, error) {
	contribution, err := r.GetContribution(ctx, communityID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return contribution.QualityComments, nil
}

func (r *ContributionRepository) CountReactionsReceivedByUser(ctx context.Context, communityID, userID string) (int, error) {
	contribution, err := r.GetContribution(ctx, communityID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return contribution.PositiveReactions, nil
}

func (r *ContributionRepository) CountEventParticipationsByUser(ctx context.Context, communityID, userID string) (int, error) {
	contribution, err := r.GetContribution(ctx, communityID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return contribution.EventParticipations, nil
}

func (r *ContributionRepository) GetRankedContributions(ctx context.Context, communityID string) ([]models.MemberContribution, error) {
	var contributions []models.MemberContribution
	err := r.db.WithContext(ctx).
		Where("community_id = ?", communityID).
		Order("contribution_score DESC, created_at ASC").
		Find(&contributions).Error
	return contributions, err
}

// AtomicIncrement atomically increments a counter column to prevent race conditions.
// Only allows safe column names to prevent SQL injection.
func (r *ContributionRepository) AtomicIncrement(ctx context.Context, communityID, userID, column string) error {
	allowed := map[string]bool{
		"valid_posts":          true,
		"quality_comments":     true,
		"positive_reactions":   true,
		"event_participations": true,
	}
	if !allowed[column] {
		return fmt.Errorf("invalid increment field: %s", column)
	}
	return r.db.WithContext(ctx).
		Model(&models.MemberContribution{}).
		Where("community_id = ? AND user_id = ?", communityID, userID).
		Update(column, gorm.Expr(column+" + ?", 1)).Error
}

// AtomicDecrement atomically decrements a counter column.
// Only allows safe column names to prevent SQL injection.
func (r *ContributionRepository) AtomicDecrement(ctx context.Context, communityID, userID, column string) error {
	allowed := map[string]bool{
		"valid_posts":          true,
		"quality_comments":     true,
		"positive_reactions":   true,
		"event_participations": true,
	}
	if !allowed[column] {
		return fmt.Errorf("invalid decrement field: %s", column)
	}
	return r.db.WithContext(ctx).
		Model(&models.MemberContribution{}).
		Where("community_id = ? AND user_id = ?", communityID, userID).
		Update(column, gorm.Expr(column+" - ?", 1)).Error
}

// UpdateContributionCalculatedFields updates only the calculated fields
// (score, badge, promotion, timestamps) — not the raw counters.
func (r *ContributionRepository) UpdateContributionCalculatedFields(ctx context.Context, contribution *models.MemberContribution) error {
	return r.db.WithContext(ctx).
		Model(&models.MemberContribution{}).
		Where("community_id = ? AND user_id = ?", contribution.CommunityID, contribution.UserID).
		Updates(map[string]interface{}{
			"contribution_score": contribution.ContributionScore,
			"badge_type":         contribution.BadgeType,
			"promoted_to_mod":    contribution.PromotedToMod,
			"last_calculated_at": contribution.LastCalculatedAt,
			"updated_at":         contribution.UpdatedAt,
		}).Error
}

// TryClaimPromotion atomically sets promoted_to_mod = true.
// Returns true if the claim succeeded (was the first caller), false if already claimed.
func (r *ContributionRepository) TryClaimPromotion(ctx context.Context, communityID, userID string) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&models.MemberContribution{}).
		Where("community_id = ? AND user_id = ? AND promoted_to_mod = ?", communityID, userID, false).
		Update("promoted_to_mod", true)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}
