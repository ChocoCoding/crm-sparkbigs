package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// ─── Mocks ──────────────────────────────────────────────────

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(u *domain.User) error {
	return m.Called(u).Error(0)
}
func (m *mockUserRepo) FindByID(id uint) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *mockUserRepo) FindByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *mockUserRepo) Update(u *domain.User) error    { return m.Called(u).Error(0) }
func (m *mockUserRepo) Delete(id uint) error           { return m.Called(id).Error(0) }
func (m *mockUserRepo) List(o, l int) ([]domain.User, int64, error) {
	args := m.Called(o, l)
	return args.Get(0).([]domain.User), args.Get(1).(int64), args.Error(2)
}

type mockTokenRepo struct{ mock.Mock }

func (m *mockTokenRepo) Create(t *domain.RefreshToken) error { return m.Called(t).Error(0) }
func (m *mockTokenRepo) FindByToken(token string) (*domain.RefreshToken, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}
func (m *mockTokenRepo) RevokeByUserID(uid uint) error  { return m.Called(uid).Error(0) }
func (m *mockTokenRepo) RevokeByToken(t string) error   { return m.Called(t).Error(0) }
func (m *mockTokenRepo) DeleteExpired() error           { return m.Called().Error(0) }

type mockLicenseRepo struct{ mock.Mock }

func (m *mockLicenseRepo) Create(l *domain.License) error { return m.Called(l).Error(0) }
func (m *mockLicenseRepo) FindByUserID(uid uint) (*domain.License, error) {
	args := m.Called(uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.License), args.Error(1)
}
func (m *mockLicenseRepo) Update(l *domain.License) error { return m.Called(l).Error(0) }

// ─── Helper ─────────────────────────────────────────────────

func hashedPassword(plain string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	return string(h)
}

const testJWTSecret = "test-secret-key-for-unit-tests-only"

// ─── Tests Login ────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	licenseRepo := new(mockLicenseRepo)

	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: hashedPassword("password123"),
		Role:         "user",
		IsActive:     true,
	}
	license := &domain.License{UserID: 1, IsActive: true}

	userRepo.On("FindByEmail", "test@example.com").Return(user, nil)
	licenseRepo.On("FindByUserID", uint(1)).Return(license, nil)
	tokenRepo.On("Create", mock.AnythingOfType("*domain.RefreshToken")).Return(nil)
	userRepo.On("Update", mock.AnythingOfType("*domain.User")).Return(nil)

	svc := services.NewAuthService(userRepo, tokenRepo, licenseRepo, testJWTSecret)
	accessToken, refreshToken, returnedUser, err := svc.Login("test@example.com", "password123")

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, user.Email, returnedUser.Email)
	userRepo.AssertExpectations(t)
	licenseRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	licenseRepo := new(mockLicenseRepo)

	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: hashedPassword("correctpass"),
		IsActive:     true,
	}

	userRepo.On("FindByEmail", "test@example.com").Return(user, nil)

	svc := services.NewAuthService(userRepo, tokenRepo, licenseRepo, testJWTSecret)
	_, _, _, err := svc.Login("test@example.com", "wrongpass")

	assert.ErrorIs(t, err, services.ErrInvalidCredentials)
}

func TestLogin_UserNotFound(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	licenseRepo := new(mockLicenseRepo)

	userRepo.On("FindByEmail", "nobody@example.com").Return(nil, errors.New("not found"))

	svc := services.NewAuthService(userRepo, tokenRepo, licenseRepo, testJWTSecret)
	_, _, _, err := svc.Login("nobody@example.com", "anypass")

	assert.ErrorIs(t, err, services.ErrInvalidCredentials)
}

func TestLogin_InactiveAccount(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	licenseRepo := new(mockLicenseRepo)

	user := &domain.User{
		ID:           2,
		Email:        "inactive@example.com",
		PasswordHash: hashedPassword("pass"),
		IsActive:     false,
	}

	userRepo.On("FindByEmail", "inactive@example.com").Return(user, nil)

	svc := services.NewAuthService(userRepo, tokenRepo, licenseRepo, testJWTSecret)
	_, _, _, err := svc.Login("inactive@example.com", "pass")

	assert.ErrorIs(t, err, services.ErrAccountInactive)
}

func TestLogin_ExpiredLicense(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	licenseRepo := new(mockLicenseRepo)

	user := &domain.User{
		ID:           3,
		Email:        "expired@example.com",
		PasswordHash: hashedPassword("pass"),
		Role:         "user",
		IsActive:     true,
	}
	past := time.Now().Add(-24 * time.Hour)
	license := &domain.License{UserID: 3, IsActive: true, ExpiresAt: &past}

	userRepo.On("FindByEmail", "expired@example.com").Return(user, nil)
	licenseRepo.On("FindByUserID", uint(3)).Return(license, nil)

	svc := services.NewAuthService(userRepo, tokenRepo, licenseRepo, testJWTSecret)
	_, _, _, err := svc.Login("expired@example.com", "pass")

	assert.ErrorIs(t, err, services.ErrLicenseExpired)
}

// ─── Tests ValidateToken ────────────────────────────────────

func TestValidateToken_InvalidToken(t *testing.T) {
	svc := services.NewAuthService(
		new(mockUserRepo), new(mockTokenRepo), new(mockLicenseRepo), testJWTSecret,
	)
	_, err := svc.ValidateToken("not.a.valid.token")
	assert.Error(t, err)
}

// ─── Tests RefreshTokens ─────────────────────────────────────

func TestRefreshTokens_Success(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	licenseRepo := new(mockLicenseRepo)

	stored := &domain.RefreshToken{
		UserID:    1,
		Token:     "old-refresh-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Revoked:   false,
	}
	user := &domain.User{ID: 1, Role: "user"}

	tokenRepo.On("FindByToken", "old-refresh-token").Return(stored, nil)
	tokenRepo.On("RevokeByToken", "old-refresh-token").Return(nil)
	userRepo.On("FindByID", uint(1)).Return(user, nil)
	tokenRepo.On("Create", mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	svc := services.NewAuthService(userRepo, tokenRepo, licenseRepo, testJWTSecret)
	newAccess, newRefresh, err := svc.RefreshTokens("old-refresh-token")

	assert.NoError(t, err)
	assert.NotEmpty(t, newAccess)
	assert.NotEmpty(t, newRefresh)
	tokenRepo.AssertExpectations(t)
}

func TestRefreshTokens_RevokedToken(t *testing.T) {
	tokenRepo := new(mockTokenRepo)

	stored := &domain.RefreshToken{
		Token:     "revoked-token",
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   true,
	}
	tokenRepo.On("FindByToken", "revoked-token").Return(stored, nil)

	svc := services.NewAuthService(new(mockUserRepo), tokenRepo, new(mockLicenseRepo), testJWTSecret)
	_, _, err := svc.RefreshTokens("revoked-token")

	assert.ErrorIs(t, err, services.ErrTokenRevoked)
}
