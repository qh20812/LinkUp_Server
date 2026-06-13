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

type roleRow struct {
	ID   int
	Name string
}

type profileSeed struct {
	UserID                     int
	Username                   string
	AvatarURI                  string
	Bio                        string
	IsPrivateProfile           bool
	IsPrivatePosts             bool
	AllowStrangerFriendRequest bool
}

type userRoleSeed struct {
	UserID int
	RoleID int
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

	if err := ensurePhase2Tables(conn); err != nil {
		log.Fatalf("ensure phase2 tables failed: %v", err)
	}

	users, err := fetchUsers(conn)
	if err != nil {
		log.Fatalf("fetch users failed: %v", err)
	}
	if len(users) == 0 {
		log.Fatalf("no users found for phase2 seeding")
	}

	roles, err := fetchRoles(conn)
	if err != nil {
		log.Fatalf("fetch roles failed: %v", err)
	}
	if len(roles) == 0 {
		log.Fatalf("no roles found for phase2 seeding")
	}

	profiles := buildProfiles(users)
	profilesInserted, err := seedProfiles(conn, profiles)
	if err != nil {
		log.Fatalf("seed profiles failed: %v", err)
	}

	userRoles := buildUserRoles(users, roles)
	userRolesInserted, err := seedUserRoles(conn, userRoles)
	if err != nil {
		log.Fatalf("seed user_roles failed: %v", err)
	}

	fmt.Printf("Seed phase2: success (profiles=%d, user_roles=%d)\n", profilesInserted, userRolesInserted)
}

func ensurePhase2Tables(conn *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS profiles (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL UNIQUE,
			username VARCHAR(50),
			avatar_uri VARCHAR(500),
			bio VARCHAR(150),
			is_private_profile BOOLEAN NOT NULL DEFAULT FALSE,
			is_private_posts BOOLEAN NOT NULL DEFAULT FALSE,
			allow_stranger_friend_request BOOLEAN NOT NULL DEFAULT TRUE,
			updated_at TIMESTAMP NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS user_roles (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			role_id INT NOT NULL,
			assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY user_role_unique (user_id, role_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	for _, q := range queries {
		if _, err := conn.Exec(q); err != nil {
			return fmt.Errorf("create phase2 table: %w", err)
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

func fetchRoles(conn *sql.DB) ([]roleRow, error) {
	rows, err := conn.Query("SELECT id, name FROM roles ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query roles: %w", err)
	}
	defer rows.Close()

	roles := make([]roleRow, 0)
	for rows.Next() {
		var r roleRow
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, r)
	}
	return roles, rows.Err()
}

func buildProfiles(users []userRow) []profileSeed {
	profiles := make([]profileSeed, 0, len(users))
	for index, user := range users {
		username := strings.SplitN(user.Email, "@", 2)[0]
		profiles = append(profiles, profileSeed{
			UserID:                     user.ID,
			Username:                   username,
			AvatarURI:                  fmt.Sprintf("/seeds/avatars/%03d.png", user.ID),
			Bio:                        fmt.Sprintf("Bio cho tài khoản %s.", username),
			IsPrivateProfile:           index%7 == 0,
			IsPrivatePosts:             index%6 == 0,
			AllowStrangerFriendRequest: index%4 != 0,
		})
	}
	return profiles
}

func buildUserRoles(users []userRow, roles []roleRow) []userRoleSeed {
	roleMap := make(map[string]int)
	for _, role := range roles {
		roleMap[strings.ToLower(role.Name)] = role.ID
	}

	userRoles := make([]userRoleSeed, 0, len(users))
	for index, user := range users {
		roleName := "user"
		if index == 0 {
			roleName = "super_admin"
		} else if index == 1 {
			roleName = "comm_chat"
		} else if index >= 2 && index < 5 {
			roleName = "admin"
		}

		roleID, ok := roleMap[roleName]
		if !ok {
			continue
		}
		userRoles = append(userRoles, userRoleSeed{UserID: user.ID, RoleID: roleID})
	}
	return userRoles
}

func seedProfiles(conn *sql.DB, profiles []profileSeed) (int64, error) {
	values := make([][]any, 0, len(profiles))
	for _, item := range profiles {
		values = append(values, []any{item.UserID, item.Username, item.AvatarURI, item.Bio, item.IsPrivateProfile, item.IsPrivatePosts, item.AllowStrangerFriendRequest})
	}
	return bulkInsertIgnore(conn, "profiles", []string{"user_id", "username", "avatar_uri", "bio", "is_private_profile", "is_private_posts", "allow_stranger_friend_request"}, values)
}

func seedUserRoles(conn *sql.DB, userRoles []userRoleSeed) (int64, error) {
	values := make([][]any, 0, len(userRoles))
	for _, item := range userRoles {
		values = append(values, []any{item.UserID, item.RoleID})
	}
	return bulkInsertIgnore(conn, "user_roles", []string{"user_id", "role_id"}, values)
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
	for index, row := range rows {
		if index > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString("(")
		for colIndex := range columns {
			if colIndex > 0 {
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
