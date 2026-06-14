package models

type Emoji struct {
	ID       int64  `json:"id" db:"id"`
	Code     string `json:"code" db:"code"`
	ImageURI string `json:"image_uri" db:"image_uri"`
}

func NewEmoji(code, imageURI string) Emoji {
	return Emoji{Code: code, ImageURI: imageURI}
}
