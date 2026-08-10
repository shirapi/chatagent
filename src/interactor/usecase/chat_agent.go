package usecase

import (
	"chatagent/domain"
	"chatagent/interactor"
	"context"
	"net/http"
)

type ChatAgent struct {
	AgentRepo        domain.ChatAgentRepository
	NotificationRepo domain.NotificationRepository
}

func NewChatAgent(agentRepo domain.ChatAgentRepository, notiRepo domain.NotificationRepository) interactor.ChatAgent {
	return &ChatAgent{
		AgentRepo:        agentRepo,
		NotificationRepo: notiRepo,
	}
}

func (ctrl *ChatAgent) Verify(r *http.Request) (string, error) {
	return ctrl.NotificationRepo.Verify(r)
}

func (ctrl *ChatAgent) Exec(ctx context.Context) error {
	// TODO
	// エージェントのコール・受け取り
	// Slackへの返信
	return nil
}
