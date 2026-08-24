## Why

Currently, `store_user` only supports account deletion (`user.deleted`), but downstream services (`store_auth` and `store_order`) have active Kafka consumers designed to handle `user.banned` events. Without an administrative ban capability in `store_user`, administrators cannot freeze abusive, fraudulent, or terms-violating accounts without permanently purging customer PII and audit trails.

## What Changes

- Add administrative ban endpoint `POST /api/admin/users/{id}/ban` protected by `RequireRole("ADMIN")` and rate limiting.
- Support optional ban metadata in the request body (e.g., `reason`, `bannedUntil`).
- Update `store_user` service layer to execute user banning and retain profile records for forensics and audit trails.
- Define `EventUserBanned = "user.banned"` and implement `PublishUserBanned` in Kafka producer with dual-compatible envelope fields (`event` and `event_type`, `userId` and `user_id`) to ensure seamless inter-service consumption by `store_auth` and `store_order`.
- Update OpenAPI documentation to reflect the new administrative ban endpoint.

## Capabilities

### New Capabilities

### Modified Capabilities
- `admin-user-management`: Add requirement for administrative account banning with reason capture, profile retention, and `user.banned` Kafka domain event publication to `user.events`.

## Impact

- **HTTP APIs**: Adds `POST /api/admin/users/{id}/ban` to `/api/admin/users` routes.
- **Kafka**: Emits `user.banned` domain events to topic `user.events`.
- **Downstream Services**:
  - `store_auth`: Consumes `user.banned` to invalidate JWTs in Redis and set `isActive = false`.
  - `store_order`: Consumes `user.banned` to auto-cancel unpaid orders while preserving PII for dispute defense.
- **Codebase**: Updates `internal/kafka`, `internal/service`, `internal/handler`, `internal/router`, and OpenAPI documentation in `store_user`.
