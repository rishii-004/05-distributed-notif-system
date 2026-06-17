// Command worker starts the background consumer process.
//
// Phase 3: runs two independent worker goroutines inside the same process:
//   1. Notifier — filters critical events, simulates SMS/email (500ms latency)
//   2. Analytics — processes every event, updates Redis metrics (DAU, counts)
//
// Each worker gets its own queue bound to the fanout exchange, so neither
// misses any events. Graceful shutdown via Ctrl+C.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/i-katta/notif-system/internal/broker"
	"github.com/i-katta/notif-system/internal/store"
	"github.com/i-katta/notif-system/internal/worker"
	"github.com/i-katta/notif-system/pkg/config"
)

const (
	notificationsQueue = "notifications"
	analyticsQueue     = "analytics"
)

func main() {
	cfg := config.Load()

	// --- RabbitMQ setup ---
	conn, err := broker.Connect(cfg.Broker)
	if err != nil {
		log.Fatalf("broker connection failed: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel open failed: %v", err)
	}
	defer ch.Close()

	if err := broker.DeclareExchange(ch); err != nil {
		log.Fatalf("exchange declaration failed: %v", err)
	}

	// Create one queue per worker type, each bound to the fanout exchange
	for _, q := range []string{notificationsQueue, analyticsQueue} {
		if err := broker.DeclareAndBind(ch, q); err != nil {
			log.Fatalf("queue %s setup failed: %v", q, err)
		}
	}

	// --- Redis setup (for analytics) ---
	rdb, err := store.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	defer rdb.Close()

	// --- Start consumers ---
	notifMsgs, err := broker.Consume(ch, notificationsQueue)
	if err != nil {
		log.Fatalf("consumer setup failed for %s: %v", notificationsQueue, err)
	}
	worker.StartNotifier(notifMsgs)

	analyticsMsgs, err := broker.Consume(ch, analyticsQueue)
	if err != nil {
		log.Fatalf("consumer setup failed for %s: %v", analyticsQueue, err)
	}
	worker.StartAnalytics(analyticsMsgs, rdb)

	// --- Wait for shutdown signal ---
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Worker started — notifier + analytics running. Press Ctrl+C to stop.")
	<-sig
	log.Println("Worker shutting down...")
}
