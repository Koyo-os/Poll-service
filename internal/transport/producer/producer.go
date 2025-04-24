package producer

import (
	"encoding/json"
	"time"

	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/Koyo-os/Poll-service/pkg/config"
	"github.com/Koyo-os/Poll-service/pkg/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Publisher struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	logger    *logger.Logger
	cfg       *config.Config
	exchanges map[string]bool
}

func Init(cfg *config.Config, logger *logger.Logger, conn *amqp.Connection) (*Publisher, error) {
	channel, err := conn.Channel()
	if err != nil {
		logger.Error("failed to open channel", zap.Error(err))
		conn.Close()
		return nil, err
	}

	return &Publisher{
		conn:      conn,
		channel:   channel,
		logger:    logger,
		cfg:       cfg,
		exchanges: make(map[string]bool),
	}, nil
}

func (p *Publisher) ensureExchange(exchangeName, exchangeType string) error {
	if p.exchanges[exchangeName] {
		return nil
	}

	err := p.channel.ExchangeDeclare(
		exchangeName,
		exchangeType,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		p.logger.Error("failed to declare exchange",
			zap.String("exchange", exchangeName),
			zap.Error(err))
		return err
	}

	p.exchanges[exchangeName] = true
	return nil
}

func (p *Publisher) Close() error {
	if err := p.channel.Close(); err != nil {
		p.logger.Error("error closing channel", zap.Error(err))
	}

	if err := p.conn.Close(); err != nil {
		p.logger.Error("error closing connection", zap.Error(err))
		return err
	}

	p.logger.Info("RabbitMQ publisher closed successfully")
	return nil
}

func (p *Publisher) Publish(poll any, exchangeName, routingKey string) error {
	if err := p.ensureExchange(exchangeName, "direct"); err != nil {
		return err
	}

	pollJson, err := json.Marshal(poll)
	if err != nil {
		p.logger.Error("failed to marshal poll",
			zapcore.Field{Key: "err", String: err.Error()})
		return err
	}

	event := entity.NewEvent(routingKey, pollJson)
	eventJson, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("failed to marshal event",
			zapcore.Field{Key: "err", String: err.Error()},
			zapcore.Field{Key: "event_UUID", String: event.ID})
		return err
	}

	err = p.channel.Publish(
		exchangeName,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         eventJson,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		p.logger.Error("failed to publish event",
			zapcore.Field{Key: "err", String: err.Error()},
			zapcore.Field{Key: "event_UUID", String: event.ID})
		return err
	}

	p.logger.Info("event published successfully",
		zapcore.Field{Key: "event_UUID", String: event.ID},
		zapcore.Field{Key: "exchange", String: exchangeName},
		zapcore.Field{Key: "routing_key", String: routingKey})

	return nil
}

func (p *Publisher) Broadcast(poll any, exchangeName string) error {
	if err := p.ensureExchange(exchangeName, "fanout"); err != nil {
		return err
	}

	pollJson, err := json.Marshal(poll)
	if err != nil {
		p.logger.Error("failed to marshal poll",
			zapcore.Field{Key: "err", String: err.Error()})
		return err
	}

	event := entity.NewEvent("broadcast", pollJson)
	eventJson, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("failed to marshal event",
			zapcore.Field{Key: "err", String: err.Error()},
			zapcore.Field{Key: "event_UUID", String: event.ID})
		return err
	}

	err = p.channel.Publish(
		exchangeName,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         eventJson,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		p.logger.Error("failed to broadcast event",
			zapcore.Field{Key: "err", String: err.Error()},
			zapcore.Field{Key: "event_UUID", String: event.ID})
		return err
	}

	p.logger.Info("event broadcasted successfully",
		zapcore.Field{Key: "event_UUID", String: event.ID},
		zapcore.Field{Key: "exchange", String: exchangeName})

	return nil
}
