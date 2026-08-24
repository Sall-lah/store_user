# User Profile Management

## Purpose
Manages retrieval, initialization, sanitization, and updating of user profile data.

## Requirements

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

### Requirement: User Profile Retrieval
The system SHALL allow authenticated users to retrieve their profile information using their verified identity (userId). If a profile record does not exist for a newly registered user, the system SHALL automatically initialize and return a baseline profile or a 404 Not Found response.

#### Scenario: Successfully retrieve existing user profile
- **WHEN** an authenticated user with a valid X-User-Id requests GET /api/v1/users/profile
- **THEN** the system returns HTTP 200 OK with the streamlined user profile JSON containing id, userId, fullName, phoneNumber, address, createdAt, and updatedAt.

#### Scenario: Retrieve profile without authentication
- **WHEN** an unauthenticated request without valid identity headers is made to GET /api/v1/users/profile
- **THEN** the system rejects the request with HTTP 401 Unauthorized.

### Requirement: User Profile Update
The system SHALL allow authenticated users to update their profile details (such as full name, phone number, and address). All input fields MUST be validated and sanitized.

#### Scenario: Successfully update user profile
- **WHEN** an authenticated user submits a valid JSON payload to PUT /api/v1/users/profile with updated fields (fullName, phoneNumber, address)
- **THEN** the system validates the inputs, updates the record in PostgreSQL via Prisma, and returns HTTP 200 OK with the updated profile data.

#### Scenario: Update profile with invalid payload or data format
- **WHEN** an authenticated user submits an invalid phone number format or empty fullName in PUT /api/v1/users/profile
- **THEN** the system rejects the update with HTTP 400 Bad Request and descriptive validation error message.

#### Scenario: Input sanitization on text fields
- **WHEN** a user includes HTML or script tags within the address or fullName fields in PUT /api/v1/users/profile
- **THEN** the system sanitizes and strips unsafe tags before storing the text in the database.
