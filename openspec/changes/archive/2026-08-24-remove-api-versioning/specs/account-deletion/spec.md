## MODIFIED Requirements

### Requirement: Synchronous Order Pre-Flight Check on Account Deletion
The system SHALL invoke store_order's gRPC OrderService.CheckActiveOrders RPC before proceeding with user account deletion. If the user has one or more orders in active lifecycle states (PENDING_PAYMENT, PAID, PROCESSING, SHIPPED), deletion MUST be aborted.

#### Scenario: Deletion blocked due to active orders
- **WHEN** an authenticated user with active in-flight orders requests DELETE /api/users/account
- **THEN** the system calls store_order via gRPC, detects has_active_orders == true, aborts deletion, and returns HTTP 409 Conflict with details of the blocking orders.

#### Scenario: Order service gRPC connection timeout or failure
- **WHEN** an authenticated user requests DELETE /api/users/account and store_order gRPC service is unavailable or times out
- **THEN** the system fails safely, preserves the user profile without deletion, and returns HTTP 503 Service Unavailable.

### Requirement: Profile Hard Deletion
The system SHALL hard-delete the user's profile record from the Supabase PostgreSQL database when the pre-flight order check confirms zero active orders.

#### Scenario: Successful hard deletion of profile
- **WHEN** an authenticated user with zero active orders requests DELETE /api/users/account
- **THEN** the system removes the user's row from the user_profiles table in Supabase PostgreSQL.

### Requirement: Domain Event Publication to Kafka
Upon successfully deleting the user's profile, the system SHALL publish a user.deleted event to the Kafka topic user.events with payload adhering to LifecycleEvent.

#### Scenario: Publish user deleted lifecycle event
- **WHEN** user profile hard-deletion succeeds
- **THEN** the system emits { event: user.deleted, userId: <userId>, timestamp: <ISO-timestamp>, reason: user_requested_deletion } to Kafka topic user.events and returns HTTP 200 OK.
