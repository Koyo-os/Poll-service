package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	EventsReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "app_events_received_total",
		Help: "Total number of events received",
	}, []string{"event_type"})

	EventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "app_events_processed_total",
		Help: "Total number of events successfully processed",
	}, []string{"event_type"})

	EventProcessingTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "app_event_processing_time_seconds",
		Help:    "Time taken to process an event",
		Buckets: []float64{0.01, 0.1, 0.5, 1, 5, 10},
	}, []string{"event_type"})

	EventProcessingErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "app_event_processing_errors_total",
		Help: "Total number of event processing errors",
	}, []string{"event_type", "error_type"})

	EventLag = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "app_event_lag_seconds",
		Help: "Lag between event creation and processing time",
	}, []string{"event_type"})
)
