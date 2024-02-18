package userdb

import (
	"fmt"
	"time"

	"github.com/dudakovict/gotify/internal/core/user"
	"github.com/google/uuid"

	"github.com/lib/pq"
)

type dbUser struct {
	ID             uuid.UUID      `db:"id"`
	Email          string         `db:"email"`
	Roles          pq.StringArray `db:"roles"`
	HashedPassword []byte         `db:"hashed_password"`
	Verified       bool           `db:"verified"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

func toDBUser(usr user.User) dbUser {
	roles := make([]string, len(usr.Roles))
	for i, role := range usr.Roles {
		roles[i] = role.Name()
	}

	return dbUser{
		ID:             usr.ID,
		Email:          usr.Email,
		Roles:          roles,
		HashedPassword: usr.HashedPassword,
		Verified:       usr.Verified,
		CreatedAt:      usr.CreatedAt.UTC(),
		UpdatedAt:      usr.UpdatedAt.UTC(),
	}
}

func toCoreUser(dbUsr dbUser) (user.User, error) {
	roles := make([]user.Role, len(dbUsr.Roles))
	for i, value := range dbUsr.Roles {
		var err error
		roles[i], err = user.ParseRole(value)
		if err != nil {
			return user.User{}, fmt.Errorf("parse role: %w", err)
		}
	}

	usr := user.User{
		ID:             dbUsr.ID,
		Email:          dbUsr.Email,
		Roles:          roles,
		HashedPassword: dbUsr.HashedPassword,
		Verified:       dbUsr.Verified,
		CreatedAt:      dbUsr.CreatedAt.In(time.Local),
		UpdatedAt:      dbUsr.UpdatedAt.In(time.Local),
	}

	return usr, nil
}

func toCoreUserSlice(dbUsrs []dbUser) ([]user.User, error) {
	usrs := make([]user.User, len(dbUsrs))

	for i, dbUsr := range dbUsrs {
		var err error
		usrs[i], err = toCoreUser(dbUsr)
		if err != nil {
			return nil, err
		}
	}

	return usrs, nil
}
