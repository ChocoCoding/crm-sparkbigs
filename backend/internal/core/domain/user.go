package domain

import "time"

// User representa un usuario del sistema CRM.
// ⚠️ Campos sensibles marcados con json:"-" para nunca filtrarse al frontend.
type User struct {
	ID             uint       `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `gorm:"index" json:"-"`
	Email          string     `gorm:"uniqueIndex;size:255" json:"email"`
	PasswordHash   string     `gorm:"size:255" json:"-"`
	Name           string     `gorm:"size:255" json:"name"`
	Role           string     `gorm:"size:20;default:user" json:"role"` // "admin" | "user"
	IsActive       bool       `gorm:"default:true" json:"is_active"`
	MustChangePass bool       `gorm:"default:true" json:"must_change_password"`
	LastLoginAt    *time.Time `json:"last_login_at"`

	// Datos sensibles encriptados en base de datos (AES-256-GCM)
	EncryptedApiKey string `gorm:"size:512" json:"-"`

	// Relaciones
	License      *License       `gorm:"foreignKey:UserID" json:"license,omitempty"`
	RefreshTokens []RefreshToken `gorm:"foreignKey:UserID" json:"-"`
}

// License representa la licencia de acceso de un usuario.
type License struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`
	UserID    uint       `gorm:"uniqueIndex" json:"user_id"`
	Plan      string     `gorm:"size:50;default:free" json:"plan"` // "free" | "pro" | "enterprise"
	IsActive  bool       `gorm:"default:true" json:"is_active"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// RefreshToken almacena los tokens de refresco activos por usuario.
// ⚠️ Todo el struct se excluye del JSON — nunca debe exponerse al cliente.
type RefreshToken struct {
	ID        uint       `gorm:"primarykey" json:"-"`
	CreatedAt time.Time  `json:"-"`
	UserID    uint       `gorm:"index" json:"-"`
	Token     string     `gorm:"uniqueIndex;size:512" json:"-"`
	ExpiresAt time.Time  `json:"-"`
	Revoked   bool       `gorm:"default:false" json:"-"`
}
