package seeddata

// Rule là một nội quy cộng đồng, hiển thị theo thứ tự Position.
type Rule struct {
	Category string // general | posting | comment | moderation
	Title    string
	Content  string
}

// Policy là cấu hình tính điểm đóng góp của cộng đồng.
type Policy struct {
	PostWeight       int
	CommentWeight    int
	ReactionWeight   int
	EventWeight      int
	TopThreshold     int
	ModThreshold     int
	AutoPromote      bool
	BadgeEnabled     bool
}

// CommunityMember là một thành viên seed kèm số điểm đóng góp ban đầu.
type CommunityMember struct {
	User   string
	Points int
}

// Community là một cộng đồng seed. Members[0] được coi là quản trị viên
// (nhận role COMMUNITY_ADMIN); các thành viên còn lại là COMMUNITY_MEMBER.
type Community struct {
	Key          string
	Name         string
	Description  string
	Avatar       string
	Background   string
	Creator      string
	Privacy      string // public | invitation_only (khớp models.CommunityPrivacy)
	AutoApprove  bool
	Members      []CommunityMember
	Rules        []Rule
	Policy       Policy
	ChatName     string
	ChatAvatar   string
	CreateChat   bool
}

// DefaultPolicy là chính sách đóng góp mặc định, dùng cho phần lớn cộng đồng.
var DefaultPolicy = Policy{
	PostWeight: 10, CommentWeight: 5, ReactionWeight: 2, EventWeight: 20,
	TopThreshold: 2500, ModThreshold: 5000, AutoPromote: true, BadgeEnabled: true,
}

