package dto

type UploadMediaResponse struct {
	ID               string  `json:"id"`
	FileURI          string  `json:"file_uri"`
	FileType         string  `json:"file_type"`
	FileSize         float64 `json:"file_size"`
	Status           string  `json:"status"`
	AvailableStorage float64 `json:"available_storage_bytes"`
}

type StorageStatusResponse struct {
	StorageQuotaBytes float64 `json:"storage_quota_bytes"`
	StorageUsedBytes  float64 `json:"storage_used_bytes"`
	AvailableBytes    float64 `json:"available_bytes"`
	UsagePercentage   float64 `json:"usage_percentage"`
}
