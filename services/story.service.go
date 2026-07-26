package services

import (
	"context"
	"errors"
	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type StoryService interface {
	CreateStory(ctx context.Context, userID string, fileHeader *multipart.FileHeader, caption string) (*dto.CreateStoryResponse, error)
	GetHomeStories() ([]dto.StoryResponse, error)
	ViewStory(storyID, viewerID string) (*models.Story, error)
	InteractWithStory(storyID, userID string, req dto.InteractStoryRequest) error
	GetAnalytics(storyID, userID string) (*dto.StoryAnalyticsResponse, error)
}

type storyService struct {
	repo         repository.StoryRepository
	mediaService MediaService // Inject MediaService
}

func NewStoryService(repo repository.StoryRepository, mediaService MediaService) StoryService {
	return &storyService{
		repo:         repo,
		mediaService: mediaService,
	}
}

func (s *storyService) CreateStory(ctx context.Context, userID string, fileHeader *multipart.FileHeader, caption string) (*dto.CreateStoryResponse, error) {
	var mediaURI string
	var mediaType models.StoryMediaType

	// Trường hợp có tải file lên
	if fileHeader != nil {
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))

		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
			mediaType = models.StoryMediaTypeImage
		} else if ext == ".mp4" || ext == ".mov" {
			mediaType = models.StoryMediaTypeVideo
		} else {
			return nil, errors.New("định dạng file không hỗ trợ, chỉ nhận ảnh (jpg, png) hoặc video (mp4, mov)")
		}

		// Upload file qua MediaService
		mediaRecord, err := s.mediaService.UploadMedia(ctx, userID, fileHeader)
		if err != nil {
			return nil, err
		}
		mediaURI = mediaRecord.FileURI
	} else {
		// Trường hợp KHÔNG CÓ file, bắt buộc phải có caption (story dạng text)
		if strings.TrimSpace(caption) == "" {
			return nil, errors.New("story phải có hình ảnh, video hoặc nội dung chữ (caption)")
		}
		mediaType = ""
	}

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

	// Nếu người xem không phải chủ story
	if story.UserID != viewerID {
		// Kiểm tra xem user này đã từng xem story này chưa
		hasViewed, err := s.repo.HasUserViewed(storyID, viewerID)
		if err != nil {
			return nil, err
		}

		// Nếu CHƯA XEM thì mới tiến hành ghi nhận tăng view
		if !hasViewed {
			if err := s.repo.LogView(storyID, viewerID); err != nil {
				return nil, err
			}
		}
		// Nếu ĐÃ XEM RỒI thì bỏ qua bước ghi nhận, giữ nguyên view hiện tại
	}

	return story, nil
}

func (s *storyService) InteractWithStory(storyID string, userID string, req dto.InteractStoryRequest) error {
	_, err := s.repo.FindByID(storyID)
	if err != nil {
		return errors.New("không tìm thấy bản tin để tương tác")
	}

	if req.Type == "react" {
		if req.EmojiID == "" {
			return errors.New("emoji_id không được để trống khi thực hiện thả cảm xúc")
		}

		existingReact, err := s.repo.FindReactByUser(storyID, userID)
		if err == nil && existingReact != nil {
			if existingReact.ClickCount >= 5 {
				return errors.New("bạn đã đạt giới hạn tối đa 5 lần biểu cảm cho story này")
			}
			existingReact.ClickCount++
			existingReact.EmojiID = &req.EmojiID
			return s.repo.UpdateInteract(existingReact)
		}

		exists, _ := s.repo.CheckEmojiExists(req.EmojiID)
		if !exists {
			return errors.New("mã hiệu ứng emoji không tồn tại trong hệ thống")
		}

		interact := &models.StoryInteract{
			StoryID:    storyID,
			UserID:     userID,
			Type:       req.Type,
			EmojiID:    &req.EmojiID,
			ClickCount: 1,
		}
		return s.repo.CreateInteract(interact)
	}

	if req.Content == "" && req.Type == "reply" {
		return errors.New("nội dung tin nhắn phản hồi không thể để trống")
	}

	interact := &models.StoryInteract{
		StoryID: storyID,
		UserID:  userID,
		Type:    req.Type,
		Content: req.Content,
	}
	return s.repo.CreateInteract(interact)
}

func (s *storyService) GetAnalytics(storyID, userID string) (*dto.StoryAnalyticsResponse, error) {
	story, err := s.repo.FindByID(storyID)
	if err != nil {
		return nil, errors.New("không tìm thấy bản tin")
	}

	if story.UserID != userID {
		return nil, errors.New("bạn không có quyền truy cập dữ liệu phân tích của story này")
	}

	views, _ := s.repo.CountViews(storyID)
	reacts, _ := s.repo.CountInteractionsByType(storyID, "react")
	replies, _ := s.repo.CountInteractionsByType(storyID, "reply")
	shares, _ := s.repo.CountInteractionsByType(storyID, "share")

	viewersRaw, interactsRaw, _ := s.repo.GetViewersDetails(storyID)

	// Dùng Map để gom nhóm duy nhất mỗi user xuất hiện 1 lần trong danh sách viewers
	viewerMap := make(map[string]*dto.ViewerDetailResponse)

	// Lưu trữ thời điểm reply mới nhất để sort hoặc lọc nếu cần
	// duyệt qua danh sách views trước để lấy thời gian xem gần nhất
	for _, v := range viewersRaw {
		if _, exists := viewerMap[v.UserID]; !exists {
			viewedAtCopy := v.ViewedAt
			viewerMap[v.UserID] = &dto.ViewerDetailResponse{
				UserID:   v.UserID,
				ViewedAt: viewedAtCopy,
				Messages: []string{},
			}
		}
	}

	// Duyệt qua các tương tác để gán React và chỉ lấy tin nhắn reply mới nhất
	for _, inter := range interactsRaw {
		viewer, exists := viewerMap[inter.UserID]
		if !exists {
			// trường hợp user tương tác nhưng chưa có trong bảng view
			viewer = &dto.ViewerDetailResponse{
				UserID:   inter.UserID,
				ViewedAt: inter.CreatedAt,
				Messages: []string{},
			}
			viewerMap[inter.UserID] = viewer
		}

		if inter.Type == "react" {
			rType := inter.Type
			viewer.ReactType = &rType
			viewer.EmojiID = inter.EmojiID
			viewer.ClickCount = inter.ClickCount
		} else if inter.Type == "reply" {
			// hiển thị tin nhắn mới nhất, đưa vào mảng chỉ chứa 1 phần tử mới nhất
			viewer.Messages = []string{inter.Content}
		}
	}

	var viewersList []dto.ViewerDetailResponse
	for _, v := range viewerMap {
		viewersList = append(viewersList, *v)
	}

	return &dto.StoryAnalyticsResponse{
		StoryID:      storyID,
		TotalViews:   views,
		TotalReacts:  reacts,
		TotalReplies: replies,
		TotalShares:  shares,
		Viewers:      viewersList,
	}, nil
}
