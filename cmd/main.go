package main

import (
	"context"

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

	producer, err := producer.Init(cfg, mainChan, logger)
	if err != nil {
		logger.Error("error init producer", zap.Error(err))
		return
	}

	logger.Info("producer init successfully")

	logger.Info("starting poll service...")

	go listener.Listen(context.Background())

	producer.ListenForMsgs()
}
