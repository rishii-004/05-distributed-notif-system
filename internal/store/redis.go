// Package store provides data-store clients used by worker services.
// Phase 3: Redis for real-time analytics.
// Phase 4: PostgreSQL for persistent event archiving.
package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedis creates a new Redis client and verifies the connection with a ping.
func NewRedis(addr string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return rdb, nil
}

// TrackDAU adds the user to today's HyperLogLog for daily-active-user counting.
// PFADD is memory-efficient (~12KB per key regardless of user count).
func TrackDAU(ctx context.Context, rdb *redis.Client, userID string) error {
	return rdb.PFAdd(ctx, dauKey(), userID).Err()
}

// IncrEventType increments the counter for a given event type.
// Used to track feature-usage metrics (e.g. how many "order_placed" events).
func IncrEventType(ctx context.Context, rdb *redis.Client, eventType string) error {
	return rdb.Incr(ctx, eventTypeKey(eventType)).Err()
}

// dauKey returns the Redis key for today's DAU HyperLogLog.
func dauKey() string {
	return fmt.Sprintf("dau:%s", time.Now().Format("2006-01-02"))
}

// eventTypeKey returns the Redis key for an event-type counter.
func eventTypeKey(eventType string) string {
	return fmt.Sprintf("event_count:%s", eventType)
}
