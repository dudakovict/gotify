package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dudakovict/gotify/internal/core/notification"
	"github.com/dudakovict/gotify/internal/core/topic"
	"github.com/dudakovict/gotify/internal/core/user"
	"github.com/dudakovict/gotify/internal/core/verification"
	"github.com/dudakovict/gotify/pkg/mailer"
	"github.com/dudakovict/gotify/pkg/util"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

// TaskProcessor defines the interface for task processing.
type TaskProcessor interface {
	Start() error
	Shutdown()
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
	ProcessTaskSendNotification(ctx context.Context, task *asynq.Task) error
}

// RedisTaskProcessor processes tasks using Redis.
type RedisTaskProcessor struct {
	server  *asynq.Server
	usrCore *user.Core
	vrfCore *verification.Core
	ntfCore *notification.Core
	tpcCore *topic.Core
	mailer  mailer.Mailer
}

// newRedisTaskProcessor initializes a new RedisTaskProcessor instance.
func newRedisTaskProcessor(
	redisOpt asynq.RedisClientOpt,
	usrCore *user.Core,
	vrfCore *verification.Core,
	ntfCore *notification.Core,
	tpcCore *topic.Core,
	mailer mailer.Mailer,
) TaskProcessor {
	logger := NewLogger()
	redis.SetLogger(logger)

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().
					Err(err).
					Str("type", task.Type()).
					Bytes("payload", task.Payload()).
					Msg("process task failed")
			}),
			Logger: logger,
		},
	)

	return &RedisTaskProcessor{
		server:  server,
		usrCore: usrCore,
		vrfCore: vrfCore,
		ntfCore: ntfCore,
		tpcCore: tpcCore,
		mailer:  mailer,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)
	mux.HandleFunc(TaskSendNotification, processor.ProcessTaskSendNotification)

	return processor.server.Start(mux)
}

func (processor *RedisTaskProcessor) Shutdown() {
	processor.server.Shutdown()
}

type PayloadSendVerifyEmail struct {
	Email string `json:"email"`
}

// ProcessTaskSendVerifyEmail processes the task to send a verification email.
func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	usr, err := processor.usrCore.QueryByEmail(payload.Email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	verifyEmail, err := processor.vrfCore.Create(verification.NewVerification{
		UserID: usr.ID,
		Email:  usr.Email,
		Code:   util.RandomString(32),
	})
	if err != nil {
		return fmt.Errorf("failed to create verify email: %w", err)
	}

	subject := "Welcome!"
	verifyURL := fmt.Sprintf("http://localhost:3000/api/v1/verify?id=%s&code=%s", verifyEmail.ID.String(), verifyEmail.Code)
	content := fmt.Sprintf(`Hello %s,<br/>
	Thank you for registering with us!<br/>
	Please <a href="%s">click here</a> to verify your email address.<br/>
	`, usr.Email, verifyURL)
	to := []string{usr.Email}

	if err := processor.mailer.SendEmail(subject, content, to, nil, nil, nil); err != nil {
		return fmt.Errorf("failed to send verify email: %w", err)
	}
	log.Info().
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("email", usr.Email).
		Msg("processed task")

	return nil
}

type PayloadSendNotification struct {
	NotificationID uuid.UUID `json:"notification_id"`
}

func (processor *RedisTaskProcessor) ProcessTaskSendNotification(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendNotification
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	ntf, err := processor.ntfCore.QueryByID(payload.NotificationID)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}

	tpc, err := processor.tpcCore.QueryByID(ntf.TopicID)
	if err != nil {
		return fmt.Errorf("failed to get topic: %w", err)
	}

	usrs, err := processor.usrCore.QueryByTopicID(ntf.TopicID)
	if err != nil {
		return fmt.Errorf("failed to get users subscribed to topic: [%s]: %w", ntf.TopicID, err)
	}

	to := make([]string, len(usrs))
	for i, usr := range usrs {
		to[i] = usr.Email
	}

	if err := processor.mailer.SendEmail(tpc.Name, ntf.Message, to, nil, nil, nil); err != nil {
		return fmt.Errorf("failed to send a notification: %w", err)
	}
	log.Info().
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("email", strings.Join(to, ",")).
		Msg("processed task")

	return nil
}
