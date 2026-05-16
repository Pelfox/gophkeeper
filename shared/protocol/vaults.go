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
