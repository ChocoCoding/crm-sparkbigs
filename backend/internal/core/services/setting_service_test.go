package services_test

import (
	"testing"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Mock ────────────────────────────────────────────────────

type mockSettingRepo struct{ mock.Mock }

func (m *mockSettingRepo) Upsert(s *domain.Setting) error {
	return m.Called(s).Error(0)
}
func (m *mockSettingRepo) FindByUserID(uid uint) ([]domain.Setting, error) {
	args := m.Called(uid)
	return args.Get(0).([]domain.Setting), args.Error(1)
}
func (m *mockSettingRepo) FindByCategory(uid uint, cat string) ([]domain.Setting, error) {
	args := m.Called(uid, cat)
	return args.Get(0).([]domain.Setting), args.Error(1)
}
func (m *mockSettingRepo) FindByKey(uid uint, key string) (*domain.Setting, error) {
	args := m.Called(uid, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Setting), args.Error(1)
}
func (m *mockSettingRepo) Delete(id uint) error {
	return m.Called(id).Error(0)
}

// ─── Tests ───────────────────────────────────────────────────

func TestSetting_Upsert(t *testing.T) {
	repo := new(mockSettingRepo)
	svc := services.NewSettingService(repo)
	st := &domain.Setting{UserID: 1, Category: "general", Key: "default_currency", Value: "USD"}
	repo.On("Upsert", st).Return(nil)
	assert.NoError(t, svc.Upsert(st))
}

func TestSetting_GetSettings(t *testing.T) {
	repo := new(mockSettingRepo)
	svc := services.NewSettingService(repo)
	expected := []domain.Setting{
		{ID: 1, UserID: 1, Key: "default_currency", Value: "EUR"},
		{ID: 2, UserID: 1, Key: "company_name", Value: "Acme"},
	}
	repo.On("FindByUserID", uint(1)).Return(expected, nil)
	result, err := svc.GetSettings(1)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestSetting_SeedDefaults_OnlyCreatesNew(t *testing.T) {
	repo := new(mockSettingRepo)
	svc := services.NewSettingService(repo)

	// Simula que ya existe "default_currency", solo debe crear los demás
	existing := []domain.Setting{
		{UserID: 1, Key: "default_currency", Value: "EUR"},
	}
	repo.On("FindByUserID", uint(1)).Return(existing, nil)
	repo.On("Upsert", mock.AnythingOfType("*domain.Setting")).Return(nil)

	err := svc.SeedDefaults(1)
	assert.NoError(t, err)
	// Debe llamar Upsert para los 5 ajustes restantes (6 total - 1 existente)
	repo.AssertNumberOfCalls(t, "Upsert", 5)
}

func TestSetting_DeleteForbidden(t *testing.T) {
	repo := new(mockSettingRepo)
	svc := services.NewSettingService(repo)
	// El setting ID=99 no pertenece al usuario 1
	repo.On("FindByUserID", uint(1)).Return([]domain.Setting{{ID: 5, UserID: 1, Key: "x"}}, nil)
	err := svc.DeleteSetting(99, 1)
	assert.ErrorIs(t, err, services.ErrSettingForbidden)
}
