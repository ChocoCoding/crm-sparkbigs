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

// ContactService define las operaciones de negocio para Contact.
type ContactService interface {
	CreateContact(contact *domain.Contact) error
	GetContact(id, userID uint) (*domain.Contact, error)
	ListContacts(userID uint, offset, limit int) ([]domain.Contact, int64, error)
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
