## 1. Schema & Model Streamlining

- [x] 1.1 Update `prisma/schema.prisma` to remove `avatar_url`, `bio`, `gender`, and `date_of_birth` from `user_profiles` model
- [x] 1.2 Run Prisma Go client generation to synchronize `internal/db` stubs
- [x] 1.3 Update `internal/repository/model.go` and `internal/service/model.go` to remove dropped attributes
- [x] 1.4 Update `internal/repository/prisma.go` and repository mocks to remove dropped fields from queries and mapping
- [x] 1.5 Update `internal/handler/handler.go`, validator rules, and tests to reflect streamlined profile fields

## 2. Service Layer Implementation

- [x] 2.1 Add `CreateUserProfile` method to `service.UserService` interface and implementation
- [x] 2.2 Implement idempotent logic in `UserServiceImpl.CreateUserProfile`: return existing profile if found (`is_created = false`), else create new profile (`is_created = true`)
- [x] 2.3 Add unit tests in `internal/service/service_test.go` covering creation, idempotent replay, and validation failures

## 3. gRPC Server & Handler

- [x] 3.1 Add `GRPC_PORT` (default `"50052"`) to `internal/config/config.go` and update configuration tests
- [x] 3.2 Implement `UserServiceServer` in `internal/grpc/user_service.go` mapping protobuf requests to `service.UserService`
- [x] 3.3 Implement `Server` in `internal/grpc/server.go` managing TCP listener, gRPC server registration, and `GracefulStop`
- [x] 3.4 Add unit tests in `internal/grpc/user_service_test.go` verifying gRPC status codes, error handling, and protobuf response structures

## 4. Main Server Lifecycle & Docs

- [x] 4.1 Update `cmd/server/main.go` to initialize and run HTTP and gRPC servers in parallel goroutines with graceful shutdown
- [x] 4.2 Update `docs/openapi.yaml` and `docs/openapi.json` to reflect streamlined profile schema
- [x] 4.3 Execute full test suite `go test ./...` and verify clean build

