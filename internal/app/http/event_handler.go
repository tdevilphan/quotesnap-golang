package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/tdevilphan/quote-snap-golang/internal/core/usecase"
)

// EventHandler wires HTTP transport with the ingest event use case.
type EventHandler struct {
	usecase        *usecase.IngestEvent
	requestTimeout time.Duration
	logger         *slog.Logger
}

// NewEventHandler builds an EventHandler instance.
func NewEventHandler(uc *usecase.IngestEvent, timeout time.Duration, logger *slog.Logger) *EventHandler {
	return &EventHandler{usecase: uc, requestTimeout: timeout, logger: logger}
}

// Register attaches handler endpoints to the provided router group.
func (h *EventHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/events", h.createEvent)
}

type createEventRequest struct {
	Name       string         `json:"name"`
	UserID     string         `json:"user_id"`
	Source     string         `json:"source"`
	Metadata   map[string]any `json:"metadata"`
	OccurredAt *time.Time     `json:"occurred_at"`
}

type createEventResponse struct {
	ID         string    `json:"id"`
	ReceivedAt time.Time `json:"received_at"`
}

func (h *EventHandler) createEvent(c *gin.Context) {
	var req createEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request payload", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	metadata, err := jsonMarshal(req.Metadata)
	if err != nil {
		h.logger.Warn("metadata marshal failed", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid metadata"})
		return
	}

	var occurredAt time.Time
	if req.OccurredAt != nil {
		occurredAt = req.OccurredAt.UTC()
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.requestTimeout)
	defer cancel()

	event, err := h.usecase.Execute(ctx, usecase.IngestEventInput{
		Name:       req.Name,
		UserID:     req.UserID,
		Source:     req.Source,
		Metadata:   metadata,
		OccurredAt: occurredAt,
	})
	if err != nil {
		h.logger.Error("event ingestion failed", "error", err)
		code := http.StatusInternalServerError
		if errors.Is(err, usecase.ErrValidation) {
			code = http.StatusBadRequest
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, createEventResponse{ID: event.ID.String(), ReceivedAt: event.ReceivedAt})
}

func jsonMarshal(v any) ([]byte, error) {
	if v == nil {
		return []byte("{}"), nil
	}
	payload, err := json.Marshal(v)
	return payload, errors.Wrap(err, "marshal metadata")
}
