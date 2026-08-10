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
	challenge, err := ctrl.Usecase.Verify(r)
	if err != nil {
		return &Response{StatusCode: http.StatusUnauthorized}, nil
	}
	if challenge != "" {
		return &Response{StatusCode: http.StatusOK, Body: challenge}, nil
	}

	// TODO ユースケースの呼び出し
	return &Response{StatusCode: http.StatusOK}, nil
}
