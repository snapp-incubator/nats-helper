package nats

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/snapp-incubator/nats-helper/pkg/config"
)

// Client represents a NATS client with JSM capabilities
type Client struct {
	Conn *nats.Conn
}

// New creates a new NATS client
func New(cfg *config.NATSConfig) (*Client, error) {
	opts := []nats.Option{}

	// Add authentication if provided
	if cfg.User != "" && cfg.Password != "" {
		opts = append(opts, nats.UserInfo(cfg.User, cfg.Password))
	}

	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &Client{
		Conn: nc,
	}, nil
}

// Close closes the NATS connection
func (c *Client) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}
