package infra

import (
	"chatagent/domain"
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

type WorkerDispatcher struct {
	client             *lambda.Client
	workerFunctionName string
}

func NewWorkerDispatcher(ctx context.Context, workerFunctionName string) (domain.WorkerDispatcher, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}
	return &WorkerDispatcher{
		client:             lambda.NewFromConfig(cfg),
		workerFunctionName: workerFunctionName,
	}, nil
}

func (d *WorkerDispatcher) Dispatch(ctx context.Context, mention domain.MentionEvent) error {
	payload, err := json.Marshal(mention)
	if err != nil {
		return fmt.Errorf("failed to marshal mention: %w", err)
	}

	_, err = d.client.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(d.workerFunctionName),
		InvocationType: types.InvocationTypeEvent,
		Payload:        payload,
	})
	if err != nil {
		return fmt.Errorf("failed to invoke worker function: %w", err)
	}
	return nil
}
