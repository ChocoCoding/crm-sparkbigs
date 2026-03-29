package services_test

import (
	"testing"
	"time"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Mock ────────────────────────────────────────────────────

type mockSubscriptionRepo struct{ mock.Mock }

func (m *mockSubscriptionRepo) Create(s *domain.Subscription) error {
	return m.Called(s).Error(0)
}
func (m *mockSubscriptionRepo) FindByID(id uint) (*domain.Subscription, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Subscription), args.Error(1)
}
func (m *mockSubscriptionRepo) FindByUserID(uid uint, o, l int) ([]domain.Subscription, int64, error) {
	args := m.Called(uid, o, l)
	return args.Get(0).([]domain.Subscription), args.Get(1).(int64), args.Error(2)
}
func (m *mockSubscriptionRepo) FindByCompanyID(cid, uid uint) ([]domain.Subscription, error) {
	args := m.Called(cid, uid)
	return args.Get(0).([]domain.Subscription), args.Error(1)
}
func (m *mockSubscriptionRepo) FindExpiringSoon(uid uint, days int) ([]domain.Subscription, error) {
	args := m.Called(uid, days)
	return args.Get(0).([]domain.Subscription), args.Error(1)
}
func (m *mockSubscriptionRepo) Update(s *domain.Subscription) error {
	return m.Called(s).Error(0)
}
func (m *mockSubscriptionRepo) Delete(id uint) error {
	return m.Called(id).Error(0)
}

// ─── Tests ───────────────────────────────────────────────────

func TestSubscription_Create(t *testing.T) {
	repo := new(mockSubscriptionRepo)
	svc := services.NewSubscriptionService(repo)

	renewal := time.Now().AddDate(1, 0, 0)
	sub := &domain.Subscription{
		UserID: 1, CompanyID: 2,
		Name: "CRM Pro", PlanType: "pro", Status: "active",
		Amount: 299.0, Currency: "EUR", BillingCycle: "annual",
		StartDate: time.Now(), RenewalDate: &renewal,
	}
	repo.On("Create", sub).Return(nil)
	assert.NoError(t, svc.CreateSubscription(sub))
	repo.AssertExpectations(t)
}

func TestSubscription_GetNotFound(t *testing.T) {
	repo := new(mockSubscriptionRepo)
	svc := services.NewSubscriptionService(repo)
	repo.On("FindByID", uint(99)).Return(nil, services.ErrSubscriptionNotFound)
	_, err := svc.GetSubscription(99, 1)
	assert.ErrorIs(t, err, services.ErrSubscriptionNotFound)
}

func TestSubscription_GetForbidden(t *testing.T) {
	repo := new(mockSubscriptionRepo)
	svc := services.NewSubscriptionService(repo)
	repo.On("FindByID", uint(1)).Return(&domain.Subscription{ID: 1, UserID: 2}, nil)
	_, err := svc.GetSubscription(1, 99)
	assert.ErrorIs(t, err, services.ErrSubscriptionForbidden)
}

func TestSubscription_UpdatePreservesUserID(t *testing.T) {
	repo := new(mockSubscriptionRepo)
	svc := services.NewSubscriptionService(repo)
	repo.On("FindByID", uint(1)).Return(&domain.Subscription{ID: 1, UserID: 5}, nil)
	repo.On("Update", mock.AnythingOfType("*domain.Subscription")).Return(nil)

	updated := &domain.Subscription{ID: 1, UserID: 99, Name: "Actualizada"}
	assert.NoError(t, svc.UpdateSubscription(updated, 5))
	assert.Equal(t, uint(5), updated.UserID)
}

func TestSubscription_ExpiringSoon(t *testing.T) {
	repo := new(mockSubscriptionRepo)
	svc := services.NewSubscriptionService(repo)
	renewal := time.Now().AddDate(0, 0, 15)
	subs := []domain.Subscription{{ID: 1, Name: "Vence pronto", RenewalDate: &renewal}}
	repo.On("FindExpiringSoon", uint(1), 30).Return(subs, nil)
	result, err := svc.ExpiringSoon(1, 30)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestSubscription_DeleteForbidden(t *testing.T) {
	repo := new(mockSubscriptionRepo)
	svc := services.NewSubscriptionService(repo)
	repo.On("FindByID", uint(1)).Return(&domain.Subscription{ID: 1, UserID: 10}, nil)
	err := svc.DeleteSubscription(1, 999)
	assert.ErrorIs(t, err, services.ErrSubscriptionForbidden)
}
