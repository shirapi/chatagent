package interactor

import (
	"chatagent/domain"
	"context"
	"net/http"
)

type ChatAgent interface {
	Verify(r *http.Request) (domain.VerifyResult, error)
	Exec(ctx context.Context, mention domain.MentionEvent) error
}