// Communities là nguồn dữ liệu duy nhất cho cộng đồng seed.
var Communities = []Community{
	{
		Key: "an-gi-hom-nay", Name: "Ăn Gì Hôm Nay",
		Description: "Chỗ để khoe bữa ăn nhà làm, quán ruột và hỏi nhau trưa nay ăn gì. Không quảng cáo, không review dìm hàng.",
		Avatar:      ImgBunCha1, Background: ImgPhoCo1,
		Creator: "ngochan_bui", Privacy: "public", AutoApprove: true,
		Members: []CommunityMember{
			{User: "ngochan_bui", Points: 4820},
			{User: "myduyen.duong", Points: 3140},
			{User: "minhanh.nguyen", Points: 2680},
			{User: "thutrang_le", Points: 2210},
			{User: "duclong.phan", Points: 1870},
			{User: "haiyen.vu", Points: 1420},
			{User: "bichha_to", Points: 980},
			{User: "camtu.trinh", Points: 760},
			{User: "tuankiet.dang", Points: 540},
			{User: "quocbao.pham", Points: 320},
		},
		Rules: []Rule{
			{Category: "general", Title: "Tôn trọng khẩu vị của nhau", Content: "Mỗi vùng miền có cách nấu khác nhau. Góp ý thì góp ý, đừng nói món của người khác là \"không đúng chuẩn\"."},
			{Category: "posting", Title: "Bài đăng kèm ảnh hoặc mô tả rõ ràng", Content: "Nếu khoe món, ghi rõ quán hoặc công thức. Bài chỉ có một câu hỏi sẽ bị nhắc nhở."},
			{Category: "posting", Title: "Không quảng cáo, không bán hàng", Content: "Bài bán đồ ăn, dịch vụ ship, hoặc kèm số điện thoại đặt hàng sẽ bị xoá."},
			{Category: "comment", Title: "Không chê bai ngoại hình", Content: "Bình luận về cân nặng hay ngoại hình của người đăng là lý do bị mute 7 ngày."},
		},
		Policy: DefaultPolicy,
		ChatName: "Ăn Gì Hôm Nay — Chợ Trưa", ChatAvatar: ImgBunCha2, CreateChat: true,
	},
	{
		Key: "chay-bo-moi-ngay", Name: "Chạy Bộ Mỗi Ngày",
		Description: "Nhóm của những người chạy vì sức khoẻ, không vì thành tích. Chia sẻ lịch tập, địa điểm chạy và cả những ngày lười.",
		Avatar:      ImgMarathon1, Background: ImgSapa1,
		Creator: "thanhtung.ngo", Privacy: "public", AutoApprove: true,
		Members: []CommunityMember{
			{User: "thanhtung.ngo", Points: 5310},
			{User: "phuongthao.mai", Points: 2960},
			{User: "minhanh.nguyen", Points: 1740},
			{User: "hoangnam.dev", Points: 1120},
			{User: "tuankiet.dang", Points: 860},
			{User: "bichha_to", Points: 620},
			{User: "duclong.phan", Points: 410},
		},
		Rules: []Rule{
			{Category: "general", Title: "Chạy vì mình, không vì người khác", Content: "Không so sánh pace, không chê người mới chạy. Mọi con số đều đáng được tôn trọng."},
			{Category: "posting", Title: "Ghi rõ địa điểm và cự ly", Content: "Bài rủ chạy phải có giờ, địa điểm và cự ly dự kiến để người khác sắp xếp."},
			{Category: "comment", Title: "Không tư vấn y tế", Content: "Chấn thương nên gặp bác sĩ. Chia sẻ kinh nghiệm cá nhân thì được, kê đơn thì không."},
		},
		Policy:    DefaultPolicy,
		ChatName:  "Chạy Bộ — Nhóm 5h30", ChatAvatar: ImgMarathon2, CreateChat: true,
	},
	{
		Key: "du-lich-bui-viet-nam", Name: "Du Lịch Bụi Việt Nam",
		Description: "Chia sẻ lịch trình, chi phí thật và những chỗ ít người biết. Ưu tiên bài có ảnh tự chụp và số liệu cụ thể.",
		Avatar:      ImgHaGiang1, Background: ImgTrangAn1,
		Creator: "tuankiet.dang", Privacy: "public", AutoApprove: true,
		Members: []CommunityMember{
			{User: "tuankiet.dang", Points: 4120},
			{User: "nhatminh.ho", Points: 3680},
			{User: "myduyen.duong", Points: 2740},
			{User: "giahuy.ly", Points: 1980},
			{User: "khanhlinh.do", Points: 1560},
			{User: "thutrang_le", Points: 1210},
			{User: "minhanh.nguyen", Points: 940},
			{User: "haiyen.vu", Points: 730},
			{User: "bichha_to", Points: 510},
		},
		Rules: []Rule{
			{Category: "general", Title: "Chia sẻ thật, không seeding", Content: "Bài giới thiệu quán, homestay phải ghi rõ bạn có được tài trợ hay không."},
			{Category: "posting", Title: "Có chi phí và thời điểm", Content: "Bài review điểm đến nên ghi chi phí thực tế và tháng đi để người khác tham khảo."},
			{Category: "posting", Title: "Không đăng ảnh không phải của mình", Content: "Ảnh sưu tầm phải ghi nguồn. Ảnh của người khác dùng làm ảnh đại diện bài sẽ bị xoá."},
			{Category: "moderation", Title: "Vi phạm ba lần sẽ bị mời ra khỏi nhóm", Content: "Cộng đồng ưu tiên chất lượng hơn số lượng bài."},
		},
		Policy:    DefaultPolicy,
		ChatName:  "Du Lịch Bụi — Rủ Nhau Đi", ChatAvatar: ImgHoiAn1, CreateChat: true,
	},
	{
		Key: "hoc-tieng-anh-moi-ngay", Name: "Học Tiếng Anh Mỗi Ngày",
		Description: "Mỗi ngày một ít từ vựng, một đoạn writing hoặc một câu hỏi. Không bán khoá học, không cam kết điểm số.",
		Avatar:      ImgBooks1, Background: ImgStudents1,
		Creator: "thutrang_le", Privacy: "public", AutoApprove: true,
		Members: []CommunityMember{
			{User: "thutrang_le", Points: 5620},
			{User: "haiyen.vu", Points: 3480},
			{User: "khanhlinh.do", Points: 2810},
			{User: "myduyen.duong", Points: 1640},
			{User: "minhanh.nguyen", Points: 880},
			{User: "bichha_to", Points: 540},
		},
		Rules: []Rule{
			{Category: "general", Title: "Học đều hơn học nhiều", Content: "Một câu mỗi ngày tốt hơn một bài dài mỗi tháng. Đừng ngại bài ngắn."},
			{Category: "posting", Title: "Ghi rõ trình độ", Content: "Bài hỏi nên ghi band điểm hoặc trình độ hiện tại để người khác trả lời đúng mức."},
			{Category: "comment", Title: "Sửa lỗi thì sửa cho tử tế", Content: "Góp ý ngữ pháp kèm giải thích, không chỉ trích. Người mới cần sự kiên nhẫn."},
		},
		Policy:    DefaultPolicy,
		ChatName:  "English Practice — 30 Phút Tối", ChatAvatar: ImgBooks2, CreateChat: true,
	},
	{
		Key: "it-viet-nam", Name: "IT Việt Nam",
		Description: "Nơi dev Việt chia sẻ kinh nghiệm thật: lương, phỏng vấn, kiến trúc hệ thống và những lần production sập lúc 2h sáng.",
		Avatar:      ImgLaptop1, Background: ImgSaiGon1,
		Creator: "hoangnam.dev", Privacy: "public", AutoApprove: true,
		Members: []CommunityMember{
			{User: "hoangnam.dev", Points: 6240},
			{User: "giahuy.ly", Points: 4380},
			{User: "minhanh.nguyen", Points: 1260},
			{User: "tuankiet.dang", Points: 940},
			{User: "thanhtung.ngo", Points: 720},
			{User: "haiyen.vu", Points: 480},
			{User: "bichha_to", Points: 260},
		},
		Rules: []Rule{
			{Category: "general", Title: "Chia sẻ kinh nghiệm, không chê công nghệ", Content: "Không có ngôn ngữ nào là \"rác\". Tranh luận dựa trên bài toán cụ thể."},
			{Category: "posting", Title: "Bài hỏi kỹ thuật phải có ngữ cảnh", Content: "Ghi rõ stack, phiên bản, thông báo lỗi. Bài chỉ có \"không chạy được\" sẽ bị bỏ qua."},
			{Category: "posting", Title: "Không tuyển dụng ẩn danh", Content: "Bài tuyển dụng phải ghi tên công ty và khoảng lương. Bài \"jobless lương cao\" sẽ bị xoá."},
			{Category: "comment", Title: "Không chia sẻ dữ liệu khách hàng", Content: "Log, ảnh chụp màn hình, mã nguồn nội bộ phải che thông tin nhận dạng."},
		},
		Policy:    DefaultPolicy,
		ChatName:  "IT Việt Nam — Hỏi Đáp", ChatAvatar: ImgLaptop2, CreateChat: true,
	},
	{
		Key: "review-phim-sach", Name: "Review Phim & Sách",
		Description: "Cộng đồng nhỏ, review có tâm. Spoiler phải có cảnh báo. Không review phim chưa xem hết.",
		Avatar:      ImgBooks2, Background: ImgHue1,
		Creator: "bichha_to", Privacy: "invitation_only", AutoApprove: false,
		Members: []CommunityMember{
			{User: "bichha_to", Points: 3240},
			{User: "thutrang_le", Points: 2680},
			{User: "minhanh.nguyen", Points: 1420},
			{User: "haiyen.vu", Points: 860},
			{User: "nhatminh.ho", Points: 420},
		},
		Rules: []Rule{
			{Category: "general", Title: "Cảnh báo spoiler", Content: "Mọi bài có tình tiết quan trọng phải ghi [SPOILER] ở đầu bài."},
			{Category: "posting", Title: "Review phải có lý do", Content: "\"Hay\" hoặc \"dở\" không phải review. Nêu cụ thể điều gì làm bạn thấy vậy."},
			{Category: "moderation", Title: "Duyệt thành viên thủ công", Content: "Nhóm riêng tư, quản trị viên duyệt từng yêu cầu tham gia để giữ chất lượng thảo luận."},
		},
		Policy:    DefaultPolicy,
		CreateChat: false,
	},
}

