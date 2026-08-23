## 1. Project Initialization & Dependencies

- [x] 1.1 Initialize Go module (github.com/Sall-lah/store_user) and configure go.mod with Chi router, Prisma Client Go, go-redis, kafka-go, and gRPC dependencies
- [x] 1.2 Define Prisma schema for Supabase PostgreSQL (user_profiles table) and generate database client
- [x] 1.3 Implement environment configuration loader (internal/config) with validation for database URL, Redis, Kafka brokers, and order gRPC endpoint

## 2. Infrastructure & Platform Components

- [x] 2.1 Implement Redis sliding-window rate limiter (internal/ratelimit) with atomic Lua script and fail-open resilience
- [x] 2.2 Implement Kafka event publisher (internal/kafka) configured for user.events topic with LifecycleEvent payload
- [x] 2.3 Implement gRPC client (internal/client/order) connecting to store_order using store_proto for CheckActiveOrders

## 3. Database Layer & Repository

- [x] 3.1 Implement UserProfile repository interface and Prisma database adapter (internal/repository) for profile retrieval, upsert, and hard deletion
- [x] 3.2 Write repository unit tests with mock database adapters

## 4. Business Logic & Service Layer

- [x] 4.1 Implement Profile service (internal/service) handling profile fetch and updates with input validation
- [x] 4.2 Implement Account Deletion service (internal/service) orchestrating gRPC pre-flight active order check, profile hard-deletion, and Kafka event publishing
- [x] 4.3 Write service unit tests for profile updates, order conflict rejection (409), and gRPC timeout failure (503)

## 5. Transport Layer, Middleware & HTTP Routing

- [x] 5.1 Implement HTTP middleware for gateway identity extraction (X-User-Id), max body size limiting (64KB), and Redis rate limiting
- [x] 5.2 Implement text sanitization and input validators for profile update payloads
- [x] 5.3 Implement HTTP REST handlers for GET /api/v1/users/profile, PUT /api/v1/users/profile, DELETE /api/v1/users/account, and /health
- [x] 5.4 Wire Chi router and server boot lifecycle in cmd/server/main.go with graceful shutdown

## 6. Verification & Documentation

- [x] 6.1 Create OpenAPI specification (docs/openapi.yaml and docs/openapi.json) for store_user endpoints matching the platform documentation standard
- [x] 6.2 Write integration tests (test/integration_test.go) and verify end-to-end user deletion with mock order service and Kafka emitter
- [x] 6.3 Create Dockerfile and .env.example with production readiness configurations
