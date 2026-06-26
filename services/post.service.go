package services

import (
	"context"
	"errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"time"
)

type PostService interface {
	CreatePost(ctx context.Context, userID, title, content string) (*models.Post, error)
	GetPostList(ctx context.Context, page, pageSize int) ([]models.Post, error)
	GetPostDetail(ctx context.Context, postID string) (*models.Post, error)
	ReactPost(ctx context.Context, userID, postID, emojiID string) (action string, emojiCode string, err error)
	CreateComment(ctx context.Context, userID, postID string, parentID *string, content string) (*models.Comment, error)
}

type postService struct {
	repo        *repository.PostRepository
	notifService *NotificationService
}

func NewPostService(repo *repository.PostRepository, notifService *NotificationService) PostService {
	return &postService{repo: repo, notifService: notifService}
}

func (s *postService) CreatePost(ctx context.Context, userID, title, content string) (*models.Post, error) {
	if title == "" || content == "" {
		return nil, errors.New("tên bài viết và nội dung không được bỏ trống")
	}

	post := models.NewPost(userID, title, content)
	post.ID = utils.GenerateUUID()
	post.CreatedAt = time.Now()
	post.ViewsCount = 0

	if err := s.repo.Create(ctx, &post); err != nil {
		return nil, err
	}
	return &post, nil
}

func (s *postService) GetPostList(ctx context.Context, page, pageSize int) ([]models.Post, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	return s.repo.FetchActive(ctx, pageSize, offset)
}

func (s *postService) GetPostDetail(ctx context.Context, postID string) (*models.Post, error) {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	_ = s.repo.IncrementViewsCount(ctx, postID)
	post.ViewsCount++

	return post, nil
}

func (s *postService) ReactPost(ctx context.Context, userID, postID, emojiID string) (string, string, error) {
	if emojiID == "" {
		return "", "", errors.New("emoji_id không được rỗng")
	}

	emoji, err := s.repo.FindEmojiByID(ctx, emojiID)
	if err != nil {
		return "", "", errors.New("emoji không tồn tại")
	}

	existingReaction, err := s.repo.FindReaction(ctx, userID, postID, emojiID)

	if err == nil && existingReaction != nil {
		if errDelete := s.repo.DeleteReaction(ctx, existingReaction.ID); errDelete != nil {
			return "", "", errDelete
		}
		return "removed", emoji.Code, nil
	}

	reaction := models.PostReaction{
		ID:        utils.GenerateUUID(),
		UserID:    userID,
		PostID:    postID,
		EmojiID:   emojiID,
		CreatedAt: time.Now(),
	}

	if errCreate := s.repo.CreateReaction(ctx, reaction); errCreate != nil {
		return "", "", errCreate
	}

	if post, err := s.repo.FindByID(ctx, postID); err == nil && post != nil && post.UserID != userID {
		s.notifService.Create(ctx, post.UserID, &userID, models.NotificationTypeLike, "đã thích bài viết của bạn", &postID, nil, nil)
	}

	return "reacted", emoji.Code, nil
}

func (s *postService) CreateComment(ctx context.Context, userID, postID string, parentID *string, content string) (*models.Comment, error) {
	if content == "" {
		return nil, errors.New("nội dung bình luận không được trống")
	}

	// Nếu là Reply (có parentID), tiến hành kiểm tra tính toàn vẹn
	if parentID != nil && *parentID != "" {
		parentComment, err := s.repo.FindCommentByID(ctx, *parentID)
		if err != nil || parentComment == nil {
			return nil, errors.New("bình luận cấp trên không tồn tại hoặc đã bị xóa")
		}

		if parentComment.PostID != postID {
			return nil, errors.New("bình luận gốc không thuộc bài viết này")
		}
	} else {
		parentID = nil // Đưa chuỗi rỗng về nil để DB lưu thành NULL
	}

	comment := models.NewComment(userID, postID, parentID, content)
	comment.ID = utils.GenerateUUID()
	comment.CreatedAt = time.Now()

	if err := s.repo.CreateComment(ctx, &comment); err != nil {
		return nil, err
	}

	if post, err := s.repo.FindByID(ctx, postID); err == nil && post != nil && post.UserID != userID {
		s.notifService.Create(ctx, post.UserID, &userID, models.NotificationTypeComment, "đã bình luận về bài viết của bạn", &postID, nil, nil)
	}

	return &comment, nil
}
