# API Documentation & Swagger UI

## Purpose
Provides interactive Swagger UI and raw OpenAPI 3.1 schema delivery directly from the microservice.

## Requirements

### Requirement: OpenAPI Specification Serving
The system SHALL expose the complete OpenAPI 3.1 schema definition at `GET /docs/openapi.yaml` and `GET /docs/openapi.json`.

#### Scenario: Retrieve OpenAPI schema in JSON format
- **WHEN** a client makes an HTTP GET request to `/docs/openapi.json`
- **THEN** the system returns HTTP 200 OK with `Content-Type: application/json` containing the OpenAPI 3.1 specification.

#### Scenario: Retrieve OpenAPI schema in YAML format
- **WHEN** a client makes an HTTP GET request to `/docs/openapi.yaml`
- **THEN** the system returns HTTP 200 OK with `Content-Type: application/x-yaml` or `text/yaml` containing the OpenAPI 3.1 specification.

### Requirement: Interactive Swagger UI Delivery
The system SHALL serve an interactive Swagger UI web interface at `GET /docs` and `GET /swagger` for interactive endpoint inspection and sandbox execution.

#### Scenario: Access Swagger UI documentation interface
- **WHEN** a client accesses `GET /docs` or `GET /swagger`
- **THEN** the system returns HTTP 200 OK with `Content-Type: text/html` rendering Swagger UI pre-configured to fetch `/docs/openapi.yaml` or `/docs/openapi.json`.

#### Scenario: Complete API coverage in documentation
- **WHEN** the Swagger UI or OpenAPI document is inspected
- **THEN** all endpoints (`/health`, `/api/users/profile`, `/api/users/account`, `/api/users/notifications`, and documentation routes) are fully documented with request/response schemas, security headers (`X-User-Id`), and status codes.
