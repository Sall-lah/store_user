package kafka

import (
	"context"
	"time"
)

// MockProducer implements Producer for testing without real Kafka brokers.
// Why: Enables validating domain event emission, keys, and payload contents in unit tests.
type MockProducer struct {
	PublishedMessages []PublishedMessage
	PublishError      error
}

// PublishedMessage stores captured messages published to the mock producer.
type PublishedMessage struct {
	Topic   string
	Key     string
	Payload []byte
}

// Publish records the published message in memory.
func (m *MockProducer) Publish(ctx context.Context, topic, key string, payload []byte) error {
	if m.PublishError != nil {
		return m.PublishError
	}
	m.PublishedMessages = append(m.PublishedMessages, PublishedMessage{
		Topic:   topic,
		Key:     key,
		Payload: payload,
	})
	return nil
}

// PublishUserDeleted creates the lifecycle event and stores it in memory.
func (m *MockProducer) PublishUserDeleted(ctx context.Context, topic, userID, reason string) error {
	event := LifecycleEvent{
		Event:     EventUserDeleted,
		UserID:    userID,
		Timestamp: time.Now().UTC(),
		Reason:    reason,
	}
	payload, err := event.Serialize()
	if err != nil {
		return err
	}
	return m.Publish(ctx, topic, userID, payload)
}

// Close marks mock producer as closed.
func (m *MockProducer) Close() error {
	return nil
}
