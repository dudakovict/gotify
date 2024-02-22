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
	CreatedAt    time.Time
}
