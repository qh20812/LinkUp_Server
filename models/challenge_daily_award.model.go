package models

import "time"

// ChallengeDailyAward theo dõi số bài viết được tính điểm trong ngày của một người
// trong một challenge — dùng để chặn farm điểm (trần mỗi ngày).
// Khóa duy nhất: (challenge_id, user_id, award_date).
type ChallengeDailyAward struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ChallengeID  string    `json:"challenge_id" gorm:"type:varchar(36);uniqueIndex:uk_challenge_daily_award,priority:1"`
	UserID       string    `json:"user_id" gorm:"type:varchar(36);uniqueIndex:uk_challenge_daily_award,priority:2"`
	AwardDate    time.Time `json:"award_date" gorm:"type:date;uniqueIndex:uk_challenge_daily_award,priority:3"`
	PostsCount   int       `json:"posts_count" gorm:"not null;default:0"`
	PointsEarned int       `json:"points_earned" gorm:"not null;default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
