package user

import (
	"time"

	"github.com/google/uuid"
)

// User represents information about an individual user.
type User struct {
	ID             uuid.UUID
	Email          string
	Roles          []Role
	Verified       bool
	HashedPassword []byte
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewUser contains information needed to create a new user.
type NewUser struct {
	Email    string
	Roles    []Role
	Password string
}

// UpdateUser contains information needed to update a user.
type UpdateUser struct {
	Email    *string
	Roles    []Role
	Verified *bool
	Password *string
}

// NewUserTx contains information needed to create a new user
// in a transaction.
type NewUserTx struct {
	NewUser
	AfterCreate func(user User) error
}
