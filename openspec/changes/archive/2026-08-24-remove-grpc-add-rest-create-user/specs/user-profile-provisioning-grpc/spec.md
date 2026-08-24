## REMOVED Requirements

### Requirement: gRPC UserService CreateUserProfile Implementation
**Reason**: Protobuf contract for `store.user.v1.UserService` removed from `store_proto`; profile provisioning is now handled via REST endpoint `POST /api/users`.
**Migration**: Call `POST /api/users` over HTTP/REST via API Gateway with `X-User-Id` header.

### Requirement: Dedicated gRPC Server Lifecycle
**Reason**: `store_user` no longer acts as a gRPC server; only requires gRPC client functionality for outbound dependency checks (`store_order`).
**Migration**: `store_user` operates exclusively as an HTTP Chi web service on `SERVER_PORT`.
