package storage

import (
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"gorm.io/gorm"
)

type mysqlScrapeAPILogRepository struct {
	db *gorm.DB
}

func NewMysqlScrapeAPILogRepository(db *gorm.DB) ports.ScrapeAPILogRepository {
	return &mysqlScrapeAPILogRepository{db: db}
}

func (r *mysqlScrapeAPILogRepository) Create(log *domain.ScrapeAPILog) error {
	return r.db.Create(log).Error
}

func (r *mysqlScrapeAPILogRepository) FindByJobID(jobID string) ([]domain.ScrapeAPILog, error) {
	var logs []domain.ScrapeAPILog
	err := r.db.Where("job_id = ?", jobID).Order("created_at ASC, step_index ASC").Find(&logs).Error
	return logs, err
}
