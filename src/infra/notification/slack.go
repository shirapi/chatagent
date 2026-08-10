package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/slack-go/slack"
)

type Slack struct {
	Client        *slack.Client
	ChannelID     string
	SigningSecret string
}

func NewSlack(token, channelID, signingSecret string) *Slack {
	return &Slack{
		Client:        slack.New(token),
		ChannelID:     channelID,
		SigningSecret: signingSecret,
	}
}

// challengeRequest はSlack Events APIのURL Verificationリクエストのペイロード。
// Slack固有のJSON形式のためinfra層に閉じ込める。
type challengeRequest struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
}

func (rep *Slack) Verify(r *http.Request) (string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read request body: %w", err)
	}
	defer r.Body.Close()

	verifier, err := slack.NewSecretsVerifier(r.Header, rep.SigningSecret)
	if err != nil {
		return "", fmt.Errorf("failed to create secrets verifier: %w", err)
	}
	if _, err := verifier.Write(body); err != nil {
		return "", fmt.Errorf("failed to write body to verifier: %w", err)
	}
	if err := verifier.Ensure(); err != nil {
		return "", fmt.Errorf("invalid slack signature: %w", err)
	}

	var req challengeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return "", fmt.Errorf("failed to unmarshal request body: %w", err)
	}
	if req.Type != "url_verification" {
		return "", nil
	}
	return req.Challenge, nil
}

func (rep *Slack) addReaction(ctx context.Context, channel, timestamp, reaction string) error {
	// TODO
	return nil
}

func (rep *Slack) PostReply(ctx context.Context, channel, userID, timestamp, message string) error {
	// TODO sendTheraedを呼ぶ想定
	return nil
}

func (rep *Slack) sendThread(msg, ts string) (string, error) {
	_, ts, err := rep.Client.PostMessage(
		rep.ChannelID,
		slack.MsgOptionTS(ts),
		slack.MsgOptionAttachments(
			slack.Attachment{
				Text: msg,
			},
		),
	)
	if err != nil {
		return ts, fmt.Errorf("failed to send message to Slack SendThread: %w", err)
	}
	return ts, nil
}
