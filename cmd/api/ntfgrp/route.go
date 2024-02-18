package ntfgrp

import (
	"github.com/dudakovict/gotify/internal/core/notification"
	notificationdb "github.com/dudakovict/gotify/internal/core/notification/store"
	"github.com/dudakovict/gotify/internal/core/user"
	userdb "github.com/dudakovict/gotify/internal/core/user/store"
	"github.com/dudakovict/gotify/internal/worker"
	"github.com/dudakovict/gotify/pkg/maker"
	"github.com/dudakovict/gotify/pkg/mid"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type Config struct {
	Log             *zerolog.Logger
	DB              *sqlx.DB
	Maker           maker.Maker
	TaskDistributor worker.TaskDistributor
}

func Routes(api fiber.Router, cfg Config) {
	usrCore := user.NewCore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB))
	ntfCore := notification.NewCore(cfg.Log, notificationdb.NewStore(cfg.Log, cfg.DB))

	hdl := new(ntfCore, cfg.TaskDistributor)

	ntfgrp := api.Group(
		"/notifications",
		mid.Authenticate(cfg.Maker, []string{user.RoleUser.Name()}),
		mid.GetUser(usrCore),
	)

	ntfgrp.Get("", hdl.query)
	ntfgrp.Post("", hdl.create)
	ntfgrp.Get("/:id", hdl.queryByID)
	ntfgrp.Delete("/:id", hdl.delete)
}

func errorResponse(err error) fiber.Map {
	return fiber.Map{
		"error": err.Error(),
	}
}
