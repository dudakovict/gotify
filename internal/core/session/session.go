// Package session provides a core business API.
package session

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound = errors.New("session not found")
)

// Storer interface declares the behavior this package needs to perist and
// retrieve data.
type Storer interface {
	Create(sessn Session) error
	QueryByID(id uuid.UUID) (Session, error)
}

// Core manages the set of APIs for session access.
type Core struct {
	log    *zerolog.Logger
	storer Storer
}

// NewCore constructs a session core API for use.
func NewCore(log *zerolog.Logger, storer Storer) *Core {
	return &Core{
		log:    log,
		storer: storer,
	}
}

// Create adds a new session to the system.
func (c *Core) Create(ns NewSession) (Session, error) {
	sessn := Session{
		ID:           uuid.New(),
		UserID:       ns.UserID,
		RefreshToken: ns.RefreshToken,
		UserAgent:    ns.UserAgent,
		ClientIP:     ns.ClientIP,
		IsBlocked:    ns.IsBlocked,
		ExpiresAt:    ns.ExpiresAt,
	}

	if err := c.storer.Create(sessn); err != nil {
		return Session{}, fmt.Errorf("create: %w", err)
	}

	return sessn, nil
}

// QueryByID finds the session by the specified ID.
func (c *Core) QueryByID(id uuid.UUID) (Session, error) {
	sessn, err := c.storer.QueryByID(id)
	if err != nil {
		return Session{}, fmt.Errorf("query: %w", err)
	}

	return sessn, nil
}
