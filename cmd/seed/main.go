package main

import (
	"log"

	"linkup/cmd/seed/internal"
	"linkup/cmd/seed/reset"
	"linkup/cmd/seed/schema"
	"linkup/cmd/seed/users"
	"linkup/cmd/seed/core"
	"linkup/cmd/seed/profiles"
	"linkup/cmd/seed/social"
	"linkup/cmd/seed/relationships"
	"linkup/cmd/seed/messaging"
	"linkup/cmd/seed/moderation"
	"linkup/cmd/seed/extended"
	"linkup/config"
)

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}
	env := config.GetEnv()

	state := &internal.SeedState{}

	steps := []struct {
		name string
		run  func() error
	}{
		{"reset", func() error { return reset.Run(env) }},
		{"schema", func() error { return schema.Run(env) }},
		{"users", func() error { return users.Run(env, state) }},
		{"core", func() error { return core.Run(env, state) }},
		{"profiles", func() error { return profiles.Run(env, state) }},
		{"social", func() error { return social.Run(env, state) }},
		{"relationships", func() error { return relationships.Run(env, state) }},
		{"messaging", func() error { return messaging.Run(env, state) }},
		{"moderation", func() error { return moderation.Run(env, state) }},
		{"extended", func() error { return extended.Run(env, state) }},
	}

	for _, step := range steps {
		log.Printf("Running seed: %s", step.name)
		if err := step.run(); err != nil {
			log.Fatalf("seed %s failed: %v", step.name, err)
		}
	}

	log.Println("All seed steps completed successfully")
}
