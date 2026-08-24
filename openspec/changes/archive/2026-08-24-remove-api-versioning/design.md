## Context

The `store_user` microservice previously supported route mounting under both `/api/v1/users` and an alias `/api/users`. In accordance with the simplified unversioned routing strategy across the store platform, we are eliminating `/api/v1/users` and establishing `/api/users` as the single canonical path for user profile and account operations.

## Goals / Non-Goals

**Goals:**
- Consolidate HTTP router endpoints exclusively under `/api/users/*` (`/api/users/profile` and `/api/users/account`).
- Update the OpenAPI 3.1 specifications (`docs/openapi.yaml` and `docs/openapi.json`) to reflect unversioned endpoints.
- Update all integration and unit tests in `internal/router` and `internal/handler` to verify unversioned routes.

**Non-Goals:**
- Changing business logic in `UserService`, PostgreSQL repository queries, Kafka event schemas, or gRPC communication with `store_order`.
- Modifying payload schemas for `GET`, `PUT`, or `DELETE` requests.

## Decisions

### Decision 1: Single Ingress Route Mount
- **Choice**: Mount user routes directly under `r.Route("/api/users", mountUserRoutes)` in Chi router and remove `r.Route("/api/v1/users", mountUserRoutes)`.
- **Rationale**: Removes ambiguity, reduces routing overhead, and aligns with the platform directive to avoid URI versioning.
- **Alternatives Considered**: Keeping `/api/v1/users` as a redirect or backward-compatible alias. Rejected to enforce strict unversioned standards across all microservices.

### Decision 2: OpenAPI Specification Synchronization
- **Choice**: Directly update `paths` in `docs/openapi.yaml` and `docs/openapi.json` from `/api/v1/users/...` to `/api/users/...`.
- **Rationale**: Keeps API documentation strictly aligned with runtime router endpoints.

## Risks / Trade-offs

- **[Risk] Existing clients calling `/api/v1/users/*` will receive HTTP 404**:
  - *Mitigation*: Upstream API gateway (`store_gateway`) routes external traffic to `/api/users/*`. If required during migration, the gateway handles path rewriting.

## Migration Plan

1. Update Chi router mount in `internal/router/router.go`.
2. Update unit and integration tests in `internal/router/router_test.go`.
3. Update OpenAPI schema definitions in `docs/openapi.yaml` and `docs/openapi.json`.
4. Run `go test ./...` to ensure all tests pass.
