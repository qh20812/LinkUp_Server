package seeddata

// ViolationRule là một quy tắc vi phạm dùng cho form báo cáo và trang quản trị.
type ViolationRule struct {
	Title        string
	Description  string
	ApplicableTo string // all | post | user | comment
	Severity     string // low | medium | high
	SortOrder    int
}

// ViolationRules là nguồn dữ liệu duy nhất cho quy tắc vi phạm.
var ViolationRules = []ViolationRule{
	{Title: "Spam / quảng cáo làm phiền", Description: "Đăng nội dung quảng cáo lặp lại, link rác hoặc nội dung không liên quan.", ApplicableTo: "all", Severity: "low", SortOrder: 1},
	{Title: "Quấy rối / bắt nạt", Description: "Tấn công, đe doạ hoặc xúc phạm người khác.", ApplicableTo: "all", Severity: "high", SortOrder: 2},
	{Title: "Ngôn từ thù ghét / phân biệt đối xử", Description: "Nội dung kỳ thị chủng tộc, giới tính, tôn giáo hoặc nhóm người.", ApplicableTo: "all", Severity: "high", SortOrder: 3},
	{Title: "Nội dung nhạy cảm / đồi trụy", Description: "Hình ảnh, video hoặc mô tả mang tính khiêu dâm, bạo lực quá mức.", ApplicableTo: "all", Severity: "high", SortOrder: 4},
	{Title: "Thông tin sai lệch", Description: "Tin giả, thông tin sai sự thật gây hiểu lầm.", ApplicableTo: "post", Severity: "medium", SortOrder: 5},
	{Title: "Lừa đảo / giả mạo", Description: "Tài khoản giả mạo người khác hoặc hành vi lừa đảo.", ApplicableTo: "user", Severity: "high", SortOrder: 6},
	{Title: "Vi phạm bản quyền", Description: "Đăng lại nội dung có bản quyền mà không được phép.", ApplicableTo: "post", Severity: "medium", SortOrder: 7},
	{Title: "Lý do khác", Description: "Vi phạm không thuộc các nhóm trên (mô tả chi tiết ở lý do).", ApplicableTo: "all", Severity: "low", SortOrder: 99},
}

// Report là một báo cáo vi phạm seed.
// Đúng một trong các Target* được đặt; RuleTitle trỏ tới ViolationRules.
type Report struct {
	Reporter      string
	Type          string // violation | other
	TargetUser    string
	TargetPost    string
	TargetCommunity string
	RuleTitle     string
	Detail        string
	Status        string // pending | reviewed | resolved | rejected
	AgeHours      float64
}

// Reports là nguồn dữ liệu duy nhất cho báo cáo vi phạm.
var Reports = []Report{
	{
		Reporter: "duclong.phan", Type: "violation", TargetPost: "anhdung-ban-hang",
		RuleTitle: "Spam / quảng cáo làm phiền",
		Detail:    "Bài này đăng đi đăng lại 4 lần trong 2 ngày, nội dung y hệt nhau và để lại số điện thoại bán hàng.",
		Status:    "pending", AgeHours: 3,
	},
	{
		Reporter: "minhanh.nguyen", Type: "violation", TargetUser: "anhdung.chu",
		RuleTitle: "Lừa đảo / giả mạo",
		Detail:    "Tài khoản này nhắn tin mời mua hàng, yêu cầu chuyển khoản trước rồi không giao. Mình có ảnh chụp tin nhắn.",
		Status:    "pending", AgeHours: 8,
	},
	{
		Reporter: "thutrang_le", Type: "violation", TargetPost: "anhdung-hidden",
		RuleTitle: "Spam / quảng cáo làm phiền",
		Detail:    "Bán tài khoản game, nội dung không liên quan tới chủ đề trang.",
		Status:    "resolved", AgeHours: 30,
	},
	{
		Reporter: "bichha_to", Type: "other", TargetCommunity: "an-gi-hom-nay",
		Detail:    "Có người bình luận chê ngoại hình người đăng ảnh món ăn. Mình không rõ nên chọn quy tắc nào.",
		Status:    "reviewed", AgeHours: 50,
	},
	{
		Reporter: "haiyen.vu", Type: "violation", TargetPost: "quocthinh-cafe",
		RuleTitle: "Thông tin sai lệch",
		Detail:    "Bài nói giá cà phê giảm 40% nhưng không dẫn nguồn, có thể gây hiểu lầm cho người mua.",
		Status:    "rejected", AgeHours: 120,
	},
	{
		Reporter: "khanhlinh.do", Type: "violation", TargetUser: "quocthinh.lam",
		RuleTitle: "Quấy rối / bắt nạt",
		Detail:    "Nhắn tin riêng nhiều lần sau khi mình đã nói không muốn tiếp tục trò chuyện.",
		Status:    "resolved", AgeHours: 160,
	},
	{
		Reporter: "myduyen.duong", Type: "violation", TargetPost: "minhanh-bigbang",
		RuleTitle: "Vi phạm bản quyền",
		Detail:    "Ảnh concert có thể lấy từ trang khác mà không ghi nguồn.",
		Status:    "rejected", AgeHours: 200,
	},
	{
		Reporter: "giahuy.ly", Type: "other", TargetPost: "hoangnam-go",
		Detail:    "Không phải vi phạm, chỉ muốn báo bài có số liệu build time có thể gây tranh cãi.",
		Status:    "rejected", AgeHours: 260,
	},
}

