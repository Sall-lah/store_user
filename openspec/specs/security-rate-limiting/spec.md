# Security & Rate Limiting

## Purpose
Provides security perimeter validation (gateway identity extraction) and API rate limiting using Redis sliding windows.

## Requirements

### Requirement: Redis Sliding-Window Rate Limiting
The system SHALL enforce rate limits using Redis Sorted Sets and Lua scripts. The rate limiter MUST apply tiered quotas per endpoint and user identity (userId).

#### Scenario: Rate limit exceeded on profile read
- **WHEN** a user exceeds 60 requests within a 1-minute window on GET /api/v1/users/profile
- **THEN** the system rejects subsequent requests with HTTP 429 Too Many Requests and includes Retry-After header.

#### Scenario: Rate limit exceeded on account deletion
- **WHEN** a user exceeds 3 deletion requests within a 1-minute window on DELETE /api/v1/users/account
- **THEN** the system rejects subsequent deletion attempts with HTTP 429 Too Many Requests.

#### Scenario: Redis failure degradation
- **WHEN** the Redis cluster is unreachable during a rate limit evaluation
- **THEN** the limiter degrades gracefully (fails open) to allow valid user traffic while logging a warning.

### Requirement: Perimeter Identity & Anti-Spoofing
The system SHALL extract and validate user identity claims from X-User-Id injected by store_gateway.

#### Scenario: Missing or malformed identity header
- **WHEN** an incoming request lacks a valid UUID format in X-User-Id and contains no valid fallback Bearer token
- **THEN** the system returns HTTP 401 Unauthorized.
