package storage

import (
	"errors"
	"time"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"gorm.io/gorm"
)

type mysqlAPIKeyRepository struct {
	db *gorm.DB
}

func NewMysqlAPIKeyRepository(db *gorm.DB) ports.APIKeyRepository {
	return &mysqlAPIKeyRepository{db: db}
}

func (r *mysqlAPIKeyRepository) Create(key *domain.APIKey) error {
	return r.db.Create(key).Error
}

func (r *mysqlAPIKeyRepository) FindByPrefix(prefix string) (*domain.APIKey, error) {
	var key domain.APIKey
	err := r.db.Where("key_prefix = ? AND is_active = true", prefix).First(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &key, err
}

func (r *mysqlAPIKeyRepository) FindByUserID(userID uint) ([]domain.APIKey, error) {
	var keys []domain.APIKey
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&keys).Error
	return keys, err
}

func (r *mysqlAPIKeyRepository) Revoke(id, userID uint) error {
	result := r.db.Model(&domain.APIKey{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("api key no encontrada o no autorizada")
	}
	return nil
}

func (r *mysqlAPIKeyRepository) UpdateLastUsed(id uint) error {
	now := time.Now()
	return r.db.Model(&domain.APIKey{}).Where("id = ?", id).Update("last_used_at", now).Error
}
