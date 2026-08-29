## Context

The `store_user` service exposes API documentation via Swagger UI and raw OpenAPI 3.1 specifications (`openapi.yaml`, `openapi.json`). Currently, Swagger UI initializes with an absolute root path `url: "/docs/openapi.yaml"`, and the OpenAPI specifications define static local server URLs (`http://localhost:8082`, `http://localhost:80`). When the service is hosted behind an API gateway (`store_gateway`) or subpath reverse proxy (e.g. `/api/users/docs`), the browser requests `/docs/openapi.yaml` from the root domain, causing 404 errors, and sandbox execution hits hardcoded localhost ports.

## Goals / Non-Goals

**Goals:**
- Enable Swagger UI and OpenAPI specifications to be served using relative paths (`./`) so they work seamlessly whether accessed directly at `http://localhost:8082/docs` or mounted behind a gateway/reverse proxy prefix (e.g., `/api/users/docs`).
- Ensure OpenAPI schema `servers:` block specifies relative URL (`./`) to support portable client execution.
- Maintain backwards compatibility for `/docs`, `/swagger`, `/docs/openapi.yaml`, and `/docs/openapi.json`.

**Non-Goals:**
- Bundling local Swagger UI offline assets (CDN links for Swagger UI CSS/JS remain unchanged).
- Re-architecting microservice auth or business endpoints.

## Decisions

### Decision 1: Relative Spec URL with Trailing-Slash Resiliency
- **Choice**: In `internal/handler/doc.go`, initialize Swagger UI with a resilient relative spec URL that works with or without trailing slash:
  ```javascript
  const basePath = window.location.pathname.replace(/\/+$/, "");
  const specUrl = basePath.endsWith("/docs") || basePath.endsWith("/swagger")
    ? "./openapi.yaml"
    : "./docs/openapi.yaml";
  ```
  Or normalize routing in Chi so `/docs` redirects to `/docs/` (or both are served), allowing `url: "./openapi.yaml"`:
  ```javascript
  url: window.location.pathname.endsWith("/") ? "./openapi.yaml" : "./docs/openapi.yaml"
  ```
- **Rationale**: RFC 3986 specifies that `./openapi.yaml` relative to `/docs` resolves to `/openapi.yaml` (root), whereas relative to `/docs/` it resolves to `/docs/openapi.yaml`. Making the client JavaScript compute the relative path or ensuring trailing slash prevents 404s across different browser address bar states.

### Decision 2: Relative Server Base in OpenAPI Schema
- **Choice**: Update `servers:` in `docs/openapi.yaml` and `docs/openapi.json`:
  ```yaml
  servers:
    - url: ./
      description: Current host relative base
  ```
- **Rationale**: In OpenAPI 3.1, relative server URLs allow the Swagger UI interactive console to dispatch API calls directly to the current host and ingress path where the documentation is viewed, eliminating hardcoded local ports.

### Decision 3: Router Registration Normalization
- **Choice**: In `internal/router/router.go`, register both `/docs` and `/docs/` (as well as `/swagger` and `/swagger/`) or route through `r.Route("/docs", ...)`.
- **Rationale**: Prevents 404 errors when developers or proxies append a trailing slash to the documentation route.

## Risks / Trade-offs

- **[RFC 3986 Base Path Mismatch]** → *Mitigation*: Client script checks `window.location.pathname` or normalizes `/docs/` so `./openapi.yaml` always targets the adjacent specification file.
- **[Reverse Proxy Ingress Rewriting]** → *Mitigation*: Because `./` is relative to the document URL in the client's address bar, upstream gateways do not need to rewrite spec paths.
