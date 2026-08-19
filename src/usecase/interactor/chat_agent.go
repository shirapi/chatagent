package interactor

import (
	"chatagent/domain"
	"chatagent/usecase"
	"context"
	"log/slog"
	"net/http"
)

type ChatAgent struct {
	AgentRepo        domain.ChatAgentRepository
	NotificationRepo domain.NotificationRepository
	WorkerDispatcher domain.WorkerDispatcher
	AllowedChannel   string
	ReactionName     string
}

func NewChatAgent(agentRepo domain.ChatAgentRepository, notiRepo domain.NotificationRepository, workerDispatcher domain.WorkerDispatcher, allowedChannel, reactionName string) usecase.ChatAgent {
	return &ChatAgent{
		AgentRepo:        agentRepo,
		NotificationRepo: notiRepo,
		WorkerDispatcher: workerDispatcher,
		AllowedChannel:   allowedChannel,
		ReactionName:     reactionName,
	}
}

func (u *ChatAgent) Verify(r *http.Request) (domain.VerifyResult, error) {
	return u.NotificationRepo.Verify(r)
}

func (u *ChatAgent) Accept(ctx context.Context, mention domain.MentionEvent) error {
	return u.WorkerDispatcher.Dispatch(ctx, mention)
}

func (u *ChatAgent) Exec(ctx context.Context, mention domain.MentionEvent) error {
	target := domain.NotificationTarget{
		Channel:   mention.Channel,
		UserID:    mention.UserID,
		Timestamp: mention.Timestamp,
	}

	if mention.Channel != u.AllowedChannel {
		return u.NotificationRepo.PostReply(ctx, target, domain.MessageInvalidChannel)
	}
	slog.Info("allowed channel check passed", "channel", mention.Channel)

	if err := u.NotificationRepo.AddReaction(ctx, target, u.ReactionName); err != nil {
		return err
	}
	slog.Info("reaction added", "reaction", u.ReactionName)

	sessionID := mention.ThreadTimestamp
	if sessionID == "" {
		sessionID = mention.Timestamp
	}

	reply, err := u.AgentRepo.Call(ctx, sessionID, mention.Text)
	if err != nil {
		return err
	}
	slog.Info("agent call done", "replyLength", len(reply))

	return u.NotificationRepo.PostReply(ctx, target, reply)
}
