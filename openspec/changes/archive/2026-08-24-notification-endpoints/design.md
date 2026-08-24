## Context

The `store_user` service is responsible for user profile management, account lifecycle, and user-centric configuration within the e-commerce microservices ecosystem. As part of enhancing user engagement and self-service capabilities, `store_user` requires dedicated endpoints for:
1. Managing in-app notifications (listing notification feeds, marking individual or all items as read, deleting notifications).
2. Managing user notification preferences (email, push, SMS, order updates, and promotional channels).
3. Exposing interactive Swagger UI documentation and raw OpenAPI 3.1 specifications (`/docs`, `/swagger`, `/docs/openapi.yaml`, `/docs/openapi.json`) for seamless developer integration, gateway routing, and API testing.

## Goals / Non-Goals

**Goals:**
- Provide REST endpoints for user notification retrieval with pagination and read-status filtering (`GET /api/users/notifications`).
- Provide REST endpoints for mutating notification status: marking single notification as read (`PATCH /api/users/notifications/{id}/read`), marking all as read (`POST /api/users/notifications/read-all`), and deleting a notification (`DELETE /api/users/notifications/{id}`).
- Provide REST endpoints for retrieving (`GET /api/users/notifications/preferences`) and updating (`PUT /api/users/notifications/preferences`) user channel preferences.
- Serve interactive Swagger UI at `/docs` and `/swagger`, and expose OpenAPI 3.1 specifications at `/docs/openapi.yaml` and `/docs/openapi.json`.
- Document all existing and new endpoints thoroughly in `docs/openapi.yaml` and `docs/openapi.json`.
- Enforce Redis sliding-window rate limiting on all notification endpoints.
- Ensure 100% modularity following repository standards: separate files for handlers, services, and repositories with comprehensive JSDoc/TSDoc/GoDoc comments.

**Non-Goals:**
- Directly dispatching external push notifications, emails, or SMS (handled downstream by `store_notification` worker service via event broker).
- Implementing WebSocket / SSE streaming connections in this change.

## Decisions

### Decision 1: Dedicated Prisma Models for Notifications and Preferences
- **Choice**: Add `user_notifications` and `user_notification_preferences` models to `prisma/schema.prisma`.
- **Rationale**: Isolates notification data and preferences per user with UUID keys. Composite indices on `(user_id, is_read)` and `(user_id, created_at DESC)` ensure fast pagination and query performance.
- **Alternatives Considered**:
  - *Storing notification settings inside user_profiles table*: Bloats the profile model with non-profile preferences and increases table write contention.
  - *External storage in Redis only*: Lacks ACID persistence, backup durability, and transactional consistency.

### Decision 2: Modular Layered Architecture (Handler, Service, Repository)
- **Choice**: Implement distinct modules:
  - `internal/handler/notification.go` & `internal/handler/doc.go`
  - `internal/service/notification.go` & `internal/service/model_notification.go`
  - `internal/repository/notification.go`
- **Rationale**: Strict separation of concerns keeps files readable, maintainable, and independently testable with mock implementations.
- **Alternatives Considered**:
  - *Merging notification logic into existing `ProfileHandler` / `UserService`*: Violates Single Responsibility Principle and causes file bloat.

### Decision 3: Interactive Swagger UI & OpenAPI Specification Delivery
- **Choice**: Mount HTTP handlers on `/docs` and `/swagger` serving an embedded Swagger UI HTML page, configured to parse `/docs/openapi.yaml` or `/docs/openapi.json` served locally by the service.
- **Rationale**: Provides zero-setup interactive documentation without requiring external tooling. Local file serving ensures documentation is always version-locked to the running binary.
- **Alternatives Considered**:
  - *Generating swagger dynamically at runtime via reflection*: High runtime overhead, fragile, and less accurate than hand-crafted OpenAPI 3.1 specifications.
  - *External documentation portal only*: Disconnects local microservice testing from API contract definitions.

### Decision 4: Tiered Redis Sliding-Window Rate Limiting
- **Choice**: Apply 60 requests/minute for notification and preference reads, and 30 requests/minute for status mutations and deletions.
- **Rationale**: Prevents accidental query spamming and polling abuse while allowing responsive user experience.
- **Alternatives Considered**:
  - *No rate limiting on notification routes*: Leaves database vulnerable to scraping and high-frequency UI polling loops.

## Risks / Trade-offs

- **[Risk] High volume of in-app notifications degrading database performance**
  → *Mitigation*: Index on `(user_id, created_at)` and enforce pagination bounds (max limit 100, default 20).
- **[Risk] Swagger UI asset loading in restricted/offline environments**
  → *Mitigation*: Raw OpenAPI spec endpoints (`/docs/openapi.yaml`, `/docs/openapi.json`) are served purely from local disk/embedded assets without external dependencies.
- **[Risk] Race condition when marking all notifications as read**
  → *Mitigation*: Atomic database update query scoped to the caller's `user_id` and unread status.
