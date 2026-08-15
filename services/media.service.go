package services

import (
	"context"
	"errors"
	"fmt"
	"linkup/models"
	errorsapp "linkup/errors"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
	"log"
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
	GetByPostIDs(ctx context.Context, postIDs []string) (map[string][]models.Media, error)
	ModerateByID(ctx context.Context, mediaID string) (*AIModerationResult, error)
}

type mediaService struct {
	repo                repository.MediaRepository
	validation          *validations.MediaValidation
	cloudinary          *cloudinary.Cloudinary
	aiModeration        AIModerationService
	moderationRepo      *repository.ModerationRepository
	notificationService *NotificationService
	storyRepo           repository.StoryRepository
}

func NewMediaService(
	repo repository.MediaRepository,
	cloudinaryURL string,
	aiModeration AIModerationService,
	notificationService *NotificationService,
) *mediaService {
	cld, _ := cloudinary.NewFromURL(cloudinaryURL)
	return &mediaService{
		repo:                repo,
		validation:          validations.NewMediaValidation(),
		cloudinary:          cld,
		aiModeration:        aiModeration,
		notificationService: notificationService,
	}
}

// SetModerationRepo gán moderation repository sau khi khởi tạo
// (tránh circular dependency nếu có).
func (s *mediaService) SetModerationRepo(repo *repository.ModerationRepository) {
	s.moderationRepo = repo
}

// SetStoryRepo gán story repository để ẩn story khi media bị reject.
func (s *mediaService) SetStoryRepo(repo repository.StoryRepository) {
	s.storyRepo = repo
}

func (s *mediaService) UploadMedia(ctx context.Context, userID string, file *multipart.FileHeader) (*models.Media, error) {
	// 1. Kiểm tra storage quota
	quota, used, err := s.repo.GetUserStorageInfo(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get storage info: %w", err)
	}
	available := quota - used
	if available < 0 {
		available = 0
	}

	// 2. Validate file type + size
	if err := s.validation.ValidateFile(file.Filename, file.Size, file.Header.Get("Content-Type")); err != nil {
		return nil, err
	}
	if err := s.validation.ValidateStorageQuota(available, file.Size); err != nil {
		return nil, err
	}

	// 3. Mở file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer src.Close()

	// 4. Upload lên Cloudinary KHÔNG chạy moderation (nhanh)
	publicID := utils.GenerateUUID()
	uploadResult, err := s.aiModeration.UploadWithoutModeration(ctx, src, publicID)
	if err != nil {
		return nil, fmt.Errorf("upload to cloudinary: %w", err)
	}

	// 5. Tạo media record — status = pending, chờ moderation chạy nền
	media := models.NewMedia(userID, nil, uploadResult.SecureURL, file.Header.Get("Content-Type"), float64(file.Size))
	media.ID = publicID
	media.CreatedAt = time.Now()
	media.Status = models.MediaStatusPending

	// 6. Lưu DB
	if err := s.repo.Create(ctx, &media); err != nil {
		return nil, fmt.Errorf("save media record: %w", err)
	}
	if err := s.repo.UpdateStorageUsage(ctx, userID, float64(file.Size)); err != nil {
		return nil, fmt.Errorf("update storage usage: %w", err)
	}

	// 7. Chạy moderation nền (không block response)
	go s.moderateInBackground(media.ID, uploadResult.SecureURL)

	return &media, nil
}

// moderateInBackground chạy AI moderation trên asset đã upload.
// Nếu reject → cập nhật status, gửi notification, ẩn story liên quan.
func (s *mediaService) moderateInBackground(mediaID, secureURL string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	publicID, resourceType, err := parseCloudinaryPublicID(secureURL)
	if err != nil {
		log.Printf("[Moderation] không thể parse public ID cho media %s: %v", mediaID, err)
		return
	}

	result, err := s.aiModeration.ModerateAsset(ctx, publicID, resourceType)
	if err != nil {
		log.Printf("[Moderation] background moderation thất bại cho media %s: %v", mediaID, err)
		return
	}

	// Nếu approved → không cần làm gì thêm (media đã ở trạng thái visible từ khi upload)
	if result.Status == models.MediaStatusApproved {
		s.writeModerationLog(ctx, mediaID, result)
		return
	}

	// Cập nhật status trong DB
	if err := s.repo.UpdateStatus(ctx, mediaID, result.Status); err != nil {
		log.Printf("[Moderation] không thể cập nhật status cho media %s: %v", mediaID, err)
		return
	}

	// Lấy media info để biết user_id và post_id
	media, err := s.repo.GetByID(ctx, mediaID)
	if err != nil {
		log.Printf("[Moderation] không thể lấy media %s: %v", mediaID, err)
		return
	}

	// Gửi notification cho user
	s.sendRejectionNotification(ctx, media.UserID, media.PostID, result)

	// Ghi moderation log
	s.writeModerationLog(ctx, mediaID, result)

	// Nếu rejected → ẩn story liên quan (nếu có)
	if result.Status == models.MediaStatusRejected && s.storyRepo != nil {
		s.hideStoriesByMediaURI(ctx, secureURL)
	}
}

