package contribution

import (
	"database/sql"
	"fmt"
	"time"

	"linkup/cmd/seed/internal"
	"linkup/config"
)

// Run tạo dữ liệu demo cho hệ contribution: challenge (đang diễn ra, sắp diễn ra,
// đã kết thúc), người tham gia và điểm đóng góp để leaderboard/badge có dữ liệu.
func Run(env config.Env, state *internal.SeedState) error {
	database, err := internal.Connect(env)
	if err != nil {
		return err
	}
	defer database.Close()

	if len(state.CommunityIDs) < 2 || len(state.UserIDs) < 8 {
		return nil
	}

	now := time.Now().UTC()
	activeChallengeIDs := map[int]string{}

	challenges := []struct {
		communityIdx int
		title        string
		description  string
		hashtag      string
		points       int
		start        time.Time
		end          time.Time
		max          *int
		status       string
	}{
		{
			communityIdx: 0,
			title:        "Ảnh đẹp mỗi ngày",
			description:  "Chia sẻ bức ảnh bạn tâm đắc nhất trong tuần này.",
			hashtag:      "#AnhDepMoiNgay",
			points:       15,
			start:        now.Add(-48 * time.Hour),
			end:          now.Add(120 * time.Hour),
			status:       "active",
		},
		{
			communityIdx: 0,
			title:        "Viết về Go",
			description:  "Kể lại một điều thú vị bạn học được với Go.",
			hashtag:      "#GoMoiNgay",
			points:       20,
			start:        now.Add(72 * time.Hour),
			end:          now.Add(240 * time.Hour),
			max:          intPtr(20),
			status:       "active",
		},
		{
			communityIdx: 0,
			title:        "Chào mừng thành viên mới",
			description:  "Giới thiệu bản thân với cộng đồng.",
			hashtag:      "#HelloLinkUp",
			points:       10,
			start:        now.Add(-360 * time.Hour),
			end:          now.Add(-240 * time.Hour),
			status:       "completed",
		},
		{
			communityIdx: 1,
			title:        "Kiến trúc cloud tuần này",
			description:  "Chia sẻ sơ đồ kiến trúc bạn đang dùng.",
			hashtag:      "#CloudWeekly",
			points:       25,
			start:        now.Add(-24 * time.Hour),
			end:          now.Add(144 * time.Hour),
			max:          intPtr(50),
			status:       "active",
		},
	}

	for _, c := range challenges {
		communityID := state.CommunityIDs[c.communityIdx]
		// Theo quy ước của bước relationships: GROUP_ADMIN của community i là user 2i+2.
		creatorID := state.UserIDs[c.communityIdx*2+2]
		challengeID := internal.UUID()

		var maxParticipants any
		if c.max != nil {
			maxParticipants = *c.max
		}

		if err := internal.Exec(database,
			`INSERT INTO community_challenges
				(id, community_id, creator_id, title, description, hashtag, points_per_post,
				 start_date, end_date, max_participants, status, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			challengeID, communityID, creatorID, c.title, c.description, c.hashtag,
			c.points, c.start, c.end, maxParticipants, c.status, now,
		); err != nil {
			return fmt.Errorf("contribution: insert challenge %s: %w", c.hashtag, err)
		}

		if c.status == "active" && !c.start.After(now) {
			activeChallengeIDs[c.communityIdx] = challengeID
		}
	}

	for communityIdx, challengeID := range activeChallengeIDs {
		memberIDs, err := loadMembers(database, state.CommunityIDs[communityIdx], 6)
		if err != nil {
			return err
		}
		for i, userID := range memberIDs {
			posts := 3 - i
			if posts < 1 {
				posts = 1
			}
			if err := internal.Exec(database,
				`INSERT INTO challenge_participants
					(id, challenge_id, user_id, posts_count, total_points_earned, joined_at)
				 VALUES (?, ?, ?, ?, ?, ?)`,
				internal.UUID(), challengeID, userID, posts, posts*15, now.Add(-time.Duration(i+1)*time.Hour),
			); err != nil {
				return fmt.Errorf("contribution: insert participant: %w", err)
			}
		}
	}

	for communityIdx := 0; communityIdx < len(state.CommunityIDs); communityIdx++ {
		memberIDs, err := loadMembers(database, state.CommunityIDs[communityIdx], 8)
		if err != nil {
			return err
		}
		for i, userID := range memberIDs {
			validPosts, comments, reactions, events := 4+i*3, 2+i*2, 8+i*4, i
			// Một thành viên đạt ngưỡng Top Contributor để badge hiển thị được.
			if communityIdx == 0 && i == 0 {
				validPosts, comments, reactions, events = 240, 60, 100, 5
			}

			score := validPosts*10 + comments*5 + reactions*2 + events*20
			var badge any
			if score >= 2500 {
				badge = "Top Contributor"
			}

			if err := internal.Exec(database,
				`INSERT INTO member_contributions
					(id, community_id, user_id, valid_posts, quality_comments, positive_reactions,
					 event_participations, contribution_score, badge_type, promoted_to_mod,
					 last_calculated_at, created_at, updated_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, ?)`,
				internal.UUID(), state.CommunityIDs[communityIdx], userID,
				validPosts, comments, reactions, events, score, badge, now, now, now,
			); err != nil {
				return fmt.Errorf("contribution: insert member contribution: %w", err)
			}
		}
	}

	return nil
}

func loadMembers(database *sql.DB, communityID string, limit int) ([]string, error) {
	rows, err := database.Query(
		`SELECT user_id FROM group_members WHERE community_id = ? ORDER BY joined_at ASC LIMIT ?`,
		communityID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("contribution: query members: %w", err)
	}
	defer rows.Close()

	ids := make([]string, 0, limit)
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("contribution: scan member: %w", err)
		}
		ids = append(ids, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("contribution: iterate members: %w", err)
	}
	return ids, nil
}

func intPtr(v int) *int {
	return &v
}
