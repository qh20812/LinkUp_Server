package repository

import (
	"linkup/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StoryView struct {
	StoryID  string    `gorm:"primaryKey;column:story_id;type:varchar(36)"`
	UserID   string    `gorm:"primaryKey;column:viewer_id;type:varchar(36)"`
	ViewedAt time.Time `gorm:"column:viewed_at"`
}

type StoryRepository interface {
	Create(story *models.Story) error
	FindByID(id string) (*models.Story, error)
	GetActiveStories() ([]models.Story, error)
	HasActiveStoryByUserID(userID string) (bool, error)
	GetActiveStoriesByUserID(userID string) ([]models.Story, error)
	HasUserViewed(storyID, viewerID string) (bool, error)
	LogView(storyID, viewerID string) error
	CountViews(storyID string) (int64, error)
	CountInteractionsByType(storyID string, interType string) (int64, error)
	CheckEmojiExists(emojiID string) (bool, error)
	CreateInteract(interact *models.StoryInteract) error
	FindReactByUser(storyID, userID string) (*models.StoryInteract, error)
	UpdateInteract(interact *models.StoryInteract) error
	GetViewersDetails(storyID string) ([]models.StoryView, []models.StoryInteract, error)
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

func (r *storyRepository) GetActiveStories() ([]models.Story, error) {
	var stories []models.Story
	err := r.db.Where("expires_at > ?", time.Now()).Order("created_at DESC").Find(&stories).Error
	return stories, err
}

func (r *storyRepository) HasActiveStoryByUserID(userID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Story{}).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Count(&count).Error
	return count > 0, err
}

func (r *storyRepository) GetActiveStoriesByUserID(userID string) ([]models.Story, error) {
	var stories []models.Story
	err := r.db.Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("created_at ASC").Find(&stories).Error
	return stories, err
}

// Kiểm tra xem user đã xem story này trước đó chưa
func (r *storyRepository) HasUserViewed(storyID, viewerID string) (bool, error) {
	var count int64
	err := r.db.Table("story_views").Where("story_id = ? AND viewer_id = ?", storyID, viewerID).Count(&count).Error
	return count > 0, err
}

func (r *storyRepository) LogView(storyID, viewerID string) error {
	view := StoryView{
		StoryID:  storyID,
		UserID:   viewerID,
		ViewedAt: time.Now(),
	}

	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "story_id"}, {Name: "viewer_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"viewed_at"}),
	}).Create(&view).Error
}

func (r *storyRepository) CountViews(storyID string) (int64, error) {
	var count int64
	err := r.db.Table("story_views").Where("story_id = ?", storyID).Count(&count).Error
	return count, err
}

func (r *storyRepository) CountInteractionsByType(storyID string, interType string) (int64, error) {
	var count int64
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

func (r *storyRepository) FindReactByUser(storyID, userID string) (*models.StoryInteract, error) {
	var inter models.StoryInteract
	err := r.db.Where("story_id = ? AND user_id = ? AND type = 'react'", storyID, userID).First(&inter).Error
	return &inter, err
}

func (r *storyRepository) UpdateInteract(interact *models.StoryInteract) error {
	return r.db.Save(interact).Error
}

func (r *storyRepository) GetViewersDetails(storyID string) ([]models.StoryView, []models.StoryInteract, error) {
	var views []models.StoryView
	var interacts []models.StoryInteract

	errViews := r.db.Table("story_views").Where("story_id = ?", storyID).Order("viewed_at DESC").Find(&views).Error
	errInteracts := r.db.Where("story_id = ?", storyID).Find(&interacts).Error

	if errViews != nil {
		return nil, nil, errViews
	}
	return views, interacts, errInteracts
}
