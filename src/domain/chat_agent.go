package domain

import (
	"context"
	"net/http"
)

type ChatAgentRepository interface {
	Call(ctx context.Context, msg string) error // TODO レスポンス型定義
}

type NotificationRepository interface {
	Verify(r *http.Request) (string, error)
	// AddReaction(ctx context.Context, channel, timestamp, reaction string) error
	PostReply(ctx context.Context, channel, userID, timestamp, message string) error
}
