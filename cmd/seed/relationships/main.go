package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"linkup/config"
	"linkup/db"
)

type userRow struct {
	ID    int
	Email string
}

type communitySeed struct {
	CreatorID   int
	Name        string
	Description string
	AvatarURI   string
}

type followSeed struct {
	FollowerID  int
	FollowingID int
}

type blockSeed struct {
	UserID        int
	BlockedUserID int
}

type friendSeed struct {
	SenderID   int
	ReceiverID int
	Status     string
}

type groupMemberSeed struct {
	CommunityID int
	UserID      int
	Role        string
	Points      int
}

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	conn, err := db.ConnectDb(config.GetEnv())
	if err != nil {
		log.Fatalf("DB connection: failed (%v)", err)
	}
	defer conn.Close()

	if err := ensurePhase4Tables(conn); err != nil {
		log.Fatalf("ensure phase4 tables failed: %v", err)
	}

	users, err := fetchUsers(conn)
	if err != nil {
		log.Fatalf("fetch users failed: %v", err)
	}
	if len(users) < 8 {
		log.Fatalf("need at least 8 users for phase4 seeding, found %d", len(users))
	}

	communities := buildCommunities(users, 8)
	commInserted, err := seedCommunities(conn, communities)
	if err != nil {
		log.Fatalf("seed communities failed: %v", err)
	}

	follows := buildFollows(users)
	followsInserted, err := seedFollows(conn, follows)
	if err != nil {
		log.Fatalf("seed follows failed: %v", err)
	}

	blocks := buildBlocks(users)
	blocksInserted, err := seedBlocks(conn, blocks)
	if err != nil {
		log.Fatalf("seed blocks failed: %v", err)
	}

	friends := buildFriends(users)
	friendsInserted, err := seedFriends(conn, friends)
	if err != nil {
		log.Fatalf("seed friends failed: %v", err)
	}

	communityIDs, err := fetchCommunityIDs(conn, 8)
	if err != nil {
		log.Fatalf("fetch communities failed: %v", err)
	}

	groupMembers := buildGroupMembers(users, communityIDs)
	membersInserted, err := seedGroupMembers(conn, groupMembers)
	if err != nil {
		log.Fatalf("seed group members failed: %v", err)
	}

	fmt.Printf("Seed phase4: success (communities=%d, follows=%d, blocks=%d, friends=%d, group_members=%d)\n",
		commInserted,
		followsInserted,
		blocksInserted,
		friendsInserted,
		membersInserted,
	)
}

