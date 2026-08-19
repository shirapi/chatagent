package controller

import (
	"chatagent/usecase"
	"context"
	"log/slog"
	"net/http"
)

type Response struct {
	StatusCode int
	Body       string
}

type Receiver struct {
	AppEnv  string
	Usecase usecase.ChatAgent
}

func NewReceiver(appEnv string, uc usecase.ChatAgent) *Receiver {
	return &Receiver{
		AppEnv:  appEnv,
		Usecase: uc,
	}
}

func (c *Receiver) Exec(ctx context.Context, r *http.Request) (*Response, error) {
	result, err := c.Usecase.Verify(r)
	if err != nil {
		slog.Error("verify failed", "err", err)
		return &Response{StatusCode: http.StatusUnauthorized}, nil
	}
	slog.Info("verify done", "hasChallenge", result.Challenge != "", "hasMention", result.Mention != nil)
	if result.Challenge != "" {
		return &Response{StatusCode: http.StatusOK, Body: result.Challenge}, nil
	}
	if result.Mention != nil {
		if err := c.Usecase.Accept(ctx, *result.Mention); err != nil {
			slog.Error("usecase accept failed", "err", err)
			return &Response{StatusCode: http.StatusInternalServerError}, nil
		}
	}

	return &Response{StatusCode: http.StatusOK}, nil
}
