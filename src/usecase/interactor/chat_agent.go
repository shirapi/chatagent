package interactor

import (
	"chatagent/domain"
	"chatagent/usecase"
	"context"
	"fmt"
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
		slog.Error("add reaction failed", "err", err)
		return u.notifyProcessingFailed(ctx, target, err)
	}
	slog.Info("reaction added", "reaction", u.ReactionName)

	sessionID := mention.ThreadTimestamp
	if sessionID == "" {
		sessionID = mention.Timestamp
	}

	reply, err := u.AgentRepo.Call(ctx, sessionID, mention.Text)
	if err != nil {
		slog.Error("agent call failed", "err", err)
		return u.notifyProcessingFailed(ctx, target, err)
	}
	slog.Info("agent call done", "replyLength", len(reply))

	if reply == "" {
		reply = domain.MessageNoResponse
	}
	return u.NotificationRepo.PostReply(ctx, target, reply)
}

// notifyProcessingFailed Slackへエラーメッセージを返信する
// エラーを返すとLambdaが自動的にリトライするので、通知に成功した場合はリトライを止めて重複通知を避けるためにnilを返す
func (u *ChatAgent) notifyProcessingFailed(ctx context.Context, target domain.NotificationTarget, cause error) error {
	if err := u.NotificationRepo.PostReply(ctx, target, domain.MessageProcessingFailed); err != nil {
		return fmt.Errorf("original error: %w; notify user failed: %w", cause, err)
	}
	return nil
}
