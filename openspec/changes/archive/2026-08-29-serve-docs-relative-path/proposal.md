## Why

Currently, the Swagger UI and OpenAPI specifications served by `store_user` hardcode root-absolute paths (`/docs/openapi.yaml`) and static host URLs (`http://localhost:8082`, `http://localhost:80`). When deployed behind an API Gateway (`store_gateway`), reverse proxy, or subpath prefix (e.g. `/api/users/docs`), the browser requests `/docs/openapi.yaml` at the root domain, bypassing the gateway prefix and resulting in 404 errors. Additionally, sandbox "Try It Out" requests fail when executed outside local development.

Serving documentation and specification files using relative paths like `./` enables seamless mounting behind any gateway, reverse proxy, or subpath without environment-specific rewrites.

## What Changes

- **Update Swagger UI spec URL**: Configure Swagger UI in `ServeSwaggerUI` to dynamically load OpenAPI specifications using relative paths (`./openapi.yaml` or `./docs/openapi.yaml`), ensuring compatibility with both direct root access and prefix-mounted gateways.
- **Normalize documentation routing**: Update router to handle `/docs` and `/docs/` consistently, ensuring relative path resolution (`./openapi.yaml`) resolves cleanly to `/docs/openapi.yaml`.
- **Set relative OpenAPI server URLs**: Update `servers:` in `docs/openapi.yaml` and `docs/openapi.json` to use relative path `./` so Swagger UI sandbox execution works across any host or ingress port.
- **Update documentation tests**: Align unit tests in `internal/handler/doc_test.go` and router tests to verify relative path resolution.

## Capabilities

### New Capabilities

### Modified Capabilities
- `api-documentation-ui`: Update Swagger UI delivery requirement to mandate relative path resolution (`./openapi.yaml`) and relative server base resolution for seamless reverse proxy and gateway ingress.

## Impact

- `internal/handler/doc.go`: Swagger UI HTML initialization template and spec loading path.
- `internal/router/router.go`: Documentation route registration and trailing-slash normalization.
- `docs/openapi.yaml` & `docs/openapi.json`: OpenAPI `servers` configuration updated to relative paths.
- `internal/handler/doc_test.go`: Test assertions for relative spec loading in Swagger UI.