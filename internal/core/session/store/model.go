package sessiondb

import (
	"time"

	"github.com/dudakovict/gotify/internal/core/session"
	"github.com/google/uuid"
)

type dbSession struct {
	ID           uuid.UUID `db:"id"`
	UserID       uuid.UUID `db:"user_id"`
	RefreshToken string    `db:"refresh_token"`
	UserAgent    string    `db:"user_agent"`
	ClientIP     string    `db:"client_ip"`
	IsBlocked    bool      `db:"is_blocked"`
	ExpiresAt    time.Time `db:"expires_at"`
	CreatedAt    time.Time `db:"created_at"`
}

func toDBSession(sessn session.Session) dbSession {
	return dbSession{
		ID:           sessn.ID,
		UserID:       sessn.UserID,
		RefreshToken: sessn.RefreshToken,
		UserAgent:    sessn.UserAgent,
		ClientIP:     sessn.ClientIP,
		IsBlocked:    sessn.IsBlocked,
		ExpiresAt:    sessn.ExpiresAt,
		CreatedAt:    sessn.CreatedAt.UTC(),
	}
}

func toCoreSession(dbSessn dbSession) session.Session {
	return session.Session{
		ID:           dbSessn.ID,
		UserID:       dbSessn.UserID,
		RefreshToken: dbSessn.RefreshToken,
		UserAgent:    dbSessn.UserAgent,
		ClientIP:     dbSessn.ClientIP,
		IsBlocked:    dbSessn.IsBlocked,
		ExpiresAt:    dbSessn.ExpiresAt,
		CreatedAt:    dbSessn.CreatedAt.In(time.Local),
	}
}
