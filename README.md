# Store User Microservice (`store_user`)

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Router](https://img.shields.io/badge/Router-Chi%20v5-blue)](https://github.com/go-chi/chi)
[![Database](https://img.shields.io/badge/Database-PostgreSQL-336791?logo=postgresql)](https://www.postgresql.org/)
[![ORM](https://img.shields.io/badge/ORM-Prisma%20Go%20Client-2D3748?logo=prisma)](https://github.com/steebchen/prisma-client-go)
[![Event Streaming](https://img.shields.io/badge/Streaming-Apache%20Kafka-231F20?logo=apachekafka)](https://kafka.apache.org/)
[![Cache & Rate Limit](https://img.shields.io/badge/Rate%20Limit-Redis-DC382D?logo=redis)](https://redis.io/)
[![RPC](https://img.shields.io/badge/RPC-gRPC-244c5a?logo=grpc)](https://grpc.io/)

A production-grade, event-driven User Profile Management, Account Lifecycle, and In-App Notifications microservice built in Go. It manages user profile synchronization, handles self-service and administrative account lifecycle events (with pre-flight active order checks via gRPC to `store_order`), dispatches domain events to Apache Kafka (`user.events`), delivers paginated in-app notification feeds, and defends against volumetric abuse using Redis sliding-window rate limiting.

---

## 📑 Table of Contents

- [Architecture Overview](#-architecture-overview)
  - [Notification Architecture & Service Separation](#-notification-architecture--service-separation)
- [Key Features](#-key-features)
- [Technology Stack](#-technology-stack)
- [Repository Structure](#-repository-structure)
- [Prerequisites & Environment Configuration](#-prerequisites--environment-configuration)
- [Database Setup & Prisma ORM](#-database-setup--prisma-orm)
- [Getting Started (Local Development)](#-getting-started-local-development)
- [API Endpoints & Documentation](#-api-endpoints--documentation)
- [Kafka Domain Events & Inter-Service Communication](#-kafka-domain-events--inter-service-communication)
- [Redis Rate Limiting Rules](#-redis-rate-limiting-rules)
- [Testing](#-testing)
- [Docker Deployment](#-docker-deployment)

---

## 🏗 Architecture Overview

```mermaid
flowchart TD
    Client[Client / Frontend] -->|HTTP Requests| Gateway[API Gateway / store_gateway]
    Gateway -->|Forward with X-User-Id / X-User-Role / X-User-Email| Router[Chi HTTP Router]
    
    subgraph store_user ["Store User Service"]
        Router --> Middleware[RateLimit & Auth Middleware]
        Middleware --> Handlers[HTTP Handlers: Profile, Admin, Notification, Doc]
        Handlers --> UserService[User Service]
        Handlers --> NotifService[Notification Service]
        UserService --> ProfileRepo[User Profile Repository]
        UserService --> NotifRepo[Notification Repository]
        NotifService --> NotifRepo
        UserService --> OrderClient[Order Service gRPC Client]
        UserService --> KafkaProducer[Kafka Event Producer]
    end

    ProfileRepo -->|Prisma Client| Postgres[(PostgreSQL DB)]
    NotifRepo -->|Prisma Client| Postgres
    Middleware -->|Sliding Window Counter| Redis[(Redis)]
    OrderClient -->|gRPC CheckActiveOrders| OrderSvc[Order Microservice / store_order]
    KafkaProducer -->|Publish: user.events| KafkaOut[Apache Kafka]
    KafkaOut -->|user.deleted / user.banned| DownstreamAuth[store_auth]
    KafkaOut -->|user.deleted / user.banned| DownstreamOrder[store_order]
```

### 🔔 Notification Architecture & Service Separation

It is important to distinguish between **in-app notifications** and **external transactional notifications**:

- **`store_user` (In-App Notification Feed)**:
  - Manages the user's persistent in-app notifications in PostgreSQL (`user_notifications` table).
  - Exposes REST endpoints (`/api/users/notifications`) for web/mobile inbox retrieval, pagination, read receipts, and deletion.
- **`store_notification` (Transactional Delivery)**:
  - Dedicated microservice whose sole responsibility is consuming Kafka events (`auth.events`, `order.events`) to send out-of-band **OTP verification codes via Email/SMS** during account registration and password resets, and transactional payment notifications.
  - Does **not** store or manage in-app user notifications.

---

## 🌟 Key Features

1. **API Gateway Offloading Authentication**: Consumes trusted gateway headers (`X-User-Id`, `X-User-Email`, `X-User-Role`) injected by upstream `store_gateway` after validating RS256 JWTs.
2. **Synchronous Pre-Flight Order Verification (gRPC)**: Enforces business constraints by verifying with `store_order` over gRPC that a customer has no active in-flight orders (`PENDING_PAYMENT`, `PROCESSING`, `SHIPPED`) before permitting self-service account deletion.
3. **Asynchronous Event Publishing (`user.events`)**: Produces dual-compatible JSON domain events (`user.deleted`, `user.banned`) to Apache Kafka to trigger token invalidation in `store_auth` and order cancellation/inventory restock in `store_order`.
4. **In-App Notification Feed**: Provides a full-featured notification inbox supporting pagination, unread filters, single/bulk read receipts, and deletion.
5. **User Governance & Forensic Retention**: Supports administrative user purge and ban operations. Administrative bans preserve profile and notification audit records while revoking access downstream.
6. **Resilient Rate Limiting**: Multi-tiered sliding-window rate limiting backed by Redis and atomic Lua scripts with a **50ms fail-open** circuit breaker.
7. **Embedded Interactive Documentation**: Self-hosted OpenAPI 3.1 schema serving and interactive **Swagger UI** (`/docs` and `/swagger`).

---

## 🛠 Technology Stack

- **Language**: Go 1.26+
- **HTTP Routing**: [Chi v5](https://github.com/go-chi/chi) with CORS, RequestID, RealIP, and Recovery middlewares
- **RPC & Inter-Service**: [gRPC](https://grpc.io/) (`google.golang.org/grpc`) for synchronous order verification
- **ORM & Data Layer**: [Prisma Client Go](https://github.com/steebchen/prisma-client-go) with PostgreSQL (Supabase)
- **Event Streaming**: [segmentio/kafka-go](https://github.com/segmentio/kafka-go)
- **Caching & Rate Limiting**: [go-redis/v9](https://github.com/redis/go-redis) (Redis 7.x+)
- **API Documentation**: OpenAPI 3.1, [Swagger UI](https://swagger.io/tools/swagger-ui/)
- **Containerization**: Multi-stage Alpine Dockerfile

---

## 📁 Repository Structure

```
store_user/
├── cmd/
│   └── server/
│       └── main.go                 # Application bootstrap & dependency injection
├── docs/
│   ├── openapi.json                # OpenAPI 3.1 specification (JSON format)
│   └── openapi.yaml                # OpenAPI 3.1 specification (YAML format)
├── internal/
│   ├── client/
│   │   └── order/                  # gRPC client for store_order active order check
│   ├── config/                     # Environment variable parsing and validation
│   ├── db/                         # Generated Prisma Client Go engine & models
│   ├── handler/                    # HTTP controllers (Profile, Admin, Notification, Doc)
│   ├── kafka/                      # Kafka producer abstraction and payload models
│   ├── middleware/                 # Auth identity, Role RBAC, Rate limiter, Body limiter
│   ├── model/                      # Domain and DTO models
│   ├── ratelimit/                  # Sliding-window Redis rate limiter with Lua scripts
│   ├── repository/                 # PostgreSQL data access layer (Profile, Notification, Admin)
│   ├── router/                     # Chi multiplexer routing and documentation endpoints
│   └── service/                    # Core business logic (User, Notification)
├── openspec/                       # OpenSpec specifications and planning artifacts
├── prisma/
│   └── schema.prisma               # Prisma schema definition
├── test/                           # Integration and end-to-end test suites
├── Dockerfile                      # Multi-stage container build definition
├── go.mod / go.sum                 # Go module definitions
└── .env.example                    # Environment variable configuration template
```

---

## ⚙️ Prerequisites & Environment Configuration

### Prerequisites
- **Go**: Version 1.26 or higher
- **PostgreSQL**: Version 14 or higher (or Supabase)
- **Apache Kafka**: Version 3.x+
- **Redis**: Version 7.x+
- **Prisma CLI**: For schema migrations (`npm install -g prisma` or `npx prisma`)

### Configuration Options (`.env`)

Copy the example configuration file:
```bash
cp .env.example .env
```

| Variable | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `SERVER_PORT` | `string` | `8082` | HTTP port for the microservice |
| `ENV` | `string` | `development` | Deployment environment (`development`, `production`, `test`) |
| `DATABASE_URL` | `string` | *(Required)* | PostgreSQL connection string (Supabase pooler / standard) |
| `REDIS_URL` | `string` | `redis://localhost:6379` | Redis connection URL for distributed rate limiting |
| `REDIS_PASSWORD` | `string` | `""` | Optional password for Redis authentication |
| `ENABLE_RATE_LIMITER` | `bool` | `true` | Set `false` to bypass Redis rate limiting |
| `RATE_LIMIT_MAX_REQUESTS` | `int` | `60` | Default sliding-window limit per window for standard endpoints |
| `RATE_LIMIT_WINDOW_SECONDS` | `int` | `60` | Sliding-window duration in seconds for general endpoints |
| `RATE_LIMIT_DELETE_MAX_REQUESTS` | `int` | `3` | Maximum account deletion attempts per window |
| `RATE_LIMIT_DELETE_WINDOW_SECONDS` | `int` | `60` | Sliding-window duration in seconds for account deletion |
| `KAFKA_BROKERS` | `string` | `localhost:9092` | Comma-separated list of Kafka broker addresses |
| `KAFKA_TOPIC_USER_EVENTS` | `string` | `user.events` | Kafka topic for publishing user lifecycle events |
| `ORDER_SERVICE_GRPC_ADDR` | `string` | `localhost:50051` | gRPC host:port address for `store_order` |
| `ORDER_SERVICE_GRPC_TIMEOUT_MS` | `int` | `3000` | gRPC dial and invocation timeout in milliseconds |

---

## 🗄 Database Setup & Prisma ORM

The project uses Prisma schema (`prisma/schema.prisma`) to maintain models and generate the Go client into `internal/db`.

1. **Push Schema to PostgreSQL Database**:
   ```bash
   npx prisma db push --schema=./prisma/schema.prisma
   ```

2. **Generate Go Client**:
   ```bash
   go run github.com/steebchen/prisma-client-go generate --schema=./prisma/schema.prisma
   ```

---

## 🚀 Getting Started (Local Development)

1. **Clone the repository**:
   ```bash
   git clone https://github.com/Sall-lah/store_user.git
   cd store_user
   ```

2. **Install Go Dependencies**:
   ```bash
   go mod download
   ```

3. **Configure Environment Variables**:
   ```bash
   cp .env.example .env
   # Edit .env to set your DATABASE_URL, REDIS_URL, KAFKA_BROKERS, and ORDER_SERVICE_GRPC_ADDR
   ```

4. **Run the Service**:
   ```bash
   go run cmd/server/main.go
   ```

   The service will start listening on `http://localhost:8082`.

---

## 📡 API Endpoints & Documentation

Interactive API documentation is accessible directly from the service:
- **Swagger UI**: [http://localhost:8082/docs](http://localhost:8082/docs) or [http://localhost:8082/swagger](http://localhost:8082/swagger)
- **OpenAPI 3.1 Specs**: [http://localhost:8082/docs/openapi.yaml](http://localhost:8082/docs/openapi.yaml) or `/docs/openapi.json`

### Endpoint Catalog

| Group | Method | Path | Auth / Headers | Rate Limit | Description |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Health** | `GET` | `/health` | None | None | Service liveness probe |
| **Documentation** | `GET` | `/docs` | None | None | Interactive Swagger UI |
| **Documentation** | `GET` | `/docs/openapi.yaml` | None | None | Raw OpenAPI 3.1 specification (YAML) |
| **Documentation** | `GET` | `/docs/openapi.json` | None | None | Raw OpenAPI 3.1 specification (JSON) |
| **User Profile** | `POST` | `/api/users` | `X-User-Id` | 15 req/min | Synchronously provision profile during OTP activation |
| **User Profile** | `GET` | `/api/users/profile` | `X-User-Id` | 60 req/min | Retrieve personal profile (auto-provisions if new user) |
| **User Profile** | `PUT` | `/api/users/profile` | `X-User-Id` | 15 req/min | Update profile fields (`fullName`, `phoneNumber`, `address`) |
| **Account Lifecycle** | `DELETE` | `/api/users/account` | `X-User-Id` | 3 req/min | Self-service account deletion guarded by gRPC active order check |
| **In-App Notifications** | `GET` | `/api/users/notifications` | `X-User-Id` | 60 req/min | Paginated notification feed with `is_read` filter |
| **In-App Notifications** | `GET` | `/api/users/notifications/{id}` | `X-User-Id` | 60 req/min | Retrieve single notification item |
| **In-App Notifications** | `PATCH` | `/api/users/notifications/{id}/read` | `X-User-Id` | 30 req/min | Mark individual notification as read |
| **In-App Notifications** | `POST` | `/api/users/notifications/read-all` | `X-User-Id` | 30 req/min | Mark all user notifications as read |
| **In-App Notifications** | `DELETE` | `/api/users/notifications/{id}` | `X-User-Id` | 30 req/min | Delete notification record |
| **Admin Governance** | `DELETE` | `/api/admin/users/{id}` | `X-User-Role: ADMIN` | 15 req/min | Forced user purge bypassing active order checks |
| **Admin Governance** | `POST` | `/api/admin/users/{id}/ban` | `X-User-Role: ADMIN` | 15 req/min | Forced user ban (preserves profile data, emits `user.banned`) |

---

## 📬 Kafka Domain Events & Inter-Service Communication

`store_user` produces lifecycle events to the `user.events` topic using dual-compatible JSON keys for heterogeneous downstream consumers:

### Outbound Domain Events (`user.events`)

#### 1. `user.deleted` Event
Published when an account is deleted (customer self-delete or admin delete):
```json
{
  "event": "user.deleted",
  "event_type": "user.deleted",
  "userId": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
  "user_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
  "timestamp": "2026-08-24T23:30:00Z",
  "reason": "user_requested_deletion"
}
```
* **`store_auth`**: Purges credentials, revokes refresh tokens, and blacklists active JWTs in Redis.
* **`store_order`**: Cancels unpaid orders, masks customer PII for GDPR compliance, and releases reserved stock.

#### 2. `user.banned` Event
Published when an administrator bans a user account:
```json
{
  "event": "user.banned",
  "event_type": "user.banned",
  "userId": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
  "user_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
  "timestamp": "2026-08-24T23:30:00Z",
  "reason": "payment_fraud"
}
```
* **`store_auth`**: Sets `isActive = false`, revokes refresh tokens, and blacklists active JWTs in Redis.
* **`store_order`**: Cancels unpaid orders and releases inventory while **strictly retaining customer PII** for chargeback and fraud defense.
* **`store_user`**: Preserves profile and notification records for forensic audit trails.

---

## 🛡 Redis Rate Limiting Rules

The service implements sliding-window counter rate limiting with the following policy definitions:

| Scope / Route | Limit | Window | Key Strategy |
| :--- | :--- | :--- | :--- |
| **Profile Provision (`POST /api/users`)** | 15 req | 1 minute | `ratelimit:user:profile:create:<userId>` |
| **Profile Retrieval (`GET /api/users/profile`)** | 60 req | 1 minute | `ratelimit:user:profile:read:<userId>` |
| **Profile Update (`PUT /api/users/profile`)** | 15 req | 1 minute | `ratelimit:user:profile:update:<userId>` |
| **Account Deletion (`DELETE /api/users/account`)** | 3 req | 1 minute | `ratelimit:user:account:delete:<userId>` |
| **Notification Feed (`GET /api/users/notifications*`)** | 60 req | 1 minute | `ratelimit:user:notif:read:<userId>` |
| **Notification Mutations (`PATCH/POST/DELETE /...`)** | 30 req | 1 minute | `ratelimit:user:notif:update:<userId>` |
| **Admin User Deletion (`DELETE /api/admin/users/{id}`)** | 15 req | 1 minute | `ratelimit:admin:user:delete:<adminId>` |
| **Admin User Ban (`POST /api/admin/users/{id}/ban`)** | 15 req | 1 minute | `ratelimit:admin:user:ban:<adminId>` |

### Resilience & Degraded Headers
- **Fail-Open Policy**: If Redis is unreachable or latency exceeds `50ms`, the limiter allows traffic to pass uninterrupted to prevent denial of service on cache outage.
- **Response Headers**:
  - `X-RateLimit-Limit`: Maximum requests permitted within the sliding window.
  - `X-RateLimit-Remaining`: Remaining request quota.
  - `Retry-After`: Seconds until quota replenishment (on `429 Too Many Requests`).

---

## 🧪 Testing

Execute unit and integration test suites:

```bash
# Run all test packages
go test -v ./...

# Run test suite with race detector and coverage
go test -race -cover ./...

# Run integration tests against test databases
go test -v ./test/...
```

---

## 🐳 Docker Deployment

A production-ready, multi-stage Docker build is provided:

1. **Build Container Image**:
   ```bash
   docker build -t store_user:latest .
   ```

2. **Run Container**:
   ```bash
   docker run -d \
     --name store_user \
     -p 8082:8082 \
     -e SERVER_PORT="8082" \
     -e DATABASE_URL="postgresql://postgres:password@postgres:5432/store_user?schema=public" \
     -e REDIS_URL="redis://redis:6379" \
     -e KAFKA_BROKERS="kafka:9092" \
     -e ORDER_SERVICE_GRPC_ADDR="store_order:50051" \
     store_user:latest
   ```

3. **Check Container Health**:
   ```bash
   docker inspect --format='{{json .State.Health}}' store_user
   ```
