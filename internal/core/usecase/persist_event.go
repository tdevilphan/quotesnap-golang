package usecase

import (
	"context"

	"github.com/pkg/errors"

	"github.com/tdevilphan/quote-snap-golang/internal/core/domain"
)

// EventRepository defines persistence operations required by the domain.
type EventRepository interface {
	Persist(ctx context.Context, event domain.Event) error
}

// PersistEvent coordinates persisting events to durable storage.
type PersistEvent struct {
	repo EventRepository
}

// NewPersistEvent constructs a PersistEvent use case instance.
func NewPersistEvent(repo EventRepository) *PersistEvent {
	return &PersistEvent{repo: repo}
}

// Execute stores the provided event using the underlying repository.
func (uc *PersistEvent) Execute(ctx context.Context, event domain.Event) error {
	if err := uc.repo.Persist(ctx, event); err != nil {
		return errors.Wrap(err, "persist event")
	}
	return nil
}
