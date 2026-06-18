package profiles

import (
	"fmt"
	"time"

	"linkup/cmd/seed/internal"
	"linkup/config"
)

func Run(env config.Env, state *internal.SeedState) error {
	database, err := internal.Connect(env)
	if err != nil {
		return fmt.Errorf("profiles: connect: %w", err)
	}
	defer database.Close()

	now := time.Now().UTC()

	bios := []string{
		"Software engineer & open source enthusiast",
		"Digital artist and photographer",
		"Full-stack developer by day, gamer by night",
		"Building the future, one line at a time",
		"Exploring the world through code",
		"Data scientist & ML enthusiast",
		"Product designer with a passion for UX",
		"DevOps engineer keeping the servers running",
		"Tech lead & mentor",
		"Frontend wizard & CSS artist",
		"Backend architect & API designer",
		"Mobile developer crafting great experiences",
		"Security researcher & ethical hacker",
		"Cloud infrastructure specialist",
		"Open source maintainer",
		"AI researcher & NLP enthusiast",
		"Game developer & 3D artist",
		"Blockchain developer & Web3 explorer",
		"Technical writer & content creator",
		"Startup founder & entrepreneur",
	}

	avatars := []string{
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user01",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user02",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user03",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user04",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user05",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user06",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user07",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user08",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user09",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user10",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user11",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user12",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user13",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user14",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user15",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user16",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user17",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user18",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user19",
		"https://api.dicebear.com/7.x/avataaars/svg?seed=user20",
	}

	for i, uid := range state.UserIDs {
		isPrivate := i >= 15
		if err := internal.Exec(database,
			`INSERT INTO profiles (id, user_id, avatar_uri, bio, is_private_profile, is_private_posts, allow_stranger_friend_request, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			internal.UUID(), uid, avatars[i], bios[i], isPrivate, false, true, now,
		); err != nil {
			return fmt.Errorf("profiles: insert for user %s: %w", uid, err)
		}
	}

	return nil
}
