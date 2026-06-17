// Command server starts the event-ingestion HTTP API.
//
// It connects to RabbitMQ, sets up the exchange/queue topology, and
// registers the POST /api/v1/events endpoint on a Gin router.
// Workers run as a separate process (cmd/worker).
package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/i-katta/notif-system/internal/broker"
	"github.com/i-katta/notif-system/internal/handler"
	"github.com/i-katta/notif-system/pkg/config"
)

func main() {
	// Load configuration from .env and environment variables
	cfg := config.Load()

	// Open a TCP connection to RabbitMQ -- persistent connection
	conn, err := broker.Connect(cfg.Broker)
	if err != nil {
		log.Fatalf("broker connection failed: %v", err)
	}
	defer conn.Close()

	// Open a lightweight AMQP channel on that connection
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel open failed: %v", err)
	}
	defer ch.Close()

	// Ensure the exchange, queue, and binding exist (idempotent)
	if err := broker.DeclareExchange(ch); err != nil {
		log.Fatalf("topology declaration failed: %v", err)
	}

	// Create the HTTP handler with the RabbitMQ channel injected
	h := handler.NewEventHandler(ch)

	// Set up Gin router with the events endpoint
	r := gin.Default()
	r.POST("/api/v1/events", h.PostEvent)

	// Start the HTTP server
	addr := ":" + cfg.Port
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
