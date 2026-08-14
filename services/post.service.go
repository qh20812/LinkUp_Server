package services

import (
	"context"
	"fmt"
	errorsapp "linkup/errors"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
	"log"
	"mime/multipart"
	"strconv"
	"strings"
	"time"
)

type PostService interface {
	CreatePost(ctx context.Context, userID, title, content, status string, communityID *string, files []*multipart.FileHeader) (*models.Post, error)
	GetPostList(ctx context.Context, cursor string, pageSize int, userID string, filter string) ([]models.Post, string, error)
	GetSavedPosts(ctx context.Context, userID string, cursor string, pageSize int) ([]models.Post, string, error)
	GetUserPosts(ctx context.Context, targetUserID, viewerID, cursor string, pageSize int) ([]models.Post, string, error)
	GetPostDetail(ctx context.Context, postID string) (*models.Post, error)
	ReactPost(ctx context.Context, userID, postID, emojiID string) (action string, emojiCode string, err error)
	CreateComment(ctx context.Context, userID, postID string, parentID *string, content string) ([]models.Comment, error)
	GetCommentList(ctx context.Context, postID string, page, pageSize int) ([]models.Comment, int64, error)
	SharePost(ctx context.Context, userID, postID, content string) error
	SavePost(ctx context.Context, userID, postID string) (action string, err error)
	DeletePost(ctx context.Context, userID, postID string) error
	GetPostsByHashtag(ctx context.Context, hashtag string, page, pageSize int) ([]models.Post, error)
	ListEmojis(ctx context.Context) ([]models.Emoji, error)
	SetMediaService(mediaService MediaService)
}

type postService struct {
	repo                *repository.PostRepository
	notifService        *NotificationService
	tagService          *TagService
	contributionService *ContributionService
	mediaService        MediaService
	validation          *validations.PostValidation
}

func NewPostService(repo *repository.PostRepository, notifService *NotificationService, tagService *TagService, validation *validations.PostValidation) *postService {
	return &postService{repo: repo, notifService: notifService, tagService: tagService, validation: validation}
}

func (s *postService) SetContributionService(contributionService *ContributionService) {
	s.contributionService = contributionService
}

func (s *postService) SetMediaService(mediaService MediaService) {
	s.mediaService = mediaService
}

func (s *postService) CreatePost(ctx context.Context, userID, title, content, status string, communityID *string, files []*multipart.FileHeader) (*models.Post, error) {
	if communityID != nil {
		if s.contributionService == nil {
			return nil, errorsapp.New(errorsapp.ErrCodePostContributionNotInit)
		}
		if err := s.contributionService.RequireMember(ctx, *communityID, userID); err != nil {
			return nil, err
		}
	}

	postStatus := models.ParsePostStatus(status)

	post := models.NewPost(userID, title, content, postStatus)
	post.ID = utils.GenerateUUID()
	post.CreatedAt = time.Now()
	post.ViewsCount = 0
	post.CommunityID = communityID

	if err := s.repo.Create(ctx, &post); err != nil {
		return nil, err
	}

	// Xử lý upload danh sách hình ảnh/video đa phần từ form-data lên Cloudinary
	if len(files) > 0 && s.mediaService != nil {
		var mediaIDs []string
		for _, file := range files {
			uploadedMedia, err := s.mediaService.UploadMedia(ctx, userID, file)
			if err == nil && uploadedMedia != nil {
				mediaIDs = append(mediaIDs, uploadedMedia.ID)
			} else if err != nil {
				log.Printf("[Media Upload Error] Lỗi tải file lên: %v", err)
			}
		}
		if len(mediaIDs) > 0 {
			_ = s.repo.LinkMediaToPost(ctx, mediaIDs, post.ID)
		}
	}

	post.Media = []models.Media{}
	if s.mediaService != nil {
		if mediaMap, errM := s.mediaService.GetByPostIDs(ctx, []string{post.ID}); errM == nil {
			if m, ok := mediaMap[post.ID]; ok {
				post.Media = m
			}
		}
	}

	if author, err := s.repo.FetchPostAuthor(ctx, userID); err == nil {
		post.Username = author.Username
		post.DisplayName = author.DisplayName
		post.AvatarURI = author.AvatarURI
	}

	if err := s.tagService.ProcessPostHashtags(ctx, nil, post.ID, content); err != nil {
		log.Printf("[Hashtag Error] không thể lưu tag cho post %s: %v", post.ID, err)
	}

	if s.contributionService != nil && communityID != nil {
		go func() {
			if err := s.contributionService.ProcessChallengePost(ctx, *communityID, userID, content); err != nil {
				log.Printf("[Contribution Error] không thể xử lý challenge cho post %s: %v", post.ID, err)
			}
		}()
	}

	if communityID != nil && s.contributionService != nil {
		go func() {
			if err := s.contributionService.IncrementValidPosts(ctx, *communityID, userID); err != nil {
				log.Printf("[Contribution Error] không thể tăng valid_posts cho user %s trong community %s: %v", userID, *communityID, err)
			}
		}()
	}

	return &post, nil
}

