package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	// EventMetadataLimit enforces an upper bound on metadata payload sizes (32KB).
	EventMetadataLimit = 32 * 1024
)

// Event captures the canonical representation of a tracking event within the domain.
type Event struct {
	ID         uuid.UUID       `json:"id"`
	Name       string          `json:"name"`
	UserID     string          `json:"user_id"`
	Source     string          `json:"source"`
	Metadata   json.RawMessage `json:"metadata"`
	OccurredAt time.Time       `json:"occurred_at"`
	ReceivedAt time.Time       `json:"received_at"`
}

// NewEvent validates input parameters and returns a fully populated Event aggregate.
func NewEvent(name, userID, source string, metadata json.RawMessage, occurredAt time.Time) (Event, error) {
	if name == "" {
		return Event{}, errors.New("name is required")
	}
	if userID == "" {
		return Event{}, errors.New("user_id is required")
	}
	if source == "" {
		return Event{}, errors.New("source is required")
	}

	if len(metadata) == 0 {
		metadata = json.RawMessage("{}")
	}
	if len(metadata) > EventMetadataLimit {
		return Event{}, errors.Errorf("metadata must be <= %d bytes", EventMetadataLimit)
	}

	occurred := occurredAt.UTC()
	if occurred.IsZero() {
		occurred = time.Now().UTC()
	}

	received := time.Now().UTC()

	return Event{
		ID:         uuid.New(),
		Name:       name,
		UserID:     userID,
		Source:     source,
		Metadata:   metadata,
		OccurredAt: occurred,
		ReceivedAt: received,
	}, nil
}
