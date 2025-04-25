package repository

import (
	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/Koyo-os/Poll-service/pkg/logger"
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

func Init(db *gorm.DB, logger *logger.Logger) PollRepository {
	return &PollRepositoryImpl{
		db:     db,
		logger: logger,
	}
}
