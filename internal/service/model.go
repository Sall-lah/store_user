package service

import "time"

// UpdateProfileRequest defines the allowed payload attributes for updating user profiles.
// Why: Enforces strict encapsulation of user-mutable fields in HTTP request bodies.
type UpdateProfileRequest struct {
	FullName    *string    `json:"fullName"`
	PhoneNumber *string    `json:"phoneNumber"`
	AvatarURL   *string    `json:"avatarUrl"`
	Bio         *string    `json:"bio"`
	Address     *string    `json:"address"`
	Gender      *string    `json:"gender"`
	DateOfBirth *time.Time `json:"dateOfBirth"`
}

// DeleteAccountRequest defines optional parameters provided when initiating account deletion.
type DeleteAccountRequest struct {
	Reason string `json:"reason"`
}
