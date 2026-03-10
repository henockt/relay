package models

import (
	"time"

	"github.com/google/uuid"
)

type Alias struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	Address         string    `json:"address"`
	Label           string    `json:"label"`
	Enabled         bool      `json:"enabled"`
	EmailsForwarded int       `json:"emails_forwarded"`
	EmailsBlocked   int       `json:"emails_blocked"`
	CreatedAt       time.Time `json:"created_at"`
}
