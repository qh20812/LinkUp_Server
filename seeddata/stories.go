package seeddata

// StoryReact là một lượt cảm xúc trên story.
type StoryReact struct {
	User  string
	Emoji string // code reaction, không có dấu hai chấm
	Click int    // 1..5
}

// StoryReply là một trả lời riêng cho story.
type StoryReply struct {
	User    string
	Content string
	AgeMin  int
}

// Story là một tin 24h. ExpiresAt được tính bằng CreatedAt + 24h ở bước seed.
type Story struct {
	Key      string
	User     string
	MediaURI string
	Type     string // image | video
	Caption  string
	AgeHours float64
	Viewers  []string
	Reacts   []StoryReact
	Replies  []StoryReply
}

// Stories là nguồn dữ liệu duy nhất cho story seed.
// Mỗi story còn hạn (dưới 24h) để story bar hiển thị đầy đủ khi chụp ảnh.
var Stories = []Story{
	{
		Key: "minhanh-cafe-sang", User: "minhanh.nguyen",
		MediaURI: ImgCaPhe1, Type: "image",
		Caption:  "6h45, hồ chưa có ai. Bắt đầu ngày bằng 20 phút đi bộ ☕",
		AgeHours: 2.5,
		Viewers:  []string{"thutrang_le", "haiyen.vu", "camtu.trinh", "myduyen.duong", "bichha_to", "khanhlinh.do"},
		Reacts:   []StoryReact{{User: "thutrang_le", Emoji: "love", Click: 2}, {User: "haiyen.vu", Emoji: "like", Click: 1}, {User: "camtu.trinh", Emoji: "heart", Click: 1}},
		Replies:  []StoryReply{{User: "camtu.trinh", Content: "Sáng nào cũng dậy sớm được, ngưỡng mộ thật.", AgeMin: 90}},
	},
	{
		Key: "nhatminh-sapa-som", User: "nhatminh.ho",
		MediaURI: ImgSapa1, Type: "image",
		Caption:  "5h20 ở Tả Van. Sương còn chưa tan, máy ảnh đã đầy 200 tấm.",
		AgeHours: 5,
		Viewers:  []string{"minhanh.nguyen", "thutrang_le", "giahuy.ly", "khanhlinh.do", "myduyen.duong"},
		Reacts:   []StoryReact{{User: "minhanh.nguyen", Emoji: "wow", Click: 1}, {User: "thutrang_le", Emoji: "love", Click: 1}},
	},
	{
		Key: "thanhtung-chay-sang", User: "thanhtung.ngo",
		MediaURI: ImgMarathon2, Type: "image",
		Caption:  "12km sáng nay. Gió Sài Gòn tháng này dễ chịu bất thường.",
		AgeHours: 7,
		Viewers:  []string{"phuongthao.mai", "minhanh.nguyen", "hoangnam.dev", "tuankiet.dang"},
		Reacts:   []StoryReact{{User: "phuongthao.mai", Emoji: "fire", Click: 2}, {User: "hoangnam.dev", Emoji: "clap", Click: 1}},
		Replies:  []StoryReply{{User: "phuongthao.mai", Content: "Mai 5h30 Tao Đàn nhé anh?", AgeMin: 300}},
	},
	{
		Key: "quocbao-quan-sang", User: "quocbao.pham",
		MediaURI: ImgCaPhe2, Type: "image",
		Caption:  "Ly đầu tiên trong ngày. Hội An sáng nay 24 độ, khách chưa tới.",
		AgeHours: 8,
		Viewers:  []string{"minhanh.nguyen", "nhatminh.ho", "thutrang_le"},
		Reacts:   []StoryReact{{User: "nhatminh.ho", Emoji: "love", Click: 1}},
	},
	{
		Key: "myduyen-cho-noi-video", User: "myduyen.duong",
		MediaURI: VideoSeaTurtle, Type: "video",
		Caption:  "Ghe chạy qua chợ nổi lúc 5h40, tiếng máy nổ hai bên. Nghe là thấy miền Tây.",
		AgeHours: 9,
		Viewers:  []string{"ngochan_bui", "minhanh.nguyen", "haiyen.vu", "bichha_to"},
		Reacts:   []StoryReact{{User: "ngochan_bui", Emoji: "love", Click: 3}, {User: "haiyen.vu", Emoji: "wow", Click: 1}},
	},
	{
		Key: "haiyen-thu-vien", User: "haiyen.vu",
		MediaURI: ImgBooks1, Type: "image",
		Caption:  "Ngồi thư viện từ 8h. Hôm nay quyết tâm xong 2 passage.",
		AgeHours: 11,
		Viewers:  []string{"minhanh.nguyen", "thutrang_le", "myduyen.duong"},
		Reacts:   []StoryReact{{User: "thutrang_le", Emoji: "clap", Click: 1}},
		Replies:  []StoryReply{{User: "thutrang_le", Content: "Cố lên em, tối cô gửi đề Writing cho.", AgeMin: 480}},
	},
	{
		Key: "bichha-truc-dem", User: "bichha_to",
		MediaURI: ImgHospital, Type: "image",
		Caption:  "2h sáng. Hành lang này quen tới mức đi không cần đèn.",
		AgeHours: 13,
		Viewers:  []string{"minhanh.nguyen", "thutrang_le", "ngochan_bui", "duclong.phan"},
		Reacts:   []StoryReact{{User: "thutrang_le", Emoji: "love", Click: 1}, {User: "ngochan_bui", Emoji: "love", Click: 1}},
	},
	{
		Key: "camtu-lookbook", User: "camtu.trinh",
		MediaURI: ImgMakeup2, Type: "image",
		Caption:  "Vừa xong look cô dâu. Da mỏng, mắt nhấn nâu, môi trầm. Ưng nhất từ đầu tháng.",
		AgeHours: 15,
		Viewers:  []string{"minhanh.nguyen", "haiyen.vu", "khanhlinh.do", "phuongthao.mai"},
		Reacts:   []StoryReact{{User: "minhanh.nguyen", Emoji: "love", Click: 2}, {User: "khanhlinh.do", Emoji: "heart", Click: 1}},
	},
	{
		Key: "tuankiet-cong-truong", User: "tuankiet.dang",
		MediaURI: ImgBridge, Type: "image",
		Caption:  "Đổ bê tông đêm. Gió sông lạnh, nhưng trời không mưa là mừng rồi.",
		AgeHours: 17,
		Viewers:  []string{"thanhtung.ngo", "hoangnam.dev", "minhanh.nguyen"},
		Reacts:   []StoryReact{{User: "hoangnam.dev", Emoji: "clap", Click: 1}},
	},
	{
		Key: "giahuy-yoga-sang", User: "giahuy.ly",
		MediaURI: ImgYoga, Type: "image",
		Caption:  "20 phút trước khi mở laptop. Ngày nào cũng vậy, 4 tháng rồi.",
		AgeHours: 19,
		Viewers:  []string{"minhanh.nguyen", "bichha_to", "thutrang_le"},
		Reacts:   []StoryReact{{User: "bichha_to", Emoji: "like", Click: 1}},
	},
	{
		Key: "khanhlinh-melbourne", User: "khanhlinh.do",
		MediaURI: ImgSaiGon1, Type: "image",
		Caption:  "Nhớ nhà. Nhưng Melbourne chiều nay cũng đẹp.",
		AgeHours: 21,
		Viewers:  []string{"minhanh.nguyen", "haiyen.vu", "myduyen.duong"},
		Reacts:   []StoryReact{{User: "haiyen.vu", Emoji: "sad", Click: 1}, {User: "minhanh.nguyen", Emoji: "love", Click: 1}},
	},
	{
		Key: "hoangnam-deploy", User: "hoangnam.dev",
		MediaURI: ImgLaptop2, Type: "image",
		Caption:  "Deploy lúc 11h đêm. Không ai muốn làm, nhưng phải làm.",
		AgeHours: 22,
		Viewers:  []string{"giahuy.ly", "minhanh.nguyen", "tuankiet.dang", "thanhtung.ngo"},
		Reacts:   []StoryReact{{User: "giahuy.ly", Emoji: "like", Click: 1}},
	},
}

// StoryMute là một lượt tắt story của người khác.
type StoryMute struct {
	User    string
	Muted   string
	AgeDays int
}

// StoryMutes là các lượt tắt story seed.
var StoryMutes = []StoryMute{
	{User: "phuongthao.mai", Muted: "quocthinh.lam", AgeDays: 4},
	{User: "khanhlinh.do", Muted: "thanhtung.ngo", AgeDays: 7},
	{User: "ngochan_bui", Muted: "anhdung.chu", AgeDays: 2},
}

// StoriesByUser nhóm story theo user, giữ nguyên thứ tự thời gian.
func StoriesByUser() map[string][]Story {
	out := make(map[string][]Story)
	for _, s := range Stories {
		out[s.User] = append(out[s.User], s)
	}
	return out
}
