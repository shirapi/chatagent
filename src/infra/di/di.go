package di

import (
	"chatagent/domain/env"
	"chatagent/infra"
	"chatagent/infra/notification"
	"chatagent/interface/controller"
	"chatagent/usecase"
	"chatagent/usecase/interactor"
	"context"
)

func newChatAgent(ctx context.Context) (usecase.ChatAgent, error) {
	agentRepo, err := infra.NewChatAgentRepository(ctx, env.GetAgentCoreRuntimeArn())
	if err != nil {
		return nil, err
	}
	workerDispatcher, err := infra.NewWorkerDispatcher(ctx, env.GetWorkerFunctionName())
	if err != nil {
		return nil, err
	}
	slackRepo := notification.NewSlack(env.GetSlackOAuthToken(), env.GetSlackSigningSecret())
	return interactor.NewChatAgent(agentRepo, slackRepo, workerDispatcher, env.GetSlackChannelID(), env.GetSlackReactionName()), nil
}

func NewReceiver(ctx context.Context) (*controller.Receiver, error) {
	uc, err := newChatAgent(ctx)
	if err != nil {
		return nil, err
	}
	return controller.NewReceiver(env.GetAppEnv(), uc), nil
}

func NewWorker(ctx context.Context) (*controller.Worker, error) {
	uc, err := newChatAgent(ctx)
	if err != nil {
		return nil, err
	}
	return controller.NewWorker(uc), nil
}
