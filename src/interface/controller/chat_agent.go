package controller

import (
	"chatagent/interactor"
	"context"
	"net/http"
)

type Response struct {
	StatusCode int
	Body       string
}

type ChatAgent struct {
	AppEnv  string
	Usecase interactor.ChatAgent
}

func NewChatAgent(appEnv string, uc interactor.ChatAgent) *ChatAgent {
	return &ChatAgent{
		AppEnv:  appEnv,
		Usecase: uc,
	}
}

func (ctrl *ChatAgent) Exec(ctx context.Context, r *http.Request) (*Response, error) {
	result, err := ctrl.Usecase.Verify(r)
	if err != nil {
		return &Response{StatusCode: http.StatusUnauthorized}, nil
	}
	if result.Challenge != "" {
		return &Response{StatusCode: http.StatusOK, Body: result.Challenge}, nil
	}
	if result.Mention != nil {
		if err := ctrl.Usecase.Exec(ctx, *result.Mention); err != nil {
			return &Response{StatusCode: http.StatusInternalServerError}, nil
		}
	}

	return &Response{StatusCode: http.StatusOK}, nil
}