// JoinRequest là một yêu cầu tham gia cộng đồng đang chờ duyệt.
type JoinRequest struct {
	Community string
	User      string
	Status    string // pending | approved | rejected
	AgeDays   int
}

// JoinRequests là các yêu cầu tham gia hiển thị ở trang quản trị cộng đồng.
var JoinRequests = []JoinRequest{
	{Community: "review-phim-sach", User: "myduyen.duong", Status: "pending", AgeDays: 2},
	{Community: "review-phim-sach", User: "hoangnam.dev", Status: "pending", AgeDays: 5},
	{Community: "review-phim-sach", User: "camtu.trinh", Status: "pending", AgeDays: 8},
	{Community: "du-lich-bui-viet-nam", User: "camtu.trinh", Status: "pending", AgeDays: 3},
	{Community: "it-viet-nam", User: "khanhlinh.do", Status: "pending", AgeDays: 4},
	{Community: "hoc-tieng-anh-moi-ngay", User: "quocbao.pham", Status: "pending", AgeDays: 6},
}

// Challenge là một thử thách đóng góp của cộng đồng.
// StartHours/EndHours tính tương đối so với thời điểm seed (âm = đã qua).
type Challenge struct {
	Community       string
	Title           string
	Description     string
	Hashtag         string
	PointsPerPost   int
	StartHours      float64
	EndHours        float64
	MaxParticipants int // 0 = không giới hạn
	Status          string
}

