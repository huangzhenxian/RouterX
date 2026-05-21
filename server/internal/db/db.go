package db

import (
	"fmt"

	"github.com/routex/routex/internal/config"
	"github.com/routex/routex/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		cfg.PGHost, cfg.PGPort, cfg.PGUser, cfg.PGPassword, cfg.PGDatabase,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Admin{},
		&model.User{},
		&model.Node{},
		&model.UserTraffic{},
		&model.ProxyProvider{},
	)
}
