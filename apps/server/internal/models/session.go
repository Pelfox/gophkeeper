package models

import (
	"time"

	"github.com/google/uuid"
)

// Session describes a single session.
type Session struct {
	// ID is a unique identifier for the session.
	ID uuid.UUID
	// UserID is the ID of the owner of this session.
	UserID uuid.UUID
	// AccessToken is an access token of the session.
	AccessToken string
	// ExpiresAt is a timestamp when this session expires.
	ExpiresAt time.Time
	// CreatedAt is a timestamp when this session was created.
	CreatedAt time.Time
}
