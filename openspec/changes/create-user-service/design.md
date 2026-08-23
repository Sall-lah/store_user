## Context

The store_user service is a Go-based microservice responsible for personal user profile data and account lifecycle operations. It operates in a distributed e-commerce microservices platform alongside store_auth, store_order, store_gateway, and store_notification.

### Current State
- store_auth stores credentials (email, password_hash, ole, is_active, otp_codes, efresh_tokens). It has a Kafka consumer on topic user.events listening for user.deleted and user.banned events.
- store_order exposes gRPC service store.order.v1.OrderService with RPC CheckActiveOrders in store_proto.
- store_gateway terminates TLS, performs subrequest token verification against store_auth, strips untrusted headers, and forwards X-User-Id, X-User-Role, X-User-Email downstream.
- Supabase PostgreSQL is provisioned for persistent profile storage.
- Redis and Apache Kafka clusters are provisioned for rate limiting and event streaming.

## Goals / Non-Goals

**Goals:**
- Provide clean REST endpoints for Profile retrieval (GET /api/v1/users/profile) and updates (PUT /api/v1/users/profile).
- Provide an account deletion endpoint (DELETE /api/v1/users/account) guarded by synchronous gRPC pre-flight validation against store_order.
- Asynchronously notify store_auth via Kafka event user.deleted on topic user.events to invalidate tokens, purge refresh tokens, and remove credentials.
- Enforce multi-tier Redis sliding-window rate limiting to prevent abuse and denial-of-service.
- Provide robust input validation, payload size guards, and XSS sanitization.
- Maintain high testability with decoupled repository, client, and handler interfaces.

**Non-Goals:**
- Password changes or email updates (these remain exclusively in store_auth).
- Managing relational order histories or payment credentials directly in store_user.
- Relational foreign keys between Supabase user_profiles and external databases.

## Decisions

### Decision 1: Supabase PostgreSQL with Standalone user_profiles Model
- **Choice**: Store user profile data in a flat user_profiles table using userId (UUID) as the primary/unique key.
- **Rationale**: Isolates PII and profile metadata from authentication credentials. Avoids cross-database foreign key coupling.
- **Alternatives Considered**:
  - *Storing profiles in store_auth database*: Violates single-responsibility principle and couples authentication with profile domain logic.
  - *Normalized multi-table address schemas*: Unnecessary complexity for the current requirements.

### Decision 2: Synchronous gRPC for Order Pre-Flight Check
- **Choice**: Use gRPC client connecting to store_order (OrderService.CheckActiveOrders) with strict 2-second timeout context.
- **Rationale**: Account deletion must be strictly blocked if active orders are in flight (PENDING_PAYMENT, PAID, PROCESSING, SHIPPED). A synchronous check guarantees real-time consistency before any profile records are deleted.
- **Alternatives Considered**:
  - *HTTP REST call to order service*: Slower, lacks schema generation, higher serialization overhead compared to protobuf.
  - *Pure asynchronous saga without pre-flight*: Risk of race conditions where an order is placed concurrently or orphaned in warehouse fulfillment.

### Decision 3: Event-Driven Account Deletion Cleanup via Kafka
- **Choice**: Publish LifecycleEvent payload (event: user.deleted, userId, 	imestamp, eason) to Kafka topic user.events.
- **Rationale**: Matches existing store_auth consumer contract (store_auth/internal/user/consumer.go). Allows multiple downstream consumers (auth, notifications, audit) to react independently.
- **Alternatives Considered**:
  - *Direct gRPC/HTTP cascade calls to all services*: Tightly couples store_user to every downstream service and introduces cascading failure modes.

### Decision 4: Sliding-Window Rate Limiting in Redis via Atomic Lua
- **Choice**: Implement sliding-window rate limiter using Redis Sorted Sets (ZSET) and Lua script evaluation.
- **Rationale**: Prevents burst exploitation and race conditions across distributed instances. Enables tiered limits (60 req/min for GET, 15 req/min for PUT, 3 req/min for DELETE).
- **Alternatives Considered**:
  - *Fixed-window counter*: Vulnerable to double-quota traffic spikes at window boundaries.
  - *In-memory token bucket*: Does not synchronize across multi-replica deployments.

### Decision 5: Perimeter Auth Offloading with Gateway Header Extraction
- **Choice**: Trust X-User-Id injected by store_gateway while providing fallback RS256 JWKS validation for direct development/testing.
- **Rationale**: Avoids redundant JWT cryptographic verification on every request while preserving zero-trust compatibility.

## Risks / Trade-offs

- **[Risk] store_order gRPC service is unavailable during account deletion**
  → *Mitigation*: Gracefully return HTTP 503 Service Unavailable with a retry-after advisory. Never delete profile data if order status cannot be verified.
- **[Risk] Kafka broker is unreachable when emitting user.deleted**
  → *Mitigation*: Fail the deletion request before returning 200 OK or utilize retry backoff so auth credentials are never orphaned.
- **[Risk] Redis outage or latency spike**
  → *Mitigation*: Fail-open rate limiter logic with warning logs to preserve core user profile access.
- **[Risk] Malicious text input / XSS in profile bio or address**
  → *Mitigation*: Sanitize text inputs using HTML tag stripping and max character limits.
