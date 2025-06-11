package eventexporter

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/jsm.go/api"
	"github.com/nats-io/nats.go"
)

// StreamEvent represents a stream-related event
type StreamEvent struct {
	Type   string            `json:"type"`
	Time   time.Time         `json:"time"`
	Stream string            `json:"stream"`
	Action string            `json:"action"`
	Config *api.StreamConfig `json:"config,omitempty"`
	State  *api.StreamState  `json:"state,omitempty"`
}

// ConsumerEvent represents a consumer-related event
type ConsumerEvent struct {
	Type     string              `json:"type"`
	Time     time.Time           `json:"time"`
	Stream   string              `json:"stream"`
	Consumer string              `json:"consumer"`
	Action   string              `json:"action"`
	Config   *api.ConsumerConfig `json:"config,omitempty"`
}

func (e *Exporter) handleStreamEvent(msg *nats.Msg) {
	subject := msg.Subject
	parts := strings.Split(subject, ".")
	if len(parts) < 5 {
		return
	}

	// Extract stream name and action
	streamName := parts[len(parts)-1]
	action := parts[len(parts)-2]

	var streamEvent StreamEvent
	if err := json.Unmarshal(msg.Data, &streamEvent); err != nil {
		return
	}

	// Create and emit event
	event := Event{
		Type:      "stream",
		Subject:   subject,
		Stream:    streamName,
		Operation: action,
		Timestamp: time.Now(),
		Data:      streamEvent,
	}

	select {
	case e.events <- event:
	default:
		fmt.Println("Channel is full, dropping event")
		// Channel is full, drop event
	}
}

func (e *Exporter) handleConsumerEvent(msg *nats.Msg) {
	subject := msg.Subject
	parts := strings.Split(subject, ".")
	if len(parts) < 6 {
		return
	}

	// Extract stream name, consumer name, and action
	consumerName := parts[len(parts)-1]
	streamName := parts[len(parts)-2]
	action := parts[len(parts)-3]

	var consumerEvent ConsumerEvent
	if err := json.Unmarshal(msg.Data, &consumerEvent); err != nil {
		return
	}

	// Create and emit event
	event := Event{
		Type:      "consumer",
		Subject:   subject,
		Stream:    streamName,
		Consumer:  consumerName,
		Operation: action,
		Timestamp: time.Now(),
		Data:      consumerEvent,
	}

	select {
	case e.events <- event:
	default:
		// Channel is full, drop event
	}
}
