## Context

The protobuf contract `store.user.v1.UserService` has been removed from `store_proto`, eliminating the synchronous gRPC dependency previously required by `store_auth` to create user profiles. To replace this, `store_user` exposes a public REST endpoint (`POST /api/users`) through the API Gateway, allowing client applications to provision their user profiles directly upon authentication.

## Goals / Non-Goals

**Goals:**
- Provide `POST /api/users` handler in Chi router, secured by `middleware.AuthIdentity` and rate-limited at 15 req/min.
- Support idempotent creation semantics returning `201 Created` for newly provisioned profiles and `200 OK` for existing profiles.
- Purge all incoming gRPC server code (`internal/grpc`), configs (`GRPC_PORT`), and multi-server lifecycle concurrency from `cmd/server/main.go`.
- Retain outbound gRPC client `internal/client/order` for active order verification during account deletion.
- Update OpenAPI definitions (`docs/openapi.yaml`, `docs/openapi.json`) and test suites.

**Non-Goals:**
- Altering the Prisma database schema or notification pipelines.
- Modifying outbound Kafka event structures (`user.deleted`).

## Decisions

### 1. Route Placement: `POST /api/users`
- **Choice**: Mount profile creation directly on the base route `POST /` inside the `/api/users` router group.
- **Alternatives Considered**:
  - `POST /api/users/profile`: Less standard REST convention for resource root instantiation.
  - `POST /api/users/register`: Overlaps with auth terminology.

### 2. Header-Driven Identity Resolution
- **Choice**: Extract `userID` from gateway-propagated `X-User-Id` header (validated UUID).
- **Rationale**: Prevents ID spoofing by ignoring user IDs passed in the body payload, maintaining consistent security with existing `GET /api/users/profile` and `PUT /api/users/profile` routes.

### 3. Idempotent Status Code Mapping
- **Choice**: Return `201 Created` with the profile body if created afresh; return `200 OK` with the existing profile if already present.
- **Rationale**: Supports seamless client retries and idempotent onboarding flows without triggering client-side error states.

### 4. Complete Removal of Incoming gRPC Server
- **Choice**: Delete `internal/grpc` package and remove gRPC listener routines from `cmd/server/main.go`.
- **Rationale**: `store_user` does not serve any other gRPC RPCs. Running a gRPC server without services wastes resources and increases maintenance overhead.

## Risks / Trade-offs

- **[Risk] Multiple rapid concurrent creation requests for the same user** → Mitigated by database unique constraints on `userId` and idempotent lookup/upsert logic in `service.CreateUserProfile`.
- **[Risk] Breaking external consumers expecting gRPC on port 50051** → Mitigated by coordinated deprecation; `store_proto` has already deleted the `store.user.v1` definitions.
