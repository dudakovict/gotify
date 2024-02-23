// Package worker provides support for distributing and processing tasks.
package worker

import (
	"github.com/dudakovict/gotify/internal/core/notification"
	notificationdb "github.com/dudakovict/gotify/internal/core/notification/store"
	"github.com/dudakovict/gotify/internal/core/topic"
	topicdb "github.com/dudakovict/gotify/internal/core/topic/store"
	"github.com/dudakovict/gotify/internal/core/user"
	userdb "github.com/dudakovict/gotify/internal/core/user/store"
	"github.com/dudakovict/gotify/internal/core/verification"
	verificationdb "github.com/dudakovict/gotify/internal/core/verification/store"
	"github.com/dudakovict/gotify/pkg/mailer"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

// Config represents the configuration needed to create a Worker instance.
type Config struct {
	Log      *zerolog.Logger
	DB       *sqlx.DB
	RedisOpt asynq.RedisClientOpt
	Mailer   mailer.Mailer
}

// Worker contains configuration for task distribution and processing.
type Worker struct {
	TaskDistributor
	TaskProcessor
}

// New initializes and returns a new Worker instance.
func New(cfg Config) *Worker {
	usrCore := user.NewCore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB))
	vrfCore := verification.NewCore(cfg.Log, usrCore, verificationdb.NewStore(cfg.Log, cfg.DB))
	ntfCore := notification.NewCore(cfg.Log, notificationdb.NewStore(cfg.Log, cfg.DB))
	tpcCore := topic.NewCore(cfg.Log, topicdb.NewStore(cfg.Log, cfg.DB), uuid.New)

	return &Worker{
		TaskDistributor: newRedisTaskDistributor(cfg.Log, cfg.RedisOpt),
		TaskProcessor:   newRedisTaskProcessor(cfg.RedisOpt, usrCore, vrfCore, ntfCore, tpcCore, cfg.Mailer),
	}
}
