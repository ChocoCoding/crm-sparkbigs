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
	FindByCompanyID(companyID, userID uint, offset, limit int) ([]domain.Contact, int64, error)
	Update(contact *domain.Contact) error
	Delete(id uint) error
	Search(userID uint, query string, offset, limit int) ([]domain.Contact, int64, error)
}

// CompanyRepository define las operaciones de persistencia para Company.
type CompanyRepository interface {
	Create(company *domain.Company) error
	FindByID(id uint) (*domain.Company, error)
	FindByUserID(userID uint, offset, limit int) ([]domain.Company, int64, error)
	Update(company *domain.Company) error
	Delete(id uint) error
	Search(userID uint, query string, offset, limit int) ([]domain.Company, int64, error)
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

// MeetingRepository define las operaciones de persistencia para Meeting.
type MeetingRepository interface {
	Create(meeting *domain.Meeting) error
	FindByID(id uint) (*domain.Meeting, error)
	FindByUserID(userID uint, offset, limit int) ([]domain.Meeting, int64, error)
	FindUpcoming(userID uint, limit int) ([]domain.Meeting, error)
	FindByCompanyID(companyID, userID uint) ([]domain.Meeting, error)
	Update(meeting *domain.Meeting) error
	Delete(id uint) error
}

// SubscriptionRepository define las operaciones de persistencia para Subscription.
type SubscriptionRepository interface {
	Create(sub *domain.Subscription) error
	FindByID(id uint) (*domain.Subscription, error)
	FindByUserID(userID uint, offset, limit int) ([]domain.Subscription, int64, error)
	FindByCompanyID(companyID, userID uint) ([]domain.Subscription, error)
	FindExpiringSoon(userID uint, days int) ([]domain.Subscription, error)
	Update(sub *domain.Subscription) error
	Delete(id uint) error
}

// APIKeyRepository gestiona la persistencia de claves de API.
type APIKeyRepository interface {
	Create(key *domain.APIKey) error
	FindByPrefix(prefix string) (*domain.APIKey, error)
	FindByUserID(userID uint) ([]domain.APIKey, error)
	Revoke(id, userID uint) error
	UpdateLastUsed(id uint) error
}

// DashboardRepository agrega métricas del CRM en consultas eficientes.
type DashboardRepository interface {
	GetStats(userID uint) (*domain.DashboardStats, error)
}

// SettingRepository define las operaciones de persistencia para Setting.
type SettingRepository interface {
	Upsert(setting *domain.Setting) error
	FindByUserID(userID uint) ([]domain.Setting, error)
	FindByCategory(userID uint, category string) ([]domain.Setting, error)
	FindByKey(userID uint, key string) (*domain.Setting, error)
	Delete(id uint) error
}
