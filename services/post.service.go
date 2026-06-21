package services

import (
	"context"
	"errors"
	"linkup/models"
	"linkup/repository"
	"time"

	"github.com/google/uuid"
)

type PostService interface {
	CreatePost(ctx context.Context, userID, title, content string) (*models.Post, error)
	GetPostList(ctx context.Context, page, pageSize int) ([]models.Post, error)
	GetPostDetail(ctx context.Context, postID string) (*models.Post, error)
	ReactPost(ctx context.Context, userID, postID, emojiID string) (string, error)
}

type postService struct {
	repo repository.PostRepository
}

func NewPostService(repo repository.PostRepository) PostService {
	return &postService{repo: repo}
}

func (s *postService) CreatePost(ctx context.Context, userID, title, content string) (*models.Post, error) {
	if title == "" || content == "" {
		return nil, errors.New("tên bài viết và nội dung không được bỏ trống")
	}

	post := models.NewPost(userID, title, content)
	post.ID = uuid.New().String()
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

func (s *postService) ReactPost(ctx context.Context, userID, postID, emojiID string) (string, error) {
	if emojiID == "" {
		return "", errors.New("emoji_id không được rỗng")
	}

	existingReaction, err := s.repo.FindReaction(ctx, userID, postID, emojiID)

	if err == nil && existingReaction != nil {
		if errDelete := s.repo.DeleteReaction(ctx, existingReaction.ID); errDelete != nil {
			return "", errDelete
		}
		return "removed", nil
	}

	reaction := models.PostReaction{
		ID:        uuid.New().String(),
		UserID:    userID,
		PostID:    postID,
		EmojiID:   emojiID,
		CreatedAt: time.Now(),
	}

	if errCreate := s.repo.CreateReaction(ctx, reaction); errCreate != nil {
		return "", errCreate
	}

	return "reacted", nil
}
