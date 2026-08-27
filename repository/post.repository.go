package repository

import (
	"context"
	"fmt"
	"linkup/config"
	"linkup/models"
	"linkup/utils"
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

func (r *PostRepository) FetchActive(ctx context.Context, limit int, userID *string, cursorScore *float64, cursorID *string, snapshotTime time.Time, filterFollowing bool) ([]models.Post, error) {
	var posts []models.Post

	w := config.DefaultFeedWeights

	// Subqueries cho engagement counts
	likesSubQuery := r.db.Table("post_reactions").Select("post_id, COUNT(*) AS likes").Group("post_id")
	commentsSubQuery := r.db.Table("comments").Select("post_id, COUNT(*) AS comments").Group("post_id")
	sharesSubQuery := r.db.Table("post_shares").Select("post_id, COUNT(*) AS shares").Group("post_id")

	q := r.db.WithContext(ctx).
		Table("posts").
		Select(`posts.*,
            users.username,
            COALESCE(profiles.display_name, users.username) AS display_name,
            COALESCE(profiles.avatar_uri, '') AS avatar_uri,
            CASE WHEN f.follower_id IS NOT NULL THEN true ELSE false END AS is_following,
            (0.4 * EXP(-? * ((UNIX_TIMESTAMP(?) - UNIX_TIMESTAMP(posts.created_at)) / 3600.0)) +
             0.35 * (LOG(1 + COALESCE(lr.likes, 0)) * ? + LOG(1 + COALESCE(cr.comments, 0)) * ? + LOG(1 + COALESCE(sr.shares, 0)) * ?) / 20.0 +
             0.25 * CASE WHEN f.follower_id IS NOT NULL THEN 1.0 ELSE 0.0 END +
             0.01 * RAND()) AS feed_score`,
			w.DecayRate, snapshotTime, w.LikeWeight, w.CommentWeight, w.ShareWeight).
		Joins("LEFT JOIN users ON users.id = posts.user_id").
		Joins("LEFT JOIN profiles ON profiles.user_id = posts.user_id").
		Joins("LEFT JOIN follows f ON f.following_id = posts.user_id AND f.follower_id = ?", userID).
		Joins("LEFT JOIN (?) lr ON lr.post_id = posts.id", likesSubQuery).
		Joins("LEFT JOIN (?) cr ON cr.post_id = posts.id", commentsSubQuery).
		Joins("LEFT JOIN (?) sr ON sr.post_id = posts.id", sharesSubQuery).
		Where("posts.status = ?", models.PostStatusPublic).
		Limit(limit)

	if filterFollowing {
		q = q.Where("f.follower_id IS NOT NULL")
	}

	if cursorScore != nil && cursorID != nil {
		q = q.Having("feed_score < ? OR (feed_score = ? AND posts.id < ?)",
			*cursorScore, *cursorScore, *cursorID)
	}

	q = q.Order("feed_score DESC, posts.id DESC")

	err := q.Find(&posts).Error
	return posts, err
}

func (r *PostRepository) FetchByUserID(ctx context.Context, targetUserID string, viewerID *string, cursorCreatedAt *time.Time, cursorID *string, limit int) ([]models.Post, error) {
	var posts []models.Post

	q := r.db.WithContext(ctx).
		Table("posts").
		Select(`posts.*,
			users.username,
			COALESCE(profiles.display_name, users.username) AS display_name,
			COALESCE(profiles.avatar_uri, '') AS avatar_uri`).
		Joins("LEFT JOIN users ON users.id = posts.user_id").
		Joins("LEFT JOIN profiles ON profiles.user_id = posts.user_id").
		Where("posts.user_id = ? AND posts.status = ?", targetUserID, models.PostStatusPublic).
		Limit(limit)

	if viewerID != nil && *viewerID != "" && *viewerID != targetUserID {
		q = q.Where(`NOT EXISTS (
			SELECT 1 FROM profiles WHERE profiles.user_id = posts.user_id
			AND profiles.is_private_posts = true
			AND NOT EXISTS (
				SELECT 1 FROM follows WHERE follows.following_id = posts.user_id AND follows.follower_id = ?
			)
		)`, *viewerID)
		q = q.Select(`posts.*,
			users.username,
			COALESCE(profiles.display_name, users.username) AS display_name,
			COALESCE(profiles.avatar_uri, '') AS avatar_uri,
			CASE WHEN f.follower_id IS NOT NULL THEN true ELSE false END AS is_following`).
			Joins("LEFT JOIN follows f ON f.following_id = posts.user_id AND f.follower_id = ?", *viewerID)
	}

	if cursorCreatedAt != nil && cursorID != nil {
		q = q.Where("posts.created_at < ? OR (posts.created_at = ? AND posts.id < ?)",
			*cursorCreatedAt, *cursorCreatedAt, *cursorID)
	}

	q = q.Order("posts.created_at DESC, posts.id DESC")

	err := q.Find(&posts).Error
	return posts, err
}

func (r *PostRepository) FetchByCommunityID(ctx context.Context, communityID string, viewerID *string, cursorCreatedAt *time.Time, cursorID *string, limit int) ([]models.Post, error) {
	var posts []models.Post

	q := r.db.WithContext(ctx).
		Table("posts").
		Select(`posts.*,
			users.username,
			COALESCE(profiles.display_name, users.username) AS display_name,
			COALESCE(profiles.avatar_uri, '') AS avatar_uri`).
		Joins("LEFT JOIN users ON users.id = posts.user_id").
		Joins("LEFT JOIN profiles ON profiles.user_id = posts.user_id").
		Where("posts.community_id = ? AND posts.status = ? AND posts.community_id IS NOT NULL",
			communityID, models.PostStatusPublic).
		Limit(limit)

	if viewerID != nil && *viewerID != "" {
		q = q.Select(`posts.*,
			users.username,
			COALESCE(profiles.display_name, users.username) AS display_name,
			COALESCE(profiles.avatar_uri, '') AS avatar_uri,
			CASE WHEN f.follower_id IS NOT NULL THEN true ELSE false END AS is_following`).
			Joins("LEFT JOIN follows f ON f.following_id = posts.user_id AND f.follower_id = ?", *viewerID)
	}

	if cursorCreatedAt != nil && cursorID != nil {
		q = q.Where("posts.created_at < ? OR (posts.created_at = ? AND posts.id < ?)",
			*cursorCreatedAt, *cursorCreatedAt, *cursorID)
	}

	q = q.Order("posts.created_at DESC, posts.id DESC")

	err := q.Find(&posts).Error
	return posts, err
}

type countRow struct {
	PostID string
	Count  int
}

func (r *PostRepository) BatchCountLikes(ctx context.Context, postIDs []string) (map[string]int, error) {
	if len(postIDs) == 0 {
		return map[string]int{}, nil
	}
	var rows []countRow
	err := r.db.WithContext(ctx).
		Table("post_reactions").
		Select("post_id, COUNT(*) AS count").
		Where("post_id IN ?", postIDs).
		Group("post_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]int, len(postIDs))
	for _, row := range rows {
		m[row.PostID] = row.Count
	}
	return m, nil
}

func (r *PostRepository) BatchCountComments(ctx context.Context, postIDs []string) (map[string]int, error) {
	if len(postIDs) == 0 {
		return map[string]int{}, nil
	}
	var rows []countRow
	err := r.db.WithContext(ctx).
		Table("comments").
		Select("post_id, COUNT(*) AS count").
		Where("post_id IN ?", postIDs).
		Group("post_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]int, len(postIDs))
	for _, row := range rows {
		m[row.PostID] = row.Count
	}
	return m, nil
}

func (r *PostRepository) BatchCountShares(ctx context.Context, postIDs []string) (map[string]int, error) {
	if len(postIDs) == 0 {
		return map[string]int{}, nil
	}
	var rows []countRow
	err := r.db.WithContext(ctx).
		Table("post_shares").
		Select("post_id, COUNT(*) AS count").
		Where("post_id IN ?", postIDs).
		Group("post_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]int, len(postIDs))
	for _, row := range rows {
		m[row.PostID] = row.Count
	}
	return m, nil
}

type boolRow struct {
	PostID string
	Found  bool
}

func (r *PostRepository) BatchCheckLiked(ctx context.Context, userID string, postIDs []string) (map[string]bool, error) {
	if len(postIDs) == 0 {
		return map[string]bool{}, nil
	}
	var rows []boolRow
	err := r.db.WithContext(ctx).
		Table("post_reactions").
		Select("post_id, true AS found").
		Where("post_id IN ? AND user_id = ?", postIDs, userID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]bool, len(postIDs))
	for _, row := range rows {
		m[row.PostID] = row.Found
	}
	return m, nil
}

func (r *PostRepository) BatchCheckSaved(ctx context.Context, userID string, postIDs []string) (map[string]bool, error) {
	if len(postIDs) == 0 {
		return map[string]bool{}, nil
	}
	var rows []boolRow
	err := r.db.WithContext(ctx).
		Table("bookmarks").
		Select("post_id, true AS found").
		Where("post_id IN ? AND user_id = ?", postIDs, userID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]bool, len(postIDs))
	for _, row := range rows {
		m[row.PostID] = row.Found
	}
	return m, nil
}

func (r *PostRepository) BatchCheckShared(ctx context.Context, userID string, postIDs []string) (map[string]bool, error) {
	if len(postIDs) == 0 {
		return map[string]bool{}, nil
	}
	var rows []boolRow
	err := r.db.WithContext(ctx).
		Table("post_shares").
		Select("post_id, true AS found").
		Where("post_id IN ? AND user_id = ?", postIDs, userID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]bool, len(postIDs))
	for _, row := range rows {
		m[row.PostID] = row.Found
	}
	return m, nil
}

// BatchLoadSharedPosts loads the original posts for reposts (shared_from_post_id IS NOT NULL).
// Returns a map[sharedPostID]originalPost.
func (r *PostRepository) BatchLoadSharedPosts(ctx context.Context, posts []models.Post) (map[string]*models.Post, error) {
	var originalIDs []string
	for _, p := range posts {
		if p.SharedFromPostID != nil && *p.SharedFromPostID != "" {
			originalIDs = append(originalIDs, *p.SharedFromPostID)
		}
	}
	if len(originalIDs) == 0 {
		return map[string]*models.Post{}, nil
	}

	var originals []models.Post
	err := r.db.WithContext(ctx).
		Table("posts").
		Select(`posts.*,
			users.username,
			COALESCE(profiles.display_name, users.username) AS display_name,
			COALESCE(profiles.avatar_uri, '') AS avatar_uri`).
		Joins("LEFT JOIN users ON users.id = posts.user_id").
		Joins("LEFT JOIN profiles ON profiles.user_id = posts.user_id").
		Where("posts.id IN ? AND posts.status = ?", originalIDs, models.PostStatusPublic).
		Find(&originals).Error
	if err != nil {
		return nil, err
	}

	m := make(map[string]*models.Post, len(originals))
	for i := range originals {
		m[originals[i].ID] = &originals[i]
	}
	return m, nil
}

func (r *PostRepository) CountActive(ctx context.Context, userID *string) (int64, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&models.Post{}).
		Where("status = ?", models.PostStatusPublic)
	if userID != nil && *userID != "" {
		q = q.Where("user_id != ?", *userID)
	}
	err := q.Count(&count).Error
	return count, err
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

func (r *PostRepository) FindByIDs(ctx context.Context, ids []string) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.WithContext(ctx).
		Table("posts").
		Select(`posts.*,
			(SELECT COUNT(*) FROM post_reactions WHERE post_reactions.post_id = posts.id) AS likes_count,
			(SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id) AS comments_count,
			(SELECT COUNT(*) FROM post_shares WHERE post_shares.post_id = posts.id) AS shares_count,
			users.username,
			COALESCE(profiles.display_name, users.username) AS display_name,
			COALESCE(profiles.avatar_uri, '') AS avatar_uri`).
		Joins("LEFT JOIN users ON users.id = posts.user_id").
		Joins("LEFT JOIN profiles ON profiles.user_id = posts.user_id").
		Where("posts.id IN ?", ids).
		Find(&posts).Error
	if err != nil {
		return nil, err
	}
	if len(posts) > 0 {
		mediaRepo := NewMediaRepository(r.db)
		mediaMap, _ := mediaRepo.GetByPostIDs(ctx, ids)
		for i := range posts {
			if media, ok := mediaMap[posts[i].ID]; ok {
				posts[i].Media = media
			}
		}
	}
	return posts, nil
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

func (r *PostRepository) ListEmojis(ctx context.Context) ([]models.Emoji, error) {
	var emojis []models.Emoji
	err := r.db.WithContext(ctx).Find(&emojis).Error
	if err != nil {
		return nil, err
	}
	return emojis, nil
}

func (r *PostRepository) CreateShare(ctx context.Context, share models.PostShare) error {
	return r.db.WithContext(ctx).Create(&share).Error
}

// Tìm share của một user cho một bài viết cụ thể
func (r *PostRepository) FindShareByUser(ctx context.Context, userID, postID string) (*models.PostShare, error) {
	var share models.PostShare
	err := r.db.WithContext(ctx).Where("user_id = ? AND post_id = ?", userID, postID).First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
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

func (r *PostRepository) FetchCommentsByPostID(ctx context.Context, postID string, limit, offset int, sort string) ([]models.Comment, error) {
	var comments []models.Comment
	query := r.db.WithContext(ctx).
		Table("comments").
		Select(`comments.*,
            users.username,
            COALESCE(profiles.display_name, users.username) AS display_name,
            COALESCE(profiles.avatar_uri, '') AS avatar_uri`).
		Joins("LEFT JOIN users ON users.id = comments.user_id").
		Joins("LEFT JOIN profiles ON profiles.user_id = comments.user_id").
		Where("comments.post_id = ? AND comments.status != ?", postID, models.CommentStatusHidden)

	switch sort {
	case "oldest":
		query = query.Order("comments.created_at ASC")
	case "relevant":
		query = query.Order(`(
			comments.likes_count * 2 +
			(SELECT COUNT(*) FROM comments c2 WHERE c2.parent_id = comments.id AND c2.status != 'hidden') * 3 +
			GREATEST(0, 168 - TIMESTAMPDIFF(HOUR, comments.created_at, NOW()))
		) DESC`)
	default:
		query = query.Order("comments.created_at DESC")
	}

	err := query.Limit(limit).Offset(offset).Find(&comments).Error
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
		Table("comments").
		Select(`comments.*,
            users.username,
            COALESCE(profiles.display_name, users.username) AS display_name,
            COALESCE(profiles.avatar_uri, '') AS avatar_uri`).
		Joins("LEFT JOIN users ON users.id = comments.user_id").
		Joins("LEFT JOIN profiles ON profiles.user_id = comments.user_id").
		Where("comments.post_id = ? AND comments.status != ?", postID, models.CommentStatusHidden).
		Order("comments.created_at DESC").
		Find(&comments).Error
	return comments, err
}

func (r *PostRepository) ListComments(ctx context.Context, keyword, status string, limit, offset int) ([]models.Comment, error) {
	var comments []models.Comment
	q := r.db.WithContext(ctx).Model(&models.Comment{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("content LIKE ?", like)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&comments).Error
	return comments, err
}

func (r *PostRepository) CountComments(ctx context.Context, keyword, status string) (int64, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&models.Comment{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("content LIKE ?", like)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Count(&count).Error
	return count, err
}

func (r *PostRepository) CountCommentsByPostID(ctx context.Context, postID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Comment{}).
		Where("post_id = ? AND status != ?", postID, models.CommentStatusHidden).
		Count(&count).Error
	return count, err
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

// Tạo bản ghi media cho GIF ngoài (Tenor/Giphy) gắn trực tiếp vào bài viết
func (r *PostRepository) CreateExternalGifMedia(ctx context.Context, userID, postID, fileURI string) error {
	media := models.Media{
		ID:        utils.GenerateUUID(),
		UserID:    userID,
		PostID:    &postID,
		FileURI:   fileURI,
		FileType:  "image/gif",
		FileSize:  0,
		Status:    models.MediaStatusApproved,
		CreatedAt: time.Now(),
	}
	return r.db.WithContext(ctx).Create(&media).Error
}

// Lấy thông tin tác giả (username, display_name, avatar_uri) theo user_id
func (r *PostRepository) FetchPostAuthor(ctx context.Context, userID string) (models.Post, error) {
	var author models.Post
	err := r.db.WithContext(ctx).
		Table("users").
		Select(`users.username,
            COALESCE(profiles.display_name, users.username) AS display_name,
            COALESCE(profiles.avatar_uri, '') AS avatar_uri`).
		Joins("LEFT JOIN profiles ON profiles.user_id = users.id").
		Where("users.id = ?", userID).
		First(&author).Error
	return author, err
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
		// Xóa các Reaction của bài viết
		if err := tx.Where("post_id = ?", postID).Delete(&models.PostReaction{}).Error; err != nil {
			return err
		}
		// Xóa các Comment của bài viết (cả reply)
		if err := tx.Where("post_id = ?", postID).Delete(&models.Comment{}).Error; err != nil {
			return err
		}
		// Ngắt liên kết Media khỏi bài viết (giữ nguyên file)
		if err := tx.Table("media").Where("post_id = ?", postID).Update("post_id", nil).Error; err != nil {
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

// Lấy danh sách bài viết đã lưu (Bookmark) của người dùng theo con trỏ
func (r *PostRepository) FetchSaved(ctx context.Context, userID string, limit int, cursorCreatedAt *time.Time, cursorID *string) ([]models.Post, error) {
	var posts []models.Post

	q := r.db.WithContext(ctx).
		Table("bookmarks b").
		Select(`p.*,
            b.id AS bookmark_id,
            b.created_at AS saved_at,
            users.username,
            COALESCE(profiles.display_name, users.username) AS display_name,
            COALESCE(profiles.avatar_uri, '') AS avatar_uri,
            CASE WHEN f.follower_id IS NOT NULL THEN true ELSE false END AS is_following`).
		Joins("JOIN posts p ON p.id = b.post_id").
		Joins("LEFT JOIN users ON users.id = p.user_id").
		Joins("LEFT JOIN profiles ON profiles.user_id = p.user_id").
		Joins("LEFT JOIN follows f ON f.following_id = p.user_id AND f.follower_id = ?", userID).
		Where("b.user_id = ?", userID).
		Where("p.status = ?", models.PostStatusPublic).
		Limit(limit)

	if cursorCreatedAt != nil && cursorID != nil {
		q = q.Where("b.created_at < ? OR (b.created_at = ? AND b.id < ?)",
			*cursorCreatedAt, *cursorCreatedAt, *cursorID)
	}

	q = q.Order("b.created_at DESC, b.id DESC")

	err := q.Find(&posts).Error
	return posts, err
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
		Where("posts.status = ?", models.PostStatusPublic).
		Order("posts.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error

	return posts, err
}

func (r *PostRepository) PinPost(ctx context.Context, postID string) error {
	now := time.Now()
	tx := r.db.WithContext(ctx).
		Model(&models.Post{}).
		Where("id = ?", postID).
		Updates(map[string]interface{}{"is_pinned": true, "pinned_at": now})
	if tx.Error != nil {
		return fmt.Errorf("pin post: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("post not found")
	}
	return nil
}

func (r *PostRepository) UnpinPost(ctx context.Context, postID string) error {
	tx := r.db.WithContext(ctx).
		Model(&models.Post{}).
		Where("id = ?", postID).
		Updates(map[string]interface{}{"is_pinned": false, "pinned_at": nil})
	if tx.Error != nil {
		return fmt.Errorf("unpin post: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("post not found")
	}
	return nil
}

func (r *PostRepository) FetchMediaByUserID(ctx context.Context, userID string, offset, limit int) ([]models.Media, error) {
	media := make([]models.Media, 0)
	tx := r.db.WithContext(ctx).
		Raw(`SELECT m.id, m.post_id, m.file_uri, m.file_type, m.file_size, m.created_at
			FROM media m
			JOIN posts p ON p.id = m.post_id
			WHERE p.user_id = ? AND p.status = ?
			ORDER BY m.created_at DESC
			LIMIT ? OFFSET ?`, userID, models.PostStatusPublic, limit, offset).
		Scan(&media)
	if tx.Error != nil {
		return nil, fmt.Errorf("fetch media by user: %w", tx.Error)
	}
	return media, nil
}

func (r *PostRepository) CountMediaByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	tx := r.db.WithContext(ctx).
		Raw(`SELECT COUNT(*)
			FROM media m
			JOIN posts p ON p.id = m.post_id
			WHERE p.user_id = ? AND p.status = ?`, userID, models.PostStatusPublic).
		Scan(&count)
	if tx.Error != nil {
		return 0, fmt.Errorf("count media by user: %w", tx.Error)
	}
	return count, nil
}

func (r *PostRepository) FindCommentReactionByUserAndComment(ctx context.Context, userID, commentID string) (*models.CommentReaction, error) {
	var reaction models.CommentReaction
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND comment_id = ?", userID, commentID).
		First(&reaction).Error
	if err != nil {
		return nil, err
	}
	return &reaction, nil
}

func (r *PostRepository) CreateCommentReaction(ctx context.Context, reaction *models.CommentReaction) error {
	return r.db.WithContext(ctx).Create(reaction).Error
}

func (r *PostRepository) DeleteCommentReaction(ctx context.Context, userID, commentID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND comment_id = ?", userID, commentID).
		Delete(&models.CommentReaction{}).Error
}

func (r *PostRepository) UpdateCommentLikesCount(ctx context.Context, commentID string, delta int) error {
	return r.db.WithContext(ctx).
		Model(&models.Comment{}).
		Where("id = ?", commentID).
		UpdateColumn("likes_count", gorm.Expr("likes_count + ?", delta)).Error
}
