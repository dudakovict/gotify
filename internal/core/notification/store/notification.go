// Package notificationdb contains notification related CRUD functionality.
package notificationdb

import (
	"errors"
	"fmt"

	"github.com/dudakovict/gotify/internal/core/notification"
	"github.com/dudakovict/gotify/platform/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

// Store manages the set of APIs for notification database access.
type Store struct {
	log    *zerolog.Logger
	db     sqlx.Ext
	inTran bool
}

// NewStore constructs the api for data access.
func NewStore(log *zerolog.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// WithinTran runs passed function and do commit/rollback at the end.
func (s *Store) WithinTran(fn func(notification.Storer) error) error {
	if s.inTran {
		return fn(s)
	}

	f := func(tx *sqlx.Tx) error {
		s := &Store{
			log:    s.log,
			db:     tx,
			inTran: true,
		}
		return fn(s)
	}

	return database.WithinTran(s.log, s.db.(*sqlx.DB), f)
}

// Create inserts a new notification into the database.
func (s *Store) Create(ntf notification.Notification) error {
	const q = `
	INSERT INTO notifications 
		(id, topic_id, message, created_at)
	VALUES 
		(:id, :topic_id, :message, :created_at)`

	if err := database.NamedExec(s.log, s.db, q, toDBNotification(ntf)); err != nil {
		return fmt.Errorf("namedexec: %w", err)
	}

	return nil
}

// Delete removes a notification from the database.
func (s *Store) Delete(ntf notification.Notification) error {
	data := struct {
		ID string `db:"id"`
	}{
		ID: ntf.ID.String(),
	}

	const q = `
	DELETE FROM
		notifications
	WHERE
		id = :id`

	if err := database.NamedExec(s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexec: %w", err)
	}

	return nil
}

// Query retrieves a list of notification from the database.
func (s *Store) Query(pageNumber int, rowsPerPage int, topicID uuid.UUID) ([]notification.Notification, error) {
	data := map[string]interface{}{
		"offset":        (pageNumber - 1) * rowsPerPage,
		"rows_per_page": rowsPerPage,
		"topic_id":      topicID.String(),
	}

	const q = `SELECT * FROM notifications`

	var dbNtfs []dbNotification
	if err := database.NamedQuerySlice(s.log, s.db, q, data, &dbNtfs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreNotificationSlice(dbNtfs), nil
}

// QueryByID gets the specified notification from the database by id.
func (s *Store) QueryByID(id uuid.UUID) (notification.Notification, error) {
	data := struct {
		ID uuid.UUID `db:"id"`
	}{
		ID: id,
	}

	const q = `SELECT * FROM notifications WHERE id = :id`

	var dbNtf dbNotification
	if err := database.NamedQueryStruct(s.log, s.db, q, data, &dbNtf); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return notification.Notification{}, notification.ErrNotFound
		}
		return notification.Notification{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreNotification(dbNtf), nil
}
