## ADDED Requirements

### Requirement: User Profile Creation via REST
The system SHALL provide a `POST /api/users` REST endpoint allowing authenticated clients to create a user profile using gateway-propagated identity headers (`X-User-Id`).

#### Scenario: Successfully create new user profile
- **WHEN** an authenticated request with a valid `X-User-Id` and non-empty `fullName` in the JSON body is sent to `POST /api/users` for a user without an existing profile
- **THEN** the system persists the profile record in PostgreSQL, and returns HTTP `201 Created` with the profile JSON object.

#### Scenario: Idempotent replay for existing profile
- **WHEN** an authenticated request is sent to `POST /api/users` for a user who already possesses a profile
- **THEN** the system does not fail or overwrite with blank attributes, and returns HTTP `200 OK` with the existing profile JSON object.

#### Scenario: Validation failure on missing required fields
- **WHEN** a request to `POST /api/users` contains an empty or whitespace-only `fullName` or an invalid phone number format
- **THEN** the system rejects the request with HTTP `400 Bad Request` and descriptive error message.

#### Scenario: Unauthorized request without valid identity header
- **WHEN** an unauthenticated request without a valid UUID `X-User-Id` header is sent to `POST /api/users`
- **THEN** the system rejects the request with HTTP `401 Unauthorized`.
