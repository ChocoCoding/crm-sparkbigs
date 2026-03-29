package services_test

import (
	"errors"
	"testing"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCompanyRepo struct{ mock.Mock }

func (m *mockCompanyRepo) Create(c *domain.Company) error { return m.Called(c).Error(0) }
func (m *mockCompanyRepo) FindByID(id uint) (*domain.Company, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Company), args.Error(1)
}
func (m *mockCompanyRepo) FindByUserID(uid uint, o, l int) ([]domain.Company, int64, error) {
	args := m.Called(uid, o, l)
	return args.Get(0).([]domain.Company), args.Get(1).(int64), args.Error(2)
}
func (m *mockCompanyRepo) Update(c *domain.Company) error { return m.Called(c).Error(0) }
func (m *mockCompanyRepo) Delete(id uint) error           { return m.Called(id).Error(0) }
func (m *mockCompanyRepo) Search(uid uint, q string, o, l int) ([]domain.Company, int64, error) {
	args := m.Called(uid, q, o, l)
	return args.Get(0).([]domain.Company), args.Get(1).(int64), args.Error(2)
}

func TestCreateCompany_Success(t *testing.T) {
	repo := new(mockCompanyRepo)
	c := &domain.Company{UserID: 1, Name: "Acme Corp"}
	repo.On("Create", c).Return(nil)

	svc := services.NewCompanyService(repo)
	assert.NoError(t, svc.CreateCompany(c))
	repo.AssertExpectations(t)
}

func TestGetCompany_NotFound(t *testing.T) {
	repo := new(mockCompanyRepo)
	repo.On("FindByID", uint(99)).Return(nil, errors.New("not found"))

	svc := services.NewCompanyService(repo)
	_, err := svc.GetCompany(99, 1)
	assert.ErrorIs(t, err, services.ErrCompanyNotFound)
}

func TestGetCompany_Forbidden(t *testing.T) {
	repo := new(mockCompanyRepo)
	repo.On("FindByID", uint(1)).Return(&domain.Company{ID: 1, UserID: 2}, nil)

	svc := services.NewCompanyService(repo)
	_, err := svc.GetCompany(1, 1)
	assert.ErrorIs(t, err, services.ErrCompanyForbidden)
}

func TestUpdateCompany_PreservesUserID(t *testing.T) {
	repo := new(mockCompanyRepo)
	existing := &domain.Company{ID: 5, UserID: 1, Name: "Old Name"}
	updated := &domain.Company{ID: 5, Name: "New Name"}
	repo.On("FindByID", uint(5)).Return(existing, nil)
	repo.On("Update", mock.AnythingOfType("*domain.Company")).Return(nil)

	svc := services.NewCompanyService(repo)
	assert.NoError(t, svc.UpdateCompany(updated, 1))
	assert.Equal(t, uint(1), updated.UserID)
}

func TestDeleteCompany_Forbidden(t *testing.T) {
	repo := new(mockCompanyRepo)
	repo.On("FindByID", uint(3)).Return(&domain.Company{ID: 3, UserID: 2}, nil)

	svc := services.NewCompanyService(repo)
	err := svc.DeleteCompany(3, 1)
	assert.ErrorIs(t, err, services.ErrCompanyForbidden)
}
