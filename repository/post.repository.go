package repository

import (
	"context"
	"linkup/models"
	"time"

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

func (r *PostRepository) UpdateCommentStatus(ctx context.Context, id string, status models.CommentStatus, reviewReason string) error {
	now := time.Now().UTC()
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": now,
	}
	if reviewReason != "" {
		updates["review_reason"] = reviewReason
	}
	return r.db.WithContext(ctx).Model(&models.Comment{}).Where("id = ?", id).Updates(updates).Error
}

func (r *PostRepository) FindDescendantCommentIDs(ctx context.Context, parentID string) ([]string, error) {
	var allIDs []string
	currentBatch := []string{parentID}
	for len(currentBatch) > 0 {
		var ids []string
		err := r.db.WithContext(ctx).
			Model(&models.Comment{}).
			Where("parent_id IN ? AND status != ?", currentBatch, models.CommentStatusHidden).
			Pluck("id", &ids).Error
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			break
		}
		allIDs = append(allIDs, ids...)
		currentBatch = ids
	}
	return allIDs, nil
}

func (r *PostRepository) HideCommentsByIDs(ctx context.Context, ids []string, reason string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&models.Comment{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"status":        models.CommentStatusHidden,
			"review_reason": reason,
			"updated_at":    now,
		}).Error
}

func (r *PostRepository) FetchCommentsByPostID(ctx context.Context, postID string, limit, offset int) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.WithContext(ctx).
		Where("post_id = ? AND status != ?", postID, models.CommentStatusHidden).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&comments).Error
	return comments, err
}

func (r *PostRepository) FindActiveCommentByID(ctx context.Context, id string) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.WithContext(ctx).Where("id = ? AND status != ?", id, models.CommentStatusHidden).First(&comment).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *PostRepository) CreateSave(ctx context.Context, bookmark models.Bookmark) error {
	return r.db.WithContext(ctx).Create(&bookmark).Error
}

func (r *PostRepository) CreateNotification(ctx context.Context, notification models.Notification) error {
	return r.db.WithContext(ctx).Create(&notification).Error
}

func (r *PostRepository) FindCommentsByPostID(ctx context.Context, postID string) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.WithContext(ctx).
		Where("post_id = ? AND status != ?", postID, models.CommentStatusHidden).
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

// Tìm kiếm Bookmark để phục vụ tính năng Toggle kiểm tra đã lưu hay chưa
func (r *PostRepository) FindBookmark(ctx context.Context, userID, postID string) (*models.Bookmark, error) {
	var bookmark models.Bookmark
	err := r.db.WithContext(ctx).Where("user_id = ? AND post_id = ?", userID, postID).First(&bookmark).Error
	if err != nil {
		return nil, err
	}
	return &bookmark, nil
}

// Xóa một Bookmark cụ thể
func (r *PostRepository) DeleteBookmark(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Bookmark{}, "id = ?", id).Error
}

// Liên kết các Media ID tạm thời vào Post ID sau khi upload xong
func (r *PostRepository) LinkMediaToPost(ctx context.Context, mediaIDs []string, postID string) error {
	return r.db.WithContext(ctx).Table("media").Where("id IN ?", mediaIDs).Update("post_id", postID).Error
}

// Xóa bài viết đồng thời xóa hàng loạt Bookmark & Share trong DB GORM Transaction
func (r *PostRepository) DeletePostWithAssociations(ctx context.Context, postID string) ([]string, error) {
	var bookmarkedUserIDs []string

	// Lấy ra danh sách ID người dùng đã lưu bài viết này để gửi thông báo ở tầng service
	r.db.WithContext(ctx).Model(&models.Bookmark{}).Where("post_id = ?", postID).Pluck("user_id", &bookmarkedUserIDs)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Xóa toàn bộ Bookmark lưu bài viết này
		if err := tx.Where("post_id = ?", postID).Delete(&models.Bookmark{}).Error; err != nil {
			return err
		}
		// Xóa toàn bộ lượt Share của bài viết này (Gốc mất -> Share mất)
		if err := tx.Where("post_id = ?", postID).Delete(&models.PostShare{}).Error; err != nil {
			return err
		}
		// Xóa các Hashtags/Tags liên quan đến bài viết
		if err := tx.Table("tags").Where("post_id = ?", postID).Delete(map[string]interface{}{}).Error; err != nil {
			return err
		}
		// Xóa bài viết chính thức khỏi hệ thống
		if err := tx.Delete(&models.Post{}, "id = ?", postID).Error; err != nil {
			return err
		}
		return nil
	})

	return bookmarkedUserIDs, err
}

// Lấy danh sách thông tin bài viết theo tập hợp các ID tìm được từ Hashtag
func (r *PostRepository) FetchByIDs(ctx context.Context, ids []string, limit, offset int) ([]models.Post, error) {
	var posts []models.Post
	if len(ids) == 0 {
		return posts, nil
	}

	err := r.db.WithContext(ctx).
		Table("posts").
		Select(`posts.*, 
            (SELECT COUNT(*) FROM post_reactions WHERE post_reactions.post_id = posts.id) AS likes_count,
            (SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id) AS comments_count,
            (SELECT COUNT(*) FROM post_shares WHERE post_shares.post_id = posts.id) AS shares_count`).
		Where("posts.id IN ?", ids).
		Where("posts.status = ?", models.PostStatusActive).
		Order("posts.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error

	return posts, err
}
