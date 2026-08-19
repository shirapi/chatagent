package usecase

import (
	"chatagent/domain"
	"context"
	"net/http"
)

type ChatAgent interface {
	Verify(r *http.Request) (domain.VerifyResult, error)
	Accept(ctx context.Context, mention domain.MentionEvent) error
	Exec(ctx context.Context, mention domain.MentionEvent) error
}
