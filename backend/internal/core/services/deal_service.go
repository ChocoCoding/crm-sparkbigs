package services

import (
	"errors"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
)

var (
	ErrDealNotFound  = errors.New("deal no encontrado")
	ErrDealForbidden = errors.New("no tienes permiso sobre este deal")
)

type dealService struct {
	dealRepo ports.DealRepository
}

// NewDealService construye el servicio de deals/oportunidades.
func NewDealService(dealRepo ports.DealRepository) ports.DealService {
	return &dealService{dealRepo: dealRepo}
}

func (s *dealService) CreateDeal(deal *domain.Deal) error {
	return s.dealRepo.Create(deal)
}

func (s *dealService) GetDeal(id, userID uint) (*domain.Deal, error) {
	deal, err := s.dealRepo.FindByID(id)
	if err != nil {
		return nil, ErrDealNotFound
	}
	if deal.UserID != userID {
		return nil, ErrDealForbidden
	}
	return deal, nil
}

func (s *dealService) ListDeals(userID uint, offset, limit int) ([]domain.Deal, int64, error) {
	return s.dealRepo.FindByUserID(userID, offset, limit)
}

func (s *dealService) UpdateDeal(deal *domain.Deal, userID uint) error {
	existing, err := s.dealRepo.FindByID(deal.ID)
	if err != nil {
		return ErrDealNotFound
	}
	if existing.UserID != userID {
		return ErrDealForbidden
	}
	deal.UserID = existing.UserID
	return s.dealRepo.Update(deal)
}

func (s *dealService) DeleteDeal(id, userID uint) error {
	existing, err := s.dealRepo.FindByID(id)
	if err != nil {
		return ErrDealNotFound
	}
	if existing.UserID != userID {
		return ErrDealForbidden
	}
	return s.dealRepo.Delete(id)
}
