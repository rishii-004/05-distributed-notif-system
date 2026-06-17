// Package worker contains background processing logic that runs in the
// consumer process (cmd/worker).
package worker

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	"github.com/i-katta/notif-system/internal/model"
	"github.com/i-katta/notif-system/internal/store"
)

// StartAnalytics runs a goroutine that consumes events from the analytics
// queue. It intercepts *all* events and updates Redis metrics:
//   - PFADD (HyperLogLog) for daily active users
//   - INCRBY for feature-usage counters
func StartAnalytics(msgs <-chan amqp.Delivery, rdb *redis.Client) {
	go func() {
		ctx := context.Background()
		for msg := range msgs {
			var event model.Event
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("[analytics] failed to unmarshal event: %v", err)
				continue
			}

			// Track this user in today's DAU HyperLogLog
			if err := store.TrackDAU(ctx, rdb, event.UserID); err != nil {
				log.Printf("[analytics] dau track failed: %v", err)
			}

			// Increment the usage counter for this event type
			if err := store.IncrEventType(ctx, rdb, event.EventType); err != nil {
				log.Printf("[analytics] incr failed: %v", err)
			}

			log.Printf("[analytics] processed event: %s (user: %s)", event.EventType, event.UserID)
		}
	}()
}
