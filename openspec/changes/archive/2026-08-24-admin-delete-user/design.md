## Context

In the store microservice architecture, customer-initiated account deletion (`DELETE /api/users/account`) enforces a strict safety guarantee: it performs a synchronous gRPC pre-flight check against `store_order` (`OrderService.CheckActiveOrders`) and blocks deletion if active orders exist (`409 Conflict`). 

However, system administrators require the authority to forcefully delete user accounts (e.g. for policy violations, fraud, GDPR/CCPA purge requests, or manual support interventions) regardless of order state. Deleting an account must also notify `store_auth` via Apache Kafka domain events (`user.events`) to invalidate active tokens in Redis, revoke refresh tokens, and remove authentication credentials.

Requests are routed through `store_gateway` using NGINX `auth-offload` subrequests (`/_auth_verify`), which populate verified `X-User-Id`, `X-User-Role`, and `X-User-Email` headers before passing to downstream microservices.

## Goals / Non-Goals

**Goals:**
- Provide an administrative endpoint `DELETE /api/admin/users/{id}` protected by `RequireRole("ADMIN")` middleware.
- Completely bypass order pre-flight verification checks during administrative deletion.
- Clean up database profile (`user_profiles`) and notifications (`user_notifications`) in `store_user`.
- Publish `{ event: "user.deleted", userId: "<id>", reason: "admin_deletion" }` to Kafka topic `user.events` for asynchronous session invalidation and credential purge in `store_auth`.
- Fully document the endpoint in OpenAPI 3.1 YAML and JSON.

**Non-Goals:**
- Modifying customer self-service deletion behavior at `DELETE /api/users/account` (which retains its active order pre-flight check).
- Implementing administrative batch user deletion in this phase (single target `{id}` only).
- Performing cross-service distributed 2PC transactions (eventual consistency via Kafka is standard across the platform).

## Decisions

### 1. Endpoint URI: `DELETE /api/admin/users/{id}`
- **Decision**: Expose the endpoint at `DELETE /api/admin/users/{id}` in `store_user` and map `/api/admin/users/` in `store_gateway` with `auth-offload.conf`.
- **Rationale**: Matches RESTful conventions and aligns with admin routes in sibling services (`/api/admin/products`, `/api/admin/orders`).
- **Alternatives Considered**:
  - `DELETE /api/users/admin/{id}`: Keeps routing under `/api/users/` without gateway changes, but breaks endpoint naming conventions established across other microservices.

### 2. Role Enforcement via Middleware (`RequireRole`)
- **Decision**: Implement a reusable `RequireRole(roles ...string)` middleware that inspects the context populated by `AuthIdentity`. Performs case-insensitive matching (`ADMIN` vs `admin`).
- **Rationale**: Centralizes role checks cleanly and prevents duplicated role-parsing logic inside handlers. Returns `403 Forbidden` if role is insufficient.
- **Alternatives Considered**:
  - Manual role check inside handler body: Leaks transport/security concerns into business handler code.

### 3. Asynchronous Coordination with `store_auth` via Kafka
- **Decision**: Publish domain event `LifecycleEvent` with `Event: "user.deleted"`, `UserID: targetUserID`, and `Reason: "admin_deletion"` onto Kafka topic `user.events`.
- **Rationale**: `store_auth`'s existing `UserEventConsumer` listens on topic `user.events` and already contains handlers to blacklist user JWTs in Redis (15m TTL), revoke refresh tokens, and delete the user row from `users`.
- **Alternatives Considered**:
  - Synchronous gRPC or REST call to `store_auth`: Introduces synchronous tight coupling and failure cascades; event-driven model matches the established architecture.

### 4. Idempotent Target User Deletion
- **Decision**: If target user's profile does not exist in `user_profiles`, database deletion silently succeeds and Kafka event is still published.
- **Rationale**: Allows administrators to clean up orphaned auth credentials in `store_auth` even if a profile was never created or was already removed in `store_user`.

## Risks / Trade-offs

- **[Risk] Active Orders in Flight** → Admin force-deleting a user with in-flight orders leaves orders in `store_order` without an active user profile.
  - *Mitigation*: This is an intended administrative override. Orders retain `userId` and snapshot customer data (`userEmail`, `shippingAddress`, order items) so fulfillment remains unbroken.
- **[Risk] Kafka Broker Transient Failure** → Database record deleted but Kafka publication fails.
  - *Mitigation*: Handled with standard error propagation (`500 Internal Server Error` / `ErrKafkaPublishFailed`) allowing the admin to retry safely.
