## Context

`store_user` currently functions as a REST HTTP microservice (listening on `:8082`) and a gRPC client (talking to `store_order`). When users authenticate in `store_auth`, an OTP verification event triggers synchronous profile provisioning via gRPC to `store_user`.

The protobuf contract in `github.com/Sall-lah/store_proto` has been streamlined under `store.user.v1.UserService` to support `CreateUserProfile`, while removing unnecessary attributes (`avatar_url`, `bio`, `gender`, `date_of_birth`).

## Goals / Non-Goals

**Goals:**
- Implement `userv1.UserServiceServer` (`CreateUserProfile`) in `internal/grpc`.
- Support idempotent replay: if `user_id` already exists, return existing profile unmodified with `is_created = false` and status `OK`.
- Run HTTP and gRPC servers concurrently with coordinated graceful shutdown on OS signals.
- Streamline database schema (`user_profiles`), repository, service, and HTTP handlers by removing deprecated fields (`avatar_url`, `bio`, `gender`, `date_of_birth`).
- Maintain thorough test coverage with unit tests for gRPC handlers, service, repository, and HTTP handlers.

**Non-Goals:**
- Handling user authentication/passwords (handled exclusively by `store_auth`).
- Emitting additional Kafka events during gRPC provisioning (Kafka events remain focused on `user.deleted`).

## Decisions

### 1. Transport Layer Isolation (`internal/grpc`)
- **Decision**: Place gRPC server initialization and RPC handler implementations in `internal/grpc` rather than combining with HTTP handlers in `internal/handler`.
- **Rationale**: Clean protocol separation. HTTP and gRPC have distinct context handling, error mapping (HTTP status codes vs gRPC codes), and interceptors.
- **Alternatives Considered**: Combining HTTP and gRPC into `internal/handler`. Rejected because it creates messy file organization and conflates transport contracts.

### 2. Streamlined Profile Schema
- **Decision**: Remove `avatar_url`, `bio`, `gender`, and `date_of_birth` columns from `user_profiles` in `prisma/schema.prisma` and domain structs.
- **Rationale**: Keeps the customer profile minimal (`user_id`, `full_name`, `phone_number`, `address`, timestamps), reducing unnecessary payload size and database storage.
- **Alternatives Considered**: Keeping nullable/unused columns in database. Rejected to avoid dead columns and schema drift.

### 3. Idempotent Provisioning Semantics ("Just Return")
- **Decision**: When `CreateUserProfile` receives a request with an existing `user_id`, retrieve the existing profile from the repository and return it immediately without updating existing fields.
- **Rationale**: Prevents accidental overwrites from replay requests or out-of-order retries while confirming profile existence to `store_auth`.
- **Alternatives Considered**: Upserting/updating fields on conflict. Rejected per business requirement to preserve existing profile state.

### 4. Unified Server Lifecycle Management
- **Decision**: Run both HTTP (`:8082`) and gRPC (`:50052`) listeners in separate goroutines in `cmd/server/main.go`. Intercept `SIGINT` / `SIGTERM` signals and invoke `srv.Shutdown(ctx)` for HTTP and `grpcServer.GracefulStop()` for gRPC.
- **Rationale**: Ensures zero dropped connections during container redeployments and unified lifecycle logging.

## Risks / Trade-offs

- **[Risk] Database Schema Migration Drift** → Prisma schema changes require running `prisma db push` or migration before running the app.
  - *Mitigation*: Update `schema.prisma`, run `prisma generate` to update Go client stubs, and document migration requirements.
- **[Risk] Port Conflict on Startup** → Default gRPC port (`50052`) might conflict with another local service.
  - *Mitigation*: Make `GRPC_PORT` configurable in `.env` and `internal/config/config.go` with safe fallback.

## Migration Plan

1. Update `prisma/schema.prisma` to remove dropped fields.
2. Run Prisma generation (`go run github.com/steebchen/prisma-client-go generate`).
3. Update repository, service, and HTTP handler models.
4. Implement `internal/grpc/server.go` and `internal/grpc/user_service.go`.
5. Update `cmd/server/main.go` and `internal/config/config.go`.
6. Run unit tests (`go test ./internal/...`).
