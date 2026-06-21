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

	displayNames := []string{
		"Nguyễn Văn An",
		"Trần Thị Bình",
		"Lê Hoàng Cường",
		"Phạm Minh Đức",
		"Hoàng Thị Mai",
		"Đỗ Gia Huy",
		"Vũ Ngọc Linh",
		"Bùi Thanh Tùng",
		"Đặng Hồng Nhung",
		"Ngô Tuấn Kiệt",
		"Phan Thùy Dương",
		"Lý Quốc Bảo",
		"Tô Minh Tâm",
		"Đinh Văn Hùng",
		"Hồ Cẩm Tú",
		"Mai Thanh Hà",
		"Dương Khánh Vy",
		"Lưu Bích Ngọc",
		"Chu Đức Minh",
		"Trịnh Hoài Nam",
	}

	phones := []string{
		"0912345678",
		"0987654321",
		"0934567890",
		"0901234567",
		"0978912345",
		"0967891234",
		"0945678901",
		"0923456789",
		"0956789012",
		"0891234567",
		"0887654321",
		"0876543210",
		"0867891234",
		"0854321098",
		"0832109876",
		"0823456789",
		"0812345678",
		"0798765432",
		"0789012345",
		"0776543210",
	}

	dobs := []string{
		"1995-03-15",
		"1998-07-22",
		"2000-11-08",
		"1993-05-30",
		"1996-09-12",
		"1999-01-25",
		"1997-04-18",
		"2001-08-03",
		"1994-12-20",
		"1992-06-14",
		"1990-10-05",
		"2002-02-28",
		"1991-07-17",
		"1998-03-09",
		"1996-11-22",
		"1993-08-15",
		"1997-05-01",
		"1999-10-19",
		"1995-12-25",
		"2000-04-07",
	}

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
		dob, _ := time.Parse("2006-01-02", dobs[i])
		if err := internal.Exec(database,
			`INSERT INTO profiles (id, user_id, display_name, phone_number, date_of_birth, avatar_uri, bio, is_private_profile, is_private_posts, allow_stranger_friend_request, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			internal.UUID(), uid, displayNames[i], phones[i], dob, avatars[i], bios[i], isPrivate, false, true, now,
		); err != nil {
			return fmt.Errorf("profiles: insert for user %s: %w", uid, err)
		}
	}

	return nil
}
