package services

import (
	"errors"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
)

var (
	ErrCompanyNotFound  = errors.New("empresa no encontrada")
	ErrCompanyForbidden = errors.New("no tienes permiso sobre esta empresa")
)

type companyService struct {
	companyRepo ports.CompanyRepository
}

func NewCompanyService(companyRepo ports.CompanyRepository) ports.CompanyService {
	return &companyService{companyRepo: companyRepo}
}

func (s *companyService) CreateCompany(company *domain.Company) error {
	return s.companyRepo.Create(company)
}

func (s *companyService) GetCompany(id, userID uint) (*domain.Company, error) {
	company, err := s.companyRepo.FindByID(id)
	if err != nil {
		return nil, ErrCompanyNotFound
	}
	if company.UserID != userID {
		return nil, ErrCompanyForbidden
	}
	return company, nil
}

func (s *companyService) ListCompanies(userID uint, offset, limit int) ([]domain.Company, int64, error) {
	return s.companyRepo.FindByUserID(userID, offset, limit)
}

func (s *companyService) UpdateCompany(company *domain.Company, userID uint) error {
	existing, err := s.companyRepo.FindByID(company.ID)
	if err != nil {
		return ErrCompanyNotFound
	}
	if existing.UserID != userID {
		return ErrCompanyForbidden
	}
	company.UserID = existing.UserID
	return s.companyRepo.Update(company)
}

func (s *companyService) DeleteCompany(id, userID uint) error {
	existing, err := s.companyRepo.FindByID(id)
	if err != nil {
		return ErrCompanyNotFound
	}
	if existing.UserID != userID {
		return ErrCompanyForbidden
	}
	return s.companyRepo.Delete(id)
}

func (s *companyService) SearchCompanies(userID uint, query string, offset, limit int) ([]domain.Company, int64, error) {
	return s.companyRepo.Search(userID, query, offset, limit)
}
