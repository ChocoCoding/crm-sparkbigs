package ports

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/sparkbigs/crm/internal/core/domain"
)

// ═══════════════════════════════════════════════════════════
// Puertos de ENTRADA (driving) — Servicios
// Implementados por internal/core/services/
// ═══════════════════════════════════════════════════════════

// JWTClaims representa el payload que se firma dentro del JWT.
type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AuthService define las operaciones del flujo de autenticación.
type AuthService interface {
	Login(email, password string) (accessToken, refreshToken string, user *domain.User, err error)
	RefreshTokens(refreshToken string) (newAccessToken, newRefreshToken string, err error)
	Logout(refreshToken string) error
	ValidateToken(tokenString string) (*JWTClaims, error)
	HashPassword(password string) (string, error)
	ChangePassword(userID uint, currentPassword, newPassword string) error
}

// AdminService define las operaciones de gestión de usuarios y licencias (rol admin).
type AdminService interface {
	CreateUser(user *domain.User, password string) error
	GetUser(id uint) (*domain.User, error)
	ListUsers(offset, limit int) ([]domain.User, int64, error)
	UpdateUser(user *domain.User) error
	DeactivateUser(id uint) error
	SetLicense(userID uint, plan string) error
}

// CompanyService define las operaciones de negocio para Company.
type CompanyService interface {
	CreateCompany(company *domain.Company) error
	GetCompany(id, userID uint) (*domain.Company, error)
	ListCompanies(userID uint, offset, limit int) ([]domain.Company, int64, error)
	UpdateCompany(company *domain.Company, userID uint) error
	DeleteCompany(id, userID uint) error
	SearchCompanies(userID uint, query string, offset, limit int) ([]domain.Company, int64, error)
}

// ContactService define las operaciones de negocio para Contact.
type ContactService interface {
	CreateContact(contact *domain.Contact) error
	GetContact(id, userID uint) (*domain.Contact, error)
	ListContacts(userID uint, offset, limit int) ([]domain.Contact, int64, error)
	SearchContacts(userID uint, query string, offset, limit int) ([]domain.Contact, int64, error)
	UpdateContact(contact *domain.Contact, userID uint) error
	DeleteContact(id, userID uint) error
}

// DealService define las operaciones de negocio para Deal.
type DealService interface {
	CreateDeal(deal *domain.Deal) error
	GetDeal(id, userID uint) (*domain.Deal, error)
	ListDeals(userID uint, offset, limit int) ([]domain.Deal, int64, error)
	UpdateDeal(deal *domain.Deal, userID uint) error
	DeleteDeal(id, userID uint) error
}

// MeetingService define las operaciones de negocio para Meeting.
type MeetingService interface {
	CreateMeeting(meeting *domain.Meeting) error
	GetMeeting(id, userID uint) (*domain.Meeting, error)
	ListMeetings(userID uint, offset, limit int) ([]domain.Meeting, int64, error)
	UpcomingMeetings(userID uint, limit int) ([]domain.Meeting, error)
	UpdateMeeting(meeting *domain.Meeting, userID uint) error
	DeleteMeeting(id, userID uint) error
}

// SubscriptionService define las operaciones de negocio para Subscription.
type SubscriptionService interface {
	CreateSubscription(sub *domain.Subscription) error
	GetSubscription(id, userID uint) (*domain.Subscription, error)
	ListSubscriptions(userID uint, offset, limit int) ([]domain.Subscription, int64, error)
	ExpiringSoon(userID uint, days int) ([]domain.Subscription, error)
	UpdateSubscription(sub *domain.Subscription, userID uint) error
	DeleteSubscription(id, userID uint) error
}

// SettingService define las operaciones de negocio para Setting.
type SettingService interface {
	GetSettings(userID uint) ([]domain.Setting, error)
	GetCategory(userID uint, category string) ([]domain.Setting, error)
	Upsert(setting *domain.Setting) error
	SeedDefaults(userID uint) error
	DeleteSetting(id, userID uint) error
}
