// Package broker wraps RabbitMQ (amqp091-go) connectivity and topology.
// It provides high-level helpers so the rest of the app never deals with
// AMQP wire-protocol details directly.
package broker

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/i-katta/notif-system/internal/model"
)

// RabbitMQ topology constants.
// Using a direct exchange with a single bound queue keeps things simple for
// Phase 2 while still being flexible enough to add routing later.
const (
	ExchangeName = "events"         // Direct exchange that receives all events
	ExchangeKind = "direct"         // Exchange type — routes by exact routing key
	QueueName    = "notifications"  // Queue that workers consume from
	RoutingKey   = "event.created"  // Key used to route events from exchange → queue
)

// Connect opens a TCP connection to RabbitMQ.
// Callers should defer conn.Close() and create ephemeral channels via conn.Channel().
func Connect(url string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	return conn, nil
}

// DeclareTopology ensures the exchange, queue, and binding exist.
// Idempotent — safe to call from both publisher and consumer processes.
func DeclareTopology(ch *amqp.Channel) error {
	// Create a direct exchange named "events"
	if err := ch.ExchangeDeclare(ExchangeName, ExchangeKind, true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Create a durable queue named "notifications"
	q, err := ch.QueueDeclare(QueueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind the queue to the exchange with the routing key
	if err := ch.QueueBind(q.Name, RoutingKey, ExchangeName, false, nil); err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	return nil
}

// PublishEvent serialises an event to JSON and publishes it to the exchange.
// The worker consumers will receive it from the bound queue.
func PublishEvent(ch *amqp.Channel, event model.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return ch.PublishWithContext(
		context.Background(),
		ExchangeName,
		RoutingKey,
		false, // mandatory — don't return if no queue is bound
		false, // immediate — don't return if no consumer is ready
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

// ConsumeEvents registers a consumer on the notifications queue and returns
// a Go channel of AMQP deliveries. The caller should range over this channel.
func ConsumeEvents(ch *amqp.Channel) (<-chan amqp.Delivery, error) {
	msgs, err := ch.Consume(
		QueueName,
		"",     // consumer tag — auto-generated
		true,   // auto-ack — acknowledge immediately after delivery
		false,  // exclusive — allow other consumers on the same queue
		false,  // no-local — not used in Go RabbitMQ client
		false,  // no-wait — wait for server confirmation
		nil,    // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start consumer: %w", err)
	}
	return msgs, nil
}
