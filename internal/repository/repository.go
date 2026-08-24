package repository

import "context"

// UserProfileRepository defines the persistence contract for personal profile records.
// Why: Abstracts the database implementation (Prisma / SQL) to permit seamless unit testing with mocks.
type UserProfileRepository interface {
	GetByUserID(ctx context.Context, userID string) (*UserProfile, error)
	Create(ctx context.Context, userID, fullName string, phoneNumber, address *string) (*UserProfile, error)
	Upsert(ctx context.Context, userID string, params UpdateProfileParams) (*UserProfile, error)
	HardDeleteByUserID(ctx context.Context, userID string) error
}

