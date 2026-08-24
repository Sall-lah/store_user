## MODIFIED Requirements

### Requirement: Interactive Swagger UI Delivery
The system SHALL serve an interactive Swagger UI web interface at `GET /docs` and `GET /swagger` for interactive endpoint inspection and sandbox execution.

#### Scenario: Access Swagger UI documentation interface
- **WHEN** a client accesses `GET /docs` or `GET /swagger`
- **THEN** the system returns HTTP 200 OK with `Content-Type: text/html` rendering Swagger UI pre-configured to fetch `/docs/openapi.yaml` or `/docs/openapi.json`.

#### Scenario: Complete API coverage in documentation
- **WHEN** the Swagger UI or OpenAPI document is inspected
- **THEN** all endpoints (`/health`, `/api/users/profile`, `/api/users/account`, `/api/users/notifications`, and documentation routes) are fully documented with request/response schemas, security headers (`X-User-Id`), and status codes, excluding removed preferences endpoints.
