package models

import (
	"time"

	"github.com/google/uuid"
)

// VaultItem represents an encrypted item stored in a vault.
type VaultItem struct {
	// ID is the unique identifier for the vault item.
	ID uuid.UUID
	// VaultID identifies the vault that contains the item.
	VaultID uuid.UUID
	// KeySalt is the salt used to derive the item encryption key.
	KeySalt []byte
	// ItemNonce is the nonce used to encrypt the item.
	ItemNonce []byte
	// Ciphertext contains the encrypted item data.
	Ciphertext []byte
	// CreatedAt is the time the item was created.
	CreatedAt time.Time
	// UpdatedAt is the time the item was last updated.
	UpdatedAt time.Time
}
