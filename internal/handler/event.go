// Package handler contains the HTTP layer — Gin handlers that receive
// incoming requests and delegate to the rest of the system.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/i-katta/notif-system/internal/broker"
	"github.com/i-katta/notif-system/internal/model"
)

// EventHandler wires a RabbitMQ channel into the HTTP handler so it can
// publish events without relying on any global/package-level state.
type EventHandler struct {
	ch *amqp.Channel // AMQP channel used to publish events to RabbitMQ
}

// NewEventHandler creates a handler with the given RabbitMQ channel.
func NewEventHandler(ch *amqp.Channel) *EventHandler {
	return &EventHandler{ch: ch}
}

// PostEvent handles POST /api/v1/events.
// It validates the JSON body, publishes it to RabbitMQ, and returns 202.
func (h *EventHandler) PostEvent(c *gin.Context) {
	var event model.Event

	// Parse and validate the incoming JSON payload
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Publish the event to RabbitMQ for async processing by workers
	if err := broker.PublishEvent(h.ch, event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish event"})
		return
	}

	// 202 Accepted — the event has been accepted for async processing
	c.JSON(http.StatusAccepted, gin.H{"status": "accepted", "event_id": event.EventID})
}
