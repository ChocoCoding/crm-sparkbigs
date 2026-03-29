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

type mockMeetingRepo struct{ mock.Mock }

func (m *mockMeetingRepo) Create(mtg *domain.Meeting) error {
	return m.Called(mtg).Error(0)
}
func (m *mockMeetingRepo) FindByID(id uint) (*domain.Meeting, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Meeting), args.Error(1)
}
func (m *mockMeetingRepo) FindByUserID(uid uint, o, l int) ([]domain.Meeting, int64, error) {
	args := m.Called(uid, o, l)
	return args.Get(0).([]domain.Meeting), args.Get(1).(int64), args.Error(2)
}
func (m *mockMeetingRepo) FindUpcoming(uid uint, limit int) ([]domain.Meeting, error) {
	args := m.Called(uid, limit)
	return args.Get(0).([]domain.Meeting), args.Error(1)
}
func (m *mockMeetingRepo) FindByCompanyID(cid, uid uint) ([]domain.Meeting, error) {
	args := m.Called(cid, uid)
	return args.Get(0).([]domain.Meeting), args.Error(1)
}
func (m *mockMeetingRepo) Update(mtg *domain.Meeting) error {
	return m.Called(mtg).Error(0)
}
func (m *mockMeetingRepo) Delete(id uint) error {
	return m.Called(id).Error(0)
}

// ─── Tests ───────────────────────────────────────────────────

func TestMeeting_Create(t *testing.T) {
	repo := new(mockMeetingRepo)
	svc := services.NewMeetingService(repo)

	contactID := uint(2)
	m := &domain.Meeting{
		UserID:      1,
		CompanyID:   1,
		ContactID:   &contactID,
		Title:       "Demo inicial",
		StartAt:     time.Now().Add(24 * time.Hour),
		DurationMin: 45,
		Status:      "scheduled",
	}
	repo.On("Create", m).Return(nil)

	assert.NoError(t, svc.CreateMeeting(m))
	repo.AssertExpectations(t)
}

func TestMeeting_CreateWithoutContact(t *testing.T) {
	repo := new(mockMeetingRepo)
	svc := services.NewMeetingService(repo)

	m := &domain.Meeting{UserID: 1, CompanyID: 3, ContactID: nil, Title: "Kick-off", StartAt: time.Now()}
	repo.On("Create", m).Return(nil)

	assert.NoError(t, svc.CreateMeeting(m))
}

func TestMeeting_GetNotFound(t *testing.T) {
	repo := new(mockMeetingRepo)
	svc := services.NewMeetingService(repo)

	repo.On("FindByID", uint(99)).Return(nil, services.ErrMeetingNotFound)

	_, err := svc.GetMeeting(99, 1)
	assert.ErrorIs(t, err, services.ErrMeetingNotFound)
}

func TestMeeting_GetForbidden(t *testing.T) {
	repo := new(mockMeetingRepo)
	svc := services.NewMeetingService(repo)

	repo.On("FindByID", uint(1)).Return(&domain.Meeting{ID: 1, UserID: 2}, nil)

	_, err := svc.GetMeeting(1, 99)
	assert.ErrorIs(t, err, services.ErrMeetingForbidden)
}

func TestMeeting_UpdatePreservesUserID(t *testing.T) {
	repo := new(mockMeetingRepo)
	svc := services.NewMeetingService(repo)

	repo.On("FindByID", uint(1)).Return(&domain.Meeting{ID: 1, UserID: 5}, nil)
	repo.On("Update", mock.AnythingOfType("*domain.Meeting")).Return(nil)

	updated := &domain.Meeting{ID: 1, UserID: 99, Title: "Actualizada"}
	assert.NoError(t, svc.UpdateMeeting(updated, 5))
	assert.Equal(t, uint(5), updated.UserID)
}

func TestMeeting_DeleteForbidden(t *testing.T) {
	repo := new(mockMeetingRepo)
	svc := services.NewMeetingService(repo)

	repo.On("FindByID", uint(1)).Return(&domain.Meeting{ID: 1, UserID: 10}, nil)

	err := svc.DeleteMeeting(1, 999)
	assert.ErrorIs(t, err, services.ErrMeetingForbidden)
}

func TestMeeting_Upcoming(t *testing.T) {
	repo := new(mockMeetingRepo)
	svc := services.NewMeetingService(repo)

	upcoming := []domain.Meeting{
		{ID: 1, Title: "Próxima", StartAt: time.Now().Add(2 * time.Hour)},
	}
	repo.On("FindUpcoming", uint(1), 5).Return(upcoming, nil)

	result, err := svc.UpcomingMeetings(1, 5)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}
