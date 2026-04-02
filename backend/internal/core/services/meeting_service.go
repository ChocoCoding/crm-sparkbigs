package services

import (
	"errors"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
)

var (
	ErrMeetingNotFound  = errors.New("reunión no encontrada")
	ErrMeetingForbidden = errors.New("no tienes permiso sobre esta reunión")
)

type meetingService struct {
	repo ports.MeetingRepository
}

func NewMeetingService(repo ports.MeetingRepository) ports.MeetingService {
	return &meetingService{repo: repo}
}

func (s *meetingService) CreateMeeting(meeting *domain.Meeting) error {
	return s.repo.Create(meeting)
}

func (s *meetingService) GetMeeting(id, userID uint) (*domain.Meeting, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrMeetingNotFound
	}
	if m.UserID != userID {
		return nil, ErrMeetingForbidden
	}
	return m, nil
}

func (s *meetingService) ListMeetings(userID uint, offset, limit int) ([]domain.Meeting, int64, error) {
	return s.repo.FindByUserID(userID, offset, limit)
}

func (s *meetingService) UpcomingMeetings(userID uint, limit int) ([]domain.Meeting, error) {
	return s.repo.FindUpcoming(userID, limit)
}

func (s *meetingService) UpdateMeeting(meeting *domain.Meeting, userID uint) error {
	existing, err := s.repo.FindByID(meeting.ID)
	if err != nil {
		return ErrMeetingNotFound
	}
	if existing.UserID != userID {
		return ErrMeetingForbidden
	}
	existing.Title       = meeting.Title
	existing.CompanyID   = meeting.CompanyID
	existing.ContactID   = meeting.ContactID
	existing.StartAt     = meeting.StartAt
	existing.DurationMin = meeting.DurationMin
	existing.Status      = meeting.Status
	existing.Notes       = meeting.Notes
	return s.repo.Update(existing)
}

func (s *meetingService) DeleteMeeting(id, userID uint) error {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return ErrMeetingNotFound
	}
	if existing.UserID != userID {
		return ErrMeetingForbidden
	}
	return s.repo.Delete(id)
}
