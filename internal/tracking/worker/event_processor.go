package worker

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"

	"quotesnap/internal/tracking/domain"
	"quotesnap/internal/tracking/queue"
	"quotesnap/internal/tracking/repository"
)

// EventProcessor consumes tracking event tasks and persists them.
type EventProcessor struct {
	repo   repository.EventRepository
	logger *slog.Logger
}

// NewEventProcessor constructs an EventProcessor instance.
func NewEventProcessor(repo repository.EventRepository, logger *slog.Logger) *EventProcessor {
	return &EventProcessor{repo: repo, logger: logger.With("component", "event_processor")}
}

// Handler returns an Asynq handler function.
func (p *EventProcessor) Handler() asynq.Handler {
	return asynq.HandlerFunc(p.ProcessTask)
}

// ProcessTask persists the event contained in the task payload.
func (p *EventProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	if task.Type() != queue.EventIngestTaskType {
		return errors.Errorf("unexpected task type: %s", task.Type())
	}

	var event domain.Event
	if err := json.Unmarshal(task.Payload(), &event); err != nil {
		p.logger.Warn("failed to decode event payload", "error", err)
		return errors.Wrap(err, "decode event payload")
	}

	if err := p.repo.Persist(ctx, event); err != nil {
		p.logger.Error("failed to persist event", "event_id", event.ID, "error", err)
		return errors.Wrap(err, "persist event")
	}

	return nil
}
