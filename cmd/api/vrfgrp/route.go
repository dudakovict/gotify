package vrfgrp

import (
	"github.com/dudakovict/gotify/internal/core/user"
	userdb "github.com/dudakovict/gotify/internal/core/user/store"
	"github.com/dudakovict/gotify/internal/core/verification"
	verificationdb "github.com/dudakovict/gotify/internal/core/verification/store"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type Config struct {
	Log *zerolog.Logger
	DB  *sqlx.DB
}

func Routes(api fiber.Router, cfg Config) {
	usrCore := user.NewCore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB))
	vrfCore := verification.NewCore(cfg.Log, usrCore, verificationdb.NewStore(cfg.Log, cfg.DB))

	hdl := new(vrfCore)

	api.Get("/verify", hdl.verify)
}

func errorResponse(err error) fiber.Map {
	return fiber.Map{
		"error": err.Error(),
	}
}
