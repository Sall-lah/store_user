# Design: README.md Standardization for Store User Microservice

## Context

The platform microservice ecosystem (`store_order`, `store_user`, `store_product`, etc.) requires a unified, high-quality documentation standard. The current `store_user` README contains basic information but lacks interactive badges, a structured table of contents, modern Mermaid architecture diagrams, a comprehensive environment variable reference table, database setup instructions, and deployment guides.

## Goals / Non-Goals

**Goals:**
- Restructure `README.md` to conform exactly to the standardized microservice layout demonstrated in `store_order`.
- Accurately document `store_user` architectural boundaries:
  - API Gateway auth header offloading (`X-User-Id`, `X-User-Role`, `X-User-Email`).
  - Synchronous gRPC active order verification client targeting `store_order`.
  - Asynchronous Kafka event publishing (`user.deleted`, `user.banned`) on `user.events`.
  - In-app notification management vs. `store_notification` transactional delivery.
  - Redis sliding-window rate limiter with fail-open circuit breaker.
  - Prisma Client Go integration and PostgreSQL schema synchronization.
- Provide clear local development, testing, and container deployment instructions.

**Non-Goals:**
- Modify backend Go application code, database schema, or OpenAPI definitions.

## Decisions

### 1. Unified Microservice Template Structure
Follow the template sequence:
1. Header Badges & Synopsis
2. Table of Contents
3. Architecture Overview (Mermaid Flowchart) & Service Boundaries
4. Key Features
5. Technology Stack
6. Repository Structure
7. Prerequisites & Environment Configuration
8. Database Setup & Prisma ORM
9. Getting Started (Local Development)
10. API Endpoints & Documentation
11. Kafka Domain Events & Inter-Service Communication
12. Redis Rate Limiting Rules
13. Testing
14. Docker Deployment

### 2. Architecture Visualization with Mermaid
Represent the runtime interaction flows:
- Ingress: Client -> Gateway -> Chi HTTP Router with AuthIdentity and RateLimit middleware.
- Internal: HTTP Handlers -> User & Notification Services.
- Outbound / Storage:
  - Persistence: PostgreSQL via Prisma Client Go (`user_profiles`, `user_notifications`).
  - gRPC: Dial `store_order` for pre-flight active order checks during account deletion.
  - Kafka: Publish `user.deleted` and `user.banned` domain events to `user.events`.
  - Downstream Consumers: `store_auth` (token & session revocation) and `store_order` (order cancellation & stock restock).

### 3. Factual Alignment with Codebase
Ensure all ports (`8082`), config keys (`ORDER_SERVICE_GRPC_ADDR`, `KAFKA_TOPIC_USER_EVENTS`, `RATE_LIMIT_MAX_REQUESTS`), endpoints (`/api/users/*`, `/api/admin/users/*`, `/health`, `/docs`), and rate limiting parameters strictly match `internal/config/config.go` and `internal/router/router.go`.

## Risks / Trade-offs

- **[Risk] Configuration & Route Drift**: Documentation might become outdated if new endpoints or configs are added later.
  - *Mitigation*: Ensure environment table matches `internal/config/config.go` and endpoint catalog matches `internal/router/router.go` and `docs/openapi.yaml`.
