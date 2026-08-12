package infra

import (
	"bufio"
	"chatagent/domain"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcore"
)

// runtimeActorID はAgentCore Runtimeの RuntimeUserId。
// SessionIDがスレッドごとに一意なので、ActorIDは固定値でよい。
const runtimeActorID = "conversation-thread"

type ChatAgentRepository struct {
	client          *bedrockagentcore.Client
	agentRuntimeArn string
}

func NewChatAgentRepository(ctx context.Context, agentRuntimeArn string) (domain.ChatAgentRepository, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}
	return &ChatAgentRepository{
		client:          bedrockagentcore.NewFromConfig(cfg),
		agentRuntimeArn: agentRuntimeArn,
	}, nil
}

func (rep *ChatAgentRepository) Call(ctx context.Context, sessionID, message string) (string, error) {
	slog.Info("chat agent repository call start", "sessionID", sessionID, "agentRuntimeArn", rep.agentRuntimeArn)

	payload, err := json.Marshal(map[string]string{"prompt": message})
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	out, err := rep.client.InvokeAgentRuntime(ctx, &bedrockagentcore.InvokeAgentRuntimeInput{
		AgentRuntimeArn:  aws.String(rep.agentRuntimeArn),
		Payload:          payload,
		RuntimeSessionId: aws.String(toRuntimeSessionID(sessionID)),
		RuntimeUserId:    aws.String(runtimeActorID),
	})
	if err != nil {
		return "", fmt.Errorf("failed to invoke agent runtime: %w", err)
	}
	defer out.Response.Close()

	text, err := extractText(out.Response)
	if err != nil {
		return "", fmt.Errorf("failed to parse agent response: %w", err)
	}
	return text, nil
}

// toRuntimeSessionID はSlackのスレッドタイムスタンプ等をAgentCoreの
// RuntimeSessionId要件（33文字以上、[a-zA-Z0-9][a-zA-Z0-9-_]*）に変換する。
func toRuntimeSessionID(sessionID string) string {
	id := "chatagent-session-" + strings.ReplaceAll(sessionID, ".", "-")
	for len(id) < 33 {
		id += "-"
	}
	return id
}

// extractText はSSE（text/event-stream）形式のレスポンスから、
// Bedrock ConverseのcontentBlockDelta.delta.textを連結して最終的な応答テキストを組み立てる。
func extractText(body io.ReadCloser) (string, error) {
	var sb strings.Builder
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()
		data, ok := strings.CutPrefix(line, "data: ")
		if !ok {
			continue
		}

		var evt struct {
			Event struct {
				ContentBlockDelta struct {
					Delta struct {
						Text string `json:"text"`
					} `json:"delta"`
				} `json:"contentBlockDelta"`
			} `json:"event"`
		}
		if err := json.Unmarshal([]byte(data), &evt); err != nil {
			continue
		}
		sb.WriteString(evt.Event.ContentBlockDelta.Delta.Text)
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return sb.String(), nil
}
