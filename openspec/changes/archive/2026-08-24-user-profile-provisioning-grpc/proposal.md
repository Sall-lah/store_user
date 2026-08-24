## Why

When users register and verify their OTP in `store_auth`, their user profile must be synchronously provisioned in `store_user` via gRPC. Furthermore, the user profile domain model is being simplified to remove unnecessary attributes (`avatar_url`, `bio`, `gender`, `date_of_birth`) and retain only core identity and contact fields (`user_id`, `full_name`, `phone_number`, `address`).

## What Changes

- **Add gRPC User Profile Provisioning**: Implement `userv1.UserServiceServer` with `CreateUserProfile` RPC in a dedicated transport package (`internal/grpc`).
- **Idempotent Profile Creation**: Provisioning a profile for a new `user_id` inserts a record and returns `is_created = true`. Replay calls for an existing `user_id` return the existing profile without mutation and `is_created = false` with status `OK`.
- **Concurrent Server Lifecycle**: Run the gRPC server (`:50052` configurable via `GRPC_PORT`) alongside the existing HTTP server (`:8082`), supporting coordinated graceful shutdown on OS termination signals.
- **Streamline Profile Schema & Domain Models**: Remove `avatar_url`, `bio`, `gender`, and `date_of_birth` columns from `user_profiles` table in Prisma schema, repository layer, service models, HTTP handlers, validation, and documentation.

## Capabilities

### New Capabilities
- `user-profile-provisioning-grpc`: Exposes `CreateUserProfile` gRPC service implementing idempotent user profile initialization for inter-service communication.

### Modified Capabilities
- `user-profile-management`: Removes `avatar_url`, `bio`, `gender`, and `date_of_birth` fields from profile models, retrieval responses, and update endpoints.

## Impact

- **Database**: `prisma/schema.prisma` `user_profiles` model simplified; Prisma client regenerated.
- **Protobuf / gRPC**: Integrates `github.com/Sall-lah/store_proto/gen/go/store/user/v1` (`userv1`).
- **Internal Layers**:
  - `internal/config`: Adds `GRPC_PORT` config parameter.
  - `internal/grpc`: New package for gRPC server and `UserService` implementation.
  - `internal/service`: Adds `CreateUserProfile` method; updates `UpdateProfileRequest` and `UserProfile` domain models.
  - `internal/repository`: Updates `UserProfile` struct, `Upsert` params, and Prisma mapping.
  - `cmd/server/main.go`: Launches gRPC server in parallel with HTTP server.
  - `internal/handler`: Updates profile HTTP handlers and request validation.
  - `docs`: Updates OpenAPI specification to reflect streamlined profile schema.
