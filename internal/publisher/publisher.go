package publisher

import (
	"context"
	"encoding/json"

	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/Koyo-os/Poll-service/pkg/config"
	"github.com/Koyo-os/Poll-service/pkg/logger"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap/zapcore"
)

type Publisher struct {
	client *kafka.Writer
	logger *logger.Logger
}

func Init(cfg *config.Config, logger *logger.Logger) *Publisher {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{cfg.KafkaUrl},
		Topic:   cfg.Topic.Producer,
	})
	defer writer.Close()

	return &Publisher{
		client: writer,
		logger: logger,
	}
}

func (p *Publisher) Publish(poll any, Type string) error {
	pollJson, err := json.Marshal(poll)
	if err != nil {
		p.logger.Error("error encode poll for publish", zapcore.Field{
			Key:    "err",
			String: err.Error(),
		},
		)

		return err
	}

	event := entity.NewEvent(Type, pollJson)

	eventJson, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("error encode event for publish", zapcore.Field{
			Key:    "err",
			String: err.Error(),
		},
			zapcore.Field{
				Key:    "event_UUID",
				String: event.ID,
			})

		return err
	}

	err = p.client.WriteMessages(context.Background(), kafka.Message{
		Value: eventJson,
	})
	if err != nil {
		p.logger.Error("error publish event", zapcore.Field{
			Key:    "err",
			String: err.Error(),
		})
	}

	p.logger.Info("successfully published event with", zapcore.Field{
		Key:    "event_UUID",
		String: event.ID,
	})

	return nil
}
