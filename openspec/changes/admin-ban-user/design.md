## Context

In the distributed store platform, `store_auth` and `store_order` already have active Kafka consumers subscribed to topic `user.events` with explicit handling for `user.banned` events:
- `store_auth` revokes all active refresh tokens, blacklists JWTs in Redis (`revoked:user:<userID>`), and sets `isActive = false` in PostgreSQL.
- `store_order` auto-cancels in-flight unpaid orders and releases inventory, while strictly preserving customer PII and completed order histories for chargeback and dispute defense.

`store_user` currently only possesses deletion logic (`user.deleted`). Administrators need a way to ban abusive or fraudulent users without deleting profile data.

## Goals / Non-Goals

**Goals:**
- Provide an admin endpoint `POST /api/admin/users/{id}/ban` requiring `ADMIN` role and rate limiting.
- Implement `AdminBanUser` in `store_user` service to orchestrate domain validation and event dispatch.
- Preserve user profile and notification records in Supabase PostgreSQL (avoiding hard deletion).
- Implement `PublishUserBanned` in Kafka producer emitting `user.banned` events with multi-consumer compatibility (`event`, `event_type`, `userId`, `user_id`, `reason`, `timestamp`).
- Update OpenAPI specification and maintain high test coverage with unit and mock tests.

**Non-Goals:**
- Direct credential management or password changes (handled by `store_auth`).
- Managing payment gateway refunds directly in `store_user` (handled by `store_order`).
- Unban / account reinstatement workflow (can be addressed in a separate proposal if required).

## Decisions

### Decision 1: Dedicated Admin Ban Endpoint `POST /api/admin/users/{id}/ban`
- **Choice**: Add `BanUser` to `AdminHandler` mounted under `/api/admin/users/{id}/ban`.
- **Rationale**: Keeps administrative actions isolated from customer routes. Reuses existing `middleware.RequireRole("ADMIN", "admin")` and UUID path parameter validation.
- **Alternatives Considered**:
  - `PUT /api/admin/users/{id}/status`: More generic, but requires implementing full status state machine schemas. A dedicated `/ban` endpoint is more explicit and aligns with existing `DELETE /api/admin/users/{id}` conventions.

### Decision 2: Profile Retention vs Hard Deletion
- **Choice**: Do not purge `user_profiles` or `user_notifications` during an admin ban.
- **Rationale**: Banned accounts require PII, phone numbers, and addresses to remain in the system for forensic lookup, chargeback investigation, and fraud pattern detection.
- **Alternatives Considered**:
  - Hard deleting profiles: Destroys legal evidence and defeats the purpose of banning vs deleting.

### Decision 3: Dual-Compatible JSON Payload for Kafka Event
- **Choice**: Produce `LifecycleEvent` with dual top-level keys (`event` and `event_type`, `userId` and `user_id`).
- **Rationale**: `store_auth`'s consumer looks for `event` and `userId`, while `store_order`'s envelope parser looks for `event_type` and `user_id`/`userId`. Emitting both in the JSON payload guarantees zero-friction inter-service compatibility without requiring synchronized deployments of other services.
- **Payload Structure**:
  ```json
  {
    "event": "user.banned",
    "event_type": "user.banned",
    "userId": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    "user_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    "timestamp": "2026-08-24T23:15:00Z",
    "reason": "payment_fraud"
  }
  ```

## Risks / Trade-offs

- **[Risk] Kafka producer failure during ban action** → **Mitigation**: Service returns `ErrKafkaPublishFailed` (HTTP 500) so the administrator can retry, while logging detailed diagnostics.
- **[Risk] Banning non-existent profile** → **Mitigation**: Treated idempotently; Kafka event is still dispatched so `store_auth` and `store_order` can freeze any associated credentials or orders even if the profile was not yet provisioned in `store_user`.
