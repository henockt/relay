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

func (s *UserStore) Delete(id uuid.UUID) error {
	// aliases are cascade-deleted
	return s.db.Delete(&models.User{}, "id = ?", id).Error
}

func (s *UserStore) FindByID(id uuid.UUID) (*models.User, error) {
	var u models.User
	if err := s.db.First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) FindByEmail(email string) (*models.User, error) {
	var u models.User
	if err := s.db.First(&u, "LOWER(email) = LOWER(?)", email).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) UpdateProvider(id uuid.UUID, provider, providerID string) error {
	return s.db.Model(&models.User{}).Where("id = ?", id).
		Updates(map[string]any{"provider": provider, "provider_id": providerID}).Error
}

func (s *UserStore) FindByProvider(provider, providerID string) (*models.User, error) {
	var u models.User
	err := s.db.Where("provider = ? AND provider_id = ?", provider, providerID).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}
