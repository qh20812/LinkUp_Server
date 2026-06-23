package services

import (
	"context"
	"fmt"
	"linkup/models"
	"linkup/repository"
	"linkup/validations"
	"mime/multipart"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
)

type MediaService interface {
	UploadMedia(ctx context.Context, userID string, file *multipart.FileHeader) (*models.Media, error)
	GetUserStorageStatus(ctx context.Context, userID string) (quota, used, available float64, err error)
	GetUserMedia(ctx context.Context, userID string) ([]models.Media, error)
}

type mediaService struct {
	repo       repository.MediaRepository
	validation *validations.MediaValidation
	cloudinary *cloudinary.Cloudinary
}

func NewMediaService(
	repo repository.MediaRepository,
	cloudinaryURL string,
) MediaService {
	cld, _ := cloudinary.NewFromURL(cloudinaryURL)
	return &mediaService{
		repo:       repo,
		validation: validations.NewMediaValidation(),
		cloudinary: cld,
	}
}

func (s *mediaService) UploadMedia(ctx context.Context, userID string, file *multipart.FileHeader) (*models.Media, error) {
	quota, used, err := s.repo.GetUserStorageInfo(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get storage info: %w", err)
	}

	available := quota - used
	if available < 0 {
		available = 0
	}

	if err := s.validation.ValidateFile(file.Filename, file.Size, file.Header.Get("Content-Type")); err != nil {
		return nil, err
	}

	if err := s.validation.ValidateStorageQuota(available, file.Size); err != nil {
		return nil, validations.ErrInsufficientStorage
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer src.Close()

	uploadResult, err := s.cloudinary.Upload.Upload(ctx, src, uploader.UploadParams{
		Folder:       fmt.Sprintf("linkup/users/%s", userID),
		PublicID:     uuid.New().String(),
		ResourceType: "auto",
	})
	if err != nil {
		return nil, fmt.Errorf("upload to cloudinary: %w", err)
	}

	media := models.NewMedia(userID, nil, uploadResult.SecureURL, file.Header.Get("Content-Type"), float64(file.Size))
	media.ID = uuid.New().String()
	media.CreatedAt = time.Now()

	if err := s.repo.Create(ctx, &media); err != nil {
		return nil, fmt.Errorf("save media record: %w", err)
	}

	if err := s.repo.UpdateStorageUsage(ctx, userID, float64(file.Size)); err != nil {
		return nil, fmt.Errorf("update storage usage: %w", err)
	}

	return &media, nil
}

func (s *mediaService) GetUserStorageStatus(ctx context.Context, userID string) (quota, used, available float64, err error) {
	quota, used, err = s.repo.GetUserStorageInfo(ctx, userID)
	if err != nil {
		return 0, 0, 0, err
	}
	available = quota - used
	if available < 0 {
		available = 0
	}
	return
}

func (s *mediaService) GetUserMedia(ctx context.Context, userID string) ([]models.Media, error) {
	return s.repo.GetByUserID(ctx, userID)
}
