package repository

import (
	"context"
	"linkup/models"

	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(ctx context.Context, post *models.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *PostRepository) FetchActive(ctx context.Context, limit, offset int) ([]models.Post, error) {
	var posts []models.Post

	err := r.db.WithContext(ctx).
		Table("posts").
		Select(`posts.*, 
            (SELECT COUNT(*) FROM post_reactions WHERE post_reactions.post_id = posts.id) AS likes_count,
            (SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id) AS comments_count,
            (SELECT COUNT(*) FROM post_shares WHERE post_shares.post_id = posts.id) AS shares_count`).
		Where("posts.status = ?", models.PostStatusActive).
		Order("RAND()").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error

	return posts, err
}

func (r *PostRepository) FindByID(ctx context.Context, id string) (*models.Post, error) {
	var post models.Post
	err := r.db.WithContext(ctx).
		Table("posts").
		Select(`posts.*, 
            (SELECT COUNT(*) FROM post_reactions WHERE post_reactions.post_id = posts.id) AS likes_count,
            (SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id) AS comments_count,
            (SELECT COUNT(*) FROM post_shares WHERE post_shares.post_id = posts.id) AS shares_count`).
		Where("posts.id = ?", id).
		First(&post).Error

	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) IncrementViewsCount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&models.Post{}).Where("id = ?", id).
		Update("views_count", gorm.Expr("views_count + ?", 1)).Error
}

func (r *PostRepository) CreateReaction(ctx context.Context, reaction models.PostReaction) error {
	return r.db.WithContext(ctx).Create(&reaction).Error
}

func (r *PostRepository) FindReaction(ctx context.Context, userID, postID, emojiID string) (*models.PostReaction, error) {
	var reaction models.PostReaction
	err := r.db.WithContext(ctx).Where("user_id = ? AND post_id = ? AND emoji_id = ?", userID, postID, emojiID).First(&reaction).Error
	if err != nil {
		return nil, err
	}
	return &reaction, nil
}

func (r *PostRepository) DeleteReaction(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.PostReaction{}, "id = ?", id).Error
}

func (r *PostRepository) FindEmojiByID(ctx context.Context, id string) (*models.Emoji, error) {
	var emoji models.Emoji
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&emoji).Error
	if err != nil {
		return nil, err
	}
	return &emoji, nil
}

func (r *PostRepository) CreateShare(ctx context.Context, share models.PostShare) error {
	return r.db.WithContext(ctx).Create(&share).Error
}

func (r *PostRepository) CreateComment(ctx context.Context, comment *models.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *PostRepository) FindCommentByID(ctx context.Context, id string) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&comment).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *PostRepository) FetchCommentsByPostID(ctx context.Context, postID string, limit, offset int) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.WithContext(ctx).
		Where("post_id = ?", postID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&comments).Error
	return comments, err
}

func (r *PostRepository) CreateSave(ctx context.Context, bookmark models.Bookmark) error {
	return r.db.WithContext(ctx).Create(&bookmark).Error
}

func (r *PostRepository) CreateNotification(ctx context.Context, notification models.Notification) error {
	return r.db.WithContext(ctx).Create(&notification).Error
}

// Lấy toàn bộ danh sách bình luận không phân trang
func (r *PostRepository) FindCommentsByPostID(ctx context.Context, postID string) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.WithContext(ctx).
		Where("post_id = ?", postID).
		Order("created_at DESC").
		Find(&comments).Error
	return comments, err
}

func (r *PostRepository) ListPosts(ctx context.Context, keyword, status string, limit, offset int) ([]models.Post, error) {
	var posts []models.Post

	q := r.db.WithContext(ctx).
		Table("posts").
		Select(`posts.*,
            (SELECT COUNT(*) FROM post_reactions WHERE post_reactions.post_id = posts.id) AS likes_count,
            (SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id) AS comments_count,
            (SELECT COUNT(*) FROM post_shares WHERE post_shares.post_id = posts.id) AS shares_count`)

	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("posts.title LIKE ? OR posts.content LIKE ?", like, like)
	}
	if status != "" {
		q = q.Where("posts.status = ?", status)
	}

	err := q.Order("posts.created_at DESC").Limit(limit).Offset(offset).Find(&posts).Error
	return posts, err
}

func (r *PostRepository) CountPosts(ctx context.Context, keyword, status string) (int64, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&models.Post{})

	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("title LIKE ? OR content LIKE ?", like, like)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}

	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PostRepository) UpdateStatus(ctx context.Context, id string, status models.PostStatus) error {
	return r.db.WithContext(ctx).Model(&models.Post{}).Where("id = ?", id).Update("status", status).Error
}
