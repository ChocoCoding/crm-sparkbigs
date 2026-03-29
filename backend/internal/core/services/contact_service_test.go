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
func (m *mockContactRepo) Update(c *domain.Contact) error { return m.Called(c).Error(0) }
func (m *mockContactRepo) Delete(id uint) error           { return m.Called(id).Error(0) }

// ─── Tests CreateContact ────────────────────────────────────

func TestCreateContact_Success(t *testing.T) {
	repo := new(mockContactRepo)
	contact := &domain.Contact{UserID: 1, Name: "Ana García", Email: "ana@example.com"}
	repo.On("Create", contact).Return(nil)

	svc := services.NewContactService(repo)
	err := svc.CreateContact(contact)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

// ─── Tests GetContact ───────────────────────────────────────

func TestGetContact_Success(t *testing.T) {
	repo := new(mockContactRepo)
	contact := &domain.Contact{ID: 10, UserID: 1, Name: "Ana García"}
	repo.On("FindByID", uint(10)).Return(contact, nil)

	svc := services.NewContactService(repo)
	result, err := svc.GetContact(10, 1)

	assert.NoError(t, err)
	assert.Equal(t, "Ana García", result.Name)
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
	// El contacto pertenece al userID=2, pero lo solicita userID=1
	contact := &domain.Contact{ID: 10, UserID: 2}
	repo.On("FindByID", uint(10)).Return(contact, nil)

	svc := services.NewContactService(repo)
	_, err := svc.GetContact(10, 1)

	assert.ErrorIs(t, err, services.ErrContactForbidden)
}

// ─── Tests UpdateContact ────────────────────────────────────

func TestUpdateContact_Forbidden(t *testing.T) {
	repo := new(mockContactRepo)
	existing := &domain.Contact{ID: 5, UserID: 2}
	repo.On("FindByID", uint(5)).Return(existing, nil)

	svc := services.NewContactService(repo)
	err := svc.UpdateContact(&domain.Contact{ID: 5, UserID: 1}, 1)

	assert.ErrorIs(t, err, services.ErrContactForbidden)
}

func TestUpdateContact_Success(t *testing.T) {
	repo := new(mockContactRepo)
	existing := &domain.Contact{ID: 5, UserID: 1, Name: "Viejo Nombre"}
	updated := &domain.Contact{ID: 5, Name: "Nuevo Nombre"}
	repo.On("FindByID", uint(5)).Return(existing, nil)
	repo.On("Update", mock.AnythingOfType("*domain.Contact")).Return(nil)

	svc := services.NewContactService(repo)
	err := svc.UpdateContact(updated, 1)

	assert.NoError(t, err)
	// El servicio debe preservar el UserID original
	assert.Equal(t, uint(1), updated.UserID)
}

// ─── Tests DeleteContact ────────────────────────────────────

func TestDeleteContact_Success(t *testing.T) {
	repo := new(mockContactRepo)
	existing := &domain.Contact{ID: 7, UserID: 1}
	repo.On("FindByID", uint(7)).Return(existing, nil)
	repo.On("Delete", uint(7)).Return(nil)

	svc := services.NewContactService(repo)
	err := svc.DeleteContact(7, 1)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteContact_Forbidden(t *testing.T) {
	repo := new(mockContactRepo)
	existing := &domain.Contact{ID: 7, UserID: 2}
	repo.On("FindByID", uint(7)).Return(existing, nil)

	svc := services.NewContactService(repo)
	err := svc.DeleteContact(7, 1)

	assert.ErrorIs(t, err, services.ErrContactForbidden)
}
