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

type ChatAgent struct {
	AppEnv  string
	Usecase usecase.ChatAgent
}

func NewChatAgent(appEnv string, uc usecase.ChatAgent) *ChatAgent {
	return &ChatAgent{
		AppEnv:  appEnv,
		Usecase: uc,
	}
}

func (c *ChatAgent) Exec(ctx context.Context, r *http.Request) (*Response, error) {
	result, err := c.Usecase.Verify(r)
	if err != nil {
		slog.Error("verify failed", "err", err)
		return &Response{StatusCode: http.StatusUnauthorized}, nil
	}
	slog.Info("verify done", "hasChallenge", result.Challenge != "", "hasMention", result.Mention != nil)
	if result.Challenge != "" {
		return &Response{StatusCode: http.StatusOK, Body: result.Challenge}, nil
	}
	// TODO 先にレスポンスは返しておく必要あり
	if result.Mention != nil {
		if err := c.Usecase.Exec(ctx, *result.Mention); err != nil {
			slog.Error("usecase exec failed", "err", err)
			return &Response{StatusCode: http.StatusInternalServerError}, nil
		}
	}

	return &Response{StatusCode: http.StatusOK}, nil
}
