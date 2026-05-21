package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// TestUsersRepositoryCreate verifies that Create stores a user and returns
// generated metadata.
func TestUsersRepositoryCreate(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewUsersRepository(pool)

	user, err := repo.Create(ctx, CreateUserInput{
		Email:        "user@example.com",
		PasswordHash: "hash-v1",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if user.ID == uuid.Nil {
		t.Fatal("expected generated user ID")
	}
	if user.Email != "user@example.com" || user.PasswordHash != "hash-v1" {
		t.Fatalf("unexpected user: %#v", user)
	}
}

// TestUsersRepositoryCreateDuplicateEmail verifies that Create maps unique
// email violations to ErrDuplicateUser.
func TestUsersRepositoryCreateDuplicateEmail(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewUsersRepository(pool)

	createTestUser(t, ctx, pool, "user@example.com")

	_, err := repo.Create(ctx, CreateUserInput{
		Email:        "user@example.com",
		PasswordHash: "hash-v2",
	})
	if !errors.Is(err, ErrDuplicateUser) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrDuplicateUser)
	}
}

// TestUsersRepositoryFindByEmail verifies that FindByEmail returns the user
// matching the supplied email address.
func TestUsersRepositoryFindByEmail(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewUsersRepository(pool)

	userID := createTestUser(t, ctx, pool, "user@example.com")

	user, err := repo.FindByEmail(ctx, "user@example.com")
	if err != nil {
		t.Fatalf("FindByEmail returned error: %v", err)
	}
	if user.ID != userID || user.Email != "user@example.com" {
		t.Fatalf("unexpected user: %#v", user)
	}
}

// TestUsersRepositoryFindByEmailNotFound verifies that FindByEmail returns
// ErrUserNotFound when no user has the supplied email.
func TestUsersRepositoryFindByEmailNotFound(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewUsersRepository(pool)

	_, err := repo.FindByEmail(ctx, "missing@example.com")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrUserNotFound)
	}
}

// TestUsersRepositoryGetByID verifies that GetByID returns the user matching
// the supplied identifier.
func TestUsersRepositoryGetByID(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewUsersRepository(pool)

	userID := createTestUser(t, ctx, pool, "user@example.com")

	user, err := repo.GetByID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if user.ID != userID || user.Email != "user@example.com" {
		t.Fatalf("unexpected user: %#v", user)
	}
}

// TestUsersRepositoryGetByIDNotFound verifies that GetByID returns
// ErrUserNotFound for an unknown identifier.
func TestUsersRepositoryGetByIDNotFound(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewUsersRepository(pool)

	_, err := repo.GetByID(ctx, uuid.New())
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrUserNotFound)
	}
}

// TestUsersRepositoryUpdate verifies that Update persists mutable user fields
// and returns the updated row.
func TestUsersRepositoryUpdate(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewUsersRepository(pool)

	userID := createTestUser(t, ctx, pool, "user@example.com")
	newEmail := "renamed@example.com"
	newPasswordHash := "hash-v2"

	user, err := repo.Update(ctx, userID, UpdateUserInput{
		Email:        &newEmail,
		PasswordHash: &newPasswordHash,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if user.ID != userID ||
		user.Email != newEmail ||
		user.PasswordHash != newPasswordHash {
		t.Fatalf("unexpected user: %#v", user)
	}
}

// TestUsersRepositoryUpdateEmailOnly verifies that Update can modify just the
// user's email address.
func TestUsersRepositoryUpdateEmailOnly(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewUsersRepository(pool)

	userID := createTestUser(t, ctx, pool, "user@example.com")
	newEmail := "renamed@example.com"

	user, err := repo.Update(ctx, userID, UpdateUserInput{Email: &newEmail})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if user.Email != newEmail || user.PasswordHash != "password-hash" {
		t.Fatalf("unexpected user: %#v", user)
	}
}

// TestUsersRepositoryUpdatePasswordHashOnly verifies that Update can modify
// just the user's password hash.
func TestUsersRepositoryUpdatePasswordHashOnly(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewUsersRepository(pool)

	userID := createTestUser(t, ctx, pool, "user@example.com")
	newPasswordHash := "hash-v2"

	user, err := repo.Update(ctx, userID, UpdateUserInput{
		PasswordHash: &newPasswordHash,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if user.Email != "user@example.com" || user.PasswordHash != newPasswordHash {
		t.Fatalf("unexpected user: %#v", user)
	}
}

// TestUsersRepositoryUpdateNoFields verifies that Update with no mutable
// fields still refreshes metadata and returns the existing user.
func TestUsersRepositoryUpdateNoFields(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewUsersRepository(pool)

	userID := createTestUser(t, ctx, pool, "user@example.com")

	user, err := repo.Update(ctx, userID, UpdateUserInput{})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if user.Email != "user@example.com" || user.PasswordHash != "password-hash" {
		t.Fatalf("unexpected user: %#v", user)
	}
}

// TestUsersRepositoryUpdateNotFound verifies that Update returns
// ErrUserNotFound for an unknown identifier.
func TestUsersRepositoryUpdateNotFound(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewUsersRepository(pool)

	newEmail := "renamed@example.com"

	_, err := repo.Update(ctx, uuid.New(), UpdateUserInput{Email: &newEmail})
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrUserNotFound)
	}
}
