package services

import (
	"context"
	"linkup/models"
	errorsapp "linkup/errors"
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
type AIModerationService interface {
	// Moderate upload file + chạy moderation trong 1 call (blocking).
	Moderate(ctx context.Context, file multipart.File, fileType string) (*AIModerationResult, error)
	// UploadWithoutModeration upload file lên Cloudinary mà KHÔNG chạy moderation.
	// Dùng cho upload-first flow: upload nhanh, chạy nền sau.
	UploadWithoutModeration(ctx context.Context, file multipart.File, publicID string) (*uploader.UploadResult, error)
	// ModerateAsset chạy moderation trên asset đã upload (gọi Cloudinary Explicit API).
	ModerateAsset(ctx context.Context, publicID string, resourceType string) (*AIModerationResult, error)
}

type cloudinaryModerationService struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryModerationService(cld *cloudinary.Cloudinary) AIModerationService {
	return &cloudinaryModerationService{cld: cld}
}

// Moderate upload file + chạy AWS Rekognition moderation trong 1 call duy nhất.
func (s *cloudinaryModerationService) Moderate(
	ctx context.Context, src multipart.File, fileType string,
) (*AIModerationResult, error) {
	if s.cld == nil {
		return nil, errorsapp.New(errorsapp.ErrCodeModCloudinaryNotInit)
	}

	uploadResult, err := s.cld.Upload.Upload(ctx, src, uploader.UploadParams{
		PublicID:     utils.GenerateUUID(),
		ResourceType: "auto",
		Moderation:   "aws_rek",
	})
	if err != nil {
		return nil, errorsapp.Wrap(errorsapp.ErrCodeModUploadFailed, err)
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

// UploadWithoutModeration upload file lên Cloudinary mà KHÔNG chạy moderation.
// Trả về UploadResult để caller lấy SecureURL và lưu vào DB.
func (s *cloudinaryModerationService) UploadWithoutModeration(
	ctx context.Context, src multipart.File, publicID string,
) (*uploader.UploadResult, error) {
	if s.cld == nil {
		return nil, errorsapp.New(errorsapp.ErrCodeModCloudinaryNotInit)
	}

	uploadResult, err := s.cld.Upload.Upload(ctx, src, uploader.UploadParams{
		PublicID:     publicID,
		ResourceType: "auto",
	})
	if err != nil {
		return nil, errorsapp.Wrap(errorsapp.ErrCodeModUploadFailed, err)
	}

	return uploadResult, nil
}

// ModerateAsset chạy AWS Rekognition moderation trên asset đã upload.
// Gọi Cloudinary Explicit API — không cần upload lại file.
func (s *cloudinaryModerationService) ModerateAsset(
	ctx context.Context, publicID string, resourceType string,
) (*AIModerationResult, error) {
	if s.cld == nil {
		return nil, errorsapp.New(errorsapp.ErrCodeModCloudinaryNotInit)
	}

	explicitResult, err := s.cld.Upload.Explicit(ctx, uploader.UploadParams{
		PublicID:     publicID,
		ResourceType: resourceType,
		Moderation:   "aws_rek",
	})
	if err != nil {
		return nil, errorsapp.Wrap(errorsapp.ErrCodeModUploadFailed, err)
	}

	// Nếu Cloudinary không trả về moderation (chưa cấu hình Rekognition)
	if len(explicitResult.Moderation) == 0 {
		return &AIModerationResult{
			Status:    models.MediaStatusFlagged,
			SecureURL: explicitResult.SecureURL,
			PublicID:  explicitResult.PublicID,
		}, nil
	}

	mod := explicitResult.Moderation[0]

	result := &AIModerationResult{
		SecureURL: explicitResult.SecureURL,
		PublicID:  explicitResult.PublicID,
		Labels:    toModerationLabels(mod.Response.ModerationLabels),
	}

	switch mod.Status {
	case api.Approved:
		result.Status = models.MediaStatusApproved
	case api.Rejected:
		result.Status = models.MediaStatusRejected
	default:
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
