package repository

import (
	"context"
	"testing"
)

func TestMockUserProfileRepository(t *testing.T) {
	repo := NewMockUserProfileRepository()
	ctx := context.Background()

	// 1. Get non-existent
	_, err := repo.GetByUserID(ctx, "usr_1")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}

	// 2. Upsert
	name := "Alice"
	phone := "+628123456789"
	p, err := repo.Upsert(ctx, "usr_1", UpdateProfileParams{
		FullName:    &name,
		PhoneNumber: &phone,
	})
	if err != nil {
		t.Fatalf("unexpected upsert error: %v", err)
	}
	if p.FullName != "Alice" || *p.PhoneNumber != "+628123456789" {
		t.Errorf("unexpected profile data: %+v", p)
	}

	// 3. Get existing
	p2, err := repo.GetByUserID(ctx, "usr_1")
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if p2.FullName != "Alice" {
		t.Errorf("expected name Alice, got: %s", p2.FullName)
	}

	// 4. Hard Delete
	if err := repo.HardDeleteByUserID(ctx, "usr_1"); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	_, err = repo.GetByUserID(ctx, "usr_1")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after deletion, got: %v", err)
	}
}
