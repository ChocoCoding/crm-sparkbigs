package domain

import (
	"time"

	"gorm.io/datatypes"
)

// Company representa una empresa cliente en el CRM B2B.
type Company struct {
	ID                uint           `gorm:"primarykey"                  json:"id"`
	CreatedAt         time.Time      `                                   json:"created_at"`
	UpdatedAt         time.Time      `                                   json:"updated_at"`
	DeletedAt         *time.Time     `gorm:"index"                       json:"-"`
	UserID            uint           `gorm:"index"                       json:"user_id"`
	Name              string         `gorm:"size:255;not null"           json:"name"`
	Sector            string         `gorm:"size:100"                    json:"sector"`
	Status            string         `gorm:"size:50;default:prospect"    json:"status"`   // "prospect" | "active" | "inactive"
	Website           string         `gorm:"size:255"                    json:"website"`
	Phone             string         `gorm:"size:50"                     json:"phone"`
	Address           string         `gorm:"size:512"                    json:"address"`
	RelationStartDate *time.Time     `                                   json:"relation_start_date"`

	// Campo flexible: almacena datos extra (NIF, notas, redes sociales...)
	// sin necesidad de migraciones de schema.
	Metadata datatypes.JSON `json:"metadata"`
}
