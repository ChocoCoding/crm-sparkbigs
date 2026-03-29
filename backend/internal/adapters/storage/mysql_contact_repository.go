package storage

import (
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"gorm.io/gorm"
)

type mysqlContactRepository struct {
	db *gorm.DB
}

func NewMysqlContactRepository(db *gorm.DB) ports.ContactRepository {
	return &mysqlContactRepository{db: db}
}

func (r *mysqlContactRepository) Create(contact *domain.Contact) error {
	return r.db.Create(contact).Error
}

func (r *mysqlContactRepository) FindByID(id uint) (*domain.Contact, error) {
	var contact domain.Contact
	err := r.db.Preload("Company").First(&contact, id).Error
	return &contact, err
}

func (r *mysqlContactRepository) FindByUserID(userID uint, offset, limit int) ([]domain.Contact, int64, error) {
	var contacts []domain.Contact
	var total int64

	r.db.Model(&domain.Contact{}).Where("user_id = ? AND deleted_at IS NULL", userID).Count(&total)
	err := r.db.
		Preload("Company").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&contacts).Error

	return contacts, total, err
}

func (r *mysqlContactRepository) FindByCompanyID(companyID, userID uint, offset, limit int) ([]domain.Contact, int64, error) {
	var contacts []domain.Contact
	var total int64

	r.db.Model(&domain.Contact{}).
		Where("company_id = ? AND user_id = ? AND deleted_at IS NULL", companyID, userID).
		Count(&total)
	err := r.db.
		Preload("Company").
		Where("company_id = ? AND user_id = ?", companyID, userID).
		Order("name ASC").
		Offset(offset).Limit(limit).
		Find(&contacts).Error

	return contacts, total, err
}

func (r *mysqlContactRepository) Update(contact *domain.Contact) error {
	return r.db.Save(contact).Error
}

func (r *mysqlContactRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Contact{}, id).Error
}

func (r *mysqlContactRepository) Search(userID uint, query string, offset, limit int) ([]domain.Contact, int64, error) {
	var contacts []domain.Contact
	var total int64
	like := "%" + query + "%"

	base := r.db.Model(&domain.Contact{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Where("name LIKE ? OR email LIKE ? OR position LIKE ?", like, like, like)

	base.Count(&total)
	err := base.
		Preload("Company").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&contacts).Error

	return contacts, total, err
}
