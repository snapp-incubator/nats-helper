package eventexporter

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	natsclient "github.com/snapp-incubator/nats-helper/pkg/nats"
)

// Event represents an audit event
type Event struct {
	Type      string    `json:"type"`
	Subject   string    `json:"subject"`
	Stream    string    `json:"stream,omitempty"`
	Consumer  string    `json:"consumer,omitempty"`
	Operation string    `json:"operation"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data,omitempty"`
}

// Exporter handles NATS stream and consumer event monitoring
type Exporter struct {
	client  *natsclient.Client
	events  chan Event
	stopCh  chan struct{}
	streams map[string]struct{}
}

// New creates a new event exporter
func New(client *natsclient.Client) *Exporter {
	return &Exporter{
		client:  client,
		events:  make(chan Event, 100),
		stopCh:  make(chan struct{}),
		streams: make(map[string]struct{}),
	}
}

// Start begins monitoring NATS streams and consumers
func (e *Exporter) Start() error {
	// Subscribe to stream events
	streamSub, err := e.client.Conn.Subscribe("$JS.EVENT.ADVISORY.>", func(msg *nats.Msg) {
		e.handleStreamEvent(msg)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to stream events: %w", err)
	}

	// Subscribe to consumer events
	consumerSub, err := e.client.Conn.Subscribe("$JS.EVENT.ADVISORY.CONSUMER.>", func(msg *nats.Msg) {
		e.handleConsumerEvent(msg)
	})
	if err != nil {
		err2 := streamSub.Unsubscribe()
		if err2 != nil {
			return fmt.Errorf("failed to unsubscribe from stream events: %w", err2)
		}
		return fmt.Errorf("failed to subscribe to consumer events: %w", err)
	}

	go func() {
		<-e.stopCh
		if err := streamSub.Unsubscribe(); err != nil {
			fmt.Printf("failed to unsubscribe from stream events: %v\n", err)
		}
		if err := consumerSub.Unsubscribe(); err != nil {
			fmt.Printf("failed to unsubscribe from consumer events: %v\n", err)
		}
	}()

	return nil
}

// Stop stops the exporter
func (e *Exporter) Stop() {
	close(e.stopCh)
	close(e.events)
}

// Events returns the channel for receiving events
func (e *Exporter) Events() <-chan Event {
	return e.events
}
