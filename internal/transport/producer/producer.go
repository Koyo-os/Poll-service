package producer

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/Koyo-os/Poll-service/pkg/config"
	"github.com/Koyo-os/Poll-service/pkg/logger"
	"go.uber.org/zap/zapcore"
)

type Producer struct {
	client     sarama.ConsumerGroup
	outputChan chan entity.Event
	logger     *logger.Logger
	ready      chan bool
	cfg        *config.Config
}

func Init(cfg *config.Config, outputChan chan entity.Event, logger *logger.Logger) (*Producer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	client, err := sarama.NewConsumerGroup([]string{cfg.KafkaUrl}, cfg.GroupID, config)
	if err != nil {
		return nil, err
	}

	return &Producer{
		client:     client,
		outputChan: outputChan,
		logger:     logger,
		cfg:        cfg,
		ready:      make(chan bool),
	}, nil
}

func (prod *Producer) ListenForMsgs() {
	prod.logger.Info("starting producer...")

	ctx := context.Background()
	for {
		if err := prod.client.Consume(ctx, []string{prod.cfg.Topic.Request}, prod); err != nil {
			prod.logger.Error("error from consumer", zapcore.Field{
				Key:    "err",
				String: err.Error(),
			})
		}

		if ctx.Err() != nil {
			return
		}
		prod.ready = make(chan bool)
	}
}

func (prod *Producer) Setup(sarama.ConsumerGroupSession) error {
	close(prod.ready)
	return nil
}

func (prod *Producer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (prod *Producer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var event entity.Event

		if err := json.Unmarshal(msg.Value, &event); err != nil {
			prod.logger.Error("error unmarshal event", zapcore.Field{
				Key:    "err",
				String: err.Error(),
			})
			continue
		}

		prod.outputChan <- event
		session.MarkMessage(msg, "")
	}
	return nil
}

func (prod *Producer) Close() error {
	return prod.client.Close()
}
