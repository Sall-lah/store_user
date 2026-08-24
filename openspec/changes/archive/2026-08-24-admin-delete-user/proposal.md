## Why

Store administrators need the ability to delete user accounts directly (e.g., for compliance, policy enforcement, or customer support) without being blocked by active order verification checks that constrain self-service user deletions. When an admin deletes a user, the system must forcefully purge the user's profile and notification records, and publish a domain event to Kafka so downstream services like `store_auth` immediately blacklist active JWTs, revoke refresh tokens, and remove login credentials.

## What Changes

- Introduce a new administrative endpoint `DELETE /api/admin/users/{id}` protected by role verification (`X-User-Role: ADMIN`).
- Implement role-checking middleware (`RequireRole`) to enforce gateway-offloaded identity roles.
- Implement an administrative deletion flow in `UserService` that explicitly bypasses the `store_order` gRPC active order check.
- Purge the user's profile (`user_profiles`) and associated notifications (`user_notifications`).
- Publish a `user.deleted` event to the Kafka topic `user.events` with payload `{ event: "user.deleted", userId: "<id>", timestamp: "<ISO-8601>", reason: "admin_deletion" }` for `store_auth` consumption.
- Update OpenAPI 3.1 YAML and JSON documentation to include `DELETE /api/admin/users/{id}` under the `Admin User Management` tag.

## Capabilities

### New Capabilities
- `admin-user-management`: Administrative user lifecycle operations, specifically forceful account deletion by user ID (`DELETE /api/admin/users/{id}`) without blocking on in-flight orders and with Kafka domain event propagation.

### Modified Capabilities
<!-- None: Existing self-service account deletion (account-deletion) requirements remain unchanged. -->

## Impact

- **Affected Code / APIs**:
  - `internal/middleware/auth.go`: add `RequireRole` middleware.
  - `internal/service/service.go`: add `AdminDeleteUser(ctx context.Context, targetUserID string) error`.
  - `internal/handler/handler.go` (or `internal/handler/admin.go`): add handler for `DELETE /api/admin/users/{id}`.
  - `internal/router/router.go`: mount `/api/admin/users` route group with rate limiting and role verification.
  - `docs/openapi.yaml` & `docs/openapi.json`: document new admin endpoint.
- **Microservice Ecosystem**:
  - `store_gateway`: routes `/api/admin/users/` through `auth-offload` to `store_user`.
  - `store_auth`: consumes `user.deleted` Kafka event from `user.events` topic to purge user credentials and blacklist JWTs in Redis.
