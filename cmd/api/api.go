// Package api binds the crud domain set of routes into the specified app.
package api

import (
	"time"

	"github.com/dudakovict/gotify/cmd/api/authgrp"
	"github.com/dudakovict/gotify/cmd/api/ntfgrp"
	"github.com/dudakovict/gotify/cmd/api/subgrp"
	"github.com/dudakovict/gotify/cmd/api/topicgrp"
	"github.com/dudakovict/gotify/cmd/api/usergrp"
	"github.com/dudakovict/gotify/cmd/api/vrfgrp"
	"github.com/dudakovict/gotify/internal/worker"
	"github.com/dudakovict/gotify/pkg/maker"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

// Config contains the configuration for setting up the API routes.
type Config struct {
	Log                  *zerolog.Logger
	DB                   *sqlx.DB
	Maker                maker.Maker
	TaskDistributor      worker.TaskDistributor
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

// Routes sets up the API routes for the application.
func Routes(app *fiber.App, cfg Config) {
	api := app.Group("/api/v1")

	authgrp.Routes(api, authgrp.Config{
		Log:                  cfg.Log,
		DB:                   cfg.DB,
		Maker:                cfg.Maker,
		TaskDistributor:      cfg.TaskDistributor,
		AccessTokenDuration:  cfg.AccessTokenDuration,
		RefreshTokenDuration: cfg.RefreshTokenDuration,
	})

	ntfgrp.Routes(api, ntfgrp.Config{
		Log:             cfg.Log,
		DB:              cfg.DB,
		Maker:           cfg.Maker,
		TaskDistributor: cfg.TaskDistributor,
	})

	subgrp.Routes(api, subgrp.Config{
		Log:   cfg.Log,
		DB:    cfg.DB,
		Maker: cfg.Maker,
	})

	topicgrp.Routes(api, topicgrp.Config{
		Log:   cfg.Log,
		DB:    cfg.DB,
		Maker: cfg.Maker,
	})

	usergrp.Routes(api, usergrp.Config{
		Log:   cfg.Log,
		DB:    cfg.DB,
		Maker: cfg.Maker,
	})

	vrfgrp.Routes(api, vrfgrp.Config{
		Log: cfg.Log,
		DB:  cfg.DB,
	})
}
