package publisher

import (
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/Koyo-os/Poll-service/pkg/config"
	"github.com/Koyo-os/Poll-service/pkg/logger"
	"go.uber.org/zap/zapcore"
)

type Publisher struct {
	producer sarama.SyncProducer
	logger   *logger.Logger
	cfg      *config.Config
}

func Init(cfg *config.Config, logger *logger.Logger) (*Publisher, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Timeout = 5 * time.Second

	producer, err := sarama.NewSyncProducer([]string{cfg.KafkaUrl}, config)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		producer: producer,
		logger:   logger,
		cfg:      cfg,
	}, nil
}

func (p *Publisher) Close() error {
	return p.producer.Close()
}

func (p *Publisher) Publish(poll any, Type string) error {
	pollJson, err := json.Marshal(poll)
	if err != nil {
		p.logger.Error("error encode poll for publish", zapcore.Field{
			Key:    "err",
			String: err.Error(),
		})
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

	msg := &sarama.ProducerMessage{
		Topic: p.cfg.Topic.Producer,
		Value: sarama.ByteEncoder(eventJson),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.logger.Error("error publish event", zapcore.Field{
			Key:    "err",
			String: err.Error(),
		})
		return err
	}

	p.logger.Info("successfully published event with",
		zapcore.Field{
			Key:    "event_UUID",
			String: event.ID,
		},
		zapcore.Field{
			Key:     "partition",
			Integer: int64(partition),
		},
		zapcore.Field{
			Key:     "offset",
			Integer: offset,
		},
	)

	return nil
}
