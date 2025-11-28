package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"

	"quotesnap/internal/tracking/domain"
	queuepkg "quotesnap/internal/tracking/queue"
)

var ErrValidation = errors.New("validation error")

// TaskEnqueuer abstracts queue interactions for SOLID-friendly testing.
type TaskEnqueuer interface {
	EnqueueContext(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

// EventService orchestrates the ingestion workflow for tracking events.
type EventService struct {
	queue TaskEnqueuer
}

// NewEventService wires queue dependencies into the service.
func NewEventService(queue TaskEnqueuer) *EventService {
	return &EventService{queue: queue}
}

// CreateEventInput captures transport-level payloads living outside the domain model.
type CreateEventInput struct {
	Name       string    `json:"name"`
	UserID     string    `json:"user_id"`
	Source     string    `json:"source"`
	Metadata   []byte    `json:"metadata"`
	OccurredAt time.Time `json:"occurred_at"`
	Queue      string    `json:"-"`
}

func (input CreateEventInput) validate() error {
	if input.Name == "" {
		return validationError("name is required")
	}
	if input.UserID == "" {
		return validationError("user_id is required")
	}
	if input.Queue == "" {
		return validationError("queue name is required")
	}
	return nil
}

// IngestEvent validates, enriches, and asynchronously persists tracking events.
func (s *EventService) IngestEvent(ctx context.Context, input CreateEventInput) (domain.Event, error) {
	if err := input.validate(); err != nil {
		return domain.Event{}, err
	}

	metadata := json.RawMessage(input.Metadata)
	event, err := domain.NewEvent(input.Name, input.UserID, input.Source, metadata, input.OccurredAt)
	if err != nil {
		return domain.Event{}, validationError(err.Error())
	}

	task, err := queuepkg.NewEventIngestTask(event)
	if err != nil {
		return domain.Event{}, errors.Wrap(err, "create queue task")
	}

	taskCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	if _, err := s.queue.EnqueueContext(taskCtx, task, asynq.Queue(input.Queue)); err != nil {
		return domain.Event{}, errors.Wrap(err, "enqueue event task")
	}

	return event, nil
}

func validationError(message string) error {
	return errors.Wrap(ErrValidation, message)
}
