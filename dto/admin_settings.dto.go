package dto

type AdminSettingsResponse struct {
	Settings map[string]string `json:"settings"`
}

type AdminSettingsUpdateInput struct {
	Settings map[string]string `json:"settings" binding:"required"`
}