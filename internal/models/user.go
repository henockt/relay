package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
    ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Email      string    
    Provider   string    
    ProviderID string
    CreatedAt  time.Time
}