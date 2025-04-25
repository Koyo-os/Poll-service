package main

import (
	"context"
	"os"

	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/Koyo-os/Poll-service/internal/publisher"
	"github.com/Koyo-os/Poll-service/internal/repository"
	"github.com/Koyo-os/Poll-service/internal/service"
	"github.com/Koyo-os/Poll-service/internal/transport/consumer"
	"github.com/Koyo-os/Poll-service/internal/transport/listener"
	"github.com/Koyo-os/Poll-service/pkg/config"
	"github.com/Koyo-os/Poll-service/pkg/logger"
	"github.com/Koyo-os/Poll-service/pkg/retrier"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go.uber.org/zap"
)

func main() {
	logger := logger.Init()

	cfg, err := config.Init("config.yaml")
	if err != nil || cfg == nil {
		logger.Error("error load config from config.yaml", zap.Error(err))
		return
	}

	logger.Info("config loaded successfully from config.yaml")

	var mainChan chan entity.Event

	conn, err := retrier.Connect(5, 10, func() (*amqp.Connection, error) {
		return amqp.Dial(cfg.RabbitmqUrl)
	})
	if err != nil {
		logger.Error("error connect to rabbitmq", zap.Error(err))
	}

	logger.Info("connect to sqlite in " + cfg.Dsn)

	if _, err := os.Stat(cfg.Dsn); os.IsNotExist(err) {
		file, err := os.Create(cfg.Dsn)
		if err != nil {
			logger.Error("error create db file", zap.Error(err))
		}

		logger.Info("created database file")

		file.Close()
	}

	db, err := gorm.Open(sqlite.Open(cfg.Dsn))
	if err != nil {
		logger.Error("error connect to db", zap.Error(err))
		return
	}

	logger.Info("connected to db")

	db.AutoMigrate(&entity.Poll{})

	publisher, err := retrier.Connect(3, 5, func() (*publisher.Publisher, error) {
		return publisher.Init(cfg, logger, conn)
	})
	if err != nil {
		logger.Error("error init publisher", zap.Error(err))
		return
	}

	logger.Info("publisher init successfully")

	service := service.Init(repository.Init(db, logger), publisher)

	consumer, err := retrier.Connect(3, 5, func() (*consumer.Consumer, error) {
		return consumer.Init(cfg, logger, conn)
	})
	if err != nil {
		logger.Error("error init producer", zap.Error(err))
		return
	}

	err = consumer.Subscribe(cfg.QueueName, cfg.RequestExchange, "request.*")
	if err != nil {
		logger.Error("error subscribe to", zap.String("queue_name", cfg.QueueName), zap.Error(err))
		return
	}

	logger.Info("consumer init successfully")

	listener := listener.Init(mainChan, logger, cfg, service)

	go listener.Listen(context.Background())
	consumer.ConsumeMessages(mainChan)
}
