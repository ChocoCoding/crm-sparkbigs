package storage

import (
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"gorm.io/gorm"
)

type mysqlDealRepository struct {
	db *gorm.DB
}

func NewMysqlDealRepository(db *gorm.DB) ports.DealRepository {
	return &mysqlDealRepository{db: db}
}

func (r *mysqlDealRepository) Create(deal *domain.Deal) error {
	return r.db.Create(deal).Error
}

func (r *mysqlDealRepository) FindByID(id uint) (*domain.Deal, error) {
	var deal domain.Deal
	err := r.db.First(&deal, id).Error
	return &deal, err
}

func (r *mysqlDealRepository) FindByUserID(userID uint, offset, limit int) ([]domain.Deal, int64, error) {
	var deals []domain.Deal
	var total int64

	r.db.Model(&domain.Deal{}).Where("user_id = ?", userID).Count(&total)
	err := r.db.Where("user_id = ?", userID).Offset(offset).Limit(limit).Find(&deals).Error
	return deals, total, err
}

func (r *mysqlDealRepository) FindByContactID(contactID uint) ([]domain.Deal, error) {
	var deals []domain.Deal
	err := r.db.Where("contact_id = ?", contactID).Find(&deals).Error
	return deals, err
}

func (r *mysqlDealRepository) Update(deal *domain.Deal) error {
	return r.db.Save(deal).Error
}

func (r *mysqlDealRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Deal{}, id).Error
}
