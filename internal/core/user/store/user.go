// Package userdb contains user related CRUD functionality.
package userdb

import (
	"errors"
	"fmt"

	"github.com/dudakovict/gotify/internal/core/user"
	"github.com/dudakovict/gotify/platform/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

// Store manages the set of APIs for user database access.
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
func (s *Store) WithinTran(fn func(user.Storer) error) error {
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

// Create inserts a new user into the database.
func (s *Store) Create(usr user.User) error {
	const q = `
	INSERT INTO users
		(id, email, roles, verified, hashed_password, created_at, updated_at)
	VALUES
		(:id, :email, :roles, :verified, :hashed_password, :created_at, :updated_at)`

	if err := database.NamedExec(s.log, s.db, q, toDBUser(usr)); err != nil {
		if errors.Is(err, database.ErrUniqueViolation) {
			return user.ErrUniqueEmail
		}
		return fmt.Errorf("namedexec: %w", err)
	}

	return nil
}

// Update replaces a user document in the database.
func (s *Store) Update(usr user.User) error {
	const q = `
	UPDATE
		users
	SET
		email = :email,
		roles = :roles,
		verified = :verified,
		hashed_password = :hashed_password,
		updated_at = :updated_at
	WHERE
		id = :id`

	if err := database.NamedExec(s.log, s.db, q, toDBUser(usr)); err != nil {
		if errors.Is(err, database.ErrUniqueViolation) {
			return user.ErrUniqueEmail
		}
		return fmt.Errorf("namedexec: %w", err)
	}

	return nil
}

// Delete removes a user from the database.
func (s *Store) Delete(usr user.User) error {
	data := struct {
		ID uuid.UUID `db:"id"`
	}{
		ID: usr.ID,
	}

	const q = `
		DELETE FROM
			users
		WHERE
			id = :id`

	if err := database.NamedExec(s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexec: %w", err)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func (s *Store) Query(pageNumber int, rowsPerPage int) ([]user.User, error) {
	data := map[string]interface{}{
		"offset":        (pageNumber - 1) * rowsPerPage,
		"rows_per_page": rowsPerPage,
	}

	const q = `SELECT * FROM users`

	var dbUsrs []dbUser
	if err := database.NamedQuerySlice(s.log, s.db, q, data, &dbUsrs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreUserSlice(dbUsrs)
}

// QueryByID gets the specified user from the database by id.
func (s *Store) QueryByID(id uuid.UUID) (user.User, error) {
	data := struct {
		ID uuid.UUID `db:"id"`
	}{
		ID: id,
	}

	const q = `SELECT * FROM users WHERE id = :id`

	var dbUsr dbUser
	if err := database.NamedQueryStruct(s.log, s.db, q, data, &dbUsr); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return user.User{}, user.ErrNotFound
		}
		return user.User{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreUser(dbUsr)
}

// QueryByEmail gets the specified user from the database by email.
func (s *Store) QueryByEmail(email string) (user.User, error) {
	data := struct {
		Email string `db:"email"`
	}{
		Email: email,
	}

	const q = `SELECT * FROM users WHERE email = :email`

	var dbUsr dbUser
	if err := database.NamedQueryStruct(s.log, s.db, q, data, &dbUsr); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return user.User{}, user.ErrNotFound
		}
		return user.User{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreUser(dbUsr)
}

func (s *Store) QueryByTopicID(topicID uuid.UUID) ([]user.User, error) {
	data := struct {
		TopicID uuid.UUID `db:"topic_id"`
	}{
		TopicID: topicID,
	}

	const q = `
	SELECT u.id, u.email, u.roles, u.verified, u.hashed_password, u.created_at, u.updated_at FROM
		users u
	JOIN
		subscriptions s
	ON
		u.id = s.user_id
	WHERE
		s.topic_id = :topic_id`

	var dbUsrs []dbUser
	if err := database.NamedQuerySlice(s.log, s.db, q, data, &dbUsrs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreUserSlice(dbUsrs)
}
