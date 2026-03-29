package domain

import "time"

// Setting representa un ajuste clave-valor del CRM, scoped por usuario.
// El índice único compuesto (user_id, key) garantiza una sola entrada por clave por usuario.
type Setting struct {
	ID        uint      `gorm:"primarykey"                                        json:"id"`
	CreatedAt time.Time `                                                         json:"created_at"`
	UpdatedAt time.Time `                                                         json:"updated_at"`

	UserID    uint   `gorm:"uniqueIndex:idx_user_key;not null"                  json:"user_id"`
	Category  string `gorm:"size:100;not null"                                  json:"category"`   // "general"|"notifications"|"integrations"
	Key       string `gorm:"size:150;not null;uniqueIndex:idx_user_key"         json:"key"`
	Value     string `gorm:"type:text"                                          json:"value"`
	Label     string `gorm:"size:255"                                           json:"label"`
	InputType string `gorm:"size:50;default:text"                               json:"input_type"` // "text"|"select"|"toggle"|"number"
}
