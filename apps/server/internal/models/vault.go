package models

import (
	"time"

	"github.com/google/uuid"
)

// Vault represents a secure container for items owned by a specific user.
type Vault struct {
	// ID is the unique identifier for the vault.
	ID uuid.UUID
	// OwnerID identifies the user who owns the vault.
	OwnerID uuid.UUID
	// Name is the human-readable name of the vault.
	Name string
	// CreatedAt is the time the vault was created.
	CreatedAt time.Time
	// UpdatedAt is the time the vault was last updated.
	UpdatedAt time.Time
}
