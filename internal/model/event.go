package model

import "encoding/json"

type Event struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	UserID    string          `json:"user_id"`
	Timestamp int64           `json:"timestamp"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}