func (s *postService) GetPostList(ctx context.Context, cursor string, pageSize int, userID string, filter string) ([]models.Post, string, error) {
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	filterFollowing := filter == "following"

	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	var cursorTier *int
	var cursorCreatedAt *time.Time
	var cursorID *string
	if cursor != "" {
		parts := strings.SplitN(cursor, "_", 3)
		if filterFollowing {
			if len(parts) == 2 {
				unixNano, err := strconv.ParseInt(parts[0], 10, 64)
				if err == nil {
					t := time.Unix(0, unixNano)
					cursorCreatedAt = &t
					cursorID = &parts[1]
				}
			}
		} else {
			if len(parts) == 3 {
				tier, err := strconv.Atoi(parts[0])
				if err == nil {
					unixNano, err := strconv.ParseInt(parts[1], 10, 64)
					if err == nil {
						t := time.Unix(0, unixNano)
						cursorTier = &tier
						cursorCreatedAt = &t
						cursorID = &parts[2]
					}
				}
			}
		}
	}

	posts, err := s.repo.FetchActive(ctx, pageSize, userIDPtr, cursorTier, cursorCreatedAt, cursorID, filterFollowing)
	if err != nil {
		return nil, "", err
	}

	if len(posts) > 0 {
		postIDs := make([]string, len(posts))
		for i, p := range posts {
			postIDs[i] = p.ID
		}

		likesMap, errL := s.repo.BatchCountLikes(ctx, postIDs)
		if errL == nil {
			for i := range posts {
				posts[i].LikesCount = likesMap[posts[i].ID]
			}
		}

		commentsMap, errC := s.repo.BatchCountComments(ctx, postIDs)
		if errC == nil {
			for i := range posts {
				posts[i].CommentsCount = commentsMap[posts[i].ID]
			}
		}

		sharesMap, errS := s.repo.BatchCountShares(ctx, postIDs)
		if errS == nil {
			for i := range posts {
				posts[i].SharesCount = sharesMap[posts[i].ID]
			}
		}

		if userID != "" {
			likedMap, errL := s.repo.BatchCheckLiked(ctx, userID, postIDs)
			if errL == nil {
				for i := range posts {
					posts[i].IsLiked = likedMap[posts[i].ID]
				}
			}

			savedMap, errS := s.repo.BatchCheckSaved(ctx, userID, postIDs)
			if errS == nil {
				for i := range posts {
					posts[i].IsSaved = savedMap[posts[i].ID]
				}
			}

			sharedMap, errSh := s.repo.BatchCheckShared(ctx, userID, postIDs)
			if errSh == nil {
				for i := range posts {
					posts[i].IsShared = sharedMap[posts[i].ID]
				}
			}
		}

		if s.mediaService != nil {
			mediaMap, errM := s.mediaService.GetByPostIDs(ctx, postIDs)
			if errM == nil {
				for i := range posts {
					if m, ok := mediaMap[posts[i].ID]; ok {
						posts[i].Media = m
					} else {
						posts[i].Media = []models.Media{}
					}
				}
			}
		}
	}

	var nextCursor string
	if len(posts) == pageSize {
		last := posts[len(posts)-1]
		if filterFollowing {
			nextCursor = fmt.Sprintf("%d_%s", last.CreatedAt.UnixNano(), last.ID)
		} else {
			tier := 1
			if last.IsFollowing {
				tier = 0
			}
			nextCursor = fmt.Sprintf("%d_%d_%s", tier, last.CreatedAt.UnixNano(), last.ID)
		}
	}

	return posts, nextCursor, nil
}

