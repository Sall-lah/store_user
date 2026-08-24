## 1. Kafka Domain Event Contract & Producer

- [x] 1.1 Add `EventUserBanned = "user.banned"` to `internal/kafka/event.go` and update `LifecycleEvent` serialization to output dual-compatible keys (`event`, `event_type`, `userId`, `user_id`, `reason`, `timestamp`).
- [x] 1.2 Add `PublishUserBanned(ctx context.Context, topic, userID, reason string) error` to `Producer` interface and concrete `KafkaProducer` in `internal/kafka/producer.go`.
- [x] 1.3 Add `PublishUserBanned` to `MockProducer` in `internal/kafka/mock.go` and add unit tests in `internal/kafka/producer_test.go`.

## 2. Service Layer Implementation

- [x] 2.1 Add `BanUserRequest` struct and `AdminBanUser(ctx context.Context, targetUserID string, req BanUserRequest) error` to `UserService` interface in `internal/service/service.go`.
- [x] 2.2 Implement `AdminBanUser` in `UserServiceImpl` in `internal/service/service.go` to validate target UUID, preserve database records, and publish `user.banned` event to Kafka topic `user.events`.
- [x] 2.3 Add unit tests for `AdminBanUser` in `internal/service/service_test.go` covering success, UUID validation, and Kafka failure handling.

## 3. Handler & Router Layer

- [x] 3.1 Implement `BanUser(w http.ResponseWriter, r *http.Request)` in `internal/handler/admin.go` to parse target UUID and request body reason, and invoke `AdminBanUser`.
- [x] 3.2 Mount route `POST /{id}/ban` in `internal/router/router.go` under `mountAdminRoutes` protected by `RequireRole("ADMIN", "admin")` and rate limiting.
- [x] 3.3 Add unit tests in `internal/handler/handler_test.go` testing authorized ban (200 OK), non-admin forbidden (403), and invalid UUIDs (400).

## 4. Documentation & Verification

- [x] 4.1 Update OpenAPI documentation in `docs/openapi.yaml` and `docs/openapi.json` to include `POST /api/admin/users/{id}/ban` endpoint and request/response schemas.
- [x] 4.2 Run full test suite (`go test ./...`) to verify all unit and integration tests pass cleanly.