func ensurePhase4Tables(conn *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS follows (
			id INT AUTO_INCREMENT PRIMARY KEY,
			follower_id INT NOT NULL,
			following_id INT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY follow_unique (follower_id, following_id),
			CONSTRAINT fk_follows_follower FOREIGN KEY (follower_id) REFERENCES users(id),
			CONSTRAINT fk_follows_following FOREIGN KEY (following_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS blocks (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			blocked_user_id INT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY block_unique (user_id, blocked_user_id),
			CONSTRAINT fk_blocks_user FOREIGN KEY (user_id) REFERENCES users(id),
			CONSTRAINT fk_blocks_blocked_user FOREIGN KEY (blocked_user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS friends (
			id INT AUTO_INCREMENT PRIMARY KEY,
			sender_id INT NOT NULL,
			receiver_id INT NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY friend_unique (sender_id, receiver_id),
			CONSTRAINT fk_friends_sender FOREIGN KEY (sender_id) REFERENCES users(id),
			CONSTRAINT fk_friends_receiver FOREIGN KEY (receiver_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS communities (
			id INT AUTO_INCREMENT PRIMARY KEY,
			creator_id INT NOT NULL,
			name VARCHAR(100) NOT NULL UNIQUE,
			description VARCHAR(1000),
			avatar_uri VARCHAR(500),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_communities_creator FOREIGN KEY (creator_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS group_members (
			id INT AUTO_INCREMENT PRIMARY KEY,
			community_id INT NOT NULL,
			user_id INT NOT NULL,
			role VARCHAR(20) NOT NULL DEFAULT 'member',
			points INT NOT NULL DEFAULT 0,
			joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY group_member_unique (community_id, user_id),
			CONSTRAINT fk_group_members_community FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE,
			CONSTRAINT fk_group_members_user FOREIGN KEY (user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	for _, q := range queries {
		if _, err := conn.Exec(q); err != nil {
			return fmt.Errorf("create phase4 table: %w", err)
		}
	}

	return nil
}

func fetchUsers(conn *sql.DB) ([]userRow, error) {
	rows, err := conn.Query("SELECT id, email FROM users ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	users := make([]userRow, 0)
	for rows.Next() {
		var u userRow
		if err := rows.Scan(&u.ID, &u.Email); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}

	return users, rows.Err()
}

func fetchCommunityIDs(conn *sql.DB, limit int) ([]int, error) {
	rows, err := conn.Query("SELECT id FROM communities ORDER BY id ASC LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("query community ids: %w", err)
	}
	defer rows.Close()

	ids := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan community id: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

func buildCommunities(users []userRow, total int) []communitySeed {
	communities := make([]communitySeed, 0, total)
	for i := 0; i < total; i++ {
		creator := users[i%len(users)]
		communities = append(communities, communitySeed{
			CreatorID:   creator.ID,
			Name:        fmt.Sprintf("Seed Community %02d", i+1),
			Description: fmt.Sprintf("Cộng đồng seed %d do %s tạo.", i+1, creator.Email),
			AvatarURI:   fmt.Sprintf("/seeds/communities/community-%02d.png", i+1),
		})
	}
	return communities
}

func buildFollows(users []userRow) []followSeed {
	items := make([]followSeed, 0)
	visited := make(map[string]struct{})
	for i, follower := range users {
		for j := 1; j <= 3; j++ {
			following := users[(i+j)%len(users)]
			if follower.ID == following.ID {
				continue
			}
			key := fmt.Sprintf("%d_%d", follower.ID, following.ID)
			if _, ok := visited[key]; ok {
				continue
			}
			visited[key] = struct{}{}
			items = append(items, followSeed{FollowerID: follower.ID, FollowingID: following.ID})
		}
	}
	return items
}

func buildBlocks(users []userRow) []blockSeed {
	items := make([]blockSeed, 0)
	for i := 0; i < len(users) && i < 20; i++ {
		blocker := users[i]
		blocked := users[(i+5)%len(users)]
		if blocker.ID == blocked.ID {
			continue
		}
		items = append(items, blockSeed{UserID: blocker.ID, BlockedUserID: blocked.ID})
	}
	return items
}

func buildFriends(users []userRow) []friendSeed {
	items := make([]friendSeed, 0)
	statuses := []string{"accepted", "pending", "rejected"}
	for i := 0; i < len(users)-1 && i < 25; i++ {
		sender := users[i]
		receiver := users[i+1]
		items = append(items, friendSeed{
			SenderID:   sender.ID,
			ReceiverID: receiver.ID,
			Status:     statuses[i%len(statuses)],
		})
	}
	return items
}

func buildGroupMembers(users []userRow, communityIDs []int) []groupMemberSeed {
	items := make([]groupMemberSeed, 0)
	for ci, communityID := range communityIDs {
		creator := users[ci%len(users)]
		items = append(items, groupMemberSeed{
			CommunityID: communityID,
			UserID:      creator.ID,
			Role:        "admin",
			Points:      100,
		})
		for j := 1; j <= 4; j++ {
			member := users[(ci+j*2)%len(users)]
			if member.ID == creator.ID {
				continue
			}
			items = append(items, groupMemberSeed{
				CommunityID: communityID,
				UserID:      member.ID,
				Role:        "member",
				Points:      10 * j,
			})
		}
	}
	return items
}

func seedCommunities(conn *sql.DB, items []communitySeed) (int64, error) {
	values := make([][]any, 0, len(items))
	for _, item := range items {
		values = append(values, []any{item.CreatorID, item.Name, item.Description, item.AvatarURI})
	}
	return bulkInsertIgnore(conn, "communities", []string{"creator_id", "name", "description", "avatar_uri"}, values)
}

func seedFollows(conn *sql.DB, items []followSeed) (int64, error) {
	values := make([][]any, 0, len(items))
	for _, item := range items {
		values = append(values, []any{item.FollowerID, item.FollowingID})
	}
	return bulkInsertIgnore(conn, "follows", []string{"follower_id", "following_id"}, values)
}

func seedBlocks(conn *sql.DB, items []blockSeed) (int64, error) {
	values := make([][]any, 0, len(items))
	for _, item := range items {
		values = append(values, []any{item.UserID, item.BlockedUserID})
	}
	return bulkInsertIgnore(conn, "blocks", []string{"user_id", "blocked_user_id"}, values)
}

func seedFriends(conn *sql.DB, items []friendSeed) (int64, error) {
	values := make([][]any, 0, len(items))
	for _, item := range items {
		values = append(values, []any{item.SenderID, item.ReceiverID, item.Status})
	}
	return bulkInsertIgnore(conn, "friends", []string{"sender_id", "receiver_id", "status"}, values)
}

func seedGroupMembers(conn *sql.DB, items []groupMemberSeed) (int64, error) {
	values := make([][]any, 0, len(items))
	for _, item := range items {
		values = append(values, []any{item.CommunityID, item.UserID, item.Role, item.Points})
	}
	return bulkInsertIgnore(conn, "group_members", []string{"community_id", "user_id", "role", "points"}, values)
}

func bulkInsertIgnore(conn *sql.DB, table string, columns []string, rows [][]any) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}

	var builder strings.Builder
	builder.WriteString("INSERT IGNORE INTO ")
	builder.WriteString(table)
	builder.WriteString(" (")
	builder.WriteString(strings.Join(columns, ", "))
	builder.WriteString(") VALUES ")

	args := make([]any, 0, len(rows)*len(columns))
	for i, row := range rows {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString("(")
		for j := range columns {
			if j > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString("?")
		}
		builder.WriteString(")")
		args = append(args, row...)
	}

	result, err := conn.Exec(builder.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("insert seed rows into %s: %w", table, err)
	}

	inserted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected for %s: %w", table, err)
	}
	return inserted, nil
}
