package storage

import (
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type mysqlSettingRepository struct{ db *gorm.DB }

func NewMysqlSettingRepository(db *gorm.DB) ports.SettingRepository {
	return &mysqlSettingRepository{db: db}
}

// Upsert: INSERT ... ON DUPLICATE KEY UPDATE usando el índice único (user_id, key).
func (r *mysqlSettingRepository) Upsert(setting *domain.Setting) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "label", "input_type", "category", "updated_at"}),
	}).Create(setting).Error
}

func (r *mysqlSettingRepository) FindByUserID(userID uint) ([]domain.Setting, error) {
	var settings []domain.Setting
	err := r.db.Where("user_id = ?", userID).Order("category, `key`").Find(&settings).Error
	return settings, err
}

func (r *mysqlSettingRepository) FindByCategory(userID uint, category string) ([]domain.Setting, error) {
	var settings []domain.Setting
	err := r.db.Where("user_id = ? AND category = ?", userID, category).Order("`key`").Find(&settings).Error
	return settings, err
}

func (r *mysqlSettingRepository) FindByKey(userID uint, key string) (*domain.Setting, error) {
	var setting domain.Setting
	err := r.db.Where("user_id = ? AND `key` = ?", userID, key).First(&setting).Error
	return &setting, err
}

func (r *mysqlSettingRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Setting{}, id).Error
}
