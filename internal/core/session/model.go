package session

import (
	"time"

	"github.com/google/uuid"
)

// Session represents information about an individual session.
type Session struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	RefreshToken string
	UserAgent    string
	ClientIP     string
	IsBlocked    bool
	ExpiresAt    time.Time
}

// NewSession contains information needed to create a new session.
type NewSession struct {
	UserID       uuid.UUID
	RefreshToken string
	UserAgent    string
	ClientIP     string
	IsBlocked    bool
	ExpiresAt    time.Time
}
