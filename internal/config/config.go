package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

		// Config encapsulates runtime configuration parameters validated on server boot.
type Config struct {
	ServerPort                 string
	Env                        string
	DatabaseURL                string
	RedisURL                   string
	RedisPassword              string
	RateLimitMaxRequests       int
	RateLimitWindow            time.Duration
	RateLimitDeleteMaxRequests int
	RateLimitDeleteWindow      time.Duration
	KafkaBrokers               []string
	KafkaTopicUserEvents       string
	OrderServiceGrpcAddr       string
	OrderServiceGrpcTimeout    time.Duration
	EnableRateLimiter          bool
}

// Load parses configuration from environment variables and validates critical service endpoints.
// Why: Enforces fail-fast startup behavior if essential credentials or network dependencies are missing.
func Load() (*Config, error) {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	dbURL := cleanEnv("DATABASE_URL", "")
	if dbURL != "" {
		_ = os.Setenv("DATABASE_URL", dbURL)
	}

	cfg := &Config{
		ServerPort:                 cleanEnv("SERVER_PORT", "8082"),
		Env:                        cleanEnv("ENV", "development"),
		DatabaseURL:                dbURL,
		RedisURL:                   cleanEnv("REDIS_URL", "redis://localhost:6379"),
		RedisPassword:              cleanEnv("REDIS_PASSWORD", ""),
		RateLimitMaxRequests:       getEnvAsInt("RATE_LIMIT_MAX_REQUESTS", 30),
		RateLimitDeleteMaxRequests: getEnvAsInt("RATE_LIMIT_DELETE_MAX_REQUESTS", 3),
		KafkaBrokers:               []string{cleanEnv("KAFKA_BROKERS", "localhost:9092")},
		KafkaTopicUserEvents:       cleanEnv("KAFKA_TOPIC_USER_EVENTS", "user.events"),
		OrderServiceGrpcAddr:       cleanEnv("ORDER_SERVICE_GRPC_ADDR", "localhost:50051"),
		OrderServiceGrpcTimeout:    time.Duration(getEnvAsInt("ORDER_SERVICE_GRPC_TIMEOUT_MS", 3000)) * time.Millisecond,
		EnableRateLimiter:          getEnvAsBool("ENABLE_RATE_LIMITER", true),
	}

	windowSec := getEnvAsInt("RATE_LIMIT_WINDOW_SECONDS", 60)
	cfg.RateLimitWindow = time.Duration(windowSec) * time.Second

	deleteWindowSec := getEnvAsInt("RATE_LIMIT_DELETE_WINDOW_SECONDS", 60)
	cfg.RateLimitDeleteWindow = time.Duration(deleteWindowSec) * time.Second

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}


// validate ensures all mandatory configurations are present and sane.
// Why: Prevents runtime panics or silent failure states when downstream connections are attempted.
func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if len(c.KafkaBrokers) == 0 || c.KafkaBrokers[0] == "" {
		return fmt.Errorf("KAFKA_BROKERS must specify at least one valid broker host")
	}
	if c.OrderServiceGrpcAddr == "" {
		return fmt.Errorf("ORDER_SERVICE_GRPC_ADDR is required for pre-flight order checks")
	}
	return nil
}


func cleanEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		trimmed := strings.TrimSpace(val)
		trimmed = strings.Trim(trimmed, "\"'")
		return trimmed
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return fallback
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return fallback
	}
	return val
}


func getEnvAsBool(key string, fallback bool) bool {
	valStr := os.Getenv(key)
	if valStr == "" {
		return fallback
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return fallback 
	}
	return val
}
