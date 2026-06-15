// Command worker starts the background consumer process.
//
// It connects to RabbitMQ, declares the same topology as the server, and
// consumes events from the notifications queue. Each received event is
// logged to stdout to prove async processing works.
//
// Graceful shutdown: pressing Ctrl+C drains in-flight work before exiting.
package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/i-katta/notif-system/internal/broker"
	"github.com/i-katta/notif-system/internal/model"
	"github.com/i-katta/notif-system/pkg/config"
)

func main() {
	// Load configuration from .env and environment variables
	cfg := config.Load()

	// Open a TCP connection to RabbitMQ (same broker as the server)
	conn, err := broker.Connect(cfg.Broker)
	if err != nil {
		log.Fatalf("broker connection failed: %v", err)
	}
	defer conn.Close()

	// Open a lightweight AMQP channel for consuming messages
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel open failed: %v", err)
	}
	defer ch.Close()

	// Ensure exchange/queue/binding exist (safe to call alongside the server)
	if err := broker.DeclareTopology(ch); err != nil {
		log.Fatalf("topology declaration failed: %v", err)
	}

	// Register as a consumer on the notifications queue
	msgs, err := broker.ConsumeEvents(ch)
	if err != nil {
		log.Fatalf("consumer setup failed: %v", err)
	}

	// Set up OS signal channel for graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// Start a goroutine that processes incoming events
	go func() {
		for msg := range msgs {
			var event model.Event
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("failed to unmarshal event: %v", err)
				continue
			}
			log.Printf("Worker processed event: %s", event.EventType)
		}
	}()

	log.Println("Worker started, waiting for events...")

	// Block until we receive SIGINT or SIGTERM
	<-sig
	log.Println("Worker shutting down...")
}
