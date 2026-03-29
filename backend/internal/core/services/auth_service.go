package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenDuration  = 60 * time.Minute
	refreshTokenDuration = 7 * 24 * time.Hour
)

var (
	ErrInvalidCredentials = errors.New("credenciales inválidas")
	ErrAccountInactive    = errors.New("cuenta desactivada")
	ErrLicenseExpired     = errors.New("licencia expirada o inactiva")
	ErrTokenInvalid       = errors.New("token inválido o expirado")
	ErrTokenRevoked       = errors.New("token revocado")
	ErrWrongPassword      = errors.New("contraseña actual incorrecta")
)

type authService struct {
	userRepo    ports.UserRepository
	tokenRepo   ports.RefreshTokenRepository
	licenseRepo ports.LicenseRepository
	jwtSecret   string
}

// NewAuthService construye el servicio de autenticación.
// Todos los parámetros son interfaces (puertos) — sin dependencia de infra.
func NewAuthService(
	userRepo ports.UserRepository,
	tokenRepo ports.RefreshTokenRepository,
	licenseRepo ports.LicenseRepository,
	jwtSecret string,
) ports.AuthService {
	return &authService{
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		licenseRepo: licenseRepo,
		jwtSecret:   jwtSecret,
	}
}

func (s *authService) Login(email, password string) (string, string, *domain.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", "", nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return "", "", nil, ErrAccountInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", nil, ErrInvalidCredentials
	}

	// Verificar licencia (solo usuarios no-admin requieren licencia activa)
	if user.Role != "admin" {
		license, err := s.licenseRepo.FindByUserID(user.ID)
		if err != nil || !license.IsActive {
			return "", "", nil, ErrLicenseExpired
		}
		if license.ExpiresAt != nil && license.ExpiresAt.Before(time.Now()) {
			return "", "", nil, ErrLicenseExpired
		}
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", nil, err
	}

	refreshToken, err := s.createRefreshToken(user.ID)
	if err != nil {
		return "", "", nil, err
	}

	// Actualizar last_login_at
	now := time.Now()
	user.LastLoginAt = &now
	_ = s.userRepo.Update(user)

	return accessToken, refreshToken, user, nil
}

func (s *authService) RefreshTokens(refreshToken string) (string, string, error) {
	stored, err := s.tokenRepo.FindByToken(refreshToken)
	if err != nil {
		return "", "", ErrTokenInvalid
	}

	if stored.Revoked {
		return "", "", ErrTokenRevoked
	}

	if time.Now().After(stored.ExpiresAt) {
		return "", "", ErrTokenInvalid
	}

	// Revocar el token antiguo (Refresh Token Rotation)
	if err := s.tokenRepo.RevokeByToken(refreshToken); err != nil {
		return "", "", err
	}

	user, err := s.userRepo.FindByID(stored.UserID)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	newAccessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.createRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *authService) Logout(refreshToken string) error {
	return s.tokenRepo.RevokeByToken(refreshToken)
}

func (s *authService) ValidateToken(tokenString string) (*ports.JWTClaims, error) {
	claims := &ports.JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

func (s *authService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *authService) ChangePassword(userID uint, currentPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrWrongPassword
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(newHash)
	user.MustChangePass = false
	return s.userRepo.Update(user)
}

// ─── helpers privados ───────────────────────────────────────

func (s *authService) generateAccessToken(user *domain.User) (string, error) {
	claims := ports.JWTClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *authService) createRefreshToken(userID uint) (string, error) {
	tokenStr := uuid.New().String()

	rt := &domain.RefreshToken{
		UserID:    userID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(refreshTokenDuration),
	}

	if err := s.tokenRepo.Create(rt); err != nil {
		return "", err
	}

	return tokenStr, nil
}
