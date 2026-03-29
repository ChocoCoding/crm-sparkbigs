package services_test

import (
	"errors"
	"testing"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Mock ContactRepository ─────────────────────────────────

type mockContactRepo struct{ mock.Mock }

func (m *mockContactRepo) Create(c *domain.Contact) error { return m.Called(c).Error(0) }
func (m *mockContactRepo) FindByID(id uint) (*domain.Contact, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contact), args.Error(1)
}
func (m *mockContactRepo) FindByUserID(uid uint, o, l int) ([]domain.Contact, int64, error) {
	args := m.Called(uid, o, l)
	return args.Get(0).([]domain.Contact), args.Get(1).(int64), args.Error(2)
}
func (m *mockContactRepo) FindByCompanyID(cid, uid uint, o, l int) ([]domain.Contact, int64, error) {
	args := m.Called(cid, uid, o, l)
	return args.Get(0).([]domain.Contact), args.Get(1).(int64), args.Error(2)
}
func (m *mockContactRepo) Update(c *domain.Contact) error { return m.Called(c).Error(0) }
func (m *mockContactRepo) Delete(id uint) error           { return m.Called(id).Error(0) }
func (m *mockContactRepo) Search(uid uint, q string, o, l int) ([]domain.Contact, int64, error) {
	args := m.Called(uid, q, o, l)
	return args.Get(0).([]domain.Contact), args.Get(1).(int64), args.Error(2)
}

// ─── Tests ───────────────────────────────────────────────────

func TestCreateContact_Success(t *testing.T) {
	repo := new(mockContactRepo)
	companyID := uint(1)
	contact := &domain.Contact{UserID: 1, CompanyID: &companyID, Name: "Ana García", Position: "Directora"}
	repo.On("Create", contact).Return(nil)

	svc := services.NewContactService(repo)
	assert.NoError(t, svc.CreateContact(contact))
	repo.AssertExpectations(t)
}

func TestCreateContact_WithoutCompany(t *testing.T) {
	repo := new(mockContactRepo)
	contact := &domain.Contact{UserID: 1, CompanyID: nil, Name: "Sin empresa"}
	repo.On("Create", contact).Return(nil)

	svc := services.NewContactService(repo)
	assert.NoError(t, svc.CreateContact(contact))
}

func TestGetContact_NotFound(t *testing.T) {
	repo := new(mockContactRepo)
	repo.On("FindByID", uint(99)).Return(nil, errors.New("not found"))

	svc := services.NewContactService(repo)
	_, err := svc.GetContact(99, 1)
	assert.ErrorIs(t, err, services.ErrContactNotFound)
}

func TestGetContact_Forbidden(t *testing.T) {
	repo := new(mockContactRepo)
	repo.On("FindByID", uint(10)).Return(&domain.Contact{ID: 10, UserID: 2}, nil)

	svc := services.NewContactService(repo)
	_, err := svc.GetContact(10, 1)
	assert.ErrorIs(t, err, services.ErrContactForbidden)
}

func TestUpdateContact_PreservesUserID(t *testing.T) {
	repo := new(mockContactRepo)
	repo.On("FindByID", uint(5)).Return(&domain.Contact{ID: 5, UserID: 1, Name: "original"}, nil)
	repo.On("Update", mock.AnythingOfType("*domain.Contact")).Return(nil)

	updated := &domain.Contact{ID: 5, UserID: 99, Name: "nuevo"}
	svc := services.NewContactService(repo)
	assert.NoError(t, svc.UpdateContact(updated, 1))
	assert.Equal(t, uint(1), updated.UserID)
}

func TestDeleteContact_Forbidden(t *testing.T) {
	repo := new(mockContactRepo)
	repo.On("FindByID", uint(7)).Return(&domain.Contact{ID: 7, UserID: 2}, nil)

	svc := services.NewContactService(repo)
	err := svc.DeleteContact(7, 999)
	assert.ErrorIs(t, err, services.ErrContactForbidden)
}
