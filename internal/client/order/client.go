package order

import (
	"context"
	"fmt"
	"strings"
	"time"

	orderv1 "github.com/Sall-lah/store_proto/gen/go/store/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client defines the contract for communicating with the Order Service over gRPC.
// Why: Decouples service handlers from concrete gRPC transport and enables deterministic mock injection in unit tests.
type Client interface {
	CheckActiveOrders(ctx context.Context, userID string) (*orderv1.CheckActiveOrdersResponse, error)
	Close() error
}

// GrpcOrderClient implements the Client interface using Google gRPC client stubs.
type GrpcOrderClient struct {
	conn    *grpc.ClientConn
	client  orderv1.OrderServiceClient
	timeout time.Duration
}

// NewOrderClient initializes a resilient gRPC connection to the target Order Service host.
// Why: Pre-establishes non-blocking connection channel and configures default timeouts to prevent thread hanging.
func NewOrderClient(target string, timeout time.Duration, opts ...grpc.DialOption) (*GrpcOrderClient, error) {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	if len(opts) == 0 {
		opts = []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
	}

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial order service grpc at %s: %w", target, err)
	}

	return &GrpcOrderClient{
		conn:    conn,
		client:  orderv1.NewOrderServiceClient(conn),
		timeout: timeout,
	}, nil
}

// CheckActiveOrders executes synchronous pre-flight inspection of customer order lifecycle status.
// Why: Enforces hard boundary check before user deletion to prevent orphaned fulfillment processes.
func (c *GrpcOrderClient) CheckActiveOrders(ctx context.Context, userID string) (*orderv1.CheckActiveOrdersResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user_id cannot be empty")
	}

	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req := &orderv1.CheckActiveOrdersRequest{
		UserId: userID,
	}

	res, err := c.client.CheckActiveOrders(callCtx, req)
	if err != nil {
		return nil, fmt.Errorf("order service CheckActiveOrders RPC failed: %w", err)
	}

	return res, nil
}

// Close terminates the underlying gRPC network connection.
// Why: Frees transport sockets and background keepalive goroutines upon service shutdown.
func (c *GrpcOrderClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
