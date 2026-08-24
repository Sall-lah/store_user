## Context

The `store_user` service previously included support for managing user communication channel preferences via the `user_notification_preferences` PostgreSQL table and two HTTP endpoints (`GET` and `PUT /api/notifications/preferences`). The requirements now mandate removing this table and all associated application code from `store_user`.

## Goals / Non-Goals

**Goals:**
- Drop the `user_notification_preferences` table from PostgreSQL and remove its model definition from `prisma/schema.prisma`.
- Regenerate the Prisma Go client (`internal/db`).
- Remove `GET /api/notifications/preferences` and `PUT /api/notifications/preferences` routes, handlers, service logic, repository queries, DTOs, and mocks.
- Clean up test files across `handler`, `service`, `repository`, and `router` packages to remove references to preferences.
- Update OpenAPI 3.1 specifications (`docs/openapi.json` and `docs/openapi.yaml`) to remove preferences paths and components.
- Ensure all remaining test suites compile and pass.

**Non-Goals:**
- Modifying core in-app notifications (`GET /api/notifications`, `PATCH /api/notifications/{id}/read`, `POST /api/notifications/read-all`, `DELETE /api/notifications/{id}`).
- Altering user profile management or account deletion logic.

## Decisions

### Decision 1: Complete End-to-End Removal Across All Layers
- **Choice**: Completely remove the endpoints and underlying code across handler, service, repository, and router rather than leaving stub/deprecated handlers.
- **Rationale**: Keeps the codebase clean, reduces dead code and technical debt, and prevents confusion regarding supported features.
- **Alternatives Considered**:
  - *Deprecating endpoints with no-op responses*: Unnecessary overhead and confusing API semantics.

### Decision 2: Prisma Schema Removal and Client Regeneration
- **Choice**: Remove `model user_notification_preferences` from `prisma/schema.prisma` and execute `prisma generate` to update `internal/db`. Provide a SQL migration to drop the table in PostgreSQL (`DROP TABLE IF EXISTS "user_notification_preferences" CASCADE;`).
- **Rationale**: Ensures Go compile-time type safety with the updated database schema.

### Decision 3: OpenAPI Documentation & Test Cleanup
- **Choice**: Remove all preferences schemas, path items, and parameters from `docs/openapi.json` and `docs/openapi.yaml`. Remove associated test cases from `notification_test.go` and `router_test.go`.
- **Rationale**: Keeps API documentation and automated test suites strictly aligned with the implemented endpoints.

## Risks / Trade-offs

- **[Risk] Clients invoking `/api/notifications/preferences` will receive 404 Not Found**
  → *Mitigation*: Update API documentation and OpenAPI specifications to reflect endpoint removal.
- **[Risk] Database schema drift if PostgreSQL table is not dropped**
  → *Mitigation*: Provide explicit migration SQL (`DROP TABLE IF EXISTS "user_notification_preferences" CASCADE;`) for deployment execution.
