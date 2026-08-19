package domain

import (
	"context"
	"net/http"
)

type ChatAgentRepository interface {
	Call(ctx context.Context, sessionID, message string) (string, error)
}

type NotificationTarget struct {
	Channel   string
	UserID    string
	Timestamp string
}

type MentionEvent struct {
	Channel         string
	UserID          string
	Timestamp       string
	ThreadTimestamp string
	Text            string
}

type VerifyResult struct {
	Challenge string
	Mention   *MentionEvent
}

const MessageInvalidChannel = "このチャンネルでは回答できません"

type NotificationRepository interface {
	Verify(r *http.Request) (VerifyResult, error)
	AddReaction(ctx context.Context, target NotificationTarget, reaction string) error
	PostReply(ctx context.Context, target NotificationTarget, message string) error
}

type WorkerDispatcher interface {
	Dispatch(ctx context.Context, mention MentionEvent) error
}
