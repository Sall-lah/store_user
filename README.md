# Store User Microservice (`store_user`)

High-performance User Profile Management, Account Lifecycle, and In-App Notifications microservice for the e-commerce platform.

---

## 🏗 Architecture Overview

```
                      ┌─────────────────────────────────────────┐
                      │              store_gateway              │
                      │  (TLS / Rate Limiting / JWT Auth)       │
                      └────────────────────┬────────────────────┘
                                           │
                        Injected Identity Headers:
                        • X-User-Id
                        • X-User-Role
                        • X-User-Email
                                           │
                                           ▼
                      ┌─────────────────────────────────────────┐
                      │               store_user                │
                      │        (Port 8082 / Go Chi)             │
                      └──────┬─────────────┬─────────────┬──────┘
                             │             │             │
              Synchronous    │             │ Persistence │ Asynchronous
              gRPC Check     │             │             │ Domain Events
                             ▼             ▼             ▼
                      ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
                      │ store_order │ │  PostgreSQL │ │ Apache Kafka│
                      │(Active Order│ │ (Prisma Go) │ │(user.events)│
                      │ Verification│ │             │ │             │
                      └─────────────┘ └─────────────┘ └──────┬──────┘
                                                             │
                                              ┌──────────────┴──────────────┐
                                              ▼                             ▼
                                       ┌─────────────┐               ┌─────────────┐
                                       │ store_auth  │               │ store_order │
                                       │(Token & PII │               │(Order Cancel│
                                       │ Cleanup)    │               │ & Stock Adj)│
                                       └─────────────┘               └─────────────┘
```

---

## ⚡ Tech Stack

- **Language & Router**: Go 1.26 with [`chi/v5`](https://github.com/go-chi/chi)
- **Database & ORM**: PostgreSQL (Supabase) via [`prisma-client-go`](https://github.com/steebchen/prisma-client-go)
- **Inter-Service Communication**:
  - **Synchronous**: Google gRPC client connecting to `store_order`
  - **Asynchronous**: Apache Kafka (`segmentio/kafka-go`) publishing domain events on topic `user.events`
- **Rate Limiting & Security**: Redis sliding-window with atomic Lua scripts and fail-open degradation
- **API Documentation**: OpenAPI 3.1.0 with embedded Swagger UI (`/docs`)

---

## 🛡️ Authentication & Authorization

All customer and administrative endpoints leverage **API Gateway Offloading**. `store_gateway` validates incoming RS256 JWTs against `store_auth` and forwards trusted headers:

| Header | Description | Required Roles |
| :--- | :--- | :--- |
| `X-User-Id` | Target User UUID | All authenticated endpoints |
| `X-User-Role` | Identity role claim (`CUSTOMER`, `ADMIN`) | `ADMIN` (or `admin`) for admin endpoints |
| `X-User-Email` | Verified user email address | Propagated downstream |

---

## 📡 API Endpoints

### 1. User Profile Management
| Method | Endpoint | Description | Rate Limit |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/users/profile` | Retrieve personal profile (auto-provisions if new user) | 60 req/min |
| `POST` | `/api/users/profile` | Synchronously provision profile during OTP activation | 30 req/min |
| `PUT` | `/api/users/profile` | Update profile fields (`fullName`, `phoneNumber`, `address`) | 15 req/min |

### 2. Account Lifecycle & Deletion
| Method | Endpoint | Description | Rate Limit |
| :--- | :--- | :--- | :--- |
| `DELETE` | `/api/users/account` | Self-service account deletion guarded by gRPC active order check | 3 req/min |

### 3. Admin User Governance
| Method | Endpoint | Description | Rate Limit |
| :--- | :--- | :--- | :--- |
| `DELETE` | `/api/admin/users/{id}` | Forced user purge bypassing active order checks | 15 req/min |
| `POST` | `/api/admin/users/{id}/ban` | Forced user ban (preserves profile data, emits `user.banned`) | 15 req/min |

### 4. In-App Notifications
| Method | Endpoint | Description | Rate Limit |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/users/notifications` | Paginated notification feed with `is_read` filter | 60 req/min |
| `GET` | `/api/users/notifications/{id}` | Single notification item lookup | 60 req/min |
| `PATCH` | `/api/users/notifications/{id}/read` | Mark individual notification as read | 30 req/min |
| `POST` | `/api/users/notifications/read-all` | Mark all user notifications as read | 30 req/min |
| `DELETE` | `/api/users/notifications/{id}` | Delete notification record | 30 req/min |

### 5. Health & Documentation
| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/health` | Service liveness probe |
| `GET` | `/docs` | Embedded interactive Swagger UI |
| `GET` | `/docs/openapi.yaml` | Raw OpenAPI 3.1.0 specification in YAML |
| `GET` | `/docs/openapi.json` | Raw OpenAPI 3.1.0 specification in JSON |

---

## 📬 Kafka Domain Event Stream (`user.events`)

`store_user` produces lifecycle events to the `user.events` topic using dual-compatible JSON keys for heterogeneous downstream consumers:

### `user.deleted` Event
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
* **`store_order`**: Cancels unpaid orders, masks PII for GDPR compliance, and releases inventory.

### `user.banned` Event
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
* **`store_user`**: Preserves profile and notification data for forensic records.

---

## ⚙️ Environment Variables

| Variable | Description | Default |
| :--- | :--- | :--- |
| `SERVER_PORT` | HTTP listening port | `8082` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://postgres:...` |
| `REDIS_ADDR` | Redis cache and rate limiter host:port | `127.0.0.1:6379` |
| `REDIS_PASSWORD` | Redis authentication password | `""` |
| `KAFKA_BROKERS` | Comma-separated Kafka broker addresses | `127.0.0.1:9092` |
| `KAFKA_TOPIC_USER_EVENTS` | Kafka topic for user lifecycle events | `user.events` |
| `ORDER_SERVICE_GRPC_TARGET` | gRPC host:port for `store_order` | `127.0.0.1:50051` |
| `ORDER_SERVICE_GRPC_TIMEOUT` | gRPC dial and invocation timeout | `3s` |
| `RATE_LIMIT_MAX_REQUESTS` | Default sliding-window limit per minute | `60` |
| `RATE_LIMIT_DELETE_MAX_REQUESTS`| Account deletion limit per minute | `3` |

---

## 🧪 Testing & Verification

Run the full automated test suite:
```bash
# Run unit, handler, and mock tests
go test -v ./...

# Run all tests without cache
go test -count=1 ./...
```

---

## 🐳 Container Build

Build a lightweight Linux production container image using Podman or Docker:
```bash
# Using Podman
podman build -t store_user:latest .

# Using Docker
docker build -t store_user:latest .
```
