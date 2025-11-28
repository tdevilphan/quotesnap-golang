package asynq

import (
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"

	"github.com/tdevilphan/quote-snap-golang/internal/core/domain"
)

const (
	// EventIngestTaskType identifies tasks that persist tracking events.
	EventIngestTaskType = "tracking:event:ingest"
)

// NewEventTask transforms a domain event into an Asynq task.
func NewEventTask(event domain.Event) (*asynq.Task, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, errors.Wrap(err, "marshal event payload")
	}
	return asynq.NewTask(EventIngestTaskType, payload, asynq.MaxRetry(5)), nil
}

// DecodeEvent recovers a domain event from an Asynq task payload.
func DecodeEvent(task *asynq.Task) (domain.Event, error) {
	var event domain.Event
	if err := json.Unmarshal(task.Payload(), &event); err != nil {
		return domain.Event{}, errors.Wrap(err, "unmarshal event payload")
	}
	return event, nil
}
