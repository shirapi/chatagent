package controller

import (
	"chatagent/domain"
	"chatagent/usecase"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

type Worker struct {
	Usecase usecase.ChatAgent
}

func NewWorker(uc usecase.ChatAgent) *Worker {
	return &Worker{
		Usecase: uc,
	}
}

func (c *Worker) Exec(ctx context.Context, payload []byte) error {
	var mention domain.MentionEvent
	if err := json.Unmarshal(payload, &mention); err != nil {
		return fmt.Errorf("failed to unmarshal mention: %w", err)
	}

	if err := c.Usecase.Exec(ctx, mention); err != nil {
		slog.Error("usecase exec failed", "err", err)
		return err
	}
	return nil
}
