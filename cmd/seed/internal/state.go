package internal

type SeedState struct {
	UserIDs         []string
	RoleIDs         []string
	EmojiIDs         []string
	// ReactionEmojiIDs là tập con của EmojiIDs (10 reaction legacy, đứng đầu) —
	// social/messaging chỉ random reaction từ đây, không random cả 3.6k emoji.
	ReactionEmojiIDs []string
	ViolationRuleIDs []string
	PostIDs         []string
	CommentIDs      []string
	CommunityIDs    []string
	ChatIDs          []string
	CommunityChatIDs []string
	MediaIDs         []string
}
