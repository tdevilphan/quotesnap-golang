package queue

import (
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"

	"github.com/tdevilphan/quote-snap-golang/internal/tracking/domain"
)

const (
	// EventIngestTaskType identifies tasks that persist tracking events.
	EventIngestTaskType = "tracking:event:ingest"
)

// NewEventIngestTask transforms a domain event into an Asynq task.
func NewEventIngestTask(event domain.Event) (*asynq.Task, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, errors.Wrap(err, "marshal event payload")
	}
	return asynq.NewTask(EventIngestTaskType, payload, asynq.MaxRetry(5)), nil
}
