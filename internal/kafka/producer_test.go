package kafka

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestLifecycleEventSerialization(t *testing.T) {
	now := time.Now().UTC()
	event := LifecycleEvent{
		Event:     EventUserDeleted,
		UserID:    "user-123",
		Timestamp: now,
		Reason:    "user_requested_deletion",
	}

	bytes, err := event.Serialize()
	if err != nil {
		t.Fatalf("failed to serialize event: %v", err)
	}

	var parsed LifecycleEvent
	if err := json.Unmarshal(bytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal serialized event: %v", err)
	}

	if parsed.Event != EventUserDeleted {
		t.Errorf("expected event '%s', got '%s'", EventUserDeleted, parsed.Event)
	}
	if parsed.UserID != "user-123" {
		t.Errorf("expected userId 'user-123', got '%s'", parsed.UserID)
	}
	if parsed.Reason != "user_requested_deletion" {
		t.Errorf("expected reason 'user_requested_deletion', got '%s'", parsed.Reason)
	}
}

func TestMockProducer(t *testing.T) {
	mock := &MockProducer{}
	err := mock.PublishUserDeleted(context.Background(), "user.events", "usr_test", "user_requested_deletion")
	if err != nil {
		t.Fatalf("unexpected error from mock producer: %v", err)
	}

	if len(mock.PublishedMessages) != 1 {
		t.Fatalf("expected 1 published message, got %d", len(mock.PublishedMessages))
	}
	if mock.PublishedMessages[0].Key != "usr_test" {
		t.Errorf("expected key 'usr_test', got '%s'", mock.PublishedMessages[0].Key)
	}
}