// Lấy danh sách bài viết đã lưu (Bookmark) của người dùng hiện tại theo con trỏ
func (s *postService) GetSavedPosts(ctx context.Context, userID string, cursor string, pageSize int) ([]models.Post, string, error) {
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var cursorCreatedAt *time.Time
	var cursorID *string
	if cursor != "" {
		parts := strings.SplitN(cursor, "_", 2)
		if len(parts) == 2 {
			unixNano, err := strconv.ParseInt(parts[0], 10, 64)
			if err == nil {
				t := time.Unix(0, unixNano)
				cursorCreatedAt = &t
				cursorID = &parts[1]
			}
		}
	}

	posts, err := s.repo.FetchSaved(ctx, userID, pageSize, cursorCreatedAt, cursorID)
	if err != nil {
		return nil, "", err
	}

	if len(posts) > 0 {
		postIDs := make([]string, len(posts))
		for i, p := range posts {
			postIDs[i] = p.ID
			posts[i].IsSaved = true
		}

		likesMap, errL := s.repo.BatchCountLikes(ctx, postIDs)
		if errL == nil {
			for i := range posts {
				posts[i].LikesCount = likesMap[posts[i].ID]
			}
		}

		commentsMap, errC := s.repo.BatchCountComments(ctx, postIDs)
		if errC == nil {
			for i := range posts {
				posts[i].CommentsCount = commentsMap[posts[i].ID]
			}
		}

		sharesMap, errS := s.repo.BatchCountShares(ctx, postIDs)
		if errS == nil {
			for i := range posts {
				posts[i].SharesCount = sharesMap[posts[i].ID]
			}
		}

		likedMap, errL := s.repo.BatchCheckLiked(ctx, userID, postIDs)
		if errL == nil {
			for i := range posts {
				posts[i].IsLiked = likedMap[posts[i].ID]
			}
		}

		sharedMap, errSh := s.repo.BatchCheckShared(ctx, userID, postIDs)
		if errSh == nil {
			for i := range posts {
				posts[i].IsShared = sharedMap[posts[i].ID]
			}
		}

		if s.mediaService != nil {
			mediaMap, errM := s.mediaService.GetByPostIDs(ctx, postIDs)
			if errM == nil {
				for i := range posts {
					if m, ok := mediaMap[posts[i].ID]; ok {
						posts[i].Media = m
					} else {
						posts[i].Media = []models.Media{}
					}
				}
			}
		}
	}

	var nextCursor string
	if len(posts) == pageSize {
		last := posts[len(posts)-1]
		if last.BookmarkID != nil && last.SavedAt != nil {
			nextCursor = fmt.Sprintf("%d_%s", last.SavedAt.UnixNano(), *last.BookmarkID)
		}
	}

	return posts, nextCursor, nil
}

