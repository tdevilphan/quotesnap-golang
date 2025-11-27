package queue

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
)

// NewClient centralises creation of an Asynq client with proper instrumentation hooks.
func NewClient(addr, password string) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{Addr: addr, Password: password})
}

// NewServer builds an Asynq server tuned for bursty event ingestion workloads.
func NewServer(addr, password, queue string, concurrency int, logger *slog.Logger) *asynq.Server {
	redisOpt := asynq.RedisClientOpt{Addr: addr, Password: password}
	config := asynq.Config{
		Concurrency: concurrency,
		Queues:      map[string]int{queue: 1},
		ErrorHandler: asynq.ErrorHandlerFunc(func(_ context.Context, task *asynq.Task, err error) {
			logger.Error("asynq task failed", "type", task.Type(), "error", err)
		}),
	}
	return asynq.NewServer(redisOpt, config)
}