// Ban là một án khoá tài khoản seed.
type Ban struct {
	User        string
	Admin       string
	Reason      string
	AgeDays     int
	DurationDays int // 0 = vĩnh viễn
}

// Bans là nguồn dữ liệu duy nhất cho án khoá.
var Bans = []Ban{
	{User: "anhdung.chu", Admin: "admin", Reason: "Spam quảng cáo lặp lại nhiều lần sau khi đã bị nhắc nhở. Tạm khoá 30 ngày.", AgeDays: 6, DurationDays: 30},
	{User: "quocthinh.lam", Admin: "admin", Reason: "Quấy rối người dùng khác qua tin nhắn riêng, nhiều báo cáo trùng khớp. Khoá vĩnh viễn.", AgeDays: 12, DurationDays: 0},
}

// ModLog là một bản ghi hành động kiểm duyệt.
type ModLog struct {
	Moderator  string
	Action     string // hide_post | ban_user | resolve_report | reject_report | delete_comment
	TargetType string // post | user | community | comment
	TargetKey  string // post key hoặc user key; rỗng nếu target là báo cáo
	Reason     string
	AgeDays    int
}

// ModLogs là nguồn dữ liệu duy nhất cho nhật ký kiểm duyệt.
var ModLogs = []ModLog{
	{Moderator: "admin", Action: "hide_post", TargetType: "post", TargetKey: "anhdung-hidden", Reason: "Bài bán tài khoản game, vi phạm quy tắc spam.", AgeDays: 27},
	{Moderator: "admin", Action: "resolve_report", TargetType: "post", TargetKey: "anhdung-hidden", Reason: "Đã xử lý báo cáo và ẩn bài viết.", AgeDays: 26},
	{Moderator: "admin", Action: "ban_user", TargetType: "user", TargetKey: "anhdung.chu", Reason: "Spam quảng cáo lặp lại, tạm khoá 30 ngày.", AgeDays: 6},
	{Moderator: "superadmin", Action: "ban_user", TargetType: "user", TargetKey: "quocthinh.lam", Reason: "Quấy rối người dùng qua tin nhắn, khoá vĩnh viễn.", AgeDays: 12},
	{Moderator: "admin", Action: "resolve_report", TargetType: "user", TargetKey: "quocthinh.lam", Reason: "Báo cáo quấy rối được xác nhận, tài khoản đã bị khoá.", AgeDays: 11},
	{Moderator: "admin", Action: "reject_report", TargetType: "post", TargetKey: "minhanh-bigbang", Reason: "Không đủ căn cứ về bản quyền, ảnh do người dùng tự chụp.", AgeDays: 8},
}
