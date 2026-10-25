package seeddata

// FollowEdge là một lượt theo dõi.
type FollowEdge struct {
	Follower  string
	Following string
}

// FriendEdge là một lời mời kết bạn. Status: accepted | pending.
type FriendEdge struct {
	Sender   string
	Receiver string
	Status   string
	AgeDays  int
}

// BlockEdge là một lượt chặn.
type BlockEdge struct {
	User    string
	Blocked string
	AgeDays int
}

// BookmarkEdge là một bài viết được lưu.
type BookmarkEdge struct {
	User    string
	Post    string
	AgeDays int
}

// ShareEdge là một lượt chia sẻ chỉ tăng bộ đếm (không tạo bài repost riêng).
// Repost có bài riêng nằm trong Posts với trường ShareOf.
type ShareEdge struct {
	User   string
	Post   string
	Text   string
	AgeMin int
}

// Shares là các lượt chia sẻ bổ sung, giúp bộ đếm chia sẻ trông tự nhiên.
var Shares = []ShareEdge{
	{User: "phuongthao.mai", Post: "minhanh-phoco-album", Text: "Phố cổ chiều thứ bảy, đúng cái không khí mình thích.", AgeMin: 200},
	{User: "thutrang_le", Post: "minhanh-ot", AgeMin: 60},
	{User: "khanhlinh.do", Post: "thutrang-dalat", Text: "Lịch trình này mình lưu rồi, Tết về đi theo đúng thế.", AgeMin: 40},
	{User: "hoangnam.dev", Post: "bichha-truc-dem", Text: "Đọc xong thấy nghề nào cũng có cái khó riêng.", AgeMin: 35},
	{User: "myduyen.duong", Post: "ngochan-bun-cha", Text: "Cô Hân nấu gì cũng ngon, nhìn là biết.", AgeMin: 50},
	{User: "bichha_to", Post: "minhanh-ot", AgeMin: 100},
	{User: "giahuy.ly", Post: "quocbao-quan", Text: "Điều 2 đúng với cả công ty phần mềm.", AgeMin: 180},
	{User: "tuankiet.dang", Post: "nhatminh-sapa", Text: "Tháng 9 năm sau mình đi, cảm ơn bài này.", AgeMin: 600},
	{User: "camtu.trinh", Post: "haiyen-ielts", Text: "Cố lên em, 7.0 là trong tầm tay.", AgeMin: 25},
	{User: "duclong.phan", Post: "thanhtung-chay-bo", AgeMin: 300},
	{User: "minhanh.nguyen", Post: "giahuy-khoi-nghiep", Text: "Chúc mừng 3 năm, đọc bài nào cũng thấy thật.", AgeMin: 70},
	{User: "haiyen.vu", Post: "thutrang-sach", Text: "Ba cuốn này em mới đọc được một.", AgeMin: 900},
}

// Notif là một thông báo seed. Content là phần đứng sau tên người gửi
// (UI hiển thị "Tên người gửi" + content).
type Notif struct {
	Receiver string
	Sender   string
	Type     string
	Content  string
	PostKey  string
	UserKey  string
	AgeMin   int
	IsRead   bool
}

// Call là một cuộc gọi seed.
type Call struct {
	Caller    string
	Callee    string
	Type      string // voice | video
	Status    string // ended | missed | rejected | cancelled
	AgeMin    int
	Duration  int
	VideoOn   bool
	MutedOne  bool
}

