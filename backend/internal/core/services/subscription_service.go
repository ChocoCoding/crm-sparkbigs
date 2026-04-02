package services

import (
	"errors"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
)

var (
	ErrSubscriptionNotFound  = errors.New("suscripción no encontrada")
	ErrSubscriptionForbidden = errors.New("no tienes permiso sobre esta suscripción")
)

type subscriptionService struct {
	repo ports.SubscriptionRepository
}

func NewSubscriptionService(repo ports.SubscriptionRepository) ports.SubscriptionService {
	return &subscriptionService{repo: repo}
}

func (s *subscriptionService) CreateSubscription(sub *domain.Subscription) error {
	return s.repo.Create(sub)
}

func (s *subscriptionService) GetSubscription(id, userID uint) (*domain.Subscription, error) {
	sub, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrSubscriptionNotFound
	}
	if sub.UserID != userID {
		return nil, ErrSubscriptionForbidden
	}
	return sub, nil
}

func (s *subscriptionService) ListSubscriptions(userID uint, offset, limit int) ([]domain.Subscription, int64, error) {
	return s.repo.FindByUserID(userID, offset, limit)
}

func (s *subscriptionService) ExpiringSoon(userID uint, days int) ([]domain.Subscription, error) {
	return s.repo.FindExpiringSoon(userID, days)
}

func (s *subscriptionService) UpdateSubscription(sub *domain.Subscription, userID uint) error {
	existing, err := s.repo.FindByID(sub.ID)
	if err != nil {
		return ErrSubscriptionNotFound
	}
	if existing.UserID != userID {
		return ErrSubscriptionForbidden
	}
	existing.Name         = sub.Name
	existing.PlanType     = sub.PlanType
	existing.Status       = sub.Status
	existing.Amount       = sub.Amount
	existing.Currency     = sub.Currency
	existing.BillingCycle = sub.BillingCycle
	existing.StartDate    = sub.StartDate
	existing.RenewalDate  = sub.RenewalDate
	existing.Notes        = sub.Notes
	existing.CompanyID    = sub.CompanyID
	return s.repo.Update(existing)
}

func (s *subscriptionService) DeleteSubscription(id, userID uint) error {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return ErrSubscriptionNotFound
	}
	if existing.UserID != userID {
		return ErrSubscriptionForbidden
	}
	return s.repo.Delete(id)
}
