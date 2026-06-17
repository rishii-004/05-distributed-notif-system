// Package model defines the core data types used throughout the system.
package model

import "encoding/json"

// Event represents a single notification event ingested by the API.
// It is serialised to JSON when published to RabbitMQ and deserialised
// by the consumer workers.
type Event struct {
	EventID   string          `json:"event_id"`             // Unique identifier for the event
	EventType string          `json:"event_type"`           // Type of event (e.g. "order_placed")
	UserID    string          `json:"user_id"`              // The user who triggered the event
	Timestamp int64           `json:"timestamp"`            // Unix timestamp in seconds
	Payload   json.RawMessage `json:"payload,omitempty"`    // Arbitrary JSON payload (passthrough)
}
