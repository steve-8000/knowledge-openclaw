// Package nats provides NATS JetStream client setup and stream management.
package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/openclaw/ki-db/internal/config"
)

// Client wraps NATS connection and JetStream context.
type Client struct {
	Conn      *nats.Conn
	JetStream jetstream.JetStream
	Stream    jetstream.Stream
	cfg       config.NATSConfig
}

// NewClient connects to NATS and initializes JetStream with the indexing stream.
func NewClient(ctx context.Context, cfg config.NATSConfig) (*Client, error) {
	nc, err := nats.Connect(cfg.URL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("create jetstream: %w", err)
	}

	// Create or update the indexing stream
	stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:       cfg.Stream,
		Subjects:   []string{"idx.doc.>"},
		Retention:  jetstream.LimitsPolicy,
		Storage:    jetstream.FileStorage,
		MaxAge:     7 * 24 * time.Hour,
		Duplicates: 2 * time.Minute, // dedup window using MsgID
		Discard:    jetstream.DiscardOld,
	})
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("create stream: %w", err)
	}

	return &Client{
		Conn:      nc,
		JetStream: js,
		Stream:    stream,
		cfg:       cfg,
	}, nil
}

// Publish sends a message to JetStream with dedup via MsgID.
func (c *Client) Publish(ctx context.Context, subject string, data []byte, msgID string) error {
	_, err := c.JetStream.Publish(ctx, subject, data,
		jetstream.WithMsgID(msgID),
	)
	if err != nil {
		return fmt.Errorf("publish to %s: %w", subject, err)
	}
	return nil
}

// CreateConsumer creates a durable pull consumer for a specific event type.
func (c *Client) CreateConsumer(ctx context.Context, name string, filterSubject string) (jetstream.Consumer, error) {
	consumer, err := c.Stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       fmt.Sprintf("%s-%s", c.cfg.ConsumerPrefix, name),
		FilterSubject: filterSubject,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       30 * time.Second,
		MaxDeliver:    8,
		BackOff: []time.Duration{
			5 * time.Second,
			30 * time.Second,
			2 * time.Minute,
			5 * time.Minute,
		},
		ReplayPolicy:  jetstream.ReplayInstantPolicy,
		MaxAckPending: 500,
	})
	if err != nil {
		return nil, fmt.Errorf("create consumer %s: %w", name, err)
	}
	return consumer, nil
}

// Close shuts down the NATS connection.
func (c *Client) Close() {
	if c.Conn != nil {
		c.Conn.Drain()
	}
}
