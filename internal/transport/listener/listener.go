package listener

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Koyo-os/Poll-service/internal/entity"
	"github.com/Koyo-os/Poll-service/internal/service"
	"github.com/Koyo-os/Poll-service/pkg/config"
	"github.com/Koyo-os/Poll-service/pkg/logger"
	"github.com/Koyo-os/Poll-service/pkg/metrics"
	"go.uber.org/zap"
)

type Listener struct {
	inputChan chan entity.Event
	logger    *logger.Logger
	service   *service.PollServiceImpl
	cfg       *config.Config
}

func Init(
	inputChan chan entity.Event,
	logger *logger.Logger,
	cfg *config.Config,
	service *service.PollServiceImpl,
) *Listener {
	return &Listener{
		inputChan: inputChan,
		service:   service,
		logger:    logger,
		cfg:       cfg,
	}
}

func (list *Listener) Listen(ctx context.Context) {
	for {
		select {
		case event := <-list.inputChan:
			then := time.Now()

			var poll entity.Poll

			if err := json.Unmarshal(event.Payload, &poll); err != nil {
				list.logger.Error("error unmarshal poll", zap.Error(err))
				metrics.EventProcessingErrors.WithLabelValues(event.Type).Inc()
				continue
			}

			switch event.Type {
			case list.cfg.Reqs.CreatePollRequestType:
				if err := list.service.Add(&poll); err != nil {
					list.logger.Error("error add poll to db", zap.Error(err))
					metrics.EventProcessingErrors.WithLabelValues(event.Type).Inc()
				}
				lag := time.Since(then)
				metrics.EventProcessingTime.WithLabelValues(event.Type).Observe(float64(lag))

			case list.cfg.Reqs.UpdatePollRequestType:
				if err := list.service.Update(poll.ID.String(), &poll); err != nil {
					list.logger.Error("error update poll", zap.Error(err))

					metrics.EventProcessingErrors.WithLabelValues(event.Type).Inc()
				}
				lag := time.Since(then)
				metrics.EventProcessingTime.WithLabelValues(event.Type).Observe(float64(lag))

			case list.cfg.Reqs.SetClosedRequestType:
				if err := list.service.SetPollClosed(poll.ID.String()); err != nil {
					list.logger.Error(
						"error set poll closed",
						zap.String("poll_id", poll.ID.String()),
						zap.Error(err),
					)

					metrics.EventProcessingErrors.WithLabelValues(event.Type).Inc()
				}
				lag := time.Since(then)
				metrics.EventProcessingTime.WithLabelValues(event.Type).Observe(float64(lag))

			default:
				list.logger.Warn("unknown event type reciewed", zap.String("type", event.Type))
			}

		case <-ctx.Done():
			list.logger.Info("stopping listeners...")
			return
		}
	}
}
