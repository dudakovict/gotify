// Package topic provides a core business API.
package topic

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound   = errors.New("topic not found")
	ErrUniqueName = errors.New("topic already exists")
)

// Storer interface declares the behavior this package needs to perists and
// retrieve data.
type Storer interface {
	WithinTran(fn func(s Storer) error) error
	Create(topic Topic) error
	Query(pageNumber int, rowsPerPage int) ([]Topic, error)
	QueryByID(id uuid.UUID) (Topic, error)
}

// Core manages the set of APIs for user access.
type Core struct {
	log    *zerolog.Logger
	storer Storer
}

// NewCore constructs a user core API for use.
func NewCore(log *zerolog.Logger, storer Storer) *Core {
	return &Core{
		log:    log,
		storer: storer,
	}
}

// Create adds a new user to the system.
func (c *Core) Create(nt NewTopic) (Topic, error) {
	topic := Topic{
		ID:   uuid.New(),
		Name: nt.Name,
	}

	if err := c.storer.Create(topic); err != nil {
		return Topic{}, fmt.Errorf("create: %w", err)
	}

	return topic, nil
}

// Query retrieves a list of existing topics.
func (c *Core) Query(pageNumber int, rowsPerPage int) ([]Topic, error) {
	topics, err := c.storer.Query(pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return topics, nil
}

// QueryByID finds the topic by the specified ID.
func (c *Core) QueryByID(id uuid.UUID) (Topic, error) {
	topic, err := c.storer.QueryByID(id)
	if err != nil {
		return Topic{}, fmt.Errorf("query: %w", err)
	}

	return topic, nil
}
