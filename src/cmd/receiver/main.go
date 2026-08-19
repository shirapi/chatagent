package main

import (
	"bytes"
	"chatagent/domain/env"
	"chatagent/infra/di"
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ctrl, err := di.NewReceiver(ctx)
	if err != nil {
		slog.Error("NewReceiver", "err", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.HTTPMethod, req.Path, bytes.NewBufferString(req.Body))
	if err != nil {
		slog.Error("build http.Request", "err", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, nil
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	res, err := ctrl.Exec(ctx, httpReq)
	if err != nil {
		slog.Error("Controller Exec", "err", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	return events.APIGatewayProxyResponse{StatusCode: res.StatusCode, Body: res.Body}, nil
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
