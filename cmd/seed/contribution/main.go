package contribution

import (
	"database/sql"
	"fmt"
	"time"

	"linkup/cmd/seed/internal"
	"linkup/config"
	"linkup/seeddata"
)

// Run tạo dữ liệu demo cho hệ contribution: thử thách của từng cộng đồng,
// người tham gia và điểm đóng góp để leaderboard/badge có dữ liệu thật.
func Run(env config.Env, state *internal.SeedState) error {
	database, err := internal.Connect(env)
	if err != nil {
		return fmt.Errorf("contribution: connect: %w", err)
	}
	defer database.Close()

	now := time.Now().UTC()

	communityIDByKey := func(key string) (string, error) {
		id, ok := state.CommunityIDByKey[key]
		if !ok {
			return "", fmt.Errorf("contribution: unknown community %q", key)
		}
		return id, nil
	}

	activeChallenges := make(map[string]string) // communityID → challengeID

	for _, c := range seeddata.Challenges {
		communityID, err := communityIDByKey(c.Community)
		if err != nil {
			return err
		}
		community, ok := seeddata.CommunityByKey(c.Community)
		if !ok {
			return fmt.Errorf("contribution: community %q missing from seeddata", c.Community)
		}
		creatorID, ok := state.UserIDByUsername[community.Creator]
		if !ok {
			return fmt.Errorf("contribution: unknown creator %q", community.Creator)
		}

		var maxParticipants any
		if c.MaxParticipants > 0 {
			maxParticipants = c.MaxParticipants
		}

		start := now.Add(time.Duration(c.StartHours * float64(time.Hour)))
		end := now.Add(time.Duration(c.EndHours * float64(time.Hour)))

		challengeID := internal.UUID()
		if err := internal.Exec(database,
			`INSERT INTO community_challenges
				(id, community_id, creator_id, title, description, hashtag, points_per_post,
				 start_date, end_date, max_participants, status, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			challengeID, communityID, creatorID, c.Title, c.Description, c.Hashtag,
			c.PointsPerPost, start, end, maxParticipants, c.Status, start,
		); err != nil {
			return fmt.Errorf("contribution: insert challenge %s: %w", c.Hashtag, err)
		}

		if c.Status == "active" && !start.After(now) {
			activeChallenges[communityID] = challengeID
		}
	}

	// Người tham gia thử thách đang chạy — lấy thành viên tích cực nhất.
	for communityID, challengeID := range activeChallenges {
		members, err := loadMembers(database, communityID, 6)
		if err != nil {
			return err
		}
		perPost := 0
		rows, err := database.Query(`SELECT points_per_post FROM community_challenges WHERE id = ?`, challengeID)
		if err != nil {
			return fmt.Errorf("contribution: query points_per_post: %w", err)
		}
		if rows.Next() {
			if err := rows.Scan(&perPost); err != nil {
				rows.Close()
				return fmt.Errorf("contribution: scan points_per_post: %w", err)
			}
		}
		rows.Close()

		for i, member := range members {
			posts := 3 - i/2
			if posts < 1 {
				posts = 1
			}
			if err := internal.Exec(database,
				`INSERT INTO challenge_participants
					(id, challenge_id, user_id, posts_count, total_points_earned, joined_at)
				 VALUES (?, ?, ?, ?, ?, ?)`,
				internal.UUID(), challengeID, member.userID, posts, posts*perPost,
				now.Add(-time.Duration(i+1)*6*time.Hour),
			); err != nil {
				return fmt.Errorf("contribution: insert participant: %w", err)
			}
		}
	}

	// Điểm đóng góp: suy ra từ số điểm thành viên, giữ hai bảng khớp nhau.
	for _, community := range seeddata.Communities {
		communityID, err := communityIDByKey(community.Key)
		if err != nil {
			return err
		}
		members, err := loadMembers(database, communityID, 500)
		if err != nil {
			return err
		}

		for _, member := range members {
			points := member.points
			validPosts := points / 55
			qualityComments := points / 35
			positiveReactions := points / 18
			events := points / 700
			score := validPosts*10 + qualityComments*5 + positiveReactions*2 + events*20

			var badge any
			if score >= community.Policy.TopThreshold {
				badge = "Top Contributor"
			}
			promoted := 0
			if score >= community.Policy.ModThreshold {
				promoted = 1
			}

			if err := internal.Exec(database,
				`INSERT INTO member_contributions
					(id, community_id, user_id, valid_posts, quality_comments, positive_reactions,
					 event_participations, contribution_score, badge_type, promoted_to_mod,
					 last_calculated_at, created_at, updated_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				internal.UUID(), communityID, member.userID,
				validPosts, qualityComments, positiveReactions, events, score, badge, promoted,
				now, now, now,
			); err != nil {
				return fmt.Errorf("contribution: insert member contribution: %w", err)
			}

			if err := internal.Exec(database,
				`UPDATE group_members SET points = ? WHERE community_id = ? AND user_id = ?`,
				score, communityID, member.userID,
			); err != nil {
				return fmt.Errorf("contribution: sync member points: %w", err)
			}
		}
	}

	return nil
}

type member struct {
	userID string
	points int
}

func loadMembers(database *sql.DB, communityID string, limit int) ([]member, error) {
	rows, err := database.Query(
		`SELECT user_id, points FROM group_members WHERE community_id = ? ORDER BY points DESC, joined_at ASC LIMIT ?`,
		communityID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("contribution: query members: %w", err)
	}
	defer rows.Close()

	members := make([]member, 0, limit)
	for rows.Next() {
		var m member
		if err := rows.Scan(&m.userID, &m.points); err != nil {
			return nil, fmt.Errorf("contribution: scan member: %w", err)
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("contribution: iterate members: %w", err)
	}
	return members, nil
}
