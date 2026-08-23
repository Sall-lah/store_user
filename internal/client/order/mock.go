package order

import (
	"context"

	orderv1 "github.com/Sall-lah/store_proto/gen/go/store/order/v1"
)

// MockClient implements Client for service layer testing.
// Why: Enables simulating order active/completed lifecycle responses and network failure scenarios in unit tests.
type MockClient struct {
	CheckActiveOrdersFunc func(ctx context.Context, userID string) (*orderv1.CheckActiveOrdersResponse, error)
	Closed                bool
}

// CheckActiveOrders dispatches to the configured mock function or returns zero active orders by default.
func (m *MockClient) CheckActiveOrders(ctx context.Context, userID string) (*orderv1.CheckActiveOrdersResponse, error) {
	if m.CheckActiveOrdersFunc != nil {
		return m.CheckActiveOrdersFunc(ctx, userID)
	}
	return &orderv1.CheckActiveOrdersResponse{
		HasActiveOrders:  false,
		ActiveOrderCount: 0,
		ActiveOrders:     []*orderv1.ActiveOrderSummary{},
		Message:          "User has no active orders",
	}, nil
}

// Close marks the mock client as closed.
func (m *MockClient) Close() error {
	m.Closed = true
	return nil
}
