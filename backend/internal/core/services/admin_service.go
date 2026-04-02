package services

import (
	"errors"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound  = errors.New("usuario no encontrado")
	ErrEmailInUse    = errors.New("el email ya está registrado")
)

type adminService struct {
	userRepo    ports.UserRepository
	licenseRepo ports.LicenseRepository
}

// NewAdminService construye el servicio de administración de usuarios.
func NewAdminService(
	userRepo ports.UserRepository,
	licenseRepo ports.LicenseRepository,
) ports.AdminService {
	return &adminService{
		userRepo:    userRepo,
		licenseRepo: licenseRepo,
	}
}

func (s *adminService) CreateUser(user *domain.User, password string) error {
	// Verificar email único
	if existing, _ := s.userRepo.FindByEmail(user.Email); existing != nil {
		return ErrEmailInUse
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hash)
	user.MustChangePass = true

	if err := s.userRepo.Create(user); err != nil {
		return err
	}

	// Crear licencia free por defecto
	license := &domain.License{
		UserID:   user.ID,
		Plan:     "free",
		IsActive: true,
	}
	return s.licenseRepo.Create(license)
}

func (s *adminService) GetUser(id uint) (*domain.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *adminService) ListUsers(offset, limit int) ([]domain.User, int64, error) {
	return s.userRepo.List(offset, limit)
}

func (s *adminService) UpdateUser(user *domain.User) error {
	existing, err := s.userRepo.FindByID(user.ID)
	if err != nil {
		return ErrUserNotFound
	}
	// Aplicar solo los campos editables al registro completo cargado de la DB
	// para evitar que GORM intente guardar timestamps con valor cero
	existing.Name     = user.Name
	existing.Role     = user.Role
	existing.IsActive = user.IsActive
	return s.userRepo.Update(existing)
}

func (s *adminService) DeactivateUser(id uint) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return ErrUserNotFound
	}
	user.IsActive = false
	return s.userRepo.Update(user)
}

func (s *adminService) SetLicense(userID uint, plan string) error {
	license, err := s.licenseRepo.FindByUserID(userID)
	if err != nil {
		// Si no existe, crear
		return s.licenseRepo.Create(&domain.License{
			UserID:   userID,
			Plan:     plan,
			IsActive: true,
		})
	}
	license.Plan = plan
	license.IsActive = true
	return s.licenseRepo.Update(license)
}
