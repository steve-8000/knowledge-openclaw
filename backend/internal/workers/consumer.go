// Package workers provides the base consumer framework for NATS JetStream workers.
package workers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	natsclient "github.com/openclaw/ki-db/internal/nats"
	"github.com/openclaw/ki-db/pkg/events"
)

// Handler processes a single event message.
type Handler func(ctx context.Context, env *events.Envelope) error

// Consumer wraps a JetStream pull consumer with retry, logging, and idempotency support.
type Consumer struct {
	name     string
	client   *natsclient.Client
	consumer jetstream.Consumer
	handler  Handler
	logger   *slog.Logger
}

// NewConsumer creates a consumer for a specific event type.
func NewConsumer(
	ctx context.Context,
	client *natsclient.Client,
	name string,
	filterSubject string,
	handler Handler,
	logger *slog.Logger,
) (*Consumer, error) {
	consumer, err := client.CreateConsumer(ctx, name, filterSubject)
	if err != nil {
		return nil, fmt.Errorf("create consumer %s: %w", name, err)
	}

	return &Consumer{
		name:     name,
		client:   client,
		consumer: consumer,
		handler:  handler,
		logger:   logger,
	}, nil
}

// Run starts consuming messages. Blocks until context is cancelled.
func (c *Consumer) Run(ctx context.Context) error {
	c.logger.Info("worker started", "name", c.name)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("worker stopping", "name", c.name)
			return ctx.Err()
		default:
			c.fetchAndProcess(ctx)
		}
	}
}

func (c *Consumer) fetchAndProcess(ctx context.Context) {
	msgs, err := c.consumer.Fetch(10, jetstream.FetchMaxWait(2*time.Second))
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		c.logger.Error("fetch error", "name", c.name, "error", err)
		return
	}

	for msg := range msgs.Messages() {
		c.processMessage(ctx, msg)
	}

	if err := msgs.Error(); err != nil {
		c.logger.Error("fetch iteration error", "name", c.name, "error", err)
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg jetstream.Msg) {
	env, err := events.UnmarshalEnvelope(msg.Data())
	if err != nil {
		c.logger.Error("unmarshal envelope failed, terminating message",
			"name", c.name,
			"error", err,
		)
		_ = msg.Term()
		return
	}

	c.logger.Info("processing event",
		"name", c.name,
		"event_id", env.EventID,
		"event_type", env.EventType,
		"tenant_id", env.TenantID,
	)

	if err := c.handler(ctx, env); err != nil {
		c.logger.Error("handler failed, nacking",
			"name", c.name,
			"event_id", env.EventID,
			"error", err,
		)
		_ = msg.NakWithDelay(5 * time.Second)
		return
	}

	if err := msg.DoubleAck(ctx); err != nil {
		c.logger.Error("double ack failed",
			"name", c.name,
			"event_id", env.EventID,
			"error", err,
		)
	}
}
