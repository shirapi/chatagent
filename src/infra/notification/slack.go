package notification

import (
	"chatagent/domain"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

const (
	headerRetryNum     = "X-Slack-Retry-Num"
	headerRetryReason  = "X-Slack-Retry-Reason"
	retryReasonTimeout = "http_timeout"
)

type Slack struct {
	Client        *slack.Client
	SigningSecret string
}

func NewSlack(token, signingSecret string) *Slack {
	return &Slack{
		Client:        slack.New(token),
		SigningSecret: signingSecret,
	}
}

func (rep *Slack) Verify(r *http.Request) (domain.VerifyResult, error) {
	if r.Header.Get(headerRetryNum) != "" && r.Header.Get(headerRetryReason) == retryReasonTimeout {
		return domain.VerifyResult{}, nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return domain.VerifyResult{}, fmt.Errorf("failed to read request body: %w", err)
	}
	defer r.Body.Close()

	verifier, err := slack.NewSecretsVerifier(r.Header, rep.SigningSecret)
	if err != nil {
		return domain.VerifyResult{}, fmt.Errorf("failed to create secrets verifier: %w", err)
	}
	if _, err := verifier.Write(body); err != nil {
		return domain.VerifyResult{}, fmt.Errorf("failed to write body to verifier: %w", err)
	}
	if err := verifier.Ensure(); err != nil {
		return domain.VerifyResult{}, fmt.Errorf("invalid slack signature: %w", err)
	}

	event, err := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken())
	if err != nil {
		return domain.VerifyResult{}, fmt.Errorf("failed to parse slack event: %w", err)
	}

	if event.Type == slackevents.URLVerification {
		var res slackevents.ChallengeResponse
		if err := json.Unmarshal(body, &res); err != nil {
			return domain.VerifyResult{}, fmt.Errorf("failed to unmarshal challenge response: %w", err)
		}
		return domain.VerifyResult{Challenge: res.Challenge}, nil
	}

	if ev, ok := event.InnerEvent.Data.(*slackevents.AppMentionEvent); ok {
		return domain.VerifyResult{
			Mention: &domain.MentionEvent{
				Channel:         ev.Channel,
				UserID:          ev.User,
				Timestamp:       ev.TimeStamp,
				ThreadTimestamp: ev.ThreadTimeStamp,
				Text:            ev.Text,
			},
		}, nil
	}

	return domain.VerifyResult{}, nil
}

func (rep *Slack) AddReaction(ctx context.Context, target domain.NotificationTarget, reaction string) error {
	item := slack.NewRefToMessage(target.Channel, target.Timestamp)
	if err := rep.Client.AddReactionContext(ctx, reaction, item); err != nil {
		return fmt.Errorf("failed to add reaction: %w", err)
	}
	return nil
}

func (rep *Slack) PostReply(ctx context.Context, target domain.NotificationTarget, message string) error {
	text := fmt.Sprintf("<@%s> %s", target.UserID, message)
	if _, err := rep.sendThread(ctx, target.Channel, target.Timestamp, text); err != nil {
		return err
	}
	return nil
}

func (rep *Slack) sendThread(ctx context.Context, channel, ts, msg string) (string, error) {
	_, resTS, err := rep.Client.PostMessageContext(
		ctx,
		channel,
		slack.MsgOptionTS(ts),
		slack.MsgOptionText(msg, false),
	)
	if err != nil {
		return "", fmt.Errorf("failed to send message to slack: %w", err)
	}
	return resTS, nil
}
