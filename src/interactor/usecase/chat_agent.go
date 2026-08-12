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
	AllowedChannel   string
	ReactionName     string
}

func NewChatAgent(agentRepo domain.ChatAgentRepository, notiRepo domain.NotificationRepository, allowedChannel, reactionName string) interactor.ChatAgent {
	return &ChatAgent{
		AgentRepo:        agentRepo,
		NotificationRepo: notiRepo,
		AllowedChannel:   allowedChannel,
		ReactionName:     reactionName,
	}
}

func (ctrl *ChatAgent) Verify(r *http.Request) (domain.VerifyResult, error) {
	return ctrl.NotificationRepo.Verify(r)
}

func (ctrl *ChatAgent) Exec(ctx context.Context, mention domain.MentionEvent) error {
	target := domain.NotificationTarget{
		Channel:   mention.Channel,
		UserID:    mention.UserID,
		Timestamp: mention.Timestamp,
	}

	if mention.Channel != ctrl.AllowedChannel {
		return ctrl.NotificationRepo.PostReply(ctx, target, domain.MessageInvalidChannel)
	}

	if err := ctrl.NotificationRepo.AddReaction(ctx, target, ctrl.ReactionName); err != nil {
		return err
	}

	// TODO
	// エージェントのコール・受け取り
	// Slackへの返信
	return nil
}
