package services

import (
	"context"
	"errors"
	"fmt"
	"linkup/models"
	"linkup/utils"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// ModerationLabel đại diện cho một nhãn từ AWS Rekognition moderation.
type ModerationLabel struct {
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
	ParentName string  `json:"parent_name"`
}

// AIModerationResult chứa kết quả moderation, URL file đã upload và danh sách nhãn.
type AIModerationResult struct {
	Status    models.MediaStatus
	SecureURL string
	PublicID  string
	Labels    []ModerationLabel
}

// AIModerationService xử lý kiểm duyệt nội dung media qua AI (Cloudinary + AWS Rekognition).
// Moderate thực hiện upload + moderation trong 1 call duy nhất, trả về URL và kết quả.
type AIModerationService interface {
	Moderate(ctx context.Context, file multipart.File, fileType string) (*AIModerationResult, error)
}

type cloudinaryModerationService struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryModerationService(cld *cloudinary.Cloudinary) AIModerationService {
	return &cloudinaryModerationService{cld: cld}
}

func (s *cloudinaryModerationService) Moderate(
	ctx context.Context, src multipart.File, fileType string,
) (*AIModerationResult, error) {
	if s.cld == nil {
		return nil, errors.New("cloudinary chưa được khởi tạo")
	}

	uploadResult, err := s.cld.Upload.Upload(ctx, src, uploader.UploadParams{
		PublicID:     utils.GenerateUUID(),
		ResourceType: "auto",
		Moderation:   "aws_rek",
	})
	if err != nil {
		return nil, fmt.Errorf("tải lên + kiểm duyệt thất bại: %w", err)
	}

	// Trường hợp moderation không được cấu hình (thiếu quyền AWS Rekognition)
	if len(uploadResult.Moderation) == 0 {
		return &AIModerationResult{
			Status:    models.MediaStatusFlagged,
			SecureURL: uploadResult.SecureURL,
			PublicID:  uploadResult.PublicID,
		}, nil
	}

	mod := uploadResult.Moderation[0]

	result := &AIModerationResult{
		SecureURL: uploadResult.SecureURL,
		PublicID:  uploadResult.PublicID,
		Labels:    toModerationLabels(mod.Response.ModerationLabels),
	}

	switch mod.Status {
	case api.Approved:
		result.Status = models.MediaStatusApproved
	case api.Rejected:
		result.Status = models.MediaStatusRejected
	default:
		// "pending" hoặc giá trị lạ → gắn cờ chờ xử lý thủ công
		result.Status = models.MediaStatusFlagged
	}

	return result, nil
}

// toModerationLabels chuyển đổi Cloudinary moderation labels sang service model.
func toModerationLabels(labels []uploader.ModerationLabel) []ModerationLabel {
	if len(labels) == 0 {
		return nil
	}
	out := make([]ModerationLabel, len(labels))
	for i, l := range labels {
		out[i] = ModerationLabel{
			Name:       l.Name,
			Confidence: l.Confidence,
			ParentName: l.ParentName,
		}
	}
	return out
}
