package infra

import (
	"chatagent/domain"
	"context"
)

type ChatAgentRepository struct {
}

func NewChatAgentRepository() domain.ChatAgentRepository {
	return &ChatAgentRepository{
		// TODO
	}
}

func (rep *ChatAgentRepository) Call(ctx context.Context, msg string) error { // TODO レスポンス
	return nil
}
