## 1. Middleware & Authentication

- [x] 1.1 Implement `RequireRole(roles ...string)` middleware in `internal/middleware/auth.go` to enforce role authorization (e.g. `ADMIN`) from context.
- [x] 1.2 Add comprehensive unit tests in `internal/middleware/middleware_test.go` verifying 401 on missing identity, 403 on non-admin roles (`CUSTOMER`), and next handler execution on valid `ADMIN` / `admin`.

## 2. Repository & Service Layer

- [x] 2.1 Add `DeleteByUserID(ctx context.Context, userID string) error` to `NotificationRepository` and Prisma implementation in `internal/repository/notification.go` to support batch notification cleanup.
- [x] 2.2 Add `AdminDeleteUser(ctx context.Context, targetUserID string) error` to `UserService` interface and `UserServiceImpl` in `internal/service/service.go` that skips `store_order` active order checks, deletes profile and notification records, and publishes a `user.deleted` event with `reason: "admin_deletion"` to Kafka topic `user.events`.
- [x] 2.3 Add unit tests for `AdminDeleteUser` in `internal/service/service_test.go` covering success with active orders bypassed, non-existent profiles (idempotent), invalid UUIDs, and Kafka dispatch.

## 3. Handler & Routing Layer

- [x] 3.1 Implement `AdminHandler` (or extend `ProfileHandler`) in `internal/handler/admin.go` with method `DeleteUser(w http.ResponseWriter, r *http.Request)` that parses and validates the target UUID from URL path parameter and delegates to `AdminDeleteUser`.
- [x] 3.2 Mount the admin route group `r.Route("/api/admin/users", ...)` in `internal/router/router.go` with `AuthIdentity`, `RequireRole("ADMIN", "admin")`, and sliding-window rate limiting.
- [x] 3.3 Add HTTP handler and router tests in `internal/handler/handler_test.go` and `internal/router/router_test.go` verifying status codes (200, 400, 401, 403, 500).

## 4. API Documentation

- [x] 4.1 Update `docs/openapi.yaml` and `docs/openapi.json` to define `DELETE /api/admin/users/{id}` under the `Admin User Management` tag with full request/response schemas.
- [x] 4.2 Verify documentation endpoint delivery and OpenAPI schema validation.
