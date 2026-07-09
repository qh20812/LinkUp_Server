package services

import (
	"errors"
	"fmt"
	"io"
	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type StoryService interface {
	CreateStory(userID string, fileHeader *multipart.FileHeader, caption string) (*dto.CreateStoryResponse, error)
	GetHomeStories() ([]dto.StoryResponse, error)
	ViewStory(storyID, viewerID string) (*models.Story, error)
	InteractWithStory(storyID, userID string, req dto.InteractStoryRequest) error
	GetAnalytics(storyID, userID string) (*dto.StoryAnalyticsResponse, error)
}

type storyService struct {
	repo repository.StoryRepository
}

func NewStoryService(repo repository.StoryRepository) StoryService {
	return &storyService{repo: repo}
}

func (s *storyService) CreateStory(userID string, fileHeader *multipart.FileHeader, caption string) (*dto.CreateStoryResponse, error) {
	// 1. Validate Định dạng & Dung lượng file
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	var mediaType models.StoryMediaType

	if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
		mediaType = models.StoryMediaTypeImage
	} else if ext == ".mp4" || ext == ".mov" {
		mediaType = models.StoryMediaTypeVideo
	} else {
		return nil, errors.New("định dạng file không hỗ trợ, chỉ nhận ảnh (jpg, png) hoặc video (mp4, mov)")
	}

	// Giới hạn dung lượng: Ảnh < 5MB, Video < 30MB
	if mediaType == models.StoryMediaTypeImage && fileHeader.Size > 5*1024*1024 {
		return nil, errors.New("dung lượng ảnh vượt quá 5MB")
	}
	if mediaType == models.StoryMediaTypeVideo && fileHeader.Size > 30*1024*1024 {
		return nil, errors.New("dung lượng video vượt quá 30MB")
	}

	// 2. Lưu file vật lý lên Server cục bộ (Thư mục ./uploads/stories)
	uploadDir := "./uploads/stories"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return nil, err
	}

	newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	finalPath := filepath.Join(uploadDir, newFileName)

	srcFile, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(finalPath)
	if err != nil {
		return nil, err
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return nil, err
	}

	// 3. Thiết lập thông tin Meta và thời gian hết hạn 24H
	mediaURI := "/static/stories/" + newFileName
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	story := models.NewStory(userID, mediaURI, mediaType, caption)
	story.ID = uuid.New().String()
	story.CreatedAt = now
	story.ExpiresAt = &expiresAt

	if err := s.repo.Create(&story); err != nil {
		return nil, err
	}

	return &dto.CreateStoryResponse{
		ID:        story.ID,
		MediaURI:  story.MediaURI,
		MediaType: story.MediaType.String(),
		Caption:   story.Caption,
		CreatedAt: story.CreatedAt,
		ExpiresAt: *story.ExpiresAt,
	}, nil
}

func (s *storyService) GetHomeStories() ([]dto.StoryResponse, error) {
	stories, err := s.repo.GetActiveStories()
	if err != nil {
		return nil, err
	}

	var res []dto.StoryResponse
	for _, story := range stories {
		res = append(res, dto.StoryResponse{
			ID:        story.ID,
			UserID:    story.UserID,
			MediaURI:  story.MediaURI,
			MediaType: story.MediaType.String(),
			Caption:   story.Caption,
			CreatedAt: story.CreatedAt,
		})
	}
	return res, nil
}

func (s *storyService) ViewStory(storyID, viewerID string) (*models.Story, error) {
	story, err := s.repo.FindByID(storyID)
	if err != nil {
		return nil, errors.New("không tìm thấy bản tin (story) này")
	}

	return story, nil
}

func (s *storyService) InteractWithStory(storyID string, userID string, req dto.InteractStoryRequest) error {
	_, err := s.repo.FindByID(storyID)
	if err != nil {
		return errors.New("không tìm thấy bản tin để tương tác")
	}

	// Khởi tạo đối tượng: Truyền thẳng cả storyID và userID dạng chuỗi string/UUID
	interact := &models.StoryInteract{
		StoryID: storyID,
		UserID:  userID,
		Type:    req.Type,
	}

	// Xử lý thông tin tùy biến theo phân loại loại hình tương tác
	if req.Type == "react" {
		if req.EmojiID == "" {
			return errors.New("emoji_id không được để trống khi thực hiện thả cảm xúc")
		}
		// Xác thực xem UUID của emoji hệ thống có tồn tại thực tế không
		exists, err := s.repo.CheckEmojiExists(req.EmojiID)
		if err != nil || !exists {
			return errors.New("mã hiệu ứng emoji không tồn tại trong hệ thống")
		}
		interact.EmojiID = &req.EmojiID
	} else {
		// Dành cho trường hợp reply văn bản chữ hoặc share bài viết
		if req.Content == "" && req.Type == "reply" {
			return errors.New("nội dung tin nhắn phản hồi không thể để trống")
		}
		interact.Content = req.Content
	}

	if err := s.repo.CreateInteract(interact); err != nil {
		return err
	}

	fmt.Printf("Thông báo: Người dùng %s đã gửi một %s đến story %s\n", userID, req.Type, storyID)
	return nil
}

func (s *storyService) GetAnalytics(storyID, userID string) (*dto.StoryAnalyticsResponse, error) {
	story, err := s.repo.FindByID(storyID)
	if err != nil {
		return nil, errors.New("không tìm thấy bản tin")
	}

	// Chỉ có chủ nhân của Story mới có quyền xem Analytics
	if story.UserID != userID {
		return nil, errors.New("bạn không có quyền truy cập dữ liệu phân tích của story này")
	}

	views, _ := s.repo.CountViews(storyID)
	reacts, _ := s.repo.CountInteractionsByType(storyID, "react")
	replies, _ := s.repo.CountInteractionsByType(storyID, "reply")
	shares, _ := s.repo.CountInteractionsByType(storyID, "share")

	return &dto.StoryAnalyticsResponse{
		StoryID:      storyID,
		TotalViews:   views,
		TotalReacts:  reacts,
		TotalReplies: replies,
		TotalShares:  shares,
	}, nil
}
