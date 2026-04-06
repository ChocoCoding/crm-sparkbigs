package storage

import (
	"time"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"gorm.io/gorm"
)

type mysqlScrapeJobRepository struct {
	db *gorm.DB
}

func NewMysqlScrapeJobRepository(db *gorm.DB) ports.ScrapeJobRepository {
	return &mysqlScrapeJobRepository{db: db}
}

func (r *mysqlScrapeJobRepository) Create(job *domain.ScrapeJob) error {
	return r.db.Create(job).Error
}

func (r *mysqlScrapeJobRepository) FindByID(id string) (*domain.ScrapeJob, error) {
	var job domain.ScrapeJob
	err := r.db.First(&job, "id = ?", id).Error
	return &job, err
}

func (r *mysqlScrapeJobRepository) FindAll(offset, limit int) ([]domain.ScrapeJob, int64, error) {
	var jobs []domain.ScrapeJob
	var total int64

	base := r.db.Model(&domain.ScrapeJob{})
	base.Count(&total)

	// Subconsulta para contar leads y avg score por job
	err := r.db.Raw(`
		SELECT j.*,
			(SELECT COUNT(*) FROM scraped_leads WHERE job_id = j.id AND deleted_at IS NULL) as lead_count,
			(SELECT ROUND(AVG(score_lead), 1) FROM scraped_leads WHERE job_id = j.id AND deleted_at IS NULL) as avg_score
		FROM scrape_jobs j
		ORDER BY j.created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset).Scan(&jobs).Error

	return jobs, total, err
}

func (r *mysqlScrapeJobRepository) UpdateStatus(id, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == "completed" || status == "error" {
		now := time.Now()
		updates["completed_at"] = &now
	}
	return r.db.Model(&domain.ScrapeJob{}).Where("id = ?", id).Updates(updates).Error
}

func (r *mysqlScrapeJobRepository) UpdateProgress(id string, total, processed int) error {
	return r.db.Model(&domain.ScrapeJob{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"total_companies":     total,
			"processed_companies": processed,
		}).Error
}

func (r *mysqlScrapeJobRepository) IncrementProcessed(id string) error {
	return r.db.Model(&domain.ScrapeJob{}).Where("id = ?", id).
		UpdateColumn("processed_companies", gorm.Expr("processed_companies + 1")).Error
}
