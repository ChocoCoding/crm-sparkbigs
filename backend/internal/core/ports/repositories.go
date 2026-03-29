package ports

import "github.com/sparkbigs/crm/internal/core/domain"

// ═══════════════════════════════════════════════════════════
// Puertos de SALIDA (driven) — Repositorios
// Implementados por internal/adapters/storage/
// ═══════════════════════════════════════════════════════════

// UserRepository define las operaciones de persistencia para User.
type UserRepository interface {
	Create(user *domain.User) error
	FindByID(id uint) (*domain.User, error)
	FindByEmail(email string) (*domain.User, error)
	Update(user *domain.User) error
	Delete(id uint) error
	List(offset, limit int) ([]domain.User, int64, error)
}

// LicenseRepository define las operaciones de persistencia para License.
type LicenseRepository interface {
	Create(license *domain.License) error
	FindByUserID(userID uint) (*domain.License, error)
	Update(license *domain.License) error
}

// RefreshTokenRepository define las operaciones de persistencia para RefreshToken.
type RefreshTokenRepository interface {
	Create(token *domain.RefreshToken) error
	FindByToken(token string) (*domain.RefreshToken, error)
	RevokeByUserID(userID uint) error
	RevokeByToken(token string) error
	DeleteExpired() error
}

// ContactRepository define las operaciones de persistencia para Contact.
type ContactRepository interface {
	Create(contact *domain.Contact) error
	FindByID(id uint) (*domain.Contact, error)
	FindByUserID(userID uint, offset, limit int) ([]domain.Contact, int64, error)
	Update(contact *domain.Contact) error
	Delete(id uint) error
}

// DealRepository define las operaciones de persistencia para Deal.
type DealRepository interface {
	Create(deal *domain.Deal) error
	FindByID(id uint) (*domain.Deal, error)
	FindByUserID(userID uint, offset, limit int) ([]domain.Deal, int64, error)
	FindByContactID(contactID uint) ([]domain.Deal, error)
	Update(deal *domain.Deal) error
	Delete(id uint) error
}
