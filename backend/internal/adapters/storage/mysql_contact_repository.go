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
	err := r.db.First(&contact, id).Error
	return &contact, err
}

func (r *mysqlContactRepository) FindByUserID(userID uint, offset, limit int) ([]domain.Contact, int64, error) {
	var contacts []domain.Contact
	var total int64

	r.db.Model(&domain.Contact{}).Where("user_id = ?", userID).Count(&total)
	err := r.db.Where("user_id = ?", userID).Offset(offset).Limit(limit).Find(&contacts).Error
	return contacts, total, err
}

func (r *mysqlContactRepository) Update(contact *domain.Contact) error {
	return r.db.Save(contact).Error
}

func (r *mysqlContactRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Contact{}, id).Error
}
