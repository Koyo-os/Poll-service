package repository

import (
	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/Koyo-os/Poll-service/pkg/config"
	"github.com/Koyo-os/Poll-service/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type PollRepository interface {
	Add(*entity.Poll) error
	Update(string, *entity.Poll) error
}

type PollRepositoryImpl struct {
	db     *gorm.DB
	logger *logger.Logger
}

func Init(cfg *config.Config) (PollRepository, error) {
	db, err := gorm.Open(sqlite.Open(cfg.DSN))
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&entity.Poll{})

	return &PollRepositoryImpl{
		db:     db,
		logger: logger.Init(),
	}, nil
}
