package main

import (
	"chatagent/domain/env"
	"chatagent/infra/di"
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, payload json.RawMessage) error {
	ctrl, err := di.NewWorker(ctx)
	if err != nil {
		slog.Error("NewWorker", "err", err)
		return err
	}

	if err := ctrl.Exec(ctx, payload); err != nil {
		slog.Error("Controller Exec", "err", err)
		return err
	}
	return nil
}

func main() {
	if env.GetAppEnv() == env.LOCAL {
		slog.Info("local start")
		return
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	lambda.Start(handler)
}
