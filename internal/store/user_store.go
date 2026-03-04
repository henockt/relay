package store

import (
	"github.com/google/uuid"
	"github.com/henockt/relay/internal/models"
	"gorm.io/gorm"
)

type UserStore struct {
	db *gorm.DB
}

func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{
		db: db,
	}
}

func (s *UserStore) Create(u *models.User) error {
	return s.db.Create(u).Error
}

func (s *UserStore) FindByID(id uuid.UUID) (*models.User, error) {
	var u models.User
	if err := s.db.First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) FindByProvider(provider, providerID string) (*models.User, error) {
	var u models.User
	err := s.db.Where("provider = ? AND provider_id = ?", provider, providerID).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}