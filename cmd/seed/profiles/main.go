package profiles

import (
	"fmt"
	"time"

	"linkup/cmd/seed/internal"
	"linkup/config"
	"linkup/seeddata"
)

// Run chèn hồ sơ đầy đủ từ seeddata.Users (địa điểm, nghề nghiệp, học vấn,
// giới tính, ảnh bìa) — mọi cột đều có trong schema seed.
func Run(env config.Env, state *internal.SeedState) error {
	database, err := internal.Connect(env)
	if err != nil {
		return fmt.Errorf("profiles: connect: %w", err)
	}
	defer database.Close()

	now := time.Now().UTC()

	for _, u := range seeddata.Users {
		userID, ok := state.UserIDByUsername[u.Username]
		if !ok {
			return fmt.Errorf("profiles: missing user id for %s", u.Username)
		}

		var dob any
		if u.DOB != "" {
			parsed, err := time.Parse("2006-01-02", u.DOB)
			if err != nil {
				return fmt.Errorf("profiles: invalid dob for %s: %w", u.Username, err)
			}
			dob = parsed
		}

		allowStranger := 1
		if u.IsPrivate {
			allowStranger = 0
		}

		if err := internal.Exec(database,
			`INSERT INTO profiles (
				id, user_id, display_name, phone_number, date_of_birth, avatar_uri, cover_uri,
				location, hometown_province, current_province, current_ward,
				work, education, work_other, website, bio, gender,
				is_private_profile, is_private_posts, allow_stranger_friend_request, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '', ?, ?, ?, ?, ?, ?, ?)`,
			internal.UUID(), userID, u.DisplayName, u.Phone, dob, u.Avatar, u.Cover,
			u.Location, u.HometownProv, u.CurrentProv, u.CurrentWard,
			u.Work, u.Education, u.Website, u.Bio, u.Gender,
			boolToInt(u.IsPrivate), boolToInt(u.IsPrivate), allowStranger, now,
		); err != nil {
			return fmt.Errorf("profiles: insert for %s: %w", u.Username, err)
		}
	}

	return nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
