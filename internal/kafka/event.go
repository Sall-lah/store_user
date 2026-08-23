package kafka

import (
	"encoding/json"
	"time"
)

const (
	// EventUserDeleted is the domain event key published when an account is permanently deleted.
	EventUserDeleted = "user.deleted"
)

// LifecycleEvent defines the structured domain event payload dispatched on user lifecycle Kafka topics.
// Why: Standardizes asynchronous communication contract across microservices for account lifecycle events.
type LifecycleEvent struct {
	Event     string    `json:"event"`
	UserID    string    `json:"userId"`
	Timestamp time.Time `json:"timestamp"`
	Reason    string    `json:"reason,omitempty"`
}

// Serialize converts the LifecycleEvent into JSON bytes.
// Why: Enforces valid JSON structure before pushing raw bytes onto Kafka partition buffers.
func (e *LifecycleEvent) Serialize() ([]byte, error) {
	return json.Marshal(e)
}
