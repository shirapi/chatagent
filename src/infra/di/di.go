package di

import (
	"chatagent/domain/env"
	"chatagent/infra"
	"chatagent/infra/notification"
	"chatagent/interface/controller"
	"chatagent/usecase/interactor"
	"context"
)

func NewController(ctx context.Context) (*controller.ChatAgent, error) {
	agentRepo, err := infra.NewChatAgentRepository(ctx, env.GetAgentCoreRuntimeArn())
	if err != nil {
		return nil, err
	}
	slackRepo := notification.NewSlack(env.GetSlackOAuthToken(), env.GetSlackSigningSecret())
	uc := interactor.NewChatAgent(agentRepo, slackRepo, env.GetSlackChannelID(), env.GetSlackReactionName())
	return controller.NewChatAgent(env.GetAppEnv(), uc), nil
}
