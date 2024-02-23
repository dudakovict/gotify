package topicgrp

import (
	"github.com/dudakovict/gotify/internal/core/topic"
	topicdb "github.com/dudakovict/gotify/internal/core/topic/store"
	"github.com/dudakovict/gotify/internal/core/user"
	userdb "github.com/dudakovict/gotify/internal/core/user/store"
	"github.com/dudakovict/gotify/pkg/maker"
	"github.com/dudakovict/gotify/pkg/mid"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	tpcCore := topic.NewCore(cfg.Log, topicdb.NewStore(cfg.Log, cfg.DB), uuid.New)

	hdl := new(tpcCore)

	tpcgrp := api.Group(
		"/topics",
		mid.Authenticate(cfg.Maker, []string{user.RoleUser.Name()}),
		mid.GetUser(usrCore),
	)

	tpcgrp.Get("", hdl.query)
	tpcgrp.Post("", hdl.create)
	tpcgrp.Get("/:id", hdl.queryByID)
}

func errorResponse(err error) fiber.Map {
	return fiber.Map{
		"error": err.Error(),
	}
}
