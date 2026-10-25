package internal

// SeedState chia sẻ dữ liệu giữa các bước seed.
//
// Các slice *IDs giữ thứ tự chèn; các map *ByKey/*ByUsername dùng để tra cứu
// ổn định từ seeddata (vốn tham chiếu bằng key, không bằng index).
type SeedState struct {
	UserIDs  []string
	RoleIDs  []string
	EmojiIDs []string
	// ReactionEmojiIDs là tập con của EmojiIDs (10 reaction legacy, đứng đầu) —
	// social/messaging chỉ random reaction từ đây, không random cả 3.6k emoji.
	ReactionEmojiIDs []string
	ViolationRuleIDs []string
	PostIDs          []string
	CommentIDs       []string
	CommunityIDs     []string
	ChatIDs          []string
	CommunityChatIDs []string
	MediaIDs         []string
	StoryIDs         []string
	AdIDs            []string

	UserIDByUsername       map[string]string
	EmojiIDByCode          map[string]string
	ViolationRuleIDByTitle map[string]string
	PostIDByKey            map[string]string
	CommentIDByKey         map[string]string
	CommunityIDByKey       map[string]string
	ChatIDByKey            map[string]string
	StoryIDByKey           map[string]string
	AdIDByKey              map[string]string
	AdPackageIDByKey       map[string]string
	MessageIDs             []string

	DemoUserID    string
	PartnerUserID string
}

// Init cấp phát các map tra cứu. Gọi một lần trước khi chạy các bước.
func (s *SeedState) Init() {
	s.UserIDByUsername = make(map[string]string)
	s.EmojiIDByCode = make(map[string]string)
	s.ViolationRuleIDByTitle = make(map[string]string)
	s.PostIDByKey = make(map[string]string)
	s.CommentIDByKey = make(map[string]string)
	s.CommunityIDByKey = make(map[string]string)
	s.ChatIDByKey = make(map[string]string)
	s.StoryIDByKey = make(map[string]string)
	s.AdIDByKey = make(map[string]string)
	s.AdPackageIDByKey = make(map[string]string)
}
