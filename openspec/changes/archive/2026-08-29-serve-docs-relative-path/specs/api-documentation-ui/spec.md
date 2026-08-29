## MODIFIED Requirements

### Requirement: Interactive Swagger UI Delivery
The system SHALL serve an interactive Swagger UI web interface at `GET /docs`, `GET /docs/`, and `GET /swagger` configured with relative path resolution for interactive endpoint inspection and sandbox execution across direct access and reverse proxy environments.

#### Scenario: Access Swagger UI documentation interface
- **WHEN** a client accesses `GET /docs`, `GET /docs/`, or `GET /swagger`
- **THEN** the system returns HTTP 200 OK with `Content-Type: text/html` rendering Swagger UI configured to fetch the OpenAPI specification using relative path `./openapi.yaml` (or normalized relative path).

#### Scenario: Complete API coverage in documentation
- **WHEN** the Swagger UI or OpenAPI document is inspected
- **THEN** all endpoints (`/health`, `/api/users/profile`, `/api/users/account`, `/api/users/notifications`, and documentation routes) are fully documented with request/response schemas, security headers (`X-User-Id`), and status codes.

### Requirement: OpenAPI Specification Serving
The system SHALL expose the complete OpenAPI 3.1 schema definition at `GET /docs/openapi.yaml` and `GET /docs/openapi.json` with relative server base URLs (`url: ./`) to support flexible gateway and subpath hosting.

#### Scenario: Retrieve OpenAPI schema in JSON format
- **WHEN** a client makes an HTTP GET request to `/docs/openapi.json`
- **THEN** the system returns HTTP 200 OK with `Content-Type: application/json` containing the OpenAPI 3.1 specification configured with relative server URLs (`./`).

#### Scenario: Retrieve OpenAPI schema in YAML format
- **WHEN** a client makes an HTTP GET request to `/docs/openapi.yaml`
- **THEN** the system returns HTTP 200 OK with `Content-Type: application/x-yaml` or `text/yaml` containing the OpenAPI 3.1 specification configured with relative server URLs (`./`).