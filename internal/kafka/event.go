package kafka

import (
	"encoding/json"
	"time"
)

const (
	// EventUserDeleted is the domain event key published when an account is permanently deleted.
	EventUserDeleted = "user.deleted"

	// EventUserBanned is the domain event key published when an administrator bans a user account.
	EventUserBanned = "user.banned"
)

// LifecycleEvent defines the structured domain event payload dispatched on user lifecycle Kafka topics.
// Why: Standardizes asynchronous communication contract across microservices for account lifecycle events.
type LifecycleEvent struct {
	Event     string    `json:"event"`
	UserID    string    `json:"userId"`
	Timestamp time.Time `json:"timestamp"`
	Reason    string    `json:"reason,omitempty"`
}

// MarshalJSON serializes LifecycleEvent with dual key compatibility for heterogeneous downstream consumers.
// Why: store_auth expects "event" and "userId", while store_order expects "event_type" and "user_id".
func (e LifecycleEvent) MarshalJSON() ([]byte, error) {
	type Alias LifecycleEvent
	return json.Marshal(&struct {
		Alias
		EventType string `json:"event_type"`
		UserIDAlt string `json:"user_id"`
	}{
		Alias:     Alias(e),
		EventType: e.Event,
		UserIDAlt: e.UserID,
	})
}

// Serialize converts the LifecycleEvent into JSON bytes.
// Why: Enforces valid JSON structure before pushing raw bytes onto Kafka partition buffers.
func (e *LifecycleEvent) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

