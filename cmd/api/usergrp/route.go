package usergrp

import (
	"github.com/dudakovict/gotify/internal/core/user"
	userdb "github.com/dudakovict/gotify/internal/core/user/store"
	"github.com/dudakovict/gotify/pkg/maker"
	"github.com/dudakovict/gotify/pkg/mid"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type Config struct {
	Log   *zerolog.Logger
	DB    *sqlx.DB
	Maker maker.Maker
}

func Routes(api fiber.Router, cfg Config) {
	usrCore := user.NewCore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB))

	hdl := new(usrCore)

	usrgrp := api.Group(
		"/users",
		mid.Authenticate(cfg.Maker, []string{user.RoleUser.Name()}),
		mid.GetUser(usrCore),
	)

	usrgrp.Get("", hdl.query)
	usrgrp.Post("", hdl.create)
	usrgrp.Get("/:id", hdl.queryByID)
	usrgrp.Put("/:id", hdl.update)
	usrgrp.Delete("/:id", hdl.delete)
}

func errorResponse(err error) fiber.Map {
	return fiber.Map{
		"error": err.Error(),
	}
}

func parseRoles(roles []string) ([]user.Role, error) {
	userRoles := make([]user.Role, len(roles))

	for i, role := range roles {
		userRole, err := user.ParseRole(role)
		if err != nil {
			return nil, err
		}
		userRoles[i] = userRole
	}

	return userRoles, nil
}
