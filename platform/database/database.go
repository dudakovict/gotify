// Package database provides support for accessing the database.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

// lib/pq errorCodeNames
// https://github.com/lib/pq/blob/master/error.go#L178
const (
	uniqueViolation = "23505"
	undefinedTable  = "42P01"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound        = sql.ErrNoRows
	ErrUniqueViolation = errors.New("duplicated entry")
	ErrUndefinedTable  = errors.New("undefined table")
)

type Config struct {
	User         string
	Password     string
	Host         string
	Name         string
	MaxIdleConns int
	MaxOpenConns int
	DisableTLS   bool
}

// Open knows how to open a database connection based on the configuration.
func Open(cfg Config) (*sqlx.DB, error) {
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:  "postgres",
		User:    url.UserPassword(cfg.User, cfg.Password),
		Host:    cfg.Host,
		Path:    cfg.Name,
		RawPath: q.Encode(),
	}

	db, err := sqlx.Open("pgx", u.String())
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)

	return db, nil
}

// Ping returns nil if it can successfully talk to the database.
// It returns a non-nil error otherwise.
func Ping(ctx context.Context, db *sqlx.DB) error {

	// If the user doesn't give us a deadline set 1 second.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer cancel()
	}

	var pingError error
	for attempts := 1; ; attempts++ {
		pingError = db.Ping()
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Microsecond)
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	// Make sure we didn't timeout or be cancelled.
	if ctx.Err() != nil {
		return ctx.Err()
	}

	const q = `SELECT true`
	var tmp bool
	return db.QueryRowContext(ctx, q).Scan(&tmp)
}

// WithinTran runs passed function and do commit/rollback at the end.
func WithinTran(log *zerolog.Logger, db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	log.Info().Msg("begin tran")

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tran: %w", err)
	}

	// We can defer the rollback since the code checks if the transaction
	// has already been committed.
	defer func() {
		if err := tx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				return
			}
			log.Error().Err(err).Msg("unable to rollback tran")
		}
		log.Info().Msg("rollback tran")
	}()

	if err := fn(tx); err != nil {
		if pqerr, ok := err.(*pgconn.PgError); ok && pqerr.Code == uniqueViolation {
			return ErrUniqueViolation
		}
		return fmt.Errorf("exec tran: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tran: %w", err)
	}
	log.Info().Msg("commit tran")

	return nil
}

func NamedExec(log *zerolog.Logger, db sqlx.Ext, query string, data any) (err error) {
	defer func() {
		if err != nil {
			log.Error().Err(err).Str("query", query).Msg("database.NamedExec")
		}
	}()

	if _, err := sqlx.NamedExec(db, query, data); err != nil {
		if pqerr, ok := err.(*pgconn.PgError); ok {
			switch pqerr.Code {
			case undefinedTable:
				return ErrUndefinedTable
			case uniqueViolation:
				return ErrUniqueViolation
			}
		}
		return err
	}

	return nil
}

func NamedQueryStruct(log *zerolog.Logger, db sqlx.Ext, query string, data any, dest any) (err error) {
	defer func() {
		if err != nil {
			log.Error().Err(err).Str("query", query).Msg("database.NamedQueryStruct")
		}
	}()

	var rows *sqlx.Rows
	rows, err = sqlx.NamedQuery(db, query, data)
	if err != nil {
		if pqerr, ok := err.(*pgconn.PgError); ok && pqerr.Code == undefinedTable {
			return ErrUndefinedTable
		}
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return ErrNotFound
	}

	if err := rows.StructScan(dest); err != nil {
		return err
	}

	return nil
}

func NamedQuerySlice[T any](log *zerolog.Logger, db sqlx.Ext, query string, data any, dest *[]T) (err error) {
	defer func() {
		if err != nil {
			log.Error().Err(err).Str("query", query).Msg("database.NamedQuerySlice")
		}
	}()

	var rows *sqlx.Rows
	rows, err = sqlx.NamedQuery(db, query, data)
	if err != nil {
		if pqerr, ok := err.(*pgconn.PgError); ok && pqerr.Code == undefinedTable {
			return ErrUndefinedTable
		}
		return err
	}
	defer rows.Close()

	var slice []T
	for rows.Next() {
		v := new(T)
		if err := rows.StructScan(v); err != nil {
			return err
		}

		slice = append(slice, *v)
	}

	*dest = slice

	return nil
}
