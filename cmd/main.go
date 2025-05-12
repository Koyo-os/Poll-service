package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/Koyo-os/Poll-service/internal/repository"
	"github.com/Koyo-os/Poll-service/internal/service"
	"github.com/Koyo-os/Poll-service/internal/transport/casher"
	"github.com/Koyo-os/Poll-service/internal/transport/consumer"
	"github.com/Koyo-os/Poll-service/internal/transport/listener"
	"github.com/Koyo-os/Poll-service/internal/transport/publisher"
	"github.com/Koyo-os/Poll-service/pkg/config"
	"github.com/Koyo-os/Poll-service/pkg/deletepull"
	"github.com/Koyo-os/Poll-service/pkg/logger"
	"github.com/Koyo-os/Poll-service/pkg/retrier"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func errorListener(ch chan error, logger *logger.Logger) {
	for err := range ch {
		logger.Error("error from deletepull", zap.Error(err))
	}
}

func main() {
	var (
		mainChan chan entity.Event
		reqChan  chan deletepull.Request
		errChan  chan error
	)

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

	go errorListener(errChan, logger)

	cfg, err := config.Init("config.yaml")
	if err != nil || cfg == nil {
		logger.Error("error load config from config.yaml", zap.Error(err))
		return
	}

	logger.Info("config loaded successfully from", zap.String("path", "config.yaml"))

	rabbitmqConns, err := retrier.MuliConnects(2, func() (*amqp.Connection, error) {
		return amqp.Dial(cfg.RabbitmqUrl)
	}, nil)
	if err != nil {
		logger.Error("error connect to rabbitmq")
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
		return publisher.Init(cfg, logger, rabbitmqConns[0])
	})
	if err != nil {
		logger.Error("error init publisher", zap.Error(err))
		return
	}

	logger.Info("publisher init successfully")

	logger.Info("connecting to redis with", zap.String("url", "master-redis:6379"))

	redisConns, err := retrier.MuliConnects(2, func() (*redis.Client, error) {
		redisClient := redis.NewClient(&redis.Options{
			Addr:     "master-redis:6379",
			Password: "",
			DB:       0,
		})

		err := redisClient.Ping(context.Background()).Err()
		if err != nil {
			return nil, err
		}

		return redisClient, nil
	}, nil)
	if err != nil {
		logger.Error("error connect to redis", zap.Error(err))
		return
	}

	logger.Info("successfully connected to redis")
	deletepull := deletepull.Init(errChan, redisConns[1], logger)

	casher := casher.Init(redisConns[0])

	service := service.Init(repository.Init(db, logger), publisher, casher)

	consumer, err := retrier.Connect(3, 5, func() (*consumer.Consumer, error) {
		return consumer.Init(cfg, logger, rabbitmqConns[1])
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

	err = consumer.Subscribe("votes", cfg.VoteExchange, "vote.*")
	if err != nil {
		logger.Error("error subscribe to", zap.String("queue_name", cfg.QueueName), zap.Error(err))
		return
	}

	logger.Info("consumer init successfully")

	listener := listener.Init(mainChan, logger, cfg, service)

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":8080", nil)
	go deletepull.Listen(context.Background())
	go listener.Listen(context.Background())
	consumer.ConsumeMessages(mainChan)
}
