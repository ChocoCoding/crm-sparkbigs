package domain

import (
	"time"

	"gorm.io/gorm"
)

// APIKey representa una clave de acceso para integraciones externas (n8n, etc.).
// La clave en texto plano se muestra UNA SOLA VEZ en la creación.
// En base de datos solo se almacena el hash bcrypt y un prefijo para lookup.
type APIKey struct {
	gorm.Model
	UserID     uint       `gorm:"index;not null"       json:"user_id"`
	Name       string     `gorm:"size:255;not null"    json:"name"`
	KeyPrefix  string     `gorm:"size:20;uniqueIndex;not null" json:"key_prefix"` // primeros 12 chars
	KeyHash    string     `gorm:"size:512;not null"    json:"-"`                  // bcrypt — NUNCA se serializa
	Scopes     string     `gorm:"size:500;default:'webhooks'" json:"scopes"`
	IsActive   bool       `gorm:"default:true"         json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at"` // nil = no expira
}
