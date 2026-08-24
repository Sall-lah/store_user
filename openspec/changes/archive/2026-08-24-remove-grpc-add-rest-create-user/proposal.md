## Why

The gRPC contract between `store_auth` and `store_user` (`store.user.v1.UserService`) has been removed from `store_proto` to simplify inter-service architecture. Instead of coupling authentication to internal gRPC provisioning calls, the system now adopts a frontend-driven REST architecture where client applications provision user profiles directly via `POST /api/users` over the API Gateway with gateway-injected identity headers (`X-User-Id`).

## What Changes

- **Add REST User Creation Endpoint**: Implement `POST /api/users` allowing authenticated users (with gateway-injected `X-User-Id`) to create their user profile.
- **Idempotent Creation Behavior**: If a user profile already exists for the supplied `X-User-Id`, `POST /api/users` returns `200 OK` with the existing profile without failing. If newly created, returns `201 Created`.
- **Remove Incoming gRPC Server**: (**BREAKING**) Remove `internal/grpc` (`server.go`, `user_service.go`, `user_service_test.go`), `GRPC_PORT` configurations, and background gRPC server listeners from `cmd/server/main.go`. `store_user` no longer acts as a gRPC server.
- **Retain gRPC Client to store_order**: Keep `internal/client/order` for checking active orders during account deletion.
- **Update Documentation and Specs**: Update OpenAPI schemas (`docs/openapi.yaml`, `docs/openapi.json`) and test suites to reflect the new REST route and removal of gRPC server stubs.

## Capabilities

### New Capabilities
<!-- None: modifications to existing capabilities -->

### Modified Capabilities
- `user-profile-management`: Adds `POST /api/users` endpoint for user profile creation, with payload validation, input sanitization, and idempotent response handling (`201 Created` vs `200 OK`).
- `user-profile-provisioning-grpc`: Removes the incoming `store.user.v1.UserService` gRPC server and dual server startup lifecycle in favor of the Chi REST routing architecture.

## Impact

- **Affected Code**: `cmd/server/main.go`, `internal/router/router.go`, `internal/handler/handler.go`, `internal/handler/handler_test.go`, `internal/service/service.go`, `internal/service/service_test.go`, `internal/config/config.go`, `docs/openapi.yaml`, `docs/openapi.json`, `test/integration_test.go`.
- **Deleted Code**: `internal/grpc/server.go`, `internal/grpc/user_service.go`, `internal/grpc/user_service_test.go`.
- **Dependencies**: Cleans up unused protobuf imports from `github.com/Sall-lah/store_proto/gen/go/store/user/v1`.
- **External Integration**: Frontend / client apps call `POST /api/users` via API Gateway following OTP verification/login.
