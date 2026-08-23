package service

import (
	"errors"
	"fmt"

	orderv1 "github.com/Sall-lah/store_proto/gen/go/store/order/v1"
)

var (
	// ErrProfileNotFound is returned when a user profile does not exist.
	ErrProfileNotFound = errors.New("user profile not found")

	// ErrInvalidUserID is returned when an empty or malformed user ID is provided.
	ErrInvalidUserID = errors.New("invalid user ID")

	// ErrValidationFailed is returned when profile input validation fails.
	ErrValidationFailed = errors.New("profile validation failed")

	// ErrOrderServiceUnavailable is returned when store_order gRPC check fails or times out.
	ErrOrderServiceUnavailable = errors.New("order service is currently unavailable")

	// ErrKafkaPublishFailed is returned when emitting the user.deleted lifecycle event fails.
	ErrKafkaPublishFailed = errors.New("failed to publish account deletion lifecycle event")
)

// ActiveOrdersConflictError represents a business rule rejection when account deletion is blocked by active orders.
// Why: Provides structured details of blocking in-flight orders for 409 Conflict HTTP responses.
type ActiveOrdersConflictError struct {
	Count        int32
	ActiveOrders []*orderv1.ActiveOrderSummary
	Message      string
}

func (e *ActiveOrdersConflictError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("cannot delete account: %d active order(s) in flight", e.Count)
}