// Follows là đồ thị theo dõi seed.
var Follows = []FollowEdge{
	{"minhanh.nguyen", "thutrang_le"}, {"minhanh.nguyen", "hoangnam.dev"},
	{"minhanh.nguyen", "nhatminh.ho"}, {"minhanh.nguyen", "camtu.trinh"},
	{"minhanh.nguyen", "quocbao.pham"}, {"minhanh.nguyen", "bichha_to"},
	{"thutrang_le", "minhanh.nguyen"}, {"thutrang_le", "haiyen.vu"},
	{"thutrang_le", "nhatminh.ho"},
	{"haiyen.vu", "thutrang_le"}, {"haiyen.vu", "minhanh.nguyen"},
	{"haiyen.vu", "khanhlinh.do"}, {"haiyen.vu", "camtu.trinh"},
	{"hoangnam.dev", "giahuy.ly"}, {"hoangnam.dev", "minhanh.nguyen"},
	{"hoangnam.dev", "thanhtung.ngo"},
	{"giahuy.ly", "hoangnam.dev"}, {"giahuy.ly", "quocbao.pham"},
	{"giahuy.ly", "nhatminh.ho"},
	{"nhatminh.ho", "thutrang_le"}, {"nhatminh.ho", "quocbao.pham"},
	{"nhatminh.ho", "tuankiet.dang"},
	{"camtu.trinh", "minhanh.nguyen"}, {"camtu.trinh", "haiyen.vu"},
	{"camtu.trinh", "ngochan_bui"},
	{"myduyen.duong", "ngochan_bui"}, {"myduyen.duong", "thutrang_le"},
	{"myduyen.duong", "minhanh.nguyen"}, {"myduyen.duong", "haiyen.vu"},
	{"ngochan_bui", "myduyen.duong"}, {"ngochan_bui", "duclong.phan"},
	{"ngochan_bui", "thutrang_le"},
	{"thanhtung.ngo", "phuongthao.mai"}, {"thanhtung.ngo", "hoangnam.dev"},
	{"phuongthao.mai", "thanhtung.ngo"}, {"phuongthao.mai", "minhanh.nguyen"},
	{"phuongthao.mai", "camtu.trinh"},
	{"tuankiet.dang", "nhatminh.ho"}, {"tuankiet.dang", "giahuy.ly"},
	{"bichha_to", "thutrang_le"}, {"bichha_to", "minhanh.nguyen"},
	{"bichha_to", "nhatminh.ho"},
	{"khanhlinh.do", "haiyen.vu"}, {"khanhlinh.do", "minhanh.nguyen"},
	{"khanhlinh.do", "thutrang_le"},
	{"duclong.phan", "ngochan_bui"}, {"duclong.phan", "myduyen.duong"},
	{"quocbao.pham", "nhatminh.ho"}, {"quocbao.pham", "minhanh.nguyen"},
	{"quocthinh.lam", "duclong.phan"},
}

// Friends là các quan hệ bạn bè seed (một dòng mỗi cặp, status accepted).
var Friends = []FriendEdge{
	{Sender: "minhanh.nguyen", Receiver: "thutrang_le", Status: "accepted", AgeDays: 220},
	{Sender: "minhanh.nguyen", Receiver: "camtu.trinh", Status: "accepted", AgeDays: 180},
	{Sender: "minhanh.nguyen", Receiver: "haiyen.vu", Status: "accepted", AgeDays: 150},
	{Sender: "thutrang_le", Receiver: "haiyen.vu", Status: "accepted", AgeDays: 300},
	{Sender: "thutrang_le", Receiver: "bichha_to", Status: "accepted", AgeDays: 120},
	{Sender: "hoangnam.dev", Receiver: "giahuy.ly", Status: "accepted", AgeDays: 260},
	{Sender: "hoangnam.dev", Receiver: "thanhtung.ngo", Status: "accepted", AgeDays: 90},
	{Sender: "ngochan_bui", Receiver: "myduyen.duong", Status: "accepted", AgeDays: 340},
	{Sender: "ngochan_bui", Receiver: "duclong.phan", Status: "accepted", AgeDays: 400},
	{Sender: "thanhtung.ngo", Receiver: "phuongthao.mai", Status: "accepted", AgeDays: 210},
	{Sender: "nhatminh.ho", Receiver: "quocbao.pham", Status: "accepted", AgeDays: 170},
	{Sender: "tuankiet.dang", Receiver: "giahuy.ly", Status: "accepted", AgeDays: 80},
	{Sender: "khanhlinh.do", Receiver: "haiyen.vu", Status: "accepted", AgeDays: 190},
	{Sender: "myduyen.duong", Receiver: "haiyen.vu", Status: "accepted", AgeDays: 60},
	// Lời mời đang chờ: xuất hiện ở trang Bạn bè và tạo thông báo chưa đọc.
	{Sender: "phuongthao.mai", Receiver: "minhanh.nguyen", Status: "pending", AgeDays: 2},
	{Sender: "myduyen.duong", Receiver: "minhanh.nguyen", Status: "pending", AgeDays: 1},
	{Sender: "quocthinh.lam", Receiver: "duclong.phan", Status: "pending", AgeDays: 4},
}

// Blocks là các lượt chặn seed.
var Blocks = []BlockEdge{
	{User: "minhanh.nguyen", Blocked: "anhdung.chu", AgeDays: 5},
	{User: "thutrang_le", Blocked: "anhdung.chu", AgeDays: 6},
	{User: "duclong.phan", Blocked: "quocthinh.lam", AgeDays: 3},
}

