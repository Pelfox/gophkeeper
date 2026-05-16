package protocol

import (
	"time"

	"github.com/google/uuid"
)

// ProtocolVault describes the vault type that can be sent over the protocol.
type ProtocolVault struct {
	// ID is vault's id.
	ID uuid.UUID `json:"id"`
	// Name is vault's name.
	Name string `json:"name"`
	// CreatedAt is a timestamp when this vault was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is a timestamp when this vault was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateVaultRequest describes the request for the vault creation.
type CreateVaultRequest struct {
	// Name is vault's name.
	Name string `json:"name" binding:"required,min=1,max=64"`
	// EncryptionSalt holds the raw bytes that were used for the encryption of
	// master key.
	EncryptionSalt []byte `json:"encryption_salt" binding:"required"`
	// EncryptionNonce holds the raw bytes that were used by the encryption
	// algorithm to encrypt master key.
	EncryptionNonce []byte `json:"encryption_nonce" binding:"required"`
	// EncryptedMasterKey holds an actual encrypted master key. This value is
	// safe to be stored in the database.
	EncryptedMasterKey []byte `json:"encrypted_master_key" binding:"required"`
	// EncryptionTimeCost describes amount of passes of the given
	// EncryptionMemoryCost.
	EncryptionTimeCost uint32 `json:"encryption_time_cost" binding:"required"`
	// EncryptionMemoryCost describes how much memory should be used.
	EncryptionMemoryCost uint32 `json:"encryption_memory_cost" binding:"required"`
	// EncryptionParallelism describes how much threads should be used.
	EncryptionParallelism uint32 `json:"encryption_parallelism" binding:"required"`
	// EncryptionKeySize describes the size of the returned byte slice.
	EncryptionKeySize uint32 `json:"encryption_key_size" binding:"required"`
}

// CreateVaultResponse describes the response after vault is created.
type CreateVaultResponse struct {
	ProtocolVault
}

// UpdateVaultRequest describes the request for updating the vault.
type UpdateVaultRequest struct {
	// Name is a new vault's name.
	Name *string `json:"name" binding:"omitempty,min=1,max=64"`
}

// UpdateVaultResponse describes the response for the vault update request.
type UpdateVaultResponse struct {
	ProtocolVault
}

// ListVaultsResponse describes the response for the list operation of vaults.
type ListVaultsResponse = []ProtocolVault
