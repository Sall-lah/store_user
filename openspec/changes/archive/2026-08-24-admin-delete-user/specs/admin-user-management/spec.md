# Admin User Management

## ADDED Requirements

### Requirement: Role-Based Authorization for Administrative Endpoints
The system SHALL require callers to present an authenticated `ADMIN` role claim (propagated via gateway identity header `X-User-Role`) to access admin user endpoints. If the caller does not possess the `ADMIN` role, the system MUST reject the request with HTTP 403 Forbidden. If the identity headers are missing or invalid, the system MUST return HTTP 401 Unauthorized.

#### Scenario: Authorized admin deletion request
- **WHEN** an authenticated caller with `X-User-Role: ADMIN` (or `admin`) sends `DELETE /api/admin/users/{id}` with a valid UUID
- **THEN** the system authorizes the request and proceeds to account deletion processing.

#### Scenario: Non-admin caller forbidden
- **WHEN** an authenticated caller with `X-User-Role: CUSTOMER` attempts to send `DELETE /api/admin/users/{id}`
- **THEN** the system aborts execution and returns HTTP 403 Forbidden with `{ "error": "forbidden: administrative role required" }`.

#### Scenario: Missing identity headers
- **WHEN** an unauthenticated caller sends `DELETE /api/admin/users/{id}` without valid identity headers
- **THEN** the system rejects the request with HTTP 401 Unauthorized.

### Requirement: Forced Account Deletion Without Order Pre-Flight Checks
The system SHALL execute user account deletion for the target user ID specified in the path parameter `DELETE /api/admin/users/{id}` without performing active order pre-flight verification against `store_order`. The system MUST purge the target user's profile from `user_profiles` and user notifications from `user_notifications`.

#### Scenario: Successful forced user deletion with active orders present
- **WHEN** an admin issues `DELETE /api/admin/users/{id}` for a user who has active in-flight orders
- **THEN** the system bypasses order checks, deletes the user's profile and notifications from PostgreSQL, publishes a lifecycle event to Kafka, and returns HTTP 200 OK with `{ "message": "User account successfully deleted by admin" }`.

#### Scenario: Target user profile does not exist in store_user
- **WHEN** an admin issues `DELETE /api/admin/users/{id}` for a user whose profile does not exist in `store_user` database
- **THEN** the system treats the deletion idempotently, publishes the `user.deleted` lifecycle event to Kafka to ensure downstream credential cleanup in `store_auth`, and returns HTTP 200 OK.

#### Scenario: Invalid user UUID path parameter
- **WHEN** an admin sends `DELETE /api/admin/users/{id}` where `{id}` is not a valid UUID format
- **THEN** the system rejects the request immediately and returns HTTP 400 Bad Request with `{ "error": "invalid target user UUID format" }`.

### Requirement: Kafka Domain Event Publication for Admin Deletions
Upon executing administrative user deletion, the system SHALL publish a `user.deleted` domain event to the Kafka topic `user.events` with `LifecycleEvent` payload containing `reason: "admin_deletion"` so `store_auth` immediately invalidates active JWT tokens, revokes refresh tokens, and purges user credentials.

#### Scenario: Dispatched Kafka user deleted event
- **WHEN** an admin successfully deletes user account `{id}`
- **THEN** the system emits `{ event: "user.deleted", userId: "<id>", timestamp: "<ISO-8601-timestamp>", reason: "admin_deletion" }` to Kafka topic `user.events`.
