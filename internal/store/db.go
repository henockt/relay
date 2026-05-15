package store

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/henockt/relay/internal/models"
)

func NewDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&models.User{}, &models.Alias{}, &models.ReplyThread{}); err != nil {
		return nil, err
	}
	return db, nil
}
