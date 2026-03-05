package store

import (
	"github.com/google/uuid"
	"github.com/henockt/relay/internal/models"
	"gorm.io/gorm"
)

type AliasStore struct {
	db *gorm.DB
}

func NewAliasStore(db *gorm.DB) *AliasStore {
	return &AliasStore{db: db}
}

func (s *AliasStore) Create(a *models.Alias) error {
	return s.db.Create(a).Error
}

func (s *AliasStore) ListByUser(userID uuid.UUID) ([]models.Alias, error) {
	var aliases []models.Alias
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&aliases).Error
	return aliases, err
}

func (s *AliasStore) FindByID(id uuid.UUID) (*models.Alias, error) {
	var a models.Alias
	if err := s.db.First(&a, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *AliasStore) Update(a *models.Alias) error {
	return s.db.Save(a).Error
}

func (s *AliasStore) Delete(id uuid.UUID) error {
	return s.db.Delete(&models.Alias{}, "id = ?", id).Error
}

// FindByAddress is used by the SMTP server to look up an alias on incoming mail.
func (s *AliasStore) FindByAddress(address string) (*models.Alias, error) {
	var a models.Alias
	if err := s.db.Where("address = ?", address).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}
