// Package topicdb contains topic related CRUD functionality.
package topicdb

import (
	"errors"
	"fmt"

	"github.com/dudakovict/gotify/internal/core/topic"
	"github.com/dudakovict/gotify/platform/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

// Store manages the set of APIs for topic database access.
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
func (s *Store) WithinTran(fn func(topic.Storer) error) error {
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

// Create inserts a new topic into the database.
func (s *Store) Create(tpc topic.Topic) error {
	const q = `
	INSERT INTO topics
		(id, name, created_at)
	VALUES
		(:id, :name, :created_at)`

	if err := database.NamedExec(s.log, s.db, q, toDBTopic(tpc)); err != nil {
		if errors.Is(err, database.ErrUniqueViolation) {
			return topic.ErrUniqueName
		}
		return fmt.Errorf("namedexec: %w", err)
	}

	return nil
}

// Query retrieves a list of existing topics from the database.
func (s *Store) Query(pageNumber int, rowsPerPage int) ([]topic.Topic, error) {
	data := map[string]interface{}{
		"offset":        (pageNumber - 1) * rowsPerPage,
		"rows_per_page": rowsPerPage,
	}

	const q = `SELECT * FROM topics`

	var dbTpcs []dbTopic
	if err := database.NamedQuerySlice(s.log, s.db, q, data, &dbTpcs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreTopicSlice(dbTpcs), nil
}

// QueryByID gets the specified topic from the database by id.
func (s *Store) QueryByID(id uuid.UUID) (topic.Topic, error) {
	data := struct {
		ID uuid.UUID `db:"id"`
	}{
		ID: id,
	}

	const q = `SELECT * FROM topics WHERE id = :id`

	var dbTpc dbTopic
	if err := database.NamedQueryStruct(s.log, s.db, q, data, &dbTpc); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return topic.Topic{}, topic.ErrNotFound
		}
		return topic.Topic{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreTopic(dbTpc), nil
}
