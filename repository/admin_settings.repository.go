package repository

import (
	"context"
	"errors"

	"linkup/models"

	"gorm.io/gorm"
)

type AdminSettingsRepository struct {
	db *gorm.DB
}

func NewAdminSettingsRepository(db *gorm.DB) *AdminSettingsRepository {
	return &AdminSettingsRepository{db: db}
}

func (r *AdminSettingsRepository) DB() *gorm.DB {
	return r.db
}

func (r *AdminSettingsRepository) GetByKey(ctx context.Context, key string) (*models.SystemConfig, error) {
	var config models.SystemConfig
	err := r.db.WithContext(ctx).Where("`key` = ?", key).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

func (r *AdminSettingsRepository) GetAll(ctx context.Context) ([]models.SystemConfig, error) {
	var configs []models.SystemConfig
	err := r.db.WithContext(ctx).Find(&configs).Error
	return configs, err
}

func (r *AdminSettingsRepository) Upsert(ctx context.Context, key, value string) error {
	return r.db.WithContext(ctx).
		Table("system_configs").
		Where("`key` = ?", key).
		Assign(map[string]interface{}{"value": value}).
		FirstOrCreate(&models.SystemConfig{Key: key, Value: value}).Error
}

func (r *AdminSettingsRepository) UpsertBatch(ctx context.Context, settings map[string]string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for key, value := range settings {
			if err := tx.
				Table("system_configs").
				Where("`key` = ?", key).
				Assign(map[string]interface{}{"value": value}).
				FirstOrCreate(&models.SystemConfig{Key: key, Value: value}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}