// Package user provides a core business API.
package user

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("user not found")
	ErrUniqueEmail           = errors.New("email is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrAuthorizationFailure  = errors.New("authorization failed")
)

// Storer interface declares the behavior this package needs to perists and
// retrieve data.
type Storer interface {
	WithinTran(fn func(s Storer) error) error
	Create(usr User) error
	Update(usr User) error
	Delete(usr User) error
	Query(pageNumber int, rowsPerPage int) ([]User, error)
	QueryByID(id uuid.UUID) (User, error)
	QueryByEmail(email string) (User, error)
	QueryByTopicID(topicID uuid.UUID) ([]User, error)
}

// Core manages the set of APIs for user access.
type Core struct {
	log    *zerolog.Logger
	storer Storer
}

// NewCore constructs a user core API for use.
func NewCore(log *zerolog.Logger, storer Storer) *Core {
	return &Core{
		log:    log,
		storer: storer,
	}
}

func (c *Core) CreateTx(nuTx NewUserTx) (User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(nuTx.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("generatefrompassword: %w", err)
	}

	usr := User{
		ID:             uuid.New(),
		Email:          nuTx.Email,
		Roles:          nuTx.Roles,
		HashedPassword: hashedPassword,
	}

	tran := func(s Storer) error {
		if err := s.Create(usr); err != nil {
			return fmt.Errorf("create: %w", err)
		}

		return nuTx.AfterCreate(usr)
	}

	if err := c.storer.WithinTran(tran); err != nil {
		return User{}, fmt.Errorf("tran: %w", err)
	}

	return usr, nil
}

// Create adds a new user to the system.
func (c *Core) Create(nu NewUser) (User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("generatefrompassword: %w", err)
	}

	usr := User{
		ID:             uuid.New(),
		Email:          nu.Email,
		Roles:          nu.Roles,
		HashedPassword: hashedPassword,
	}

	if err := c.storer.Create(usr); err != nil {
		return User{}, fmt.Errorf("create: %w", err)
	}

	return usr, nil
}

// Update modifies information about a user.
func (c *Core) Update(usr User, uu UpdateUser) (User, error) {
	if uu.Email != nil {
		usr.Email = *uu.Email
	}

	if uu.Roles != nil {
		usr.Roles = uu.Roles
	}

	if uu.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*uu.Password), bcrypt.DefaultCost)
		if err != nil {
			return User{}, fmt.Errorf("generatefrompassword: %w", err)
		}
		usr.HashedPassword = hashedPassword
	}

	if uu.Verified != nil {
		usr.Verified = *uu.Verified
	}

	usr.UpdatedAt = time.Now()

	if err := c.storer.Update(usr); err != nil {
		return User{}, fmt.Errorf("update: %w", err)
	}

	return usr, nil
}

// Delete removes the specified user.
func (c *Core) Delete(usr User) error {
	if err := c.storer.Delete(usr); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of existing users.
func (c *Core) Query(pageNumber int, rowsPerPage int) ([]User, error) {
	users, err := c.storer.Query(pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return users, nil
}

// QueryByID finds the user by the specified ID.
func (c *Core) QueryByID(id uuid.UUID) (User, error) {
	user, err := c.storer.QueryByID(id)
	if err != nil {
		return User{}, fmt.Errorf("query: %w", err)
	}

	return user, nil
}

// QueryByEmail finds the user by a specified user email.
func (c *Core) QueryByEmail(email string) (User, error) {
	user, err := c.storer.QueryByEmail(email)
	if err != nil {
		return User{}, fmt.Errorf("query: email[%s]: %w", email, err)
	}

	return user, nil
}

func (c *Core) QueryByTopicID(topicID uuid.UUID) ([]User, error) {
	users, err := c.storer.QueryByTopicID(topicID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return users, nil
}

func (c *Core) Authenticate(email string, password string) (User, error) {
	usr, err := c.QueryByEmail(email)
	if err != nil {
		return User{}, err
	}

	if err := bcrypt.CompareHashAndPassword(usr.HashedPassword, []byte(password)); err != nil {
		return User{}, fmt.Errorf("comparehashandpassword: %w", ErrAuthenticationFailure)
	}

	return usr, nil
}
