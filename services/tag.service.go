package services

import (
	"context"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"

	"gorm.io/gorm"
)

type TagService struct {
	tagRepo *repository.TagRepository
}

func NewTagService(tagRepo *repository.TagRepository) *TagService {
	return &TagService{tagRepo: tagRepo}
}

func (s *TagService) ProcessPostHashtags(ctx context.Context, tx *gorm.DB, postID, content string) error {
	hashtagNames := utils.ExtractHashtags(content)
	if len(hashtagNames) == 0 {
		return nil
	}

	var tags []models.Tag
	for _, name := range hashtagNames {
		tag := models.NewTag(postID, nil, models.TagTypeHashtag, nil, name)
		tag.ID = utils.GenerateUUID()
		tags = append(tags, tag)
	}

	return s.tagRepo.CreateTagsInTx(ctx, tx, tags)
}

func (s *TagService) ProcessCommentHashtags(ctx context.Context, tx *gorm.DB, postID, commentID, content string) error {
	hashtagNames := utils.ExtractHashtags(content)
	if len(hashtagNames) == 0 {
		return nil
	}

	var tags []models.Tag
	for _, name := range hashtagNames {
		tag := models.NewTag(postID, &commentID, models.TagTypeHashtag, nil, name)
		tag.ID = utils.GenerateUUID()
		tags = append(tags, tag)
	}

	return s.tagRepo.CreateTagsInTx(ctx, tx, tags)
}

func (s *TagService) GetPostIDsByHashtag(ctx context.Context, hashtagName string) ([]string, error) {
	return s.tagRepo.GetPostIDsByHashtag(ctx, hashtagName)
}
