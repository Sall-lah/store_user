package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Sall-lah/store_user/internal/db"
)

// ErrNotFound indicates a requested user profile does not exist.
var ErrNotFound = errors.New("user profile not found")

// PrismaUserProfileRepository implements UserProfileRepository backed by Prisma Client Go.
type PrismaUserProfileRepository struct {
	client *db.PrismaClient
}

// NewPrismaUserProfileRepository creates a new instance of the repository with Prisma client dependencies.
// Why: Injects database connection pool into repository operations.
func NewPrismaUserProfileRepository(client *db.PrismaClient) *PrismaUserProfileRepository {
	return &PrismaUserProfileRepository{client: client}
}

// GetByUserID retrieves a profile by the user's UUID.
// Why: Provides primary key identity lookup for authenticated user queries.
func (r *PrismaUserProfileRepository) GetByUserID(ctx context.Context, userID string) (*UserProfile, error) {
	m, err := r.client.UserProfiles.FindUnique(
		db.UserProfiles.UserID.Equals(userID),
	).Exec(ctx)

	if err != nil {
		if db.IsErrNotFound(err) || errors.Is(err, db.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user profile by user_id: %w", err)
	}

	return mapPrismaModelToUserProfile(m), nil
}

// Create inserts a new user profile record for a user.
// Why: Provides explicit profile provisioning for newly registered users.
func (r *PrismaUserProfileRepository) Create(ctx context.Context, userID, fullName string, phoneNumber, address *string) (*UserProfile, error) {
	var createOptional []db.UserProfilesSetParam
	if phoneNumber != nil {
		createOptional = append(createOptional, db.UserProfiles.PhoneNumber.Set(*phoneNumber))
	}
	if address != nil {
		createOptional = append(createOptional, db.UserProfiles.Address.Set(*address))
	}

	m, err := r.client.UserProfiles.CreateOne(
		db.UserProfiles.UserID.Set(userID),
		db.UserProfiles.FullName.Set(fullName),
		createOptional...,
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create user profile: %w", err)
	}

	return mapPrismaModelToUserProfile(m), nil
}

// Upsert creates a baseline profile if none exists or updates specified mutable attributes.
// Why: Provides idempotent profile creation and mutation in a single roundtrip.
func (r *PrismaUserProfileRepository) Upsert(ctx context.Context, userID string, params UpdateProfileParams) (*UserProfile, error) {
	fullName := "User"
	if params.FullName != nil && *params.FullName != "" {
		fullName = *params.FullName
	}

	var createOptional []db.UserProfilesSetParam
	var updateParams []db.UserProfilesSetParam

	if params.FullName != nil {
		updateParams = append(updateParams, db.UserProfiles.FullName.Set(*params.FullName))
	}
	if params.PhoneNumber != nil {
		createOptional = append(createOptional, db.UserProfiles.PhoneNumber.Set(*params.PhoneNumber))
		updateParams = append(updateParams, db.UserProfiles.PhoneNumber.Set(*params.PhoneNumber))
	}
	if params.Address != nil {
		createOptional = append(createOptional, db.UserProfiles.Address.Set(*params.Address))
		updateParams = append(updateParams, db.UserProfiles.Address.Set(*params.Address))
	}

	m, err := r.client.UserProfiles.UpsertOne(
		db.UserProfiles.UserID.Equals(userID),
	).Create(
		db.UserProfiles.UserID.Set(userID),
		db.UserProfiles.FullName.Set(fullName),
		createOptional...,
	).Update(
		updateParams...,
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to upsert user profile: %w", err)
	}

	return mapPrismaModelToUserProfile(m), nil
}

// HardDeleteByUserID permanently removes the user's profile record from the database.
// Why: Enforces privacy compliance and purges all PII upon verified account deletion.
func (r *PrismaUserProfileRepository) HardDeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.client.UserProfiles.FindUnique(
		db.UserProfiles.UserID.Equals(userID),
	).Delete().Exec(ctx)

	if err != nil {
		if db.IsErrNotFound(err) || errors.Is(err, db.ErrNotFound) || strings.Contains(err.Error(), "Record to delete does not exist") {
			return nil // Already deleted or not found; idempotent
		}
		return fmt.Errorf("failed to hard delete user profile: %w", err)
	}

	return nil
}

func mapPrismaModelToUserProfile(m *db.UserProfilesModel) *UserProfile {
	if m == nil {
		return nil
	}
	return &UserProfile{
		ID:          m.ID,
		UserID:      m.UserID,
		FullName:    m.FullName,
		PhoneNumber: m.InnerUserProfiles.PhoneNumber,
		Address:     m.InnerUserProfiles.Address,
		CreatedAt:   time.Time(m.CreatedAt),
		UpdatedAt:   time.Time(m.UpdatedAt),
	}
}
