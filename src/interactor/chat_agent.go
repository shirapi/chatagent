package interactor

import (
	"context"
	"net/http"
)

type ChatAgent interface {
	Verify(r *http.Request) (string, error)
	Exec(ctx context.Context) error
}
