package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
)

// TaskDistributor defines the interface for task distribution.
type TaskDistributor interface {
	DistributeTaskSendVerifyEmail(
		ctx context.Context,
		payload *PayloadSendVerifyEmail,
		opts ...asynq.Option,
	) error
	DistributeTaskSendNotification(
		ctx context.Context,
		payload *PayloadSendNotification,
		opts ...asynq.Option,
	) error
}

// RedisTaskDistributor distributes tasks using Redis.
type RedisTaskDistributor struct {
	log    *zerolog.Logger
	client *asynq.Client
}

// newRedisTaskDistributor initializes a new RedisTaskDistributor instance.
func newRedisTaskDistributor(log *zerolog.Logger, redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		log:    log,
		client: client,
	}
}

const TaskSendVerifyEmail = "task:send_verify_email"

// DistributeTaskSendVerifyEmail enqueues a task to send a verification email.
func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(ctx context.Context, payload *PayloadSendVerifyEmail, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	distributor.log.Info().
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("queue", info.Queue).
		Int("max_retry", info.MaxRetry).
		Msg("enqueued task")

	return nil
}

const TaskSendNotification = "task:send_notification"

func (distributor *RedisTaskDistributor) DistributeTaskSendNotification(ctx context.Context, payload *PayloadSendNotification, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendNotification, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	distributor.log.Info().
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("queue", info.Queue).
		Int("max_retry", info.MaxRetry).
		Msg("enqueued task")

	return nil
}
