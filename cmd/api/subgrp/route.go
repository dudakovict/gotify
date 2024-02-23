package subgrp

import (
	"github.com/dudakovict/gotify/internal/core/subscription"
	subscriptiondb "github.com/dudakovict/gotify/internal/core/subscription/store"
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
	subCore := subscription.NewCore(cfg.Log, subscriptiondb.NewStore(cfg.Log, cfg.DB))
	tpcCore := topic.NewCore(cfg.Log, topicdb.NewStore(cfg.Log, cfg.DB), uuid.New)

	hdl := new(subCore, tpcCore)

	subgrp := api.Group(
		"/subscriptions",
		mid.Authenticate(cfg.Maker, []string{user.RoleUser.Name()}),
		mid.GetUser(usrCore),
	)

	subgrp.Get("", hdl.query)
	subgrp.Post("", hdl.create)
	subgrp.Get("/:id", hdl.queryByID)
	subgrp.Delete("/:id", hdl.delete)
}

func errorResponse(err error) fiber.Map {
	return fiber.Map{
		"error": err.Error(),
	}
}
