package domain

import (
	"time"

	"gorm.io/datatypes"
)

// Meeting representa una reunión en el CRM.
// Pertenece obligatoriamente a una Company y opcionalmente a un Contact.
type Meeting struct {
	ID        uint       `gorm:"primarykey"                    json:"id"`
	CreatedAt time.Time  `                                     json:"created_at"`
	UpdatedAt time.Time  `                                     json:"updated_at"`
	DeletedAt *time.Time `gorm:"index"                        json:"-"`

	UserID    uint  `gorm:"index;not null"                   json:"user_id"`
	CompanyID uint  `gorm:"index;not null"                   json:"company_id"` // FK obligatoria → companies
	ContactID *uint `gorm:"index"                            json:"contact_id"` // FK nullable   → contacts

	Title       string    `gorm:"size:255;not null"            json:"title"`
	StartAt     time.Time `gorm:"not null"                     json:"start_at"`      // Fecha + Hora (UTC)
	DurationMin int       `gorm:"default:60"                   json:"duration_min"`  // Duración en minutos
	Status      string    `gorm:"size:50;default:scheduled"    json:"status"`        // "scheduled"|"completed"|"cancelled"
	Notes       string    `gorm:"type:text"                    json:"notes"`

	Metadata datatypes.JSON `json:"metadata"`

	// Relaciones cargadas con Preload
	Company *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Contact *Contact `gorm:"foreignKey:ContactID" json:"contact,omitempty"`
}
