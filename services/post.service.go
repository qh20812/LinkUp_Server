package services

import (
	"context"
	"errors"
	"linkup/models"
	"linkup/repository"
	"time"
	"github.com/google/uuid" // go get github.com/google/uuid
)

type PostService interface {
	CreatePost(ctx context.Context, userID, title, content string) (*models.Post, error)
	GetPostList(ctx context.Context, page, pageSize int) ([]models.Post, error)
}

type postService struct {
	repo repository.PostRepository
}

func NewPostService(repo repository.PostRepository) PostService {
	return &postService{repo: repo}
}

func (s *postService) CreatePost(ctx context.Context, userID, title, content string) (*models.Post, error) {
	if title == "" || content == "" {
		return nil, errors.New("title and content cannot be empty")
	}

	// Sử dụng hàm NewPost có sẵn của bạn để lấy cấu trúc mặc định ban đầu
	post := models.NewPost(userID, title, content)

	// Bổ sung các thông tin hệ thống tự tạo
	post.ID = uuid.New().String()
	post.CreatedAt = time.Now()
	post.ViewsCount = 0

	err := s.repo.Create(ctx, &post)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (s *postService) GetPostList(ctx context.Context, page, pageSize int) ([]models.Post, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10 // Số lượng bài mặc định trên 1 trang
	}

	offset := (page - 1) * pageSize
	return s.repo.FetchActive(ctx, pageSize, offset)
}
