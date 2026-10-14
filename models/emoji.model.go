package models

type Emoji struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	ImageURI  string `json:"image_uri"`
	// Render native: ký tự unicode thật (vd "😀"). Client mới dùng field này,
	// ImageURI giữ lại cho nội dung/reaction cũ.
	Character  string `json:"character"`
	Name       string `json:"name"`
	Keywords   string `json:"keywords"`
	Category   string `json:"category"`
	SortOrder  int    `json:"sort_order"`
	IsReaction bool   `json:"is_reaction"`
}

func NewEmoji(code, imageURI string) Emoji {
	return Emoji{Code: code, ImageURI: imageURI}
}
