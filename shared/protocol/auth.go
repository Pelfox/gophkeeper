package protocol

import (
	"time"
)

// SessionCreatedResponse describes a response from the server when new session
// is created.
type SessionCreatedResponse struct {
	// User is the object of a user that owns this session.
	User ProtocolUser `json:"user"`
	// AccessToken is a JWT token that is used to access the API.
	AccessToken string `json:"access_token"`
	// ExpiresAt is a timestamp when this session expires.
	ExpiresAt time.Time `json:"expires_at"`
}

// RegisterRequest describes the request that is sent when user registers.
type RegisterRequest struct {
	// Email is user's email.
	Email string `json:"email" binding:"required,email"`
	// Password is user's password.
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// RegisterResponse describes the response after user is registered. Basically
// it's a new created session.
type RegisterResponse struct {
	SessionCreatedResponse
}

// LoginRequest describes the request that is sent when user logins into the
// account.
type LoginRequest struct {
	// Email is user's email.
	Email string `json:"email" binding:"required,email"`
	// Password is user's password.
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// LoginResponse describes the response when user successfully logged into the
// account. Basically it's a new created session.
type LoginResponse struct {
	SessionCreatedResponse
}
