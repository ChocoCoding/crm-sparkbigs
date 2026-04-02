package services

import (
	"errors"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
)

var (
	ErrContactNotFound  = errors.New("contacto no encontrado")
	ErrContactForbidden = errors.New("no tienes permiso sobre este contacto")
)

type contactService struct {
	contactRepo ports.ContactRepository
}

func NewContactService(contactRepo ports.ContactRepository) ports.ContactService {
	return &contactService{contactRepo: contactRepo}
}

func (s *contactService) CreateContact(contact *domain.Contact) error {
	return s.contactRepo.Create(contact)
}

func (s *contactService) GetContact(id, userID uint) (*domain.Contact, error) {
	contact, err := s.contactRepo.FindByID(id)
	if err != nil {
		return nil, ErrContactNotFound
	}
	if contact.UserID != userID {
		return nil, ErrContactForbidden
	}
	return contact, nil
}

func (s *contactService) ListContacts(userID uint, offset, limit int) ([]domain.Contact, int64, error) {
	return s.contactRepo.FindByUserID(userID, offset, limit)
}

func (s *contactService) SearchContacts(userID uint, query string, offset, limit int) ([]domain.Contact, int64, error) {
	return s.contactRepo.Search(userID, query, offset, limit)
}

func (s *contactService) UpdateContact(contact *domain.Contact, userID uint) error {
	existing, err := s.contactRepo.FindByID(contact.ID)
	if err != nil {
		return ErrContactNotFound
	}
	if existing.UserID != userID {
		return ErrContactForbidden
	}
	existing.Name      = contact.Name
	existing.Email     = contact.Email
	existing.Phone     = contact.Phone
	existing.Position  = contact.Position
	existing.Status    = contact.Status
	existing.CompanyID = contact.CompanyID // *uint nullable
	return s.contactRepo.Update(existing)
}

func (s *contactService) DeleteContact(id, userID uint) error {
	existing, err := s.contactRepo.FindByID(id)
	if err != nil {
		return ErrContactNotFound
	}
	if existing.UserID != userID {
		return ErrContactForbidden
	}
	return s.contactRepo.Delete(id)
}
