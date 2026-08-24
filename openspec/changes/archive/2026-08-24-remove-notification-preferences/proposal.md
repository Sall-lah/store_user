## Why

The `user_notification_preferences` table and its associated endpoints (`GET` and `PUT /api/notifications/preferences`) are no longer needed in `store_user`. Removing them simplifies the data model, eliminates redundant database storage and maintenance overhead, and streamlines notification logic down to core in-app feed operations.

## What Changes

- **BREAKING**: Remove database model `user_notification_preferences` from `prisma/schema.prisma` and drop the corresponding table from PostgreSQL.
- **BREAKING**: Remove `GET /api/notifications/preferences` and `PUT /api/notifications/preferences` HTTP endpoints from router, handler, service, and repository layers.
- Remove DTOs (`NotificationPreferencesDTO`, `UpdateNotificationPreferencesRequest`) and repository models (`NotificationPreferencesRecord`, `UpdateNotificationPreferencesParams`).
- Remove mock repository methods (`GetPreferences`, `UpsertPreferences`) and all associated unit tests across handler, service, repository, and router packages.
- Update OpenAPI definitions (`docs/openapi.json`, `docs/openapi.yaml`) to remove `/api/notifications/preferences` routes and schema components.

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- `user-notifications`: Remove notification preferences management requirement (`GET` and `PUT /api/notifications/preferences`).
- `security-rate-limiting`: Remove rate limit rules and scenarios referencing notification preferences endpoints.
- `api-documentation-ui`: Update complete API coverage requirements to exclude the removed preferences endpoints.

## Impact

- **Database**: `user_notification_preferences` table dropped; Prisma client regenerated.
- **API Endpoints**: `GET /api/notifications/preferences` and `PUT /api/notifications/preferences` return 404 Not Found.
- **Internal Layers**: `internal/repository`, `internal/service`, `internal/handler`, and `internal/router` cleaned of all preferences structs, interfaces, and methods.
- **Documentation**: OpenAPI 3.1 specifications updated.
