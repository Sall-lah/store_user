## 1. Schema & Database Cleanup

- [x] 1.1 Remove `model user_notification_preferences` from `prisma/schema.prisma`
- [x] 1.2 Run Prisma client generation to regenerate `internal/db` without preferences models

## 2. Repository Layer Cleanup

- [x] 2.1 Remove `NotificationPreferencesRecord` and `UpdateNotificationPreferencesParams` structs from `internal/repository/model_notification.go`
- [x] 2.2 Remove `GetPreferences`, `UpsertPreferences`, and `mapPrismaPreferencesModel` from `internal/repository/notification.go` and `NotificationRepository` interface
- [x] 2.3 Remove preferences mock methods (`GetPreferences`, `UpsertPreferences`) from `internal/repository/mock_notification.go`

## 3. Service Layer Cleanup

- [x] 3.1 Remove `NotificationPreferencesDTO` and `UpdateNotificationPreferencesRequest` structs from `internal/service/model_notification.go`
- [x] 3.2 Remove `GetPreferences`, `UpdatePreferences`, and `mapRepoPreferencesToDTO` from `internal/service/notification.go` and `NotificationService` interface

## 4. Handler & Router Cleanup

- [x] 4.1 Remove `GetPreferences` and `UpdatePreferences` handler functions from `internal/handler/notification.go`
- [x] 4.2 Remove `GET` and `PUT /api/notifications/preferences` route registrations from `internal/router/router.go`

## 5. OpenAPI Documentation Updates

- [x] 5.1 Remove `/api/notifications/preferences` paths and schemas from `docs/openapi.yaml`
- [x] 5.2 Remove `/api/notifications/preferences` paths and schemas from `docs/openapi.json`

## 6. Test Suite & Verification

- [x] 6.1 Clean up preferences test cases in `internal/repository/notification_test.go`
- [x] 6.2 Clean up preferences test cases in `internal/service/notification_test.go`
- [x] 6.3 Clean up preferences test cases in `internal/handler/notification_test.go`
- [x] 6.4 Clean up preferences test cases in `internal/router/router_test.go`
- [x] 6.5 Run `go test -v ./...` to verify all remaining packages compile and pass

