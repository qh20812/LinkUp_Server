package services

import (
	"context"
	"errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"log"
	"time"
)

type PostService interface {
	CreatePost(ctx context.Context, userID, title, content, status string) (*models.Post, error)
	GetPostList(ctx context.Context, page, pageSize int) ([]models.Post, error)
	GetPostDetail(ctx context.Context, postID string) (*models.Post, error)
	ReactPost(ctx context.Context, userID, postID, emojiID string) (action string, emojiCode string, err error)
	CreateComment(ctx context.Context, userID, postID string, parentID *string, content string) ([]models.Comment, error)
	GetCommentList(ctx context.Context, postID string, page, pageSize int) ([]models.Comment, error)
	SharePost(ctx context.Context, userID, postID string) error
	SavePost(ctx context.Context, userID, postID string) error
}

type postService struct {
	repo         *repository.PostRepository
	notifService *NotificationService
	tagService   *TagService // 🌟 Đã cập nhật
}

func NewPostService(repo *repository.PostRepository, notifService *NotificationService, tagService *TagService) PostService {
	return &postService{
		repo:         repo,
		notifService: notifService,
		tagService:   tagService, // 🌟 Đã cập nhật
	}
}

func (s *postService) CreatePost(ctx context.Context, userID, title, content, status string) (*models.Post, error) {
	postStatus := models.ParsePostStatus(status)

	post := models.NewPost(userID, title, content, postStatus)
	post.ID = utils.GenerateUUID()
	post.CreatedAt = time.Now()
	post.ViewsCount = 0

	if err := s.repo.Create(ctx, &post); err != nil {
		return nil, err
	}

	// 🌟 Đã cập nhật: Tự động bóc tách và lưu hashtag từ bài viết mới
	if err := s.tagService.ProcessPostHashtags(ctx, nil, post.ID, content); err != nil {
		log.Printf("[Hashtag Error] không thể lưu tag cho post %s: %v", post.ID, err)
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
		return nil, errors.New("bài viết không tồn tại")
	}

	if post.Status == models.PostStatusHidden || post.Status == models.PostStatusPrivate {
		return nil, errors.New("bài viết đã bị ẩn hoặc ở chế độ riêng tư")
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

func (s *postService) CreateComment(ctx context.Context, userID, postID string, parentID *string, content string) ([]models.Comment, error) {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return nil, errors.New("bài viết không tồn tại")
	}

	if post.Status == models.PostStatusHidden || post.Status == models.PostStatusPrivate {
		return nil, errors.New("không thể bình luận vào bài viết đã bị ẩn hoặc ở chế độ riêng tư")
	}

	if parentID != nil && *parentID != "" {
		parentComment, err := s.repo.FindCommentByID(ctx, *parentID)
		if err != nil || parentComment == nil {
			return nil, errors.New("bình luận cấp trên không tồn tại hoặc đã bị xóa")
		}
		if parentComment.PostID != postID {
			return nil, errors.New("bình luận gốc không thuộc bài viết này")
		}
	} else {
		parentID = nil
	}

	comment := models.NewComment(userID, postID, parentID, content)
	comment.ID = utils.GenerateUUID()
	comment.CreatedAt = time.Now()

	if err := s.repo.CreateComment(ctx, &comment); err != nil {
		return nil, err
	}

	// 🌟 Đã cập nhật: Tự động bóc tách và lưu hashtag từ bình luận mới
	if err := s.tagService.ProcessCommentHashtags(ctx, nil, postID, comment.ID, content); err != nil {
		log.Printf("[Hashtag Error] không thể lưu tag cho comment %s: %v", comment.ID, err)
	}

	if post.UserID != userID {
		senderIDPtr := userID
		postIDPtr := postID
		commentIDPtr := comment.ID

		notification := models.NewNotification(
			post.UserID,
			&senderIDPtr,
			models.NotificationTypeComment,
			"Bạn vừa có 1 comment mới trên bài viết của mình.",
		)

		notification.ID = utils.GenerateUUID()
		notification.RedirectPostID = &postIDPtr
		notification.RedirectCommentID = &commentIDPtr
		notification.IsRead = false
		notification.CreatedAt = time.Now()

		_ = s.repo.CreateNotification(ctx, notification)
	}

	return s.repo.FindCommentsByPostID(ctx, postID)
}

func (s *postService) GetCommentList(ctx context.Context, postID string, page, pageSize int) ([]models.Comment, error) {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil || post.Status == models.PostStatusHidden || post.Status == models.PostStatusPrivate {
		return nil, errors.New("bài viết không tồn tại hoặc không thể truy cập")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	return s.repo.FetchCommentsByPostID(ctx, postID, pageSize, offset)
}

func (s *postService) SharePost(ctx context.Context, userID, postID string) error {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil || post.Status == models.PostStatusHidden || post.Status == models.PostStatusPrivate {
		return errors.New("bài viết không tồn tại hoặc không cho phép chia sẻ")
	}

	share := models.NewPostShare(userID, postID)
	share.ID = utils.GenerateUUID()
	share.CreatedAt = time.Now()

	return s.repo.CreateShare(ctx, share)
}

func (s *postService) SavePost(ctx context.Context, userID, postID string) error {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil || post.Status == models.PostStatusHidden || post.Status == models.PostStatusPrivate {
		return errors.New("bài viết không tồn tại hoặc đã bị ẩn")
	}

	bookmark := models.Bookmark{
		ID:        utils.GenerateUUID(),
		UserID:    userID,
		PostID:    postID,
		CreatedAt: time.Now(),
	}

	return s.repo.CreateSave(ctx, bookmark)
}
