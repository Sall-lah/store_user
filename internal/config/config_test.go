package config

import (
	"os"
	"testing"
	"time"
)

func TestConfigLoadSuccess(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/user_db?sslmode=disable")
	os.Setenv("SERVER_PORT", "8082")
	os.Setenv("ORDER_SERVICE_GRPC_ADDR", "localhost:50051")
	defer os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.ServerPort != "8082" {
		t.Errorf("expected ServerPort '8082', got: %s", cfg.ServerPort)
	}
	if cfg.OrderServiceGrpcAddr != "localhost:50051" {
		t.Errorf("expected OrderServiceGrpcAddr 'localhost:50051', got: %s", cfg.OrderServiceGrpcAddr)
	}
	if cfg.RateLimitWindow != 60*time.Second {
		t.Errorf("expected 60s window, got: %v", cfg.RateLimitWindow)
	}
}

func TestConfigLoadMissingDatabaseURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE_URL", "")
	os.Setenv("ORDER_SERVICE_GRPC_ADDR", "localhost:50051")
	os.Setenv("KAFKA_BROKERS", "localhost:9092")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL, got nil")
	}
}
