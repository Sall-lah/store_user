## Why

The store microservices platform currently lacks a dedicated user profile and account lifecycle service. To decouple authentication and credential storage (store_auth) from personal profile data, and to provide a secure, synchronous pre-flight guard against deleting accounts with active orders in store_order, the store_user service is established.

## What Changes

- **User Profile Management**: Implement HTTP endpoints to retrieve and update personal profile attributes (e.g., name, phone number, bio, avatar, address, gender, date of birth) stored in Supabase PostgreSQL (user_profiles table).
- **Account Deletion Lifecycle with Order Verification**: Implement a secure account deletion workflow that performs a synchronous gRPC pre-flight check against store_order (OrderService.CheckActiveOrders), prevents deletion if active orders exist (HTTP 409 Conflict), hard-deletes the profile record upon approval, and publishes an asynchronous user.deleted domain event to Kafka topic user.events.
- **Downstream Authentication Synchronization**: The published user.deleted Kafka event triggers store_auth to blacklist active JWTs in Redis, revoke refresh tokens, and purge auth credentials.
- **Perimeter & Rate Limiting Security**: Integrate Redis sliding-window rate limiting per endpoint/user, gateway identity header ingestion (X-User-Id), max payload size limits, and HTML/XSS sanitization.

## Capabilities

### New Capabilities
- user-profile-management: Retrieve and update user profile information in Supabase PostgreSQL with strict input validation and sanitization.
- ccount-deletion: Secure account deletion with synchronous gRPC check against store_order, profile hard deletion, and Kafka user.deleted event dispatch.
- security-rate-limiting: Redis-backed sliding-window rate limiting, max body size guards, and gateway identity verification.

### Modified Capabilities
<!-- None: This is a greenfield service -->

## Impact

- **Database**: New user_profiles table in Supabase PostgreSQL.
- **Inter-service Dependencies**:
  - store_proto (gRPC client for OrderService.CheckActiveOrders)
  - store_order (gRPC server target)
  - store_auth (Kafka consumer for user.deleted events on user.events)
  - store_gateway (Reverse proxy routing /api/v1/users/* with X-User-Id injection)
- **Infrastructure**: Apache Kafka broker and Redis cache instance.