// Challenges là nguồn dữ liệu duy nhất cho thử thách cộng đồng.
var Challenges = []Challenge{
	{
		Community: "an-gi-hom-nay", Title: "Bữa sáng nhà làm",
		Description:   "Đăng một bữa sáng bạn tự nấu trong tuần này, kèm ảnh và công thức ngắn.",
		Hashtag:       "#BuoiSangNhaLam",
		PointsPerPost: 15, StartHours: -48, EndHours: 120, Status: "active",
	},
	{
		Community: "chay-bo-moi-ngay", Title: "100km trong tháng",
		Description:   "Cộng dồn 100km trong tháng này, đăng ảnh đồng hồ hoặc Strava sau mỗi buổi chạy.",
		Hashtag:       "#100kmTrongThang",
		PointsPerPost: 20, StartHours: -72, EndHours: 200, MaxParticipants: 50, Status: "active",
	},
	{
		Community: "du-lich-bui-viet-nam", Title: "Một ngày một điểm đến",
		Description:   "Mỗi bài là một điểm đến trong ngày, ghi rõ chi phí và cách di chuyển.",
		Hashtag:       "#MotNgayMotDiemDen",
		PointsPerPost: 25, StartHours: -24, EndHours: 144, Status: "active",
	},
	{
		Community: "hoc-tieng-anh-moi-ngay", Title: "30 ngày viết mỗi ngày",
		Description:   "Mỗi ngày một đoạn writing ngắn, không cần dài, miễn là viết đều.",
		Hashtag:       "#30NgayViet",
		PointsPerPost: 10, StartHours: -100, EndHours: 80, Status: "active",
	},
	{
		Community: "review-phim-sach", Title: "Đọc lại một cuốn cũ",
		Description:   "Chọn một cuốn từng đọc, đọc lại và viết cảm nhận ở hiện tại.",
		Hashtag:       "#DocLaiSachCu",
		PointsPerPost: 15, StartHours: -48, EndHours: 264, MaxParticipants: 20, Status: "active",
	},
	{
		Community: "it-viet-nam", Title: "Lần production sập đáng nhớ",
		Description:   "Kể lại sự cố và cách xử lý, che hết thông tin khách hàng.",
		Hashtag:       "#ProductionSap",
		PointsPerPost: 20, StartHours: -360, EndHours: -240, Status: "completed",
	},
}

// CommunityByKey tra cứu cộng đồng theo key.
func CommunityByKey(key string) (Community, bool) {
	for _, c := range Communities {
		if c.Key == key {
			return c, true
		}
	}
	return Community{}, false
}
