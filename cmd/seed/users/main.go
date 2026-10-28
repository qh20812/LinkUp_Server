package users

import (
	"fmt"
	"time"

	"linkup/cmd/seed/internal"
	"linkup/config"
	"linkup/seeddata"

	"golang.org/x/crypto/bcrypt"
)

// Run chèn toàn bộ tài khoản từ seeddata.Users và nạp map username → ID
// để các bước sau tham chiếu theo key thay vì index mảng.
func Run(env config.Env, state *internal.SeedState) error {
	database, err := internal.Connect(env)
	if err != nil {
		return fmt.Errorf("users: connect: %w", err)
	}
	defer database.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("users: hash: %w", err)
	}
	passwordHash := string(hash)
	now := time.Now().UTC()

	for _, u := range seeddata.Users {
		userID := internal.UUID()
		if err := internal.Exec(database,
			`INSERT INTO users (id, username, email, password_hash, status, email_verified_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, NULL)`,
			userID, u.Username, u.Email, passwordHash, u.Status, now, now,
		); err != nil {
			return fmt.Errorf("users: insert %s: %w", u.Username, err)
		}
		state.UserIDs = append(state.UserIDs, userID)
		state.UserIDByUsername[u.Username] = userID
		if u.Key == seeddata.DemoUser {
			state.DemoUserID = userID
		}
		if u.Key == seeddata.PartnerUser {
			state.PartnerUserID = userID
		}
	}

	if state.DemoUserID == "" {
		return fmt.Errorf("users: demo user %q not found in seeddata", seeddata.DemoUser)
	}
	if state.PartnerUserID == "" {
		return fmt.Errorf("users: partner user %q not found in seeddata", seeddata.PartnerUser)
	}

	return nil
}