// Bookmarks là các bài viết được lưu seed.
var Bookmarks = []BookmarkEdge{
	{User: "minhanh.nguyen", Post: "thutrang-dalat", AgeDays: 1},
	{User: "minhanh.nguyen", Post: "giahuy-khoi-nghiep", AgeDays: 2},
	{User: "minhanh.nguyen", Post: "hoangnam-laptop", AgeDays: 4},
	{User: "haiyen.vu", Post: "thutrang-dalat", AgeDays: 1},
	{User: "haiyen.vu", Post: "hoangnam-laptop", AgeDays: 3},
	{User: "myduyen.duong", Post: "ngochan-bun-cha", AgeDays: 2},
	{User: "myduyen.duong", Post: "ngochan-com-tam", AgeDays: 5},
	{User: "khanhlinh.do", Post: "minhanh-ot", AgeDays: 1},
	{User: "tuankiet.dang", Post: "nhatminh-sapa", AgeDays: 6},
	{User: "camtu.trinh", Post: "minhanh-phoco-album", AgeDays: 3},
	{User: "phuongthao.mai", Post: "thanhtung-chay-bo", AgeDays: 7},
}

// Notifications là thông báo seed. 6 thông báo đầu của tài khoản demo
// (minhanh.nguyen) để chưa đọc nhằm thấy badge trên navbar.
var Notifications = []Notif{
	{Receiver: "minhanh.nguyen", Sender: "thutrang_le", Type: "follow", Content: "đã bắt đầu theo dõi bạn", UserKey: "thutrang_le", AgeMin: 12, IsRead: false},
	{Receiver: "minhanh.nguyen", Sender: "phuongthao.mai", Type: "friend_request", Content: "đã gửi lời mời kết bạn", UserKey: "phuongthao.mai", AgeMin: 40, IsRead: false},
	{Receiver: "minhanh.nguyen", Sender: "myduyen.duong", Type: "friend_request", Content: "đã gửi lời mời kết bạn", UserKey: "myduyen.duong", AgeMin: 90, IsRead: false},
	{Receiver: "minhanh.nguyen", Sender: "haiyen.vu", Type: "like", Content: "đã thích bài viết của bạn", PostKey: "minhanh-ot", AgeMin: 120, IsRead: false},
	{Receiver: "minhanh.nguyen", Sender: "thutrang_le", Type: "comment", Content: "đã bình luận về bài viết của bạn", PostKey: "minhanh-ot", AgeMin: 150, IsRead: false},
	{Receiver: "minhanh.nguyen", Sender: "phuongthao.mai", Type: "share", Content: "đã chia sẻ bài viết của bạn", PostKey: "minhanh-phoco-album", AgeMin: 200, IsRead: false},
	{Receiver: "minhanh.nguyen", Sender: "hoangnam.dev", Type: "message", Content: "đã gửi cho bạn một tin nhắn", AgeMin: 240, IsRead: true},
	{Receiver: "minhanh.nguyen", Sender: "khanhlinh.do", Type: "story_react", Content: "đã bày tỏ cảm xúc về tin của bạn", AgeMin: 300, IsRead: true},
	{Receiver: "minhanh.nguyen", Sender: "bichha_to", Type: "like", Content: "đã thích bài viết của bạn", PostKey: "minhanh-phoco-album", AgeMin: 420, IsRead: true},

	{Receiver: "thutrang_le", Sender: "minhanh.nguyen", Type: "follow", Content: "đã bắt đầu theo dõi bạn", UserKey: "minhanh.nguyen", AgeMin: 60, IsRead: false},
	{Receiver: "thutrang_le", Sender: "haiyen.vu", Type: "comment", Content: "đã bình luận về bài viết của bạn", PostKey: "thutrang-dalat", AgeMin: 180, IsRead: false},
	{Receiver: "thutrang_le", Sender: "nhatminh.ho", Type: "like", Content: "đã thích bài viết của bạn", PostKey: "thutrang-dalat", AgeMin: 210, IsRead: true},

	{Receiver: "haiyen.vu", Sender: "thutrang_le", Type: "comment", Content: "đã bình luận về bài viết của bạn", PostKey: "haiyen-ielts", AgeMin: 45, IsRead: false},
	{Receiver: "haiyen.vu", Sender: "myduyen.duong", Type: "follow", Content: "đã bắt đầu theo dõi bạn", UserKey: "myduyen.duong", AgeMin: 260, IsRead: true},

	{Receiver: "ngochan_bui", Sender: "myduyen.duong", Type: "comment", Content: "đã bình luận về bài viết của bạn", PostKey: "ngochan-bun-cha", AgeMin: 70, IsRead: false},
	{Receiver: "ngochan_bui", Sender: "thutrang_le", Type: "like", Content: "đã thích bài viết của bạn", PostKey: "ngochan-bun-cha", AgeMin: 150, IsRead: true},

	{Receiver: "hoangnam.dev", Sender: "giahuy.ly", Type: "comment", Content: "đã bình luận về bài viết của bạn", PostKey: "hoangnam-go", AgeMin: 130, IsRead: false},
	{Receiver: "hoangnam.dev", Sender: "minhanh.nguyen", Type: "like", Content: "đã thích bài viết của bạn", PostKey: "hoangnam-go", AgeMin: 400, IsRead: true},

	{Receiver: "quocbao.pham", Sender: "nhatminh.ho", Type: "comment", Content: "đã bình luận về bài viết của bạn", PostKey: "quocbao-quan", AgeMin: 200, IsRead: false},
	{Receiver: "quocbao.pham", Sender: "giahuy.ly", Type: "share", Content: "đã chia sẻ bài viết của bạn", PostKey: "quocbao-quan", AgeMin: 500, IsRead: true},

	{Receiver: "thanhtung.ngo", Sender: "phuongthao.mai", Type: "message", Content: "đã gửi cho bạn một tin nhắn", AgeMin: 100, IsRead: false},
	{Receiver: "thanhtung.ngo", Sender: "hoangnam.dev", Type: "like", Content: "đã thích bài viết của bạn", PostKey: "thanhtung-chay-bo", AgeMin: 300, IsRead: true},

	{Receiver: "bichha_to", Sender: "minhanh.nguyen", Type: "comment", Content: "đã bình luận về bài viết của bạn", PostKey: "bichha-truc-dem", AgeMin: 90, IsRead: false},
	{Receiver: "minhanh.nguyen", Sender: "admin", Type: "media_approved", Content: "đã duyệt ảnh bạn tải lên", AgeMin: 600, IsRead: true},
	{Receiver: "duclong.phan", Sender: "ngochan_bui", Type: "like", Content: "đã thích bài viết của bạn", PostKey: "ngochan-com-tam", AgeMin: 280, IsRead: false},
	{Receiver: "quocthinh.lam", Sender: "admin", Type: "media_rejected", Content: "đã từ chối ảnh bạn tải lên vì chất lượng thấp", AgeMin: 3000, IsRead: false},
}

