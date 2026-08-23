package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Sall-lah/store_user/internal/client/order"
	"github.com/Sall-lah/store_user/internal/kafka"
	"github.com/Sall-lah/store_user/internal/repository"
)

// UserService defines the business logic contract for user profile and account operations.
// Why: Encapsulates transactional boundaries, cross-service orchestrations, and data validation rules.
type UserService interface {
	GetProfile(ctx context.Context, userID string) (*repository.UserProfile, error)
	UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*repository.UserProfile, error)
	DeleteAccount(ctx context.Context, userID string, req DeleteAccountRequest) error
}

// UserServiceImpl implements UserService with persistence, gRPC, and messaging dependencies.
type UserServiceImpl struct {
	repo          repository.UserProfileRepository
	orderClient   order.Client
	kafkaProducer kafka.Producer
	kafkaTopic    string
}

// NewUserService constructs a new UserServiceImpl with required collaborators.
// Why: Injects decoupled repository, gRPC client, and event publisher dependencies.
func NewUserService(
	repo repository.UserProfileRepository,
	orderClient order.Client,
	kafkaProducer kafka.Producer,
	kafkaTopic string,
) *UserServiceImpl {
	if kafkaTopic == "" {
		kafkaTopic = "user.events"
	}
	return &UserServiceImpl{
		repo:          repo,
		orderClient:   orderClient,
		kafkaProducer: kafkaProducer,
		kafkaTopic:    kafkaTopic,
	}
}

// GetProfile retrieves a user's personal profile by their unique identifier.
// Why: Returns existing profile data or auto-initializes a clean default profile for newly registered users.
func (s *UserServiceImpl) GetProfile(ctx context.Context, userID string) (*repository.UserProfile, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}

	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// Auto-initialize default profile for new user
			defaultName := "User"
			return s.repo.Upsert(ctx, userID, repository.UpdateProfileParams{
				FullName: &defaultName,
			})
		}
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
	}

	return profile, nil
}

// UpdateProfile validates incoming fields and updates the user's profile attributes.
// Why: Enforces field sanity constraints (e.g. max string lengths) before persisting mutations.
func (s *UserServiceImpl) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*repository.UserProfile, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}

	if req.FullName != nil && strings.TrimSpace(*req.FullName) == "" {
		return nil, fmt.Errorf("%w: fullName cannot be empty", ErrValidationFailed)
	}
	if req.Bio != nil && len(*req.Bio) > 500 {
		return nil, fmt.Errorf("%w: bio exceeds 500 characters limit", ErrValidationFailed)
	}

	params := repository.UpdateProfileParams{
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
		AvatarURL:   req.AvatarURL,
		Bio:         req.Bio,
		Address:     req.Address,
		Gender:      req.Gender,
		DateOfBirth: req.DateOfBirth,
	}

	profile, err := s.repo.Upsert(ctx, userID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return profile, nil
}

// DeleteAccount orchestrates account deletion with synchronous order verification and async Kafka notification.
// Why: Prevents orphaned shipments via gRPC pre-flight, deletes profile PII, and notifies auth service to revoke sessions.
func (s *UserServiceImpl) DeleteAccount(ctx context.Context, userID string, req DeleteAccountRequest) error {
	if strings.TrimSpace(userID) == "" {
		return ErrInvalidUserID
	}

	// 1. Synchronous Pre-flight Check against store_order
	if s.orderClient != nil {
		orderRes, err := s.orderClient.CheckActiveOrders(ctx, userID)
		if err != nil {
			log.Printf("[UserService] Order service pre-flight check failed for user %s: %v", userID, err)
			return ErrOrderServiceUnavailable
		}

		if orderRes != nil && orderRes.GetHasActiveOrders() {
			return &ActiveOrdersConflictError{
				Count:        orderRes.GetActiveOrderCount(),
				ActiveOrders: orderRes.GetActiveOrders(),
				Message:      orderRes.GetMessage(),
			}
		}
	}

	// 2. Hard-delete profile data from database
	if err := s.repo.HardDeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user profile: %w", err)
	}

	// 3. Publish asynchronous user.deleted domain event to Kafka
	if s.kafkaProducer != nil {
		reason := req.Reason
		if strings.TrimSpace(reason) == "" {
			reason = "user_requested_deletion"
		}

		if err := s.kafkaProducer.PublishUserDeleted(ctx, s.kafkaTopic, userID, reason); err != nil {
			log.Printf("[UserService] Warning: Failed to publish user.deleted event for user %s: %v", userID, err)
			return ErrKafkaPublishFailed
		}
	}

	return nil
}
