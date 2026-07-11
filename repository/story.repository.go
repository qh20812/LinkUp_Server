package repository

import (
	"linkup/models"
	"time"

	"gorm.io/gorm"
)

type StoryView struct {
	StoryID  string    `gorm:"primaryKey;column:story_id;type:varchar(36)"`
	UserID   string    `gorm:"primaryKey;column:user_id;type:varchar(36)"`
	ViewedAt time.Time `gorm:"column:viewed_at"`
}

type StoryRepository interface {
	Create(story *models.Story) error
	FindByID(id string) (*models.Story, error)
	GetActiveStories() ([]models.Story, error)
	LogView(storyID, viewerID string) error
	CountViews(storyID string) (int64, error)
	CountInteractionsByType(storyID string, interType string) (int64, error)
	CheckEmojiExists(emojiID string) (bool, error)
	CreateInteract(interact *models.StoryInteract) error
}

type storyRepository struct {
	db *gorm.DB
}

func NewStoryRepository(db *gorm.DB) StoryRepository {
	return &storyRepository{db: db}
}

func (r *storyRepository) Create(story *models.Story) error {
	return r.db.Create(story).Error
}

func (r *storyRepository) FindByID(id string) (*models.Story, error) {
	var story models.Story
	err := r.db.First(&story, "id = ?", id).Error
	return &story, err
}

// Lấy danh sách các story còn hạn (chưa quá 24h) xếp theo thời gian mới nhất
func (r *storyRepository) GetActiveStories() ([]models.Story, error) {
	var stories []models.Story
	err := r.db.Where("expires_at > ?", time.Now()).Order("created_at DESC").Find(&stories).Error
	return stories, err
}

// Ghi nhận lượt xem (Dùng câu lệnh Save/Upsert để tránh lỗi trùng lặp khi xem lại nhiều lần)
func (r *storyRepository) LogView(storyID, viewerID string) error {
	view := StoryView{
		StoryID:  storyID,
		UserID:   viewerID,
		ViewedAt: time.Now(),
	}
	return r.db.Table("story_views").Create(&view).Error
}

func (r *storyRepository) CountViews(storyID string) (int64, error) {
	var count int64
	err := r.db.Table("story_views").Where("story_id = ?", storyID).Count(&count).Error
	return count, err
}

func (r *storyRepository) CountInteractionsByType(storyID string, interType string) (int64, error) {
	var count int64
	// Đếm trực tiếp dựa trên Model StoryInteract mới hướng tới bảng story_interacts
	err := r.db.Model(&models.StoryInteract{}).Where("story_id = ? AND type = ?", storyID, interType).Count(&count).Error
	return count, err
}

func (r *storyRepository) CheckEmojiExists(emojiID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Emoji{}).Where("id = ?", emojiID).Count(&count).Error
	return count > 0, err
}

func (r *storyRepository) CreateInteract(interact *models.StoryInteract) error {
	return r.db.Create(interact).Error
}
