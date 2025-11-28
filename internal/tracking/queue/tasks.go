package queue

import (
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"

	"quotesnap/internal/tracking/domain"
)

const (
	// EventIngestTaskType identifies tasks responsible for persisting tracking events.
	EventIngestTaskType = "tracking:event:ingest"
)

// NewEventIngestTask transforms a tracking event into an Asynq task payload.
func NewEventIngestTask(event domain.Event) (*asynq.Task, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, errors.Wrap(err, "marshal event payload")
	}
	return asynq.NewTask(EventIngestTaskType, payload, asynq.MaxRetry(5)), nil
}

// DecodeEvent reconstructs an event from an Asynq task payload.
func DecodeEvent(task *asynq.Task) (domain.Event, error) {
	var event domain.Event
	if err := json.Unmarshal(task.Payload(), &event); err != nil {
		return domain.Event{}, errors.Wrap(err, "unmarshal event payload")
	}
	return event, nil
}
