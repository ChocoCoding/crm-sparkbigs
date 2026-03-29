package storage

import (
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"gorm.io/gorm"
)

type mysqlRefreshTokenRepository struct {
	db *gorm.DB
}

func NewMysqlRefreshTokenRepository(db *gorm.DB) ports.RefreshTokenRepository {
	return &mysqlRefreshTokenRepository{db: db}
}

func (r *mysqlRefreshTokenRepository) Create(token *domain.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *mysqlRefreshTokenRepository) FindByToken(token string) (*domain.RefreshToken, error) {
	var rt domain.RefreshToken
	err := r.db.Where("token = ?", token).First(&rt).Error
	return &rt, err
}

func (r *mysqlRefreshTokenRepository) RevokeByUserID(userID uint) error {
	return r.db.Model(&domain.RefreshToken{}).
		Where("user_id = ? AND revoked = false", userID).
		Update("revoked", true).Error
}

func (r *mysqlRefreshTokenRepository) RevokeByToken(token string) error {
	return r.db.Model(&domain.RefreshToken{}).
		Where("token = ?", token).
		Update("revoked", true).Error
}

func (r *mysqlRefreshTokenRepository) DeleteExpired() error {
	return r.db.Where("expires_at < NOW()").Delete(&domain.RefreshToken{}).Error
}
