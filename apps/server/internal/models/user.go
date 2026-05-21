package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents an application user stored in the database.
type User struct {
	// ID is the unique identifier for the user.
	ID uuid.UUID
	// Email is the unique email address the user registered with.
	Email string
	// PasswordHash is the Argon2id hash of the user's password.
	PasswordHash string
	// CreatedAt is the time the user was created.
	CreatedAt time.Time
	// UpdatedAt is the time the user was last updated.
	UpdatedAt time.Time
}
