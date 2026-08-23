package order

import (
	"context"
	"testing"

	orderv1 "github.com/Sall-lah/store_proto/gen/go/store/order/v1"
)

func TestMockOrderClient(t *testing.T) {
	mock := &MockClient{
		CheckActiveOrdersFunc: func(ctx context.Context, userID string) (*orderv1.CheckActiveOrdersResponse, error) {
			return &orderv1.CheckActiveOrdersResponse{
				HasActiveOrders:  true,
				ActiveOrderCount: 1,
				ActiveOrders: []*orderv1.ActiveOrderSummary{
					{
						OrderId:     "ord-123",
						OrderNumber: "ORD-2026-001",
						Status:      orderv1.OrderStatus_ORDER_STATUS_PROCESSING,
						TotalAmount: 150000,
					},
				},
				Message: "Active order blocking deletion",
			}, nil
		},
	}

	res, err := mock.CheckActiveOrders(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.HasActiveOrders {
		t.Errorf("expected HasActiveOrders=true")
	}
	if res.ActiveOrderCount != 1 {
		t.Errorf("expected 1 active order, got %d", res.ActiveOrderCount)
	}
	if len(res.ActiveOrders) != 1 || res.ActiveOrders[0].OrderId != "ord-123" {
		t.Errorf("unexpected active orders payload")
	}
}
