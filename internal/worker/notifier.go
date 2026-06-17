// Package worker contains background processing logic that runs in the
// consumer process (cmd/worker).
package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/i-katta/notif-system/internal/model"
)

// CriticalEventTypes lists the events that trigger a notification.
// Extend this slice as new notification-worthy events are added.
var CriticalEventTypes = []string{
	"order_placed",
	"payment_failed",
	"account_suspended",
}

// StartNotifier runs a goroutine that consumes events from the notifications
// queue. It filters for critical event types and simulates sending an
// email/SMS notification with a 500ms latency.
func StartNotifier(msgs <-chan amqp.Delivery) {
	go func() {
		for msg := range msgs {
			var event model.Event
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("[notifier] failed to unmarshal event: %v", err)
				continue
			}

			if !isCritical(event.EventType) {
				continue
			}

			// Simulate network latency for the external notification API call
			time.Sleep(500 * time.Millisecond)

			notification := fmt.Sprintf(
				"\n=== NOTIFICATION ===\nTo: user-%s\nType: %s\nEvent: %s\nBody: %s\n====================\n",
				event.UserID,
				notificationType(event.EventType),
				event.EventID,
				string(event.Payload),
			)
			log.Println(notification)
		}
	}()
}

// isCritical returns true if the event type should trigger a notification.
func isCritical(eventType string) bool {
	for _, ct := range CriticalEventTypes {
		if eventType == ct {
			return true
		}
	}
	return false
}

// notificationType maps event types to a human-readable notification channel.
func notificationType(eventType string) string {
	switch eventType {
	case "order_placed":
		return "EMAIL"
	case "payment_failed":
		return "EMAIL"
	case "account_suspended":
		return "SMS"
	default:
		return "EMAIL"
	}
}
