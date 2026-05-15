package store

import (
	"github.com/henockt/relay/internal/models"
	"gorm.io/gorm"
)

type ReplyThreadStore struct {
	db *gorm.DB
}

func NewReplyThreadStore(db *gorm.DB) *ReplyThreadStore {
	return &ReplyThreadStore{db: db}
}

func (s *ReplyThreadStore) Create(replyThread *models.ReplyThread) error {
	return s.db.Create(replyThread).Error
}

func (s *ReplyThreadStore) FindByToken(token string) (*models.ReplyThread, error) {
	var replyThread models.ReplyThread
	if err := s.db.First(&replyThread, "reply_token = ?", token).Error; err != nil {
		return nil, err
	}
	return &replyThread, nil
}
