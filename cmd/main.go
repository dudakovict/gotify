package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/dudakovict/gotify/cmd/api"
	_ "github.com/dudakovict/gotify/docs/swagger"
	"github.com/dudakovict/gotify/internal/worker"
	"github.com/dudakovict/gotify/pkg/config"
	"github.com/dudakovict/gotify/pkg/mailer"
	"github.com/dudakovict/gotify/pkg/maker"
	"github.com/dudakovict/gotify/platform/database"
	"github.com/dudakovict/gotify/platform/logger"
	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

var build = "develop"

// @title Gotify API
// @version 1.0
// @description Auth & Notifications API.
// @host localhost:3000
// @BasePath /api/v1
func main() {
	log := logger.New()

	log.Info().
		Str("VERSION", build).
		Int("GOMAXPROCS", runtime.GOMAXPROCS(0))

	if err := run(log); err != nil {
		log.Error().Err(err)
		os.Exit(1)
	}
}

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func run(log *zerolog.Logger) error {
	cfg, err := config.Load(".")
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	db, err := database.Open(database.Config{
		User:         cfg.User,
		Password:     cfg.Password,
		Host:         cfg.Host,
		Name:         cfg.Name,
		MaxIdleConns: cfg.MaxIdleConns,
		MaxOpenConns: cfg.MaxOpenConns,
		DisableTLS:   cfg.DisableTLS,
	})
	if err != nil {
		return err
	}
	defer db.Close()

	if err := database.Ping(ctx, db); err != nil {
		return err
	}

	log.Info().Msg("connected to database")

	migration, err := migrate.New(cfg.MigrationURL, cfg.DatabaseURL)
	if err != nil {
		log.Error().Err(err)
		return err
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Info().Msg("db migrated successfully")

	maker, err := maker.NewPasetoMaker(cfg.TokenSymmetricKey)
	if err != nil {
		return err
	}

	redisOpt := asynq.RedisClientOpt{
		Addr: cfg.RedisHost,
	}

	mailer := mailer.NewGmailMailer(cfg.MailerName, cfg.EmailAddress, cfg.EmailPassword)

	worker := worker.New(worker.Config{
		Log:      log,
		DB:       db,
		RedisOpt: redisOpt,
		Mailer:   mailer,
	})

	app := fiber.New()

	app.Use(cors.New())
	app.Use(pprof.New())
	app.Use(recover.New())
	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: log,
	}))

	app.Get("/swagger/*", swagger.HandlerDefault)
	app.Get("/metrics", monitor.New(monitor.Config{
		Title: "Gotify Metrics",
	}))

	api.Routes(app, api.Config{
		Log:                  log,
		DB:                   db,
		Maker:                maker,
		TaskDistributor:      worker.TaskDistributor,
		AccessTokenDuration:  cfg.AccessTokenDuration,
		RefreshTokenDuration: cfg.RefreshTokenDuration,
	})

	wg, ctx := errgroup.WithContext(ctx)

	processTasks(log, ctx, wg, worker)
	listen(log, ctx, wg, app, cfg.HTTPServerAddress)

	if err := wg.Wait(); err != nil {
		return err
	}

	return nil
}

func listen(log *zerolog.Logger, ctx context.Context, wg *errgroup.Group, app *fiber.App, addr string) {
	wg.Go(func() error {
		if err := app.Listen(addr); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			log.Error().Err(err).Msg("HTTP server failed to serve")
		}
		return nil
	})

	wg.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("gracefully shutdown HTTP server")

		if err := app.Shutdown(); err != nil {
			log.Error().Err(err).Msg("failed to shutdown HTTP server")
			return err
		}

		log.Info().Msg("HTTP server is shutdown")
		return nil
	})
}

func processTasks(log *zerolog.Logger, ctx context.Context, wg *errgroup.Group, worker *worker.Worker) {
	log.Info().Msg("start task processor")
	if err := worker.TaskProcessor.Start(); err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}

	wg.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("gracefully shutting down task processor")

		worker.TaskProcessor.Shutdown()
		log.Info().Msg("task processor is shutdown")

		return nil
	})
}
