## Why

The platform architecture standardizes on clean, unversioned API paths routed via API gateways and microservices. Removing redundant `/api/v1` route prefixes simplifies endpoint routing, eliminates duplicate route definitions, and standardizes ingress under `/api/users/profile` and `/api/users/account`.

## What Changes

- **BREAKING**: Remove `/api/v1/users` route prefixes and standardize all HTTP endpoints strictly under `/api/users`.
- Update the HTTP Chi router configuration in `internal/router/router.go` to mount routes solely under `/api/users`.
- Update route references across automated unit/integration tests in `internal/router/router_test.go` and `internal/handler/handler_test.go`.
- Update OpenAPI 3.1 specifications in `docs/openapi.yaml` and `docs/openapi.json` to expose unversioned paths `/api/users/profile` and `/api/users/account`.

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- `user-profile-management`: Update endpoint paths from `/api/v1/users/profile` to `/api/users/profile` for retrieval and updates.
- `account-deletion`: Update endpoint paths from `/api/v1/users/account` to `/api/users/account` for account deletion workflows.
- `security-rate-limiting`: Update endpoint path references for rate-limited routes to `/api/users/profile` and `/api/users/account`.

## Impact

- **Affected Code**: `internal/router/router.go`, `internal/router/router_test.go`, `internal/handler/handler_test.go`.
- **API Surface**: API clients and gateway ingress must target `/api/users/*` instead of `/api/v1/users/*`.
- **Documentation**: `docs/openapi.yaml` and `docs/openapi.json` updated with unversioned path definitions.