// hideStoriesByMediaURI xóa các story có media_uri trùng với URL bị reject.
func (s *mediaService) hideStoriesByMediaURI(ctx context.Context, mediaURI string) {
	if s.storyRepo == nil {
		return
	}
	// Không có method xóa story theo mediaURI trong repo hiện tại.
	// Story sẽ tự hết hạn sau 24h. Nếu cần ẩn ngay, thêm method vào StoryRepository.
	log.Printf("[Moderation] story có media %s cần bị ẩn (chờ hết hạn 24h hoặc thêm story repo method)", mediaURI)
}

func (s *mediaService) DeleteMedia(ctx context.Context, userID string, mediaID string) error {
	media, err := s.repo.GetByID(ctx, mediaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorsapp.New(errorsapp.ErrCodeMediaNotFound)
		}
		return fmt.Errorf("get media by id: %w", err)
	}

	if media.UserID != userID {
		return errorsapp.New(errorsapp.ErrCodeMediaForbidden)
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

func (s *mediaService) GetByPostIDs(ctx context.Context, postIDs []string) (map[string][]models.Media, error) {
	return s.repo.GetByPostIDs(ctx, postIDs)
}

// ModerateByID chạy AI moderation cho media đã tồn tại (dùng khi cần re-moderate).
func (s *mediaService) ModerateByID(ctx context.Context, mediaID string) (*AIModerationResult, error) {
	media, err := s.repo.GetByID(ctx, mediaID)
	if err != nil {
		return nil, fmt.Errorf("get media: %w", err)
	}

	publicID, resourceType, err := parseCloudinaryPublicID(media.FileURI)
	if err != nil {
		return nil, fmt.Errorf("parse cloudinary url: %w", err)
	}

	result, err := s.aiModeration.ModerateAsset(ctx, publicID, resourceType)
	if err != nil {
		return nil, fmt.Errorf("moderate asset: %w", err)
	}

	if err := s.repo.UpdateStatus(ctx, mediaID, result.Status); err != nil {
		log.Printf("[Moderation] không thể cập nhật status cho media %s: %v", mediaID, err)
	}

	return result, nil
}

// sendRejectionNotification gửi thông báo khi media bị reject.
// Không gửi cho approved (silent success) vì user không cần biết.
func (s *mediaService) sendRejectionNotification(ctx context.Context, userID string, postID *string, result *AIModerationResult) {
	if s.notificationService == nil {
		return
	}

	var message string
	var notifType models.NotificationType
	switch result.Status {
	case models.MediaStatusRejected:
		message = "Ảnh/video của bạn bị từ chối do vi phạm tiêu chuẩn cộng đồng."
		notifType = models.NotificationTypeMediaRejected
	case models.MediaStatusFlagged:
		message = "Ảnh/video của bạn đang chờ admin kiểm duyệt."
		notifType = models.NotificationTypeMediaFlagged
	default:
		return
	}

	if _, err := s.notificationService.Create(ctx, userID, nil, notifType,
		message, postID, nil, nil); err != nil {
		log.Printf("[Media] không thể gửi thông báo moderation cho user %s: %v", userID, err)
	}
}

// writeModerationLog ghi log kiểm duyệt vào bảng moderation_logs.
func (s *mediaService) writeModerationLog(ctx context.Context, mediaID string, result *AIModerationResult) {
	if s.moderationRepo == nil {
		return
	}

	reason := buildModerationReason(result)

	logEntry := models.NewModerationLog("AI_SYSTEM", models.ModerationActionUpdate,
		models.ModerationTargetMedia, mediaID, reason)
	logEntry.ID = utils.GenerateUUID()
	logEntry.CreatedAt = time.Now()

	if err := s.moderationRepo.CreateLog(ctx, &logEntry); err != nil {
		log.Printf("[Media] không thể ghi moderation log cho %s: %v", mediaID, err)
	}
}

// buildModerationReason tạo chuỗi mô tả lý do từ kết quả AI moderation.
func buildModerationReason(result *AIModerationResult) string {
	if len(result.Labels) == 0 {
		return fmt.Sprintf("AI moderation: %s", string(result.Status))
	}
	parts := make([]string, len(result.Labels))
	for i, l := range result.Labels {
		parts[i] = fmt.Sprintf("%s (%.1f%%)", l.Name, l.Confidence)
	}
	return fmt.Sprintf("AI moderation: %s — %s", string(result.Status), strings.Join(parts, ", "))
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
