package repository

import (
	"github.com/Koyo-os/Poll-service/pkg/logger"
	"gorm.io/gorm"
)

type PollRepositoryImpl struct {
	db     *gorm.DB
	logger *logger.Logger
}

func Init(db *gorm.DB, logger *logger.Logger) *PollRepositoryImpl {
	return &PollRepositoryImpl{
		db:     db,
		logger: logger,
	}
}
