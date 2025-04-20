package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/Koyo-os/Poll-service/internal/transport/listener"
	"github.com/Koyo-os/Poll-service/internal/transport/producer"
	"github.com/Koyo-os/Poll-service/pkg/config"
	"github.com/Koyo-os/Poll-service/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	logger := logger.Init()

	logger.Info("start initialazying...")

	cfg, err := config.Init("config.yaml")
	if err != nil {
		logger.Error("error load config from config.yaml", zap.Error(err))
		return
	}

	logger.Info("config loaded successfully from config.yaml")

	var mainChan chan entity.Event

	listener, err := listener.Init(mainChan, logger, cfg)
	if err != nil {
		logger.Error("error init listener", zap.Error(err))
		return
	}

	logger.Info("listener init successfully")

	producer := producer.Init(cfg, mainChan, logger)

	logger.Info("producer init successfully")

	logger.Info("starting poll service...")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go listener.Listen(ctx)

	producer.ListenForMsgs()
}
