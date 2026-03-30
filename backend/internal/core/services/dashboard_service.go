package services

import (
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
)

type dashboardService struct {
	repo ports.DashboardRepository
}

func NewDashboardService(repo ports.DashboardRepository) ports.DashboardService {
	return &dashboardService{repo: repo}
}

func (s *dashboardService) GetStats(userID uint) (*domain.DashboardStats, error) {
	return s.repo.GetStats(userID)
}
