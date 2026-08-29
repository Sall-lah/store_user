## 1. Documentation Handler Updates

- [x] 1.1 Update Swagger UI template in `internal/handler/doc.go` to initialize with relative spec URL resolution (`./openapi.yaml` / `./docs/openapi.yaml`)
- [x] 1.2 Update unit tests in `internal/handler/doc_test.go` to assert relative spec loading and verify 200 OK responses

## 2. OpenAPI Specification Updates

- [x] 2.1 Update `servers:` section in `docs/openapi.yaml` to use relative path (`url: ./`)
- [x] 2.2 Update `servers:` section in `docs/openapi.json` to use relative path (`url: ./`)

## 3. Router Normalization

- [x] 3.1 Update documentation routes in `internal/router/router.go` to handle `/docs` and `/docs/` trailing slash consistency
- [x] 3.2 Update `internal/router/router_test.go` to test documentation endpoints with and without trailing slashes

## 4. Verification & Regression Testing

- [x] 4.1 Run `go test ./...` across all packages to ensure zero test regressions
- [x] 4.2 Verify Swagger UI HTML renders with valid relative paths
