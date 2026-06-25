package repository

import (
	"context"
	"gorm.io/gorm"
	"linkup/models"
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
		Where("posts.id = ? AND posts.status = ?", id, models.PostStatusActive).
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
