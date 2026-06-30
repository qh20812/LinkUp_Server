package services

import (
	"context"
	"errors"
	"fmt"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
	"mime/multipart"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"gorm.io/gorm"
)

type MediaService interface {
	UploadMedia(ctx context.Context, userID string, file *multipart.FileHeader) (*models.Media, error)
	DeleteMedia(ctx context.Context, userID string, mediaID string) error
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
		PublicID:     utils.GenerateUUID(),
		ResourceType: "auto",
	})
	if err != nil {
		return nil, fmt.Errorf("upload to cloudinary: %w", err)
	}

	media := models.NewMedia(userID, nil, uploadResult.SecureURL, file.Header.Get("Content-Type"), float64(file.Size))
	media.ID = utils.GenerateUUID()
	media.CreatedAt = time.Now()

	if err := s.repo.Create(ctx, &media); err != nil {
		return nil, fmt.Errorf("save media record: %w", err)
	}

	if err := s.repo.UpdateStorageUsage(ctx, userID, float64(file.Size)); err != nil {
		return nil, fmt.Errorf("update storage usage: %w", err)
	}

	return &media, nil
}

func (s *mediaService) DeleteMedia(ctx context.Context, userID string, mediaID string) error {
	media, err := s.repo.GetByID(ctx, mediaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return validations.ErrMediaNotFound
		}
		return fmt.Errorf("get media by id: %w", err)
	}

	if media.UserID != userID {
		return validations.ErrMediaForbidden
	}

	if s.cloudinary != nil {
		publicID, resourceType, parseErr := parseCloudinaryPublicID(media.FileURI)
		if parseErr == nil {
			if _, err = s.cloudinary.Upload.Destroy(ctx, uploader.DestroyParams{
				PublicID:     publicID,
				ResourceType: resourceType,
			}); err != nil {
				return fmt.Errorf("delete from cloudinary: %w", err)
			}
		}
	}

	if err := s.repo.DeleteWithStorageAdjustment(ctx, userID, media); err != nil {
		return fmt.Errorf("delete media record: %w", err)
	}

	return nil
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

func parseCloudinaryPublicID(fileURI string) (string, string, error) {
	parsedURL, err := url.Parse(fileURI)
	if err != nil {
		return "", "", err
	}
	if parsedURL.Path == "" {
		return "", "", fmt.Errorf("invalid cloudinary path")
	}

	segments := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	idx := -1
	for i, seg := range segments {
		if seg == "upload" {
			idx = i
			break
		}
	}

	if idx <= 0 || idx+1 >= len(segments) {
		return "", "", fmt.Errorf("invalid cloudinary path")
	}

	resourceType := segments[idx-1]
	partsAfterUpload := segments[idx+1:]

	if len(partsAfterUpload) == 0 {
		return "", "", fmt.Errorf("missing public id")
	}

	if strings.HasPrefix(partsAfterUpload[0], "v") || strings.HasPrefix(partsAfterUpload[0], "f") || strings.HasPrefix(partsAfterUpload[0], "c") {
		partsAfterUpload = partsAfterUpload[1:]
	}

	if len(partsAfterUpload) == 0 {
		return "", "", fmt.Errorf("missing public id")
	}

	publicID := strings.Join(partsAfterUpload, "/")
	if ext := path.Ext(publicID); ext != "" {
		publicID = strings.TrimSuffix(publicID, ext)
	}

	return publicID, resourceType, nil
}
