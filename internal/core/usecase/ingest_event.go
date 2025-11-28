package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"

	"github.com/tdevilphan/quote-snap-golang/internal/core/domain"
)

var (
	// ErrValidation indicates that the provided input cannot be processed.
	ErrValidation = errors.New("validation error")
)

// EventQueue defines the outbound dependency required to dispatch events for asynchronous processing.
type EventQueue interface {
	Enqueue(ctx context.Context, event domain.Event) error
}

// IngestEvent orchestrates validation and dispatch of tracking events.
type IngestEvent struct {
	queue EventQueue
}

// NewIngestEvent constructs an IngestEvent use case instance.
func NewIngestEvent(queue EventQueue) *IngestEvent {
	return &IngestEvent{queue: queue}
}

// IngestEventInput models the information required to create a new event.
type IngestEventInput struct {
	Name       string
	UserID     string
	Source     string
	Metadata   json.RawMessage
	OccurredAt time.Time
}

// Execute validates the input, constructs a domain event, and enqueues it for processing.
func (uc *IngestEvent) Execute(ctx context.Context, input IngestEventInput) (domain.Event, error) {
	event, err := domain.NewEvent(input.Name, input.UserID, input.Source, input.Metadata, input.OccurredAt)
	if err != nil {
		return domain.Event{}, validationError(err.Error())
	}

	if err := uc.queue.Enqueue(ctx, event); err != nil {
		return domain.Event{}, errors.Wrap(err, "enqueue event task")
	}

	return event, nil
}

func validationError(message string) error {
	return errors.Wrap(ErrValidation, message)
}
