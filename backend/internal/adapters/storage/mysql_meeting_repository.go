package storage

import (
	"time"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"gorm.io/gorm"
)

type mysqlMeetingRepository struct {
	db *gorm.DB
}

func NewMysqlMeetingRepository(db *gorm.DB) ports.MeetingRepository {
	return &mysqlMeetingRepository{db: db}
}

func (r *mysqlMeetingRepository) Create(meeting *domain.Meeting) error {
	return r.db.Create(meeting).Error
}

func (r *mysqlMeetingRepository) FindByID(id uint) (*domain.Meeting, error) {
	var m domain.Meeting
	err := r.db.Preload("Company").Preload("Contact").First(&m, id).Error
	return &m, err
}

func (r *mysqlMeetingRepository) FindByUserID(userID uint, offset, limit int) ([]domain.Meeting, int64, error) {
	var meetings []domain.Meeting
	var total int64

	r.db.Model(&domain.Meeting{}).Where("user_id = ? AND deleted_at IS NULL", userID).Count(&total)
	err := r.db.
		Preload("Company").Preload("Contact").
		Where("user_id = ?", userID).
		Order("start_at DESC").
		Offset(offset).Limit(limit).
		Find(&meetings).Error

	return meetings, total, err
}

func (r *mysqlMeetingRepository) FindUpcoming(userID uint, limit int) ([]domain.Meeting, error) {
	var meetings []domain.Meeting
	err := r.db.
		Preload("Company").Preload("Contact").
		Where("user_id = ? AND start_at > ? AND status = 'scheduled'", userID, time.Now()).
		Order("start_at ASC").
		Limit(limit).
		Find(&meetings).Error
	return meetings, err
}

func (r *mysqlMeetingRepository) FindByCompanyID(companyID, userID uint) ([]domain.Meeting, error) {
	var meetings []domain.Meeting
	err := r.db.
		Preload("Contact").
		Where("company_id = ? AND user_id = ?", companyID, userID).
		Order("start_at DESC").
		Find(&meetings).Error
	return meetings, err
}

func (r *mysqlMeetingRepository) Update(meeting *domain.Meeting) error {
	return r.db.Save(meeting).Error
}

func (r *mysqlMeetingRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Meeting{}, id).Error
}
