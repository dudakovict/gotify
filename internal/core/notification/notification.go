// Package notification provides a core business API.
package notification

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound = errors.New("notification not found")
)

// Storer interface declares the behavior this package needs to perists and
// retrieve data.
type Storer interface {
	WithinTran(fn func(s Storer) error) error
	Create(ntf Notification) error
	Delete(ntf Notification) error
	Query(pageNumber int, rowsPerPage int, topicID uuid.UUID) ([]Notification, error)
	QueryByID(id uuid.UUID) (Notification, error)
}

// Core manages the set of APIs for notification access.
type Core struct {
	log    *zerolog.Logger
	storer Storer
}

// NewCore constructs a notification core API for use.
func NewCore(log *zerolog.Logger, storer Storer) *Core {
	return &Core{
		log:    log,
		storer: storer,
	}
}

func (c *Core) CreateTx(nnTx NewNotificationTx) (Notification, error) {
	ntf := Notification{
		ID:      uuid.New(),
		TopicID: nnTx.TopicID,
		Message: nnTx.Message,
	}

	tran := func(s Storer) error {
		if err := s.Create(ntf); err != nil {
			return fmt.Errorf("create: %w", err)
		}

		return nnTx.AfterCreate(ntf)
	}

	if err := c.storer.WithinTran(tran); err != nil {
		return Notification{}, fmt.Errorf("tran: %w", err)
	}

	return ntf, nil
}

// Create adds a new notification to the system.
func (c *Core) Create(nn NewNotification) (Notification, error) {
	ntf := Notification{
		ID:      uuid.New(),
		TopicID: nn.TopicID,
		Message: nn.Message,
	}

	if err := c.storer.Create(ntf); err != nil {
		return Notification{}, fmt.Errorf("create: %w", err)
	}

	return ntf, nil
}

// Delete removes the specified notification.
func (c *Core) Delete(ntf Notification) error {
	if err := c.storer.Delete(ntf); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of notifications.
func (c *Core) Query(pageNumber int, rowsPerPage int, topicID uuid.UUID) ([]Notification, error) {
	ntfs, err := c.storer.Query(pageNumber, rowsPerPage, topicID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return ntfs, nil
}

// QueryByID finds the notification by the specified ID.
func (c *Core) QueryByID(id uuid.UUID) (Notification, error) {
	ntf, err := c.storer.QueryByID(id)
	if err != nil {
		return Notification{}, fmt.Errorf("query: %w", err)
	}

	return ntf, nil
}
