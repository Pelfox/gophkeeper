package models

import (
	"time"

	"github.com/google/uuid"
)

// Keyring represents a user's encrypted access key for a vault.
type Keyring struct {
	// UserID identifies the user who owns this keyring.
	UserID uuid.UUID
	// VaultID identifies the vault this keyring unlocks.
	VaultID uuid.UUID
	// EncryptionSalt is the salt used to derive the key-encryption key.
	EncryptionSalt []byte
	// EncryptionNonce is the nonce used to encrypt the master key.
	EncryptionNonce []byte
	// EncryptedMasterKey is the vault master key encrypted for the user.
	EncryptedMasterKey []byte
	// EncryptionTimeCost is the number of iterations used to derive the
	// encryption key.
	EncryptionTimeCost uint32
	// EncryptionMemoryCost is the amount of memory used to derive the
	// encryption key.
	EncryptionMemoryCost uint32
	// EncryptionParallelism is the parallelism factor used for key derivation.
	EncryptionParallelism uint32
	// EncryptionKeySize is the size, in bytes, of the derived encryption key.
	EncryptionKeySize uint32
	// CreatedAt is the time the keyring was created.
	CreatedAt time.Time
	// UpdatedAt is the time the keyring was last updated.
	UpdatedAt time.Time
}