func (s *postService) GetUserPosts(ctx context.Context, targetUserID, viewerID, cursor string, pageSize int) ([]models.Post, string, error) {
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var cursorCreatedAt *time.Time
	var cursorID *string
	if cursor != "" {
		parts := strings.SplitN(cursor, "_", 2)
		if len(parts) == 2 {
			unixNano, err := strconv.ParseInt(parts[0], 10, 64)
			if err == nil {
				t := time.Unix(0, unixNano)
				cursorCreatedAt = &t
				cursorID = &parts[1]
			}
		}
	}

	var viewerIDPtr *string
	if viewerID != "" {
		viewerIDPtr = &viewerID
	}

	posts, err := s.repo.FetchByUserID(ctx, targetUserID, viewerIDPtr, cursorCreatedAt, cursorID, pageSize)
	if err != nil {
		return nil, "", err
	}

	if len(posts) > 0 {
		postIDs := make([]string, len(posts))
		for i, p := range posts {
			postIDs[i] = p.ID
		}

		likesMap, errL := s.repo.BatchCountLikes(ctx, postIDs)
		if errL == nil {
			for i := range posts {
				posts[i].LikesCount = likesMap[posts[i].ID]
			}
		}

		commentsMap, errC := s.repo.BatchCountComments(ctx, postIDs)
		if errC == nil {
			for i := range posts {
				posts[i].CommentsCount = commentsMap[posts[i].ID]
			}
		}

		sharesMap, errS := s.repo.BatchCountShares(ctx, postIDs)
		if errS == nil {
			for i := range posts {
				posts[i].SharesCount = sharesMap[posts[i].ID]
			}
		}

		if viewerID != "" {
			likedMap, errL := s.repo.BatchCheckLiked(ctx, viewerID, postIDs)
			if errL == nil {
				for i := range posts {
					posts[i].IsLiked = likedMap[posts[i].ID]
				}
			}

			savedMap, errS := s.repo.BatchCheckSaved(ctx, viewerID, postIDs)
			if errS == nil {
				for i := range posts {
					posts[i].IsSaved = savedMap[posts[i].ID]
				}
			}

			sharedMap, errSh := s.repo.BatchCheckShared(ctx, viewerID, postIDs)
			if errSh == nil {
				for i := range posts {
					posts[i].IsShared = sharedMap[posts[i].ID]
				}
			}
		}

		if s.mediaService != nil {
			mediaMap, errM := s.mediaService.GetByPostIDs(ctx, postIDs)
			if errM == nil {
				for i := range posts {
					if m, ok := mediaMap[posts[i].ID]; ok {
						posts[i].Media = m
					} else {
						posts[i].Media = []models.Media{}
					}
				}
			}
		}
	}

	var nextCursor string
	if len(posts) == pageSize {
		last := posts[len(posts)-1]
		nextCursor = fmt.Sprintf("%d_%s", last.CreatedAt.UnixNano(), last.ID)
	}

	return posts, nextCursor, nil
}

func (s *postService) GetPostDetail(ctx context.Context, postID string) (*models.Post, error) {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodePostNotFound)
	}

	if post.Status == models.PostStatusHidden || post.Status == models.PostStatusPrivate {
		return nil, errorsapp.New(errorsapp.ErrCodePostHiddenOrPrivate)
	}

	_ = s.repo.IncrementViewsCount(ctx, postID)
	post.ViewsCount++

	post.Media = []models.Media{}
	if s.mediaService != nil {
		if mediaMap, errM := s.mediaService.GetByPostIDs(ctx, []string{postID}); errM == nil {
			if m, ok := mediaMap[postID]; ok {
				post.Media = m
			}
		}
	}

	return post, nil
}

