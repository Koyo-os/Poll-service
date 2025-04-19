package producer

import (
	"context"
	"encoding/json"

	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/Koyo-os/Poll-service/pkg/config"
	"github.com/Koyo-os/Poll-service/pkg/logger"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap/zapcore"
)

type Producer struct {
	client     *kafka.Reader
	outputChan chan *entity.Event
	logger     *logger.Logger
}

func Init(cfg *config.Config, outputchan chan *entity.Event, logger *logger.Logger) *Producer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{cfg.KafkaUrl},
		GroupID:        cfg.GroupID,
		Topic:          cfg.Topic.Request,
		CommitInterval: 0,
	})

	defer reader.Close()

	return &Producer{
		client:     reader,
		outputChan: outputchan,
	}
}

func (prod *Producer) ListenForMsgs() {
	prod.logger.Info("starting producer...")

	for {
		msg, err := prod.client.ReadMessage(context.Background())
		if err != nil {
			prod.logger.Error("error handle message", zapcore.Field{
				Key:    "err",
				String: err.Error(),
			})

			continue
		}

		var event entity.Event

		if err = json.Unmarshal(msg.Value, &event); err != nil {
			prod.logger.Error("error unmarshal event", zapcore.Field{
				Key:    "err",
				String: err.Error(),
			})

			continue
		}

		if err = prod.client.CommitMessages(context.Background(), msg); err != nil {
			prod.logger.Error("error commit messages", zapcore.Field{
				Key:    "err",
				String: err.Error(),
			})
		}
	}
}
