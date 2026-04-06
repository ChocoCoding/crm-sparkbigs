package storage

import (
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"gorm.io/gorm"
)

type mysqlScrapedLeadRepository struct {
	db *gorm.DB
}

func NewMysqlScrapedLeadRepository(db *gorm.DB) ports.ScrapedLeadRepository {
	return &mysqlScrapedLeadRepository{db: db}
}

func (r *mysqlScrapedLeadRepository) Create(lead *domain.ScrapedLead) error {
	return r.db.Create(lead).Error
}

func (r *mysqlScrapedLeadRepository) FindByID(id uint) (*domain.ScrapedLead, error) {
	var lead domain.ScrapedLead
	err := r.db.First(&lead, id).Error
	return &lead, err
}

func (r *mysqlScrapedLeadRepository) FindByJobID(jobID string) ([]domain.ScrapedLead, error) {
	var leads []domain.ScrapedLead
	err := r.db.Where("job_id = ?", jobID).Order("score_lead DESC").Find(&leads).Error
	return leads, err
}

func (r *mysqlScrapedLeadRepository) FindAll(offset, limit int) ([]domain.ScrapedLead, int64, error) {
	var leads []domain.ScrapedLead
	var total int64

	base := r.db.Model(&domain.ScrapedLead{})
	base.Count(&total)
	err := base.Order("score_lead DESC").Offset(offset).Limit(limit).Find(&leads).Error
	return leads, total, err
}

func (r *mysqlScrapedLeadRepository) FindExisting(googlePlaceID, name, city string) (*domain.ScrapedLead, error) {
	var lead domain.ScrapedLead
	err := r.db.
		Where("google_place_id = ? OR (nombre_empresa = ? AND ciudad = ?)", googlePlaceID, name, city).
		Order("created_at DESC").
		First(&lead).Error
	if err != nil {
		return nil, err
	}
	return &lead, nil
}

func (r *mysqlScrapedLeadRepository) UpdateEstado(id uint, estado string) error {
	return r.db.Model(&domain.ScrapedLead{}).Where("id = ?", id).Update("estado", estado).Error
}

func (r *mysqlScrapedLeadRepository) Delete(id uint) error {
	return r.db.Delete(&domain.ScrapedLead{}, id).Error
}
