package metricexporter

import (
	"fmt"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/snapp-incubator/nats-helper/internal/eventexporter"
)

// Exporter handles conversion of NATS events to Prometheus metrics
type Exporter struct {
	// Stream metrics
	streamOperationsTotal *prometheus.CounterVec
	streamConfigChanges   *prometheus.CounterVec
	streamStateMetrics    *prometheus.GaugeVec

	// Consumer metrics
	consumerOperationsTotal *prometheus.CounterVec
	consumerConfigChanges   *prometheus.CounterVec
	consumerStateMetrics    *prometheus.GaugeVec

	// Event channel
	events chan eventexporter.Event

	// Stop channel
	stopCh chan struct{}

	// Mutex for thread safety
	mu sync.RWMutex
}

// New creates a new metric exporter
func New() *Exporter {
	return &Exporter{
		streamOperationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "nats_stream_operations_total",
				Help: "Total number of stream operations by type",
			},
			[]string{"stream", "operation"},
		),
		streamConfigChanges: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "nats_stream_config_changes_total",
				Help: "Total number of stream configuration changes",
			},
			[]string{"stream", "config_field"},
		),
		streamStateMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nats_stream_state",
				Help: "Current state of NATS streams",
			},
			[]string{"stream", "metric"},
		),
		consumerOperationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "nats_consumer_operations_total",
				Help: "Total number of consumer operations by type",
			},
			[]string{"stream", "consumer", "operation"},
		),
		consumerConfigChanges: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "nats_consumer_config_changes_total",
				Help: "Total number of consumer configuration changes",
			},
			[]string{"stream", "consumer", "config_field"},
		),
		consumerStateMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nats_consumer_state",
				Help: "Current state of NATS consumers",
			},
			[]string{"stream", "consumer", "metric"},
		),
		events: make(chan eventexporter.Event, 100),
		stopCh: make(chan struct{}),
	}
}

// Start begins processing events and updating metrics
func (e *Exporter) Start() error {
	// Register metrics
	prometheus.MustRegister(
		e.streamOperationsTotal,
		e.streamConfigChanges,
		e.streamStateMetrics,
		e.consumerOperationsTotal,
		e.consumerConfigChanges,
		e.consumerStateMetrics,
	)

	// Start processing events
	go e.processEvents()

	return nil
}

// Stop stops the metric exporter
func (e *Exporter) Stop() {
	close(e.stopCh)
	close(e.events)

	// Unregister metrics
	prometheus.Unregister(e.streamOperationsTotal)
	prometheus.Unregister(e.streamConfigChanges)
	prometheus.Unregister(e.streamStateMetrics)
	prometheus.Unregister(e.consumerOperationsTotal)
	prometheus.Unregister(e.consumerConfigChanges)
	prometheus.Unregister(e.consumerStateMetrics)
}

// Events returns the channel for receiving events
func (e *Exporter) Events() chan<- eventexporter.Event {
	return e.events
}

func (e *Exporter) processEvents() {
	for {
		select {
		case <-e.stopCh:
			return
		case event, ok := <-e.events:
			if !ok {
				return
			}
			e.handleEvent(event)
		}
	}
}

func (e *Exporter) handleEvent(event eventexporter.Event) {
	e.mu.Lock()
	defer e.mu.Unlock()
	switch event.Type {
	case "stream":
		e.handleStreamEvent(event)
	case "consumer":
		e.handleConsumerEvent(event)
	}
}

func (e *Exporter) handleStreamEvent(event eventexporter.Event) {
	fmt.Printf("Handling stream event: %+v\n", event)

	// Extract stream name from subject for stream events
	subject := event.Subject
	parts := strings.Split(subject, ".")
	if len(parts) < 5 {
		return
	}
	var streamName string
	// Check if this is a consumer-related stream event
	if len(parts) >= 6 && parts[3] == "CONSUMER" {
		// For consumer-related stream events, the stream name is in the second-to-last part
		// and the consumer name is in the last part
		streamName := parts[len(parts)-2]
		operation := "NEW_CONSUMER" // Special operation for consumer creation

		// Increment operation counter with correct labels
		e.streamOperationsTotal.WithLabelValues(streamName, operation).Inc()

	} else {
		// For regular stream events, the stream name is the last part of the subject
		streamName := parts[len(parts)-1]
		operation := parts[len(parts)-2]

		// Increment operation counter with correct labels
		e.streamOperationsTotal.WithLabelValues(streamName, operation).Inc()

	}

	// Handle stream event data
	if streamEvent, ok := event.Data.(eventexporter.StreamEvent); ok {
		// Update stream state metrics if available
		if streamEvent.State != nil {
			e.streamStateMetrics.WithLabelValues(streamName, "messages").Set(float64(streamEvent.State.Msgs))
			e.streamStateMetrics.WithLabelValues(streamName, "bytes").Set(float64(streamEvent.State.Bytes))
			e.streamStateMetrics.WithLabelValues(streamName, "first_seq").Set(float64(streamEvent.State.FirstSeq))
			e.streamStateMetrics.WithLabelValues(streamName, "last_seq").Set(float64(streamEvent.State.LastSeq))
			e.streamStateMetrics.WithLabelValues(streamName, "consumer_count").Set(float64(streamEvent.State.Consumers))
		}

		// Update config change metrics if available
		if streamEvent.Config != nil {
			if streamEvent.Config.MaxMsgs > 0 {
				e.streamConfigChanges.WithLabelValues(streamName, "max_msgs").Inc()
			}
			if streamEvent.Config.MaxBytes > 0 {
				e.streamConfigChanges.WithLabelValues(streamName, "max_bytes").Inc()
			}
			if streamEvent.Config.MaxAge > 0 {
				e.streamConfigChanges.WithLabelValues(streamName, "max_age").Inc()
			}
		}
	}
}

func (e *Exporter) handleConsumerEvent(event eventexporter.Event) {
	fmt.Printf("Handling consumer event: %+v\n", event)
	// Extract stream and consumer names from subject for consumer events
	subject := event.Subject
	parts := strings.Split(subject, ".")
	if len(parts) < 6 {
		return
	}

	// For consumer events, the consumer name is the last part, stream name is second to last
	consumerName := parts[len(parts)-1]
	streamName := parts[len(parts)-2]
	operation := parts[len(parts)-3]

	// Increment operation counter with correct labels
	e.consumerOperationsTotal.WithLabelValues(streamName, consumerName, operation).Inc()

	// Handle consumer event data
	if consumerEvent, ok := event.Data.(eventexporter.ConsumerEvent); ok {
		// Update consumer config change metrics if available
		if consumerEvent.Config != nil {
			e.consumerConfigChanges.WithLabelValues(streamName, consumerName, "max_deliver").Inc()
			e.consumerConfigChanges.WithLabelValues(streamName, consumerName, "ack_policy").Inc()
			e.consumerConfigChanges.WithLabelValues(streamName, consumerName, "replay_policy").Inc()
		}
	}
}
