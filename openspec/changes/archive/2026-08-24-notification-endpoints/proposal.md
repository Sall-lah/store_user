## Why

Users require visibility into real-time account and order notifications within the platform, as well as granular control over their notification channel preferences. Additionally, developers and gateway operators need interactive Swagger UI and OpenAPI specification endpoints directly exposed by the user service for automated API discovery and testing.

## What Changes

- Add notification REST endpoints under `/api/users/notifications` for retrieving paginated user notifications, marking notifications as read, and deleting notifications.
- Add notification preference endpoints under `/api/users/notifications/preferences` to retrieve and update user communication channels (email, push, SMS, promotional).
- Add interactive Swagger UI and OpenAPI specification HTTP routes (`GET /docs`, `GET /docs/openapi.yaml`, `GET /docs/openapi.json`, and `/swagger`) to serve live API documentation.
- Update `docs/openapi.yaml` and `docs/openapi.json` to define all notification models, parameters, request/response schemas, security schemes, and rate limiting headers.
- Extend Prisma database schema with models for notifications and notification preferences.
- Implement repository, service, and handler layers with modular architecture, input validation, XSS sanitization, and full unit test coverage.
- Apply Redis sliding-window rate limiting on all notification endpoints.

## Capabilities

### New Capabilities
- `user-notifications`: Retrieval, status mutation (read/unread), deletion of in-app user notifications, and user notification channel preferences management.
- `api-documentation-ui`: Interactive Swagger UI and raw OpenAPI specification delivery served via Chi router endpoints.

### Modified Capabilities
- `security-rate-limiting`: Extend sliding-window rate limiting policies and route guards to cover new notification endpoints.

## Impact

- **Affected Code**: `internal/router/router.go`, `internal/handler/`, `internal/service/`, `internal/repository/`, `prisma/schema.prisma`.
- **API Surface**: New endpoints under `/api/users/notifications`, `/api/users/notifications/preferences`, and `/docs` / `/swagger`.
- **Dependencies**: Embed or serve Swagger UI assets (e.g. `http.FS` with static Swagger UI or swagger embed handler).
- **Documentation**: `docs/openapi.yaml` and `docs/openapi.json` updated with notification and documentation endpoints.
