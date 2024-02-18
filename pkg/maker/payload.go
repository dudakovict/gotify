package maker

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

// Payload represents the information stored within a token.
type Payload struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Roles     []string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// NewPayload creates a new Payload.
func NewPayload(userID uuid.UUID, roles []string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := Payload{
		ID:        tokenID,
		UserID:    userID,
		Roles:     roles,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(duration),
	}

	return &payload, nil
}

// Valid checks if the token is still valid.
func (p *Payload) Valid() error {
	if time.Now().After(p.ExpiresAt) {
		return ErrExpiredToken
	}

	return nil
}
