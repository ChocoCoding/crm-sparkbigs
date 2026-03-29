package storage

import (
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"gorm.io/gorm"
)

type mysqlLicenseRepository struct {
	db *gorm.DB
}

func NewMysqlLicenseRepository(db *gorm.DB) ports.LicenseRepository {
	return &mysqlLicenseRepository{db: db}
}

func (r *mysqlLicenseRepository) Create(license *domain.License) error {
	return r.db.Create(license).Error
}

func (r *mysqlLicenseRepository) FindByUserID(userID uint) (*domain.License, error) {
	var license domain.License
	err := r.db.Where("user_id = ?", userID).First(&license).Error
	return &license, err
}

func (r *mysqlLicenseRepository) Update(license *domain.License) error {
	return r.db.Save(license).Error
}
