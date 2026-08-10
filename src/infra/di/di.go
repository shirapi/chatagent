package di

import (
	"chatagent/domain/env"
	"chatagent/infra"
	"chatagent/infra/notification"
	"chatagent/interactor/usecase"
	"chatagent/interface/controller"
	"context"
)

func NewController(ctx context.Context) (*controller.ChatAgent, error) {
	agentRepo := infra.NewChatAgentRepository()
	slackRepo := notification.NewSlack(env.GetSlackOAuthToken(), env.GetStackChannelID(), env.GetSlackSigningSecret())
	uc := usecase.NewChatAgent(agentRepo, slackRepo)
	return controller.NewChatAgent(env.GetAppEnv(), uc), nil
}
