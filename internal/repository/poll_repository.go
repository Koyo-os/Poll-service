package repository

import (
	"fmt"
	"os"

	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/Koyo-os/Poll-service/pkg/config"
	"github.com/Koyo-os/Poll-service/pkg/logger"
	"go.uber.org/zap"
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
	logger := logger.Init()

	logger.Info("connect to sqlite in " + cfg.Dsn)

	if _, err := os.Stat(cfg.Dsn); os.IsNotExist(err) {
		file, err := os.Create(cfg.Dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to create SQLite file: %w", err)
		}

		logger.Info("created database file")

		file.Close()
	}

	db, err := gorm.Open(sqlite.Open(cfg.Dsn))
	if err != nil {
		logger.Error("error connect to db", zap.Error(err))
		return nil, err
	}

	logger.Info("connected to db")

	db.AutoMigrate(&entity.Poll{})

	logger.Info("migrate")

	return &PollRepositoryImpl{
		db:     db,
		logger: logger,
	}, nil
}
