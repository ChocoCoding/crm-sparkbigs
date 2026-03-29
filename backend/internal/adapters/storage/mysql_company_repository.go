package storage

import (
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"gorm.io/gorm"
)

type mysqlCompanyRepository struct {
	db *gorm.DB
}

func NewMysqlCompanyRepository(db *gorm.DB) ports.CompanyRepository {
	return &mysqlCompanyRepository{db: db}
}

func (r *mysqlCompanyRepository) Create(company *domain.Company) error {
	return r.db.Create(company).Error
}

func (r *mysqlCompanyRepository) FindByID(id uint) (*domain.Company, error) {
	var company domain.Company
	err := r.db.First(&company, id).Error
	return &company, err
}

func (r *mysqlCompanyRepository) FindByUserID(userID uint, offset, limit int) ([]domain.Company, int64, error) {
	var companies []domain.Company
	var total int64

	base := r.db.Model(&domain.Company{}).Where("user_id = ?", userID)
	base.Count(&total)
	err := base.Order("created_at DESC").Offset(offset).Limit(limit).Find(&companies).Error
	return companies, total, err
}

func (r *mysqlCompanyRepository) Update(company *domain.Company) error {
	return r.db.Save(company).Error
}

func (r *mysqlCompanyRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Company{}, id).Error
}

func (r *mysqlCompanyRepository) Search(userID uint, query string, offset, limit int) ([]domain.Company, int64, error) {
	var companies []domain.Company
	var total int64

	like := "%" + query + "%"
	base := r.db.Model(&domain.Company{}).
		Where("user_id = ? AND (name LIKE ? OR sector LIKE ? OR address LIKE ?)", userID, like, like, like)

	base.Count(&total)
	err := base.Order("name ASC").Offset(offset).Limit(limit).Find(&companies).Error
	return companies, total, err
}
