## 1. Deprecate and Remove Incoming gRPC Server

- [x] 1.1 Remove `internal/grpc` package (`server.go`, `user_service.go`, `user_service_test.go`)
- [x] 1.2 Clean up `GrpcPort` and `GRPC_PORT` environment variables in `internal/config/config.go` and `internal/config/config_test.go`
- [x] 1.3 Refactor `cmd/server/main.go` to run exclusively as an HTTP Chi server and remove gRPC background server routines

## 2. Implement REST Profile Creation Endpoint

- [x] 2.1 Refine `CreateProfileRequest` and `CreateProfileResult` in `internal/service/model.go` and `internal/service/service.go` for REST usage
- [x] 2.2 Implement `CreateProfile` handler in `internal/handler/handler.go` with input validation, sanitization, and `201 Created` / `200 OK` response mapping
- [x] 2.3 Mount `POST /` in `internal/router/router.go` under `/api/users` with `middleware.AuthIdentity` and sliding-window rate limiting

## 3. Update API Documentation

- [x] 3.1 Add `POST /api/users` endpoint documentation and schemas to `docs/openapi.yaml`
- [x] 3.2 Update `docs/openapi.json` with the new endpoint schema and remove legacy gRPC references

## 4. Verification and Testing

- [x] 4.1 Update unit tests in `internal/handler/handler_test.go` and `internal/service/service_test.go`
- [x] 4.2 Update end-to-end integration tests in `test/integration_test.go`
- [x] 4.3 Execute `go test ./...` to ensure all tests pass with zero regressions
