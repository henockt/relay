package models

import (
	"time"

	"github.com/google/uuid"
)

type ReplyThread struct {
	ReplyToken        string    `gorm:"primaryKey;size:64" json:"reply_token"`
	AliasID           uuid.UUID `gorm:"type:uuid;index;not null" json:"alias_id"`
	OriginalFrom      string    `gorm:"not null" json:"original_from"`
	OriginalMessageID string    `json:"original_message_id"`
	ExpiresAt         time.Time `gorm:"index;not null" json:"expires_at"`
	CreatedAt         time.Time `json:"created_at"`
}
