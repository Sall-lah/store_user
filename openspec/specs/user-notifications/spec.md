# User Notifications

## Purpose
Provides management of in-app user notification feeds, read/unread state updates, and deletion.

## Requirements

### Requirement: Notification Feed Retrieval
The system SHALL provide an endpoint `GET /api/users/notifications` to retrieve paginated notification records for the authenticated user, supporting optional filtering by read status (`is_read`) and pagination query parameters (`limit`, `page`).

#### Scenario: Successful notification list retrieval
- **WHEN** an authenticated user calls `GET /api/users/notifications?page=1&limit=20`
- **THEN** the system returns HTTP 200 OK with the paginated list of notifications, unread count, and pagination metadata.

#### Scenario: Unauthorized notification list retrieval
- **WHEN** a request to `GET /api/users/notifications` lacks valid authentication headers
- **THEN** the system returns HTTP 401 Unauthorized.

### Requirement: Notification Read Status Updates
The system SHALL provide endpoints `PATCH /api/users/notifications/{id}/read` to mark a specific notification as read and `POST /api/users/notifications/read-all` to mark all unread notifications of the caller as read.

#### Scenario: Successfully mark a single notification as read
- **WHEN** an authenticated user calls `PATCH /api/users/notifications/{id}/read` for a notification they own
- **THEN** the system marks `is_read` to true, records `read_at` timestamp, and returns HTTP 200 OK with the updated notification object.

#### Scenario: Marking non-existent or foreign notification as read
- **WHEN** an authenticated user calls `PATCH /api/users/notifications/{id}/read` for an ID that does not exist or belongs to another user
- **THEN** the system returns HTTP 404 Not Found.

#### Scenario: Successfully mark all notifications as read
- **WHEN** an authenticated user calls `POST /api/users/notifications/read-all`
- **THEN** the system marks all unread notifications for that user as read and returns HTTP 200 OK with the count of updated records.

### Requirement: Notification Deletion
The system SHALL provide an endpoint `DELETE /api/users/notifications/{id}` allowing users to permanently remove a notification record.

#### Scenario: Successful notification deletion
- **WHEN** an authenticated user calls `DELETE /api/users/notifications/{id}` for an owned notification
- **THEN** the system permanently removes the notification and returns HTTP 200 OK with a success message.

#### Scenario: Deleting non-existent or foreign notification
- **WHEN** an authenticated user calls `DELETE /api/users/notifications/{id}` for an ID not belonging to the user
- **THEN** the system returns HTTP 404 Not Found without modifying any records.