// Calls là lịch sử cuộc gọi seed giữa các tài khoản thường.
var Calls = []Call{
	{Caller: "minhanh.nguyen", Callee: "thutrang_le", Type: "voice", Status: "ended", AgeMin: 180, Duration: 742},
	{Caller: "thutrang_le", Callee: "minhanh.nguyen", Type: "video", Status: "ended", AgeMin: 1200, Duration: 1435, VideoOn: true},
	{Caller: "minhanh.nguyen", Callee: "hoangnam.dev", Type: "voice", Status: "missed", AgeMin: 640, Duration: 0},
	{Caller: "hoangnam.dev", Callee: "minhanh.nguyen", Type: "voice", Status: "ended", AgeMin: 900, Duration: 320, MutedOne: true},
	{Caller: "ngochan_bui", Callee: "myduyen.duong", Type: "video", Status: "ended", AgeMin: 420, Duration: 980, VideoOn: true},
	{Caller: "myduyen.duong", Callee: "ngochan_bui", Type: "voice", Status: "rejected", AgeMin: 1500, Duration: 0},
	{Caller: "thanhtung.ngo", Callee: "phuongthao.mai", Type: "voice", Status: "ended", AgeMin: 300, Duration: 215},
	{Caller: "phuongthao.mai", Callee: "thanhtung.ngo", Type: "video", Status: "cancelled", AgeMin: 2600, Duration: 0},
	{Caller: "quocbao.pham", Callee: "nhatminh.ho", Type: "voice", Status: "ended", AgeMin: 800, Duration: 465},
	{Caller: "khanhlinh.do", Callee: "haiyen.vu", Type: "video", Status: "ended", AgeMin: 220, Duration: 1820, VideoOn: true},
	{Caller: "haiyen.vu", Callee: "thutrang_le", Type: "voice", Status: "missed", AgeMin: 3600, Duration: 0},
	{Caller: "giahuy.ly", Callee: "hoangnam.dev", Type: "voice", Status: "ended", AgeMin: 1900, Duration: 1260},
	{Caller: "tuankiet.dang", Callee: "giahuy.ly", Type: "voice", Status: "rejected", AgeMin: 2400, Duration: 0},
	{Caller: "bichha_to", Callee: "minhanh.nguyen", Type: "voice", Status: "ended", AgeMin: 2000, Duration: 540},
}
