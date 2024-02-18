// Package subscriptiondb contains subscription related CRUD functionality.
package subscriptiondb

import (
	"errors"
	"fmt"

	"github.com/dudakovict/gotify/internal/core/subscription"
	"github.com/dudakovict/gotify/platform/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

// Store manages the set of APIs for subscription database access.
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
func (s *Store) WithinTran(fn func(subscription.Storer) error) error {
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

// Create inserts a new subscription into the database.
func (s *Store) Create(sub subscription.Subscription) error {
	const q = `
	INSERT INTO subscriptions
		(id, topic_id, user_id, created_at)
	VALUES
		(:id, :topic_id, :user_id, :created_at)`

	if err := database.NamedExec(s.log, s.db, q, toDBSubscription(sub)); err != nil {
		if errors.Is(err, database.ErrUniqueViolation) {
			return subscription.ErrUniqueViolation
		}
		return fmt.Errorf("namedexec: %w", err)
	}

	return nil
}

// Delete removes a subscription from the database.
func (s *Store) Delete(sub subscription.Subscription) error {
	data := struct {
		ID string `db:"id"`
	}{
		ID: sub.ID.String(),
	}

	const q = `
	DELETE FROM
		subscriptions
	WHERE
		id = :id`

	if err := database.NamedExec(s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexec: %w", err)
	}

	return nil
}

// Query retrieves a list of existing subscriptions from the database.
func (s *Store) Query(pageNumber int, rowsPerPage int) ([]subscription.Subscription, error) {
	data := map[string]interface{}{
		"offset":        (pageNumber - 1) * rowsPerPage,
		"rows_per_page": rowsPerPage,
	}

	const q = `SELECT * FROM subscriptions`

	var dbSubs []dbSubscription
	if err := database.NamedQuerySlice(s.log, s.db, q, data, &dbSubs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreSubscriptionSlice(dbSubs), nil
}

// QueryByID gets the specified subscription from the database by id.
func (s *Store) QueryByID(id uuid.UUID) (subscription.Subscription, error) {
	data := struct {
		ID uuid.UUID `db:"id"`
	}{
		ID: id,
	}

	const q = `SELECT * FROM subscriptions WHERE id = :id`

	var dbSub dbSubscription
	if err := database.NamedQueryStruct(s.log, s.db, q, data, &dbSub); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return subscription.Subscription{}, subscription.ErrNotFound
		}
		return subscription.Subscription{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreSubscription(dbSub), nil
}
