// Package verificationdb contains verification related CRUD functionality.
package verificationdb

import (
	"errors"
	"fmt"

	"github.com/dudakovict/gotify/internal/core/verification"
	"github.com/dudakovict/gotify/platform/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

// Store manages the set of APIs for verification database access.
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
func (s *Store) WithinTran(fn func(verification.Storer) error) error {
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

// Create inserts a new verification into the database.
func (s *Store) Create(vrf verification.Verificiation) error {
	const q = `
	INSERT INTO verifications
		(id, user_id, email, code, used, created_at, expired_at)
	VALUES
		(:id, :user_id, :email, :code, :used, :created_at, :expired_at)`

	if err := database.NamedExec(s.log, s.db, q, toDBVerification(vrf)); err != nil {
		return fmt.Errorf("namedexec: %w", err)
	}

	return nil
}

func (s *Store) Update(vrf verification.Verificiation) error {
	const q = `
	UPDATE
		verifications
	SET
		used = :used
	WHERE
		id = :id
		AND code = :code
		AND used = FALSE
		AND expired_at > now()`

	if err := database.NamedExec(s.log, s.db, q, toDBVerification(vrf)); err != nil {
		return fmt.Errorf("namedexec: %w", err)
	}

	return nil
}

func (s *Store) QueryByID(id uuid.UUID) (verification.Verificiation, error) {
	const q = `SELECT * FROM verifications WHERE id = :id`

	data := struct {
		ID uuid.UUID `db:"id"`
	}{
		ID: id,
	}

	var dbVrf dbVerification
	if err := database.NamedQueryStruct(s.log, s.db, q, data, &dbVrf); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return verification.Verificiation{}, verification.ErrNotFound
		}
		return verification.Verificiation{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreVerification(dbVrf), nil
}
