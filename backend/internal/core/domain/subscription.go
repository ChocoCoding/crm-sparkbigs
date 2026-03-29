package domain

import (
	"time"

	"gorm.io/datatypes"
)

// Subscription representa una licencia/suscripción contratada por una Empresa cliente.
// ⚠️ Distinto de domain.License, que gestiona el acceso de los usuarios al CRM.
type Subscription struct {
	ID        uint       `gorm:"primarykey"                    json:"id"`
	CreatedAt time.Time  `                                     json:"created_at"`
	UpdatedAt time.Time  `                                     json:"updated_at"`
	DeletedAt *time.Time `gorm:"index"                        json:"-"`

	UserID    uint `gorm:"index;not null"                   json:"user_id"`    // Dueño (usuario CRM)
	CompanyID uint `gorm:"index;not null"                   json:"company_id"` // FK → companies

	Name         string     `gorm:"size:255;not null"            json:"name"`          // Nombre del producto/servicio
	PlanType     string     `gorm:"size:100"                     json:"plan_type"`     // "basic"|"pro"|"enterprise"|libre
	Status       string     `gorm:"size:50;default:active"       json:"status"`        // "active"|"trial"|"expired"|"cancelled"
	Amount       float64    `gorm:"type:decimal(15,2);default:0" json:"amount"`
	Currency     string     `gorm:"size:3;default:EUR"           json:"currency"`
	BillingCycle string     `gorm:"size:20;default:monthly"      json:"billing_cycle"` // "monthly"|"quarterly"|"annual"|"one_time"
	StartDate    time.Time  `gorm:"not null"                     json:"start_date"`
	RenewalDate  *time.Time `                                     json:"renewal_date"`  // nullable
	Notes        string     `gorm:"type:text"                    json:"notes"`

	Metadata datatypes.JSON `json:"metadata"`

	// Relación Preload
	Company *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
}
