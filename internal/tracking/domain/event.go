package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// Event captures an immutable tracking event emitted by clients.
type Event struct {
	ID         uuid.UUID       `json:"id"`
	Name       string          `json:"name"`
	UserID     string          `json:"user_id"`
	Source     string          `json:"source"`
	Metadata   json.RawMessage `json:"metadata"`
	OccurredAt time.Time       `json:"occurred_at"`
	ReceivedAt time.Time       `json:"received_at"`
}

const maxMetadataSize = 32 * 1024 // 32KB keeps payloads bounded.

// NewEvent validates input and builds a well-formed Event instance.
func NewEvent(name, userID, source string, metadata json.RawMessage, occurredAt time.Time) (Event, error) {
	if name == "" {
		return Event{}, errors.New("event name is required")
	}
	if userID == "" {
		return Event{}, errors.New("event user_id is required")
	}
	if source == "" {
		return Event{}, errors.New("event source is required")
	}
	if len(metadata) > maxMetadataSize {
		return Event{}, errors.New("metadata exceeds 32KB limit")
	}

	now := time.Now().UTC()
	if metadata == nil {
		metadata = json.RawMessage("{}")
	}
	if occurredAt.IsZero() {
		occurredAt = now
	} else {
		occurredAt = occurredAt.UTC()
	}

	return Event{
		ID:         uuid.New(),
		Name:       name,
		UserID:     userID,
		Source:     source,
		Metadata:   metadata,
		OccurredAt: occurredAt,
		ReceivedAt: now,
	}, nil
}
