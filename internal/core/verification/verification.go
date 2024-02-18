// Package verification provides a core business API.
package verification

import (
	"errors"
	"fmt"

	"github.com/dudakovict/gotify/internal/core/user"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound = errors.New("verification not found")
)

// Storer interface declares the behavior this package needs to perists and
// retrieve data.
type Storer interface {
	WithinTran(fn func(s Storer) error) error
	Create(vrf Verificiation) error
	Update(vrf Verificiation) error
	QueryByID(id uuid.UUID) (Verificiation, error)
}

// Core manages the set of APIs for verification access.
type Core struct {
	log     *zerolog.Logger
	usrCore *user.Core
	storer  Storer
}

// NewCore constructs a verification core API for use.
func NewCore(log *zerolog.Logger, usrCore *user.Core, storer Storer) *Core {
	return &Core{
		log:     log,
		usrCore: usrCore,
		storer:  storer,
	}
}

func (c *Core) Verify(id uuid.UUID) (user.User, error) {
	var updUsr user.User
	tran := func(s Storer) error {
		vrf, err := s.QueryByID(id)
		if err != nil {
			return err
		}

		vrf.Used = true

		if err := s.Update(vrf); err != nil {
			return err
		}

		usr, err := c.usrCore.QueryByEmail(vrf.Email)
		if err != nil {
			return err
		}

		verified := true

		updUsr, err = c.usrCore.Update(usr, user.UpdateUser{
			Verified: &verified,
		})
		if err != nil {
			return err
		}

		return nil
	}

	if err := c.storer.WithinTran(tran); err != nil {
		return user.User{}, fmt.Errorf("tran: %w", err)
	}

	return updUsr, nil
}

// Create adds a new verification to the system.
func (c *Core) Create(nv NewVerification) (Verificiation, error) {
	vrf := Verificiation{
		ID:     uuid.New(),
		UserID: nv.UserID,
		Email:  nv.Email,
		Code:   nv.Code,
	}

	if err := c.storer.Create(vrf); err != nil {
		return Verificiation{}, fmt.Errorf("create: %w", err)
	}

	return vrf, nil
}

func (c *Core) Update(vrf Verificiation, used *bool) (Verificiation, error) {
	if used != nil {
		vrf.Used = *used
	}

	if err := c.storer.Update(vrf); err != nil {
		return Verificiation{}, fmt.Errorf("update: %w", err)
	}

	return vrf, nil
}

// QueryByID finds the verification by the specified ID.
func (c *Core) QueryByID(id uuid.UUID) (Verificiation, error) {
	vrf, err := c.storer.QueryByID(id)
	if err != nil {
		return Verificiation{}, fmt.Errorf("query: %w", err)
	}

	return vrf, nil
}
