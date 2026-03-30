package storage

import (
	"time"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"gorm.io/gorm"
)

type mysqlDashboardRepository struct {
	db *gorm.DB
}

func NewMysqlDashboardRepository(db *gorm.DB) ports.DashboardRepository {
	return &mysqlDashboardRepository{db: db}
}

func (r *mysqlDashboardRepository) GetStats(userID uint) (*domain.DashboardStats, error) {
	stats := &domain.DashboardStats{}

	// ── Empresas ──────────────────────────────────────────────
	if err := r.db.Model(&domain.Company{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalCompanies).Error; err != nil {
		return nil, err
	}
	if err := r.db.Model(&domain.Company{}).
		Where("user_id = ? AND status = 'active'", userID).
		Count(&stats.ActiveCompanies).Error; err != nil {
		return nil, err
	}

	// ── Contactos ─────────────────────────────────────────────
	if err := r.db.Model(&domain.Contact{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalContacts).Error; err != nil {
		return nil, err
	}

	// ── Suscripciones ─────────────────────────────────────────
	if err := r.db.Model(&domain.Subscription{}).
		Where("user_id = ? AND status = 'active'", userID).
		Count(&stats.ActiveSubscriptions).Error; err != nil {
		return nil, err
	}

	// MRR — suma de suscripciones activas con ciclo mensual
	r.db.Model(&domain.Subscription{}).
		Where("user_id = ? AND status = 'active' AND billing_cycle = 'monthly'", userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.MRR)

	// ARR — suma de suscripciones activas con ciclo anual
	r.db.Model(&domain.Subscription{}).
		Where("user_id = ? AND status = 'active' AND billing_cycle = 'annual'", userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.ARR)

	// ── Reuniones este mes ────────────────────────────────────
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)
	if err := r.db.Model(&domain.Meeting{}).
		Where("user_id = ? AND start_at >= ? AND start_at < ?", userID, monthStart, monthEnd).
		Count(&stats.MeetingsThisMonth).Error; err != nil {
		return nil, err
	}

	// ── Próximas reuniones (top 5) ────────────────────────────
	if err := r.db.Where("user_id = ? AND start_at > ? AND status = 'scheduled'", userID, now).
		Order("start_at ASC").
		Limit(5).
		Preload("Company").
		Preload("Contact").
		Find(&stats.UpcomingMeetings).Error; err != nil {
		return nil, err
	}

	// ── Suscripciones por vencer en 30 días (top 5) ──────────
	cutoff := now.AddDate(0, 0, 30)
	if err := r.db.Where(
		"user_id = ? AND status = 'active' AND renewal_date IS NOT NULL AND renewal_date <= ?",
		userID, cutoff,
	).
		Order("renewal_date ASC").
		Limit(5).
		Preload("Company").
		Find(&stats.ExpiringSoon).Error; err != nil {
		return nil, err
	}

	// ── Empresas recientes (top 5) ────────────────────────────
	if err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(5).
		Find(&stats.RecentCompanies).Error; err != nil {
		return nil, err
	}

	return stats, nil
}
