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

	// Verify dual key compatibility for store_order and store_auth
	var rawMap map[string]interface{}
	if err := json.Unmarshal(bytes, &rawMap); err != nil {
		t.Fatalf("failed to unmarshal raw map: %v", err)
	}
	if rawMap["event"] != EventUserDeleted || rawMap["event_type"] != EventUserDeleted {
		t.Errorf("expected dual event/event_type keys, got %+v", rawMap)
	}
	if rawMap["userId"] != "user-123" || rawMap["user_id"] != "user-123" {
		t.Errorf("expected dual userId/user_id keys, got %+v", rawMap)
	}
}

func TestLifecycleEventBannedSerialization(t *testing.T) {
	now := time.Now().UTC()
	event := LifecycleEvent{
		Event:     EventUserBanned,
		UserID:    "user-456",
		Timestamp: now,
		Reason:    "fraudulent_activity",
	}

	bytes, err := event.Serialize()
	if err != nil {
		t.Fatalf("failed to serialize banned event: %v", err)
	}

	var rawMap map[string]interface{}
	if err := json.Unmarshal(bytes, &rawMap); err != nil {
		t.Fatalf("failed to unmarshal raw map: %v", err)
	}
	if rawMap["event"] != EventUserBanned || rawMap["event_type"] != EventUserBanned {
		t.Errorf("expected dual event/event_type keys for banned, got %+v", rawMap)
	}
	if rawMap["userId"] != "user-456" || rawMap["user_id"] != "user-456" {
		t.Errorf("expected dual userId/user_id keys for banned, got %+v", rawMap)
	}
	if rawMap["reason"] != "fraudulent_activity" {
		t.Errorf("expected reason 'fraudulent_activity', got '%v'", rawMap["reason"])
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

	// Test PublishUserBanned
	err = mock.PublishUserBanned(context.Background(), "user.events", "usr_banned", "chargeback_abuse")
	if err != nil {
		t.Fatalf("unexpected error from mock publish banned: %v", err)
	}
	if len(mock.PublishedMessages) != 2 {
		t.Fatalf("expected 2 published messages, got %d", len(mock.PublishedMessages))
	}
	if mock.PublishedMessages[1].Key != "usr_banned" {
		t.Errorf("expected key 'usr_banned', got '%s'", mock.PublishedMessages[1].Key)
	}
}
