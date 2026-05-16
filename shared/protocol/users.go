package protocol

import (
	"time"

	"github.com/google/uuid"
)

// ProtocolUser describes the user object that can be sent over the protocol.
type ProtocolUser struct {
	// ID is user's ID.
	ID uuid.UUID `json:"id"`
	// Email is user's email.
	Email string `json:"email"`
	// CreatedAt is a timestamp when this user was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is a timestamp when this user was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest describes the request that is sent when user registers.
type CreateUserRequest struct {
	// Email is user's email.
	Email string `json:"email" validate:"required,email"`
	// Password is user's password.
	Password string `json:"password" validate:"required,min=6,max=64"`
}

// CreateUserResponse describes the response after user is registered.
// Basically it's a new created session.
type CreateUserResponse struct {
	SessionCreatedResponse
}
