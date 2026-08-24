package service

import "github.com/Sall-lah/store_user/internal/repository"

// CreateProfileRequest defines payload parameters for provisioning a new user profile.
// Why: Encapsulates incoming RPC identity parameters needed to initialize a profile.
type CreateProfileRequest struct {
	UserID      string  `json:"userId"`
	FullName    string  `json:"fullName"`
	Email       string  `json:"email,omitempty"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
	Address     *string `json:"address,omitempty"`
}

// CreateProfileResult encapsulates the provisioned profile and creation status metadata.
// Why: Conveys whether the profile was freshly inserted or resolved idempotently from an existing record.
type CreateProfileResult struct {
	Profile   *repository.UserProfile
	IsCreated bool
	Message   string
}

// UpdateProfileRequest defines the allowed payload attributes for updating user profiles.
// Why: Enforces strict encapsulation of user-mutable fields in HTTP request bodies.
type UpdateProfileRequest struct {
	FullName    *string `json:"fullName"`
	PhoneNumber *string `json:"phoneNumber"`
	Address     *string `json:"address"`
}

// DeleteAccountRequest defines optional parameters provided when initiating account deletion.
type DeleteAccountRequest struct {
	Reason string `json:"reason"`
}

// BanUserRequest defines optional parameters provided when an administrator bans a user account.
type BanUserRequest struct {
	Reason string `json:"reason"`
}



