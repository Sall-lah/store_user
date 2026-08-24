## ADDED Requirements

### Requirement: Administrative Account Banning
The system SHALL provide an administrative endpoint `POST /api/admin/users/{id}/ban` requiring `ADMIN` role. The system MUST validate that `{id}` is a valid UUID, accept an optional JSON request body containing `reason`, preserve the user profile and notification records in the database without hard deletion, and publish a `user.banned` domain event to Kafka topic `user.events`.

#### Scenario: Authorized admin bans user
- **WHEN** an authenticated caller with `X-User-Role: ADMIN` sends `POST /api/admin/users/{id}/ban` with a valid UUID and body `{"reason": "payment_fraud"}`
- **THEN** the system accepts the ban, preserves the profile data, dispatches a `user.banned` event to Kafka topic `user.events`, and returns HTTP 200 OK with `{"message": "User account successfully banned by admin"}`.

#### Scenario: Non-admin caller forbidden from banning user
- **WHEN** an authenticated caller with `X-User-Role: CUSTOMER` attempts to send `POST /api/admin/users/{id}/ban`
- **THEN** the system rejects the request with HTTP 403 Forbidden and error message `{"error": "forbidden: administrative role required"}`.

#### Scenario: Missing identity headers
- **WHEN** an unauthenticated caller sends `POST /api/admin/users/{id}/ban` without valid identity headers
- **THEN** the system rejects the request with HTTP 401 Unauthorized.

#### Scenario: Invalid target user UUID path parameter
- **WHEN** an admin sends `POST /api/admin/users/{id}/ban` where `{id}` is not a valid UUID format
- **THEN** the system rejects the request with HTTP 400 Bad Request and error message `{"error": "invalid target user UUID format"}`.

### Requirement: Kafka Domain Event Publication for Admin Bans
Upon successfully processing an administrative ban, the system SHALL publish a `user.banned` lifecycle domain event to Kafka topic `user.events` with a payload containing `event: "user.banned"`, `event_type: "user.banned"`, `userId`, `user_id`, `timestamp`, and `reason` to enable downstream session revocation in `store_auth` and order cancellation in `store_order`.

#### Scenario: Dispatched Kafka user banned event
- **WHEN** an admin successfully bans user account `{id}` with reason `"chargeback_abuse"`
- **THEN** the system emits `{ "event": "user.banned", "event_type": "user.banned", "userId": "<id>", "user_id": "<id>", "timestamp": "<ISO-8601-timestamp>", "reason": "chargeback_abuse" }` to Kafka topic `user.events`.
