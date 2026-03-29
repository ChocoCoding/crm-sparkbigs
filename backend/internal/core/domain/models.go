package domain

import (
	"time"

	"gorm.io/datatypes"
)

// Contact representa un contacto en el CRM.
// CompanyID es FK nullable hacia Company (relación 1:N).
type Contact struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`

	UserID    uint   `gorm:"index;not null" json:"user_id"`
	CompanyID *uint  `gorm:"index" json:"company_id"`          // FK nullable
	Name      string `gorm:"size:255;not null" json:"name"`
	Email     string `gorm:"size:255;index" json:"email"`
	Phone     string `gorm:"size:50" json:"phone"`
	Position  string `gorm:"size:150" json:"position"`          // Cargo / Puesto
	Status    string `gorm:"size:50;default:active" json:"status"` // "active" | "inactive" | "lead"

	Metadata datatypes.JSON `json:"metadata"`

	// Relaciones (cargadas con Preload)
	Company *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Deals   []Deal   `gorm:"foreignKey:ContactID" json:"deals,omitempty"`
}

// Deal representa una oportunidad de negocio asociada a un contacto.
type Deal struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`

	UserID    uint    `gorm:"index" json:"user_id"`
	ContactID uint    `gorm:"index" json:"contact_id"`
	Title     string  `gorm:"size:255" json:"title"`
	Value     float64 `gorm:"type:decimal(15,2);default:0" json:"value"`
	Currency  string  `gorm:"size:3;default:EUR" json:"currency"`
	Stage     string  `gorm:"size:50;default:prospect" json:"stage"`
	ClosedAt  *time.Time `json:"closed_at"`

	Metadata datatypes.JSON `json:"metadata"`
}
