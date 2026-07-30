package users

import (
	"fmt"
	"time"

	"linkup/cmd/seed/internal"
	"linkup/config"

	"golang.org/x/crypto/bcrypt"
)

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

	type user struct {
		id       string
		username string
		email    string
		status   string
	}

	users := []user{
		{internal.UUID(), "superadmin", "superadmin@example.com", "active"},
		{internal.UUID(), "admin1", "admin1@example.com", "active"},
		{internal.UUID(), "alice_wonder", "alice@example.com", "active"},
		{internal.UUID(), "bob_builder", "bob@example.com", "active"},
		{internal.UUID(), "charlie_dev", "charlie@example.com", "active"},
		{internal.UUID(), "diana_prince", "diana@example.com", "active"},
		{internal.UUID(), "eve_artist", "eve@example.com", "banned"},
		{internal.UUID(), "frank_castle", "frank@example.com", "active"},
		{internal.UUID(), "grace_hopper", "grace@example.com", "active"},
		{internal.UUID(), "hank_pym", "hank@example.com", "active"},
		{internal.UUID(), "ivy_poison", "ivy@example.com", "suspended"},
		{internal.UUID(), "jack_sparrow", "jack@example.com", "active"},
		{internal.UUID(), "kate_bishop", "kate@example.com", "active"},
		{internal.UUID(), "leo_da_vinci", "leo@example.com", "active"},
		{internal.UUID(), "mila_kunis", "mila@example.com", "active"},
		{internal.UUID(), "neo_matrix", "neo@example.com", "active"},
		{internal.UUID(), "olivia_dunham", "olivia@example.com", "active"},
		{internal.UUID(), "peter_parker", "peter@example.com", "banned"},
		{internal.UUID(), "quincy_adams", "quincy@example.com", "active"},
		{internal.UUID(), "rachel_green", "rachel@example.com", "active"},
	}

	for _, u := range users {
		if err := internal.Exec(database,
			`INSERT INTO users (id, username, email, password_hash, status, email_verified_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, NULL)`,
			u.id, u.username, u.email, passwordHash, u.status, now, now,
		); err != nil {
			return fmt.Errorf("users: insert %s: %w", u.username, err)
		}
		state.UserIDs = append(state.UserIDs, u.id)
	}

	return nil
}

