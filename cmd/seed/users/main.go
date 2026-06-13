package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"linkup/config"
	"linkup/db"
)

type seedUser struct {
	Email        string
	PasswordHash string
	Status       string
	Username     string
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

	if err := ensureUsersTable(conn); err != nil {
		log.Fatalf("ensure users table failed: %v", err)
	}

	seedUsers := buildSeedUsers(50)
	inserted, err := seedUsersToDatabase(conn, seedUsers)
	if err != nil {
		log.Fatalf("seed users failed: %v", err)
	}

	fmt.Printf("Seed users: success (%d inserted, %d total)\n", inserted, len(seedUsers))
}

func buildSeedUsers(total int) []seedUser {
	users := make([]seedUser, 0, total)
	for index := 1; index <= total; index++ {
		email := fmt.Sprintf("seed.user%02d@example.com", index)
		username := fmt.Sprintf("seed_user_%02d", index)
		users = append(users, seedUser{
			Email:        email,
			PasswordHash: hashPassword(fmt.Sprintf("SeedPass@%02d", index)),
			Status:       "active",
			Username:     username,
		})
	}
	return users
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func ensureUsersTable(conn *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		email VARCHAR(255) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'active',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`

	if _, err := conn.Exec(query); err != nil {
		return fmt.Errorf("create users table: %w", err)
	}

	return nil
}

func seedUsersToDatabase(conn *sql.DB, users []seedUser) (int64, error) {
	if len(users) == 0 {
		return 0, nil
	}

	query := strings.Builder{}
	query.WriteString("INSERT IGNORE INTO users (email, password_hash, status) VALUES ")

	args := make([]any, 0, len(users)*3)
	for index, user := range users {
		if index > 0 {
			query.WriteString(", ")
		}
		query.WriteString("(?, ?, ?)")
		args = append(args, user.Email, user.PasswordHash, user.Status)
	}

	result, err := conn.Exec(query.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("insert seed users: %w", err)
	}

	inserted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}

	return inserted, nil
}
