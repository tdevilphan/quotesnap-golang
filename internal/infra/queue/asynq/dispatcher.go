package asynq

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"

	"quotesnap/internal/core/domain"
	"quotesnap/internal/core/usecase"
)

// Dispatcher converts domain events into Asynq tasks and enqueues them for processing.
type Dispatcher struct {
	client *asynq.Client
	queue  string
}

// NewDispatcher constructs a new Dispatcher instance.
func NewDispatcher(client *asynq.Client, queue string) *Dispatcher {
	return &Dispatcher{client: client, queue: queue}
}

// Enqueue pushes the event onto the configured Asynq queue.
func (d *Dispatcher) Enqueue(ctx context.Context, event domain.Event) error {
	task, err := NewEventTask(event)
	if err != nil {
		return err
	}
	if _, err := d.client.EnqueueContext(ctx, task, asynq.Queue(d.queue)); err != nil {
		return errors.Wrap(err, "enqueue event task")
	}
	return nil
}

// Ensure Dispatcher satisfies the EventQueue dependency.
var _ usecase.EventQueue = (*Dispatcher)(nil)
