package models

import (
	"time"

	"github.com/google/uuid"
)

type Alias struct {
	ID              uuid.UUID `gorm:"primaryKey"`
	UserID          uuid.UUID
	Address         string
	Label           string
	Enabled         bool
	EmailsForwarded int
	EmailsBlocked   int
	CreatedAt       time.Time
}
