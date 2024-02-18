// Package subscription provides a core business API.
package subscription

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound        = errors.New("subscription not found")
	ErrUniqueViolation = errors.New("subscription already exists")
)

// Storer interface declares the behavior this package needs to perists and
// retrieve data.
type Storer interface {
	WithinTran(fn func(s Storer) error) error
	Create(sub Subscription) error
	Delete(sub Subscription) error
	Query(pageNumber int, rowsPerPage int) ([]Subscription, error)
	QueryByID(id uuid.UUID) (Subscription, error)
}

// Core manages the set of APIs for subscription access.
type Core struct {
	log    *zerolog.Logger
	storer Storer
}

// NewCore constructs a subscription core API for use.
func NewCore(log *zerolog.Logger, storer Storer) *Core {
	return &Core{
		log:    log,
		storer: storer,
	}
}

// Create adds a new subscription to the system.
func (c *Core) Create(ns NewSubscription) (Subscription, error) {
	sub := Subscription{
		ID:      uuid.New(),
		TopicID: ns.TopicID,
		UserID:  ns.UserID,
	}

	if err := c.storer.Create(sub); err != nil {
		return Subscription{}, fmt.Errorf("create: %w", err)
	}

	return sub, nil
}

// Delete removes the specified subscription.
func (c *Core) Delete(sub Subscription) error {
	if err := c.storer.Delete(sub); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of existing subscriptions.
func (c *Core) Query(pageNumber int, rowsPerPage int) ([]Subscription, error) {
	subs, err := c.storer.Query(pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return subs, nil
}

// QueryByID finds the subscription by the specified ID.
func (c *Core) QueryByID(id uuid.UUID) (Subscription, error) {
	sub, err := c.storer.QueryByID(id)
	if err != nil {
		return Subscription{}, fmt.Errorf("query: %w", err)
	}

	return sub, nil
}
