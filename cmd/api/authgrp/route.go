package authgrp

import (
	"time"

	"github.com/dudakovict/gotify/internal/core/session"
	sessiondb "github.com/dudakovict/gotify/internal/core/session/store"
	"github.com/dudakovict/gotify/internal/core/user"
	userdb "github.com/dudakovict/gotify/internal/core/user/store"
	"github.com/dudakovict/gotify/internal/worker"
	"github.com/dudakovict/gotify/pkg/maker"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type Config struct {
	Log                  *zerolog.Logger
	DB                   *sqlx.DB
	Maker                maker.Maker
	TaskDistributor      worker.TaskDistributor
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

func Routes(api fiber.Router, cfg Config) {
	usrCore := user.NewCore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB))
	sessnCore := session.NewCore(cfg.Log, sessiondb.NewStore(cfg.Log, cfg.DB))

	hdl := new(usrCore, sessnCore, cfg.Maker, cfg.TaskDistributor, cfg.AccessTokenDuration, cfg.RefreshTokenDuration)

	api.Post("/register", hdl.register)
	api.Post("/login", hdl.login)
	api.Post("/token", hdl.token)
}

func errorResponse(err error) fiber.Map {
	return fiber.Map{
		"error": err.Error(),
	}
}
