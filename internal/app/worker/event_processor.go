package worker

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"

	"github.com/tdevilphan/quote-snap-golang/internal/core/usecase"
	queueinfra "github.com/tdevilphan/quote-snap-golang/internal/infra/queue/asynq"
)

// EventProcessor consumes tracking event tasks and persists them via the provided use case.
type EventProcessor struct {
	usecase *usecase.PersistEvent
	logger  *slog.Logger
}

// NewEventProcessor constructs an EventProcessor instance.
func NewEventProcessor(usecase *usecase.PersistEvent, logger *slog.Logger) *EventProcessor {
	return &EventProcessor{usecase: usecase, logger: logger.With("component", "event_processor")}
}

// Handler returns an Asynq handler function.
func (p *EventProcessor) Handler() asynq.Handler {
	return asynq.HandlerFunc(p.ProcessTask)
}

// ProcessTask persists the event contained in the task payload.
func (p *EventProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	if task.Type() != queueinfra.EventIngestTaskType {
		return errors.Errorf("unexpected task type: %s", task.Type())
	}

	event, err := queueinfra.DecodeEvent(task)
	if err != nil {
		p.logger.Warn("failed to decode event payload", "error", err)
		return errors.Wrap(err, "decode event payload")
	}

	if err := p.usecase.Execute(ctx, event); err != nil {
		p.logger.Error("failed to persist event", "event_id", event.ID, "error", err)
		return err
	}

	return nil
}
