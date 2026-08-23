package repository

import (
	"context"
	"time"
)

// MockUserProfileRepository provides an in-memory test double for repository operations.
// Why: Enables fast, isolated unit testing of domain services without requiring live PostgreSQL or Prisma connections.
type MockUserProfileRepository struct {
	Profiles map[string]*UserProfile
}

// NewMockUserProfileRepository initializes an empty in-memory repository mock.
// Why: Provides a ready-to-use in-memory repository for unit test fixtures.
func NewMockUserProfileRepository() *MockUserProfileRepository {
	return &MockUserProfileRepository{
		Profiles: make(map[string]*UserProfile),
	}
}

// GetByUserID retrieves a mock profile from memory.
func (m *MockUserProfileRepository) GetByUserID(ctx context.Context, userID string) (*UserProfile, error) {
	p, ok := m.Profiles[userID]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

// Upsert creates or mutates a mock profile in memory.
func (m *MockUserProfileRepository) Upsert(ctx context.Context, userID string, params UpdateProfileParams) (*UserProfile, error) {
	now := time.Now()
	p, exists := m.Profiles[userID]
	if !exists {
		fullName := "User"
		if params.FullName != nil && *params.FullName != "" {
			fullName = *params.FullName
		}
		p = &UserProfile{
			ID:          "prof-" + userID,
			UserID:      userID,
			FullName:    fullName,
			PhoneNumber: params.PhoneNumber,
			AvatarURL:   params.AvatarURL,
			Bio:         params.Bio,
			Address:     params.Address,
			Gender:      params.Gender,
			DateOfBirth: params.DateOfBirth,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		m.Profiles[userID] = p
		return p, nil
	}

	if params.FullName != nil {
		p.FullName = *params.FullName
	}
	if params.PhoneNumber != nil {
		p.PhoneNumber = params.PhoneNumber
	}
	if params.AvatarURL != nil {
		p.AvatarURL = params.AvatarURL
	}
	if params.Bio != nil {
		p.Bio = params.Bio
	}
	if params.Address != nil {
		p.Address = params.Address
	}
	if params.Gender != nil {
		p.Gender = params.Gender
	}
	if params.DateOfBirth != nil {
		p.DateOfBirth = params.DateOfBirth
	}
	p.UpdatedAt = now
	return p, nil
}

// HardDeleteByUserID removes a mock profile from memory.
func (m *MockUserProfileRepository) HardDeleteByUserID(ctx context.Context, userID string) error {
	delete(m.Profiles, userID)
	return nil
}
