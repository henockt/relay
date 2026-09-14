package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email      string    `gorm:"uniqueIndex;not null" json:"email"`
	Provider   string    `gorm:"not null;index:idx_users_provider,priority:1" json:"provider"`
	ProviderID string    `gorm:"not null;index:idx_users_provider,priority:2" json:"provider_id"`
	CreatedAt  time.Time `json:"created_at"`

	// declared only so AutoMigrate emits the ON DELETE CASCADE foreign key
	// never populated or serialised
	Aliases []Alias `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}