func (s *postService) ReactPost(ctx context.Context, userID, postID, emojiID string) (string, string, error) {
	if err := s.validation.ValidateReactPost(emojiID); err != nil {
		return "", "", err
	}

	emoji, err := s.repo.FindEmojiByID(ctx, emojiID)
	if err != nil {
		return "", "", errorsapp.New(errorsapp.ErrCodeEmojiNotFound)
	}

	existingReaction, err := s.repo.FindReaction(ctx, userID, postID, emojiID)

	if err == nil && existingReaction != nil {
		if errDelete := s.repo.DeleteReaction(ctx, existingReaction.ID); errDelete != nil {
			return "", "", errDelete
		}

		if post, err := s.repo.FindByID(ctx, postID); err == nil && post != nil && post.UserID != userID {
			if s.contributionService != nil && post.CommunityID != nil {
				go func() {
					if err := s.contributionService.DecrementPositiveReactions(ctx, *post.CommunityID, post.UserID); err != nil {
						log.Printf("[Contribution Error] không thể giảm positive_reactions cho user %s: %v", post.UserID, err)
					}
				}()
			}
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

		if s.contributionService != nil && post.CommunityID != nil {
			go func() {
				if err := s.contributionService.IncrementPositiveReactions(ctx, *post.CommunityID, post.UserID); err != nil {
					log.Printf("[Contribution Error] không thể tăng positive_reactions cho user %s: %v", post.UserID, err)
				}
			}()
		}
	}

	return "reacted", emoji.Code, nil
}

func (s *postService) CreateComment(ctx context.Context, userID, postID string, parentID *string, content string) ([]models.Comment, error) {
	if err := s.validation.ValidateCreateComment(content); err != nil {
		return nil, err
	}

	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return nil, errorsapp.New(errorsapp.ErrCodePostNotFound)
	}

	if post.Status == models.PostStatusHidden || post.Status == models.PostStatusPrivate {
		return nil, errorsapp.New(errorsapp.ErrCodeCommentHiddenPrivate)
	}

	var parentComment *models.Comment
	if parentID != nil && *parentID != "" {
		parentComment, err = s.repo.FindCommentByID(ctx, *parentID)
		if err != nil || parentComment == nil {
			return nil, errorsapp.New(errorsapp.ErrCodeCommentNotFound)
		}
		if parentComment.PostID != postID {
			return nil, errorsapp.New(errorsapp.ErrCodeCommentWrongPost)
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

	if err := s.tagService.ProcessCommentHashtags(ctx, nil, postID, comment.ID, content); err != nil {
		log.Printf("[Hashtag Error] không thể lưu tag cho comment %s: %v", comment.ID, err)
	}

	if s.contributionService != nil && post.CommunityID != nil {
		go func() {
			if err := s.contributionService.IncrementQualityComments(ctx, *post.CommunityID, userID); err != nil {
				log.Printf("[Contribution Error] không thể tăng quality_comments cho user %s: %v", userID, err)
			}
		}()
	}

	if post.UserID != userID {
		postIDPtr := postID
		commentIDPtr := comment.ID
		s.notifService.Create(ctx, post.UserID, &userID, models.NotificationTypeComment, "đã bình luận bài viết của bạn", &postIDPtr, nil, &commentIDPtr)
	}

	if parentComment != nil && parentComment.UserID != userID && parentComment.UserID != post.UserID {
		postIDPtr := postID
		commentIDPtr := comment.ID
		s.notifService.Create(ctx, parentComment.UserID, &userID, models.NotificationTypeComment, "đã trả lời bình luận của bạn", &postIDPtr, nil, &commentIDPtr)
	}

	return s.repo.FindCommentsByPostID(ctx, postID)
}

func (s *postService) GetCommentList(ctx context.Context, postID string, page, pageSize int) ([]models.Comment, int64, error) {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil || post.Status == models.PostStatusHidden || post.Status == models.PostStatusPrivate {
		return nil, 0, errorsapp.New(errorsapp.ErrCodePostNotAccessible)
	}

	page, pageSize = s.validation.NormalizePagination(page, pageSize)
	offset := (page - 1) * pageSize

	total, err := s.repo.CountCommentsByPostID(ctx, postID)
	if err != nil {
		return nil, 0, err
	}

	comments, err := s.repo.FetchCommentsByPostID(ctx, postID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func (s *postService) SharePost(ctx context.Context, userID, postID, content string) error {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil || post.Status == models.PostStatusHidden || post.Status == models.PostStatusPrivate {
		return errorsapp.New(errorsapp.ErrCodePostNotShareable)
	}

	if post.UserID == userID {
		return errorsapp.New(errorsapp.ErrCodePostCannotShareOwn)
	}

	existing, err := s.repo.FindShareByUser(ctx, userID, postID)
	if err == nil && existing != nil {
		return errorsapp.New(errorsapp.ErrCodePostAlreadyShared)
	}

	share := models.NewPostShare(userID, postID, content)
	share.ID = utils.GenerateUUID()
	share.CreatedAt = time.Now()

	if err := s.repo.CreateShare(ctx, share); err != nil {
		return err
	}

	if post.UserID != userID {
		s.notifService.Create(ctx, post.UserID, &userID, models.NotificationTypeShare, "đã chia sẻ bài viết của bạn", &postID, nil, nil)
	}

	return nil
}

// Bấm lưu bài viết lần nữa hệ thống tự hiểu và xóa (Toggle Bookmark)
func (s *postService) SavePost(ctx context.Context, userID, postID string) (string, error) {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil || post.Status == models.PostStatusHidden || post.Status == models.PostStatusPrivate {
		return "", errorsapp.New(errorsapp.ErrCodePostNotFound)
	}

	if post.UserID == userID {
		return "", errorsapp.New(errorsapp.ErrCodePostCannotSaveOwn)
	}

	existingBookmark, err := s.repo.FindBookmark(ctx, userID, postID)
	if err == nil && existingBookmark != nil {
		// Đã lưu từ trước -> Thực hiện xóa
		if errDelete := s.repo.DeleteBookmark(ctx, existingBookmark.ID); errDelete != nil {
			return "", errDelete
		}
		return "removed", nil
	}

	// Chưa từng lưu -> Tạo mới
	bookmark := models.Bookmark{
		ID:        utils.GenerateUUID(),
		UserID:    userID,
		PostID:    postID,
		CreatedAt: time.Now(),
	}

	if errCreate := s.repo.CreateSave(ctx, bookmark); errCreate != nil {
		return "", errCreate
	}
	return "saved", nil
}

// Xóa bài viết và thông báo tới toàn bộ những tài khoản đã lưu bài viết này
func (s *postService) DeletePost(ctx context.Context, userID, postID string) error {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return errorsapp.New(errorsapp.ErrCodePostNotFound)
	}

	if post.UserID != userID {
		return errorsapp.New(errorsapp.ErrCodePostCannotDeleteOthers)
	}

	bookmarkedUserIDs, err := s.repo.DeletePostWithAssociations(ctx, postID)
	if err != nil {
		return err
	}

	// SỬA TẠI ĐÂY: Đẩy luồng thông báo bất đồng bộ an toàn tuyệt đối (Đã sửa kiểu dữ liệu ID về dạng string)
	go func(targetIDs []string, currentUserID string) {
		// Dùng Context độc lập để luồng ngầm không bị hủy khi HTTP Request của Gin kết thúc
		bgCtx := context.Background()

		for _, targetUserID := range targetIDs {
			if targetUserID != currentUserID { // Không gửi cho chính chủ bài viết

				// Tạo biến cục bộ trong mỗi vòng lặp để lấy con trỏ an toàn, chống dữ liệu bị ghi đè (Data Race)
				senderID := currentUserID

				s.notifService.Create(
					bgCtx,
					targetUserID,
					&senderID,
					models.NotificationTypeMessage,
					"Một bài viết trong danh sách dấu trang (Bookmark) của bạn đã bị chủ sở hữu xóa.",
					nil, nil, nil,
				)
			}
		}
	}(bookmarkedUserIDs, userID)

	return nil
}

// Tìm bài viết theo từ khóa Hashtag
func (s *postService) GetPostsByHashtag(ctx context.Context, hashtag string, page, pageSize int) ([]models.Post, error) {
	page, pageSize = s.validation.NormalizePagination(page, pageSize)
	offset := (page - 1) * pageSize
	postIDs, err := s.tagService.GetPostIDsByHashtag(ctx, hashtag)
	if err != nil || len(postIDs) == 0 {
		return []models.Post{}, nil
	}
	posts, err := s.repo.FetchByIDs(ctx, postIDs, pageSize, offset)
	return posts, err
}

func (s *postService) ListEmojis(ctx context.Context) ([]models.Emoji, error) {
	return s.repo.ListEmojis(ctx)
}
