package models

type Emoji struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	ImageURI string `json:"image_uri"`
}

func NewEmoji(code, imageURI string) Emoji {
	return Emoji{Code: code, ImageURI: imageURI}
}
