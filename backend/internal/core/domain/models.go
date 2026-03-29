package domain

import (
	"time"

	"gorm.io/datatypes"
)

// Contact representa un contacto/cliente en el CRM.
// ⚠️ Usa datatypes.JSON para Metadata, permitiendo añadir campos sin migraciones.
type Contact struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`
	UserID    uint       `gorm:"index" json:"user_id"`
	Name      string     `gorm:"size:255" json:"name"`
	Email     string     `gorm:"size:255;index" json:"email"`
	Phone     string     `gorm:"size:50" json:"phone"`
	Company   string     `gorm:"size:255" json:"company"`
	Status    string     `gorm:"size:50;default:active" json:"status"` // "active" | "inactive" | "lead"

	// Campo flexible: permite almacenar datos extra (redes sociales, notas, etc.)
	// sin necesidad de migraciones de schema.
	Metadata datatypes.JSON `json:"metadata"`

	// Relaciones
	Deals []Deal `gorm:"foreignKey:ContactID" json:"deals,omitempty"`
}

// Deal representa una oportunidad de negocio asociada a un contacto.
type Deal struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`
	UserID    uint       `gorm:"index" json:"user_id"`
	ContactID uint       `gorm:"index" json:"contact_id"`
	Title     string     `gorm:"size:255" json:"title"`
	Value     float64    `gorm:"type:decimal(15,2);default:0" json:"value"`
	Currency  string     `gorm:"size:3;default:EUR" json:"currency"`
	Stage     string     `gorm:"size:50;default:prospect" json:"stage"` // "prospect" | "qualified" | "proposal" | "won" | "lost"
	ClosedAt  *time.Time `json:"closed_at"`

	Metadata datatypes.JSON `json:"metadata"`
}
