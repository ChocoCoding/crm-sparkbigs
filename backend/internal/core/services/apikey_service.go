package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/sparkbigs/crm/internal/core/domain"
	"github.com/sparkbigs/crm/internal/core/ports"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAPIKeyInvalid  = errors.New("api key inválida o inactiva")
	ErrAPIKeyExpired  = errors.New("api key expirada")
	ErrAPIKeyNotFound = errors.New("api key no encontrada")
)

// keyPrefix extrae los primeros 12 caracteres de la clave para lookup en DB.
// Formato de clave: skb_ + 64 hex chars = 68 chars totales.
// Prefijo guardado: "skb_" + primeros 8 hex chars = 12 chars.
const keyStaticPrefix = "skb_"
const prefixLen = 12 // "skb_" + 8 chars del hex

type apiKeyService struct {
	repo ports.APIKeyRepository
}

func NewAPIKeyService(repo ports.APIKeyRepository) ports.APIKeyService {
	return &apiKeyService{repo: repo}
}

// GenerateKey genera una clave segura, la hashea con bcrypt y persiste.
// El texto plano se devuelve solo en esta llamada.
func (s *apiKeyService) GenerateKey(userID uint, name, scopes string) (string, *domain.APIKey, error) {
	// 32 bytes = 256 bits de entropía
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return "", nil, fmt.Errorf("error generando entropía: %w", err)
	}

	hexPart := hex.EncodeToString(rawBytes) // 64 chars
	plaintext := keyStaticPrefix + hexPart  // "skb_" + 64 chars = 68 chars

	prefix := plaintext[:prefixLen] // "skb_" + primeros 8 hex chars

	// bcrypt cost=12: ~300ms en hardware moderno — adecuado para producción
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return "", nil, fmt.Errorf("error hasheando clave: %w", err)
	}

	key := &domain.APIKey{
		UserID:    userID,
		Name:      name,
		KeyPrefix: prefix,
		KeyHash:   string(hash),
		Scopes:    scopes,
		IsActive:  true,
	}

	if err := s.repo.Create(key); err != nil {
		return "", nil, fmt.Errorf("error persistiendo api key: %w", err)
	}

	return plaintext, key, nil
}

// VerifyKey valida la clave recibida en el header X-API-Key.
// 1. Extrae el prefijo para lookup eficiente (evita full-table scan)
// 2. Compara con bcrypt (timing-safe)
// 3. Verifica estado y expiración
// 4. Actualiza LastUsedAt de forma asíncrona
func (s *apiKeyService) VerifyKey(rawKey string) (*domain.APIKey, error) {
	if len(rawKey) < prefixLen {
		return nil, ErrAPIKeyInvalid
	}

	prefix := rawKey[:prefixLen]

	key, err := s.repo.FindByPrefix(prefix)
	if err != nil {
		// Mismo error genérico para no revelar si la clave existe
		return nil, ErrAPIKeyInvalid
	}

	// Comparación en tiempo constante (bcrypt es timing-safe por diseño)
	if err := bcrypt.CompareHashAndPassword([]byte(key.KeyHash), []byte(rawKey)); err != nil {
		return nil, ErrAPIKeyInvalid
	}

	if !key.IsActive {
		return nil, ErrAPIKeyInvalid
	}

	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, ErrAPIKeyExpired
	}

	// Actualizar LastUsedAt sin bloquear la petición
	go func() {
		_ = s.repo.UpdateLastUsed(key.ID)
	}()

	return key, nil
}

func (s *apiKeyService) ListKeys(userID uint) ([]domain.APIKey, error) {
	return s.repo.FindByUserID(userID)
}

func (s *apiKeyService) RevokeKey(id, userID uint) error {
	return s.repo.Revoke(id, userID)
}
