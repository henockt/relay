package models

import (
	"time"

	"github.com/google/uuid"
)

type Alias struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          uuid.UUID
	Address         string
	Label           string
	Enabled         bool
	EmailsForwarded int
	EmailsBlocked   int
	CreatedAt       time.Time
}
