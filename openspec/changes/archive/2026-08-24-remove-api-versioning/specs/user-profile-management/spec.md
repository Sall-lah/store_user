## MODIFIED Requirements

### Requirement: User Profile Retrieval
The system SHALL allow authenticated users to retrieve their profile information using their verified identity (userId). If a profile record does not exist for a newly registered user, the system SHALL automatically initialize and return a baseline profile or a 404 Not Found response.

#### Scenario: Successfully retrieve existing user profile
- **WHEN** an authenticated user with a valid X-User-Id requests GET /api/users/profile
- **THEN** the system returns HTTP 200 OK with the user profile JSON containing userId, fullName, phoneNumber, avatarUrl, bio, address, gender, dateOfBirth, and timestamp fields.

#### Scenario: Retrieve profile without authentication
- **WHEN** an unauthenticated request without valid identity headers is made to GET /api/users/profile
- **THEN** the system rejects the request with HTTP 401 Unauthorized.

### Requirement: User Profile Update
The system SHALL allow authenticated users to update their profile details (such as full name, phone number, bio, avatar URL, address, gender, and date of birth). All input fields MUST be validated and sanitized.

#### Scenario: Successfully update user profile
- **WHEN** an authenticated user submits a valid JSON payload to PUT /api/users/profile with updated fields
- **THEN** the system validates the inputs, updates the record in Supabase PostgreSQL, and returns HTTP 200 OK with the updated profile data.

#### Scenario: Update profile with invalid payload or data format
- **WHEN** an authenticated user submits an invalid phone number format or payload exceeding size limits
- **THEN** the system rejects the update with HTTP 400 Bad Request and detailed error messages.

#### Scenario: Input sanitization on text fields
- **WHEN** a user includes HTML or script tags within the bio or address fields in PUT /api/users/profile
- **THEN** the system sanitizes and strips the unsafe tags before storing the text in the database.
