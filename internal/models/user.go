package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
    ID         uuid.UUID `gorm:"primaryKey"`
    Email      string    
    Provider   string    
    ProviderID string
    CreatedAt  time.Time
}