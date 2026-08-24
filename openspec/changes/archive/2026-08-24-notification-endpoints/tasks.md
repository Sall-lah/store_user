## 1. Database Schema & Data Models

- [x] 1.1 Update `prisma/schema.prisma` with `user_notifications` and `user_notification_preferences` models
- [x] 1.2 Define notification domain structures, request/response DTOs, and error types in `internal/service/model_notification.go` and `internal/repository/model_notification.go`

## 2. Repository Layer

- [x] 2.1 Implement `NotificationRepository` interface and Prisma operations in `internal/repository/notification.go`
- [x] 2.2 Create mock repository in `internal/repository/mock_notification.go` for isolated unit testing
- [x] 2.3 Implement repository unit tests in `internal/repository/notification_test.go`

## 3. Service Layer

- [x] 3.1 Implement `NotificationService` interface and business validation in `internal/service/notification.go`
- [x] 3.2 Implement comprehensive service layer unit tests in `internal/service/notification_test.go`

## 4. HTTP Handlers & Documentation UI

- [x] 4.1 Implement `NotificationHandler` in `internal/handler/notification.go` covering notification feed, read-status mutations, deletion, and preferences
- [x] 4.2 Implement `DocHandler` in `internal/handler/doc.go` to serve Swagger UI HTML and raw OpenAPI specifications
- [x] 4.3 Implement handler unit tests in `internal/handler/notification_test.go` and `internal/handler/doc_test.go`

## 5. Router, Middleware & Server Integration

- [x] 5.1 Mount documentation endpoints (`/docs`, `/swagger`, `/docs/openapi.yaml`, `/docs/openapi.json`) in `internal/router/router.go`
- [x] 5.2 Mount notification and preference routes under `/api/users/notifications` with identity extraction and Redis rate limiting in `internal/router/router.go`
- [x] 5.3 Wire notification repository, service, and handlers into `cmd/server/main.go`
- [x] 5.4 Update router integration tests in `internal/router/router_test.go`

## 6. OpenAPI & Swagger Contracts

- [x] 6.1 Update `docs/openapi.yaml` with schemas, endpoints, response codes, and security definitions for notification and doc endpoints
- [x] 6.2 Update `docs/openapi.json` to synchronize with OpenAPI 3.1 specification
