package models

import (
	"time"

	"github.com/google/uuid"
)

type Alias struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID          uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Address         string    `gorm:"uniqueIndex;not null" json:"address"`
	Label           string    `json:"label"`
	Enabled         bool      `gorm:"not null;default:true" json:"enabled"`
	EmailsForwarded int       `gorm:"not null;default:0" json:"emails_forwarded"`
	EmailsBlocked   int       `gorm:"not null;default:0" json:"emails_blocked"`
	CreatedAt       time.Time `json:"created_at"`

	// see the note on User.Aliases
	ReplyThreads []ReplyThread `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}
