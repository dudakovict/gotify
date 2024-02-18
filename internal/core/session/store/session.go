// Package sessiondb contains session related CRUD functionality.
package sessiondb

import (
	"errors"
	"fmt"

	"github.com/dudakovict/gotify/internal/core/session"
	"github.com/dudakovict/gotify/platform/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

// Store manages the set of APIs for session database access.
type Store struct {
	log *zerolog.Logger
	db  sqlx.Ext
}

// NewStore constructs the api for data access.
func NewStore(log *zerolog.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Create inserts a new session into the database.
func (s *Store) Create(sessn session.Session) error {
	const q = `
	INSERT INTO sessions
		(id, user_id, refresh_token, user_agent, client_ip, is_blocked, expires_at)
	VALUES
		(:id, :user_id, :refresh_token, :user_agent, :client_ip, :is_blocked, :expires_at)`

	if err := database.NamedExec(s.log, s.db, q, toDBSession(sessn)); err != nil {
		return fmt.Errorf("creating session: %w", err)
	}

	return nil
}

// QueryByID gets the specified session from the database by id.
func (s *Store) QueryByID(id uuid.UUID) (session.Session, error) {
	data := struct {
		ID uuid.UUID `db:"id"`
	}{
		ID: id,
	}

	const q = `SELECT * FROM sessions WHERE id = :id`

	var dbSessn dbSession
	if err := database.NamedQueryStruct(s.log, s.db, q, data, &dbSessn); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return session.Session{}, session.ErrNotFound
		}
		return session.Session{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreSession(dbSessn), nil
}
