## 1. HTTP Router & Middleware Updates

- [x] 1.1 Update `internal/router/router.go` to remove `/api/v1/users` mount and retain solely `r.Route("/api/users", mountUserRoutes)`.

## 2. Test Suite Route Adjustments

- [x] 2.1 Update test requests in `internal/router/router_test.go` to use `/api/users/profile` and `/api/users/account`.
- [x] 2.2 Verify tests in `internal/handler/handler_test.go` and run `go test ./...` to ensure all unit and integration tests pass.

## 3. OpenAPI Documentation Updates

- [x] 3.1 Update endpoint paths and descriptions in `docs/openapi.yaml` from `/api/v1/users/*` to `/api/users/*`.
- [x] 3.2 Update endpoint paths and descriptions in `docs/openapi.json` from `/api/v1/users/*` to `/api/users/*`.
