## ADDED Requirements

### Requirement: gRPC UserService CreateUserProfile Implementation
The system SHALL implement the `store.user.v1.UserService` gRPC service interface exposing `CreateUserProfile` RPC method in `internal/grpc`.

#### Scenario: Successfully create new user profile via gRPC
- **WHEN** `CreateUserProfile` is invoked with valid `user_id`, `full_name`, `email`, and optional `phone_number` for a non-existent user profile
- **THEN** the system persists the new user profile in the database, returns `is_created = true`, `message = "User profile created successfully"`, and the populated `UserProfile` protobuf entity.

#### Scenario: Idempotent replay for existing user profile
- **WHEN** `CreateUserProfile` is invoked for a `user_id` that already has an existing profile record
- **THEN** the system does not mutate the existing profile, returns `is_created = false`, `message = "User profile already exists"`, and the existing `UserProfile` protobuf entity with gRPC status `OK`.

#### Scenario: Validation failure on required fields
- **WHEN** `CreateUserProfile` is invoked with empty `user_id` or empty `full_name`
- **THEN** the system returns an `InvalidArgument` gRPC status error.

### Requirement: Dedicated gRPC Server Lifecycle
The system SHALL run a gRPC server on a dedicated network port (configurable via `GRPC_PORT`, default `:50052`) concurrently with the HTTP server.

#### Scenario: Concurrent server startup
- **WHEN** the application starts up
- **THEN** both the HTTP server (on `SERVER_PORT`) and gRPC server (on `GRPC_PORT`) initialize listeners and accept connections concurrently.

#### Scenario: Coordinated graceful shutdown
- **WHEN** an OS interrupt (`SIGINT`, `SIGTERM`) is received
- **THEN** both the HTTP server and gRPC server initiate graceful stop procedures, waiting for active requests to complete within the configured shutdown timeout window.
