package storage

import (
	"time"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"gorm.io/gorm"
)

type mysqlSubscriptionRepository struct{ db *gorm.DB }

func NewMysqlSubscriptionRepository(db *gorm.DB) ports.SubscriptionRepository {
	return &mysqlSubscriptionRepository{db: db}
}

func (r *mysqlSubscriptionRepository) Create(sub *domain.Subscription) error {
	return r.db.Create(sub).Error
}

func (r *mysqlSubscriptionRepository) FindByID(id uint) (*domain.Subscription, error) {
	var sub domain.Subscription
	err := r.db.Preload("Company").First(&sub, id).Error
	return &sub, err
}

func (r *mysqlSubscriptionRepository) FindByUserID(userID uint, offset, limit int) ([]domain.Subscription, int64, error) {
	var subs []domain.Subscription
	var total int64
	r.db.Model(&domain.Subscription{}).Where("user_id = ? AND deleted_at IS NULL", userID).Count(&total)
	err := r.db.Preload("Company").
		Where("user_id = ?", userID).
		Order("renewal_date ASC").
		Offset(offset).Limit(limit).
		Find(&subs).Error
	return subs, total, err
}

func (r *mysqlSubscriptionRepository) FindByCompanyID(companyID, userID uint) ([]domain.Subscription, error) {
	var subs []domain.Subscription
	err := r.db.Where("company_id = ? AND user_id = ?", companyID, userID).
		Order("renewal_date ASC").Find(&subs).Error
	return subs, err
}

func (r *mysqlSubscriptionRepository) FindExpiringSoon(userID uint, days int) ([]domain.Subscription, error) {
	var subs []domain.Subscription
	cutoff := time.Now().AddDate(0, 0, days)
	err := r.db.Preload("Company").
		Where("user_id = ? AND status = 'active' AND renewal_date IS NOT NULL AND renewal_date <= ? AND deleted_at IS NULL",
			userID, cutoff).
		Order("renewal_date ASC").
		Find(&subs).Error
	return subs, err
}

func (r *mysqlSubscriptionRepository) Update(sub *domain.Subscription) error {
	return r.db.Save(sub).Error
}

func (r *mysqlSubscriptionRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Subscription{}, id).Error
}
