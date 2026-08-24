## REMOVED Requirements

### Requirement: Notification Preferences Management
**Reason**: Communication channel preferences (email, push, SMS, order updates, promotions) are no longer managed by store_user.
**Migration**: Clients must discontinue calls to `GET /api/users/notifications/preferences` and `PUT /api/users/notifications/preferences`.
