package repository

import "time"

// UserProfile represents the core domain model for a user's personal profile information.
// Why: Decouples domain logic and HTTP representations from specific ORM / Prisma generated models.
type UserProfile struct {
	ID          string     `json:"id"`
	UserID      string     `json:"userId"`
	FullName    string     `json:"fullName"`
	PhoneNumber *string    `json:"phoneNumber,omitempty"`
	AvatarURL   *string    `json:"avatarUrl,omitempty"`
	Bio         *string    `json:"bio,omitempty"`
	Address     *string    `json:"address,omitempty"`
	Gender      *string    `json:"gender,omitempty"`
	DateOfBirth *time.Time `json:"dateOfBirth,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// UpdateProfileParams specifies mutable fields when updating an existing user profile.
// Why: Distinguishes between explicitly provided values, omitted values, and non-updatable identifiers.
type UpdateProfileParams struct {
	FullName    *string
	PhoneNumber *string
	AvatarURL   *string
	Bio         *string
	Address     *string
	Gender      *string
	DateOfBirth *time.Time
}
