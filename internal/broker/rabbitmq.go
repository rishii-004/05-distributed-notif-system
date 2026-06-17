// Package broker wraps RabbitMQ (amqp091-go) connectivity and topology.
// Phase 2 used a direct exchange with a single queue.
// Phase 3 switches to a fanout exchange so multiple worker types each get
// their own queue and receive every event independently.
package broker

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/i-katta/notif-system/internal/model"
)

const (
	ExchangeName = "notif.events" // Fanout exchange — broadcasts every event to all bound queues
	ExchangeKind = "fanout"       // Exchange type: routes messages to every bound queue
)

// Connect opens a TCP connection to RabbitMQ.
func Connect(url string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	return conn, nil
}

// DeclareExchange creates the fanout exchange (idempotent).
// Must be called before any DeclareAndBind or PublishEvent call.
func DeclareExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		ExchangeName,
		ExchangeKind,
		true,  // durable — survives broker restart
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	)
}

// DeclareAndBind creates a named, durable queue and binds it to the fanout
// exchange. Each worker type should use its own queue name.
func DeclareAndBind(ch *amqp.Channel, queueName string) error {
	_, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	return ch.QueueBind(queueName, "", ExchangeName, false, nil)
}

// PublishEvent serialises an event to JSON and publishes it to the exchange.
// With a fanout exchange, every bound queue receives the event.
func PublishEvent(ch *amqp.Channel, event model.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return ch.PublishWithContext(
		context.Background(),
		ExchangeName,
		"",    // routing key is ignored by fanout exchanges
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

// Consume registers a consumer on the given queue and returns a Go channel
// of AMQP deliveries.
func Consume(ch *amqp.Channel, queueName string) (<-chan amqp.Delivery, error) {
	msgs, err := ch.Consume(
		queueName,
		"",     // consumer tag — auto-generated
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start consumer on %s: %w", queueName, err)
	}
	return msgs, nil
}
