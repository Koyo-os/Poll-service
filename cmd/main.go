package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/Koyo-os/Poll-service/internal/repository"
	"github.com/Koyo-os/Poll-service/internal/service"
	"github.com/Koyo-os/Poll-service/internal/transport/casher"
	"github.com/Koyo-os/Poll-service/internal/transport/consumer"
	"github.com/Koyo-os/Poll-service/internal/transport/listener"
	"github.com/Koyo-os/Poll-service/internal/transport/publisher"
	"github.com/Koyo-os/Poll-service/pkg/config"
	"github.com/Koyo-os/Poll-service/pkg/logger"
	"github.com/Koyo-os/Poll-service/pkg/retrier"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"go.uber.org/zap"
)

func main() {
	cfgLog := logger.Config{
		LogFile:   "app.log",
		LogLevel:  "debug",
		AppName:   "poll-service",
		AddCaller: true,
	}

	if err := logger.Init(cfgLog); err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger := logger.Get()

	cfg, err := config.Init("config.yaml")
	if err != nil || cfg == nil {
		logger.Error("error load config from config.yaml", zap.Error(err))
		return
	}

	logger.Info("config loaded successfully from", zap.String("path", "config.yaml"))

	var mainChan chan entity.Event

	conn, err := retrier.Connect(5, 10, func() (*amqp.Connection, error) {
		return amqp.Dial(cfg.RabbitmqUrl)
	})
	if err != nil {
		logger.Error("error connect to rabbitmq", zap.Error(err))
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	logger.Info("connecting to mariadb...", zap.String("dsn", dsn))

	db, err := retrier.Connect(10, 10, func() (*gorm.DB, error) {
		return gorm.Open(mysql.Open(dsn))
	})
	if err != nil {
		logger.Error("error connect to mariadb", zap.Error(err))

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

	logger.Info("connecting to redis with", zap.String("url", "master-redis:6379"))

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "master-redis:6379",
		Password: "",
		DB:       0,
	})

	err = retrier.Do(3, 5, func() error {
		return redisClient.Ping(context.Background()).Err()
	})
	if err != nil {
		logger.Error("error connect to redis", zap.Error(err))
		return
	}

	logger.Info("successfully connected to redis")

	casher := casher.Init(redisClient)

	service := service.Init(repository.Init(db, logger), publisher, casher)

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
