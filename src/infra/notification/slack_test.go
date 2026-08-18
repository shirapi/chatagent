package notification

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

const testSigningSecret = "test-signing-secret"

func signBody(secret, timestamp, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(fmt.Appendf(nil, "v0:%s:%s", timestamp, body))
	return "v0=" + hex.EncodeToString(mac.Sum(nil))
}

func newSignedRequest(t *testing.T, secret, body string) *http.Request {
	t.Helper()
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	req := httptest.NewRequest(http.MethodPost, "/slack/events", strings.NewReader(body))
	req.Header.Set("X-Slack-Request-Timestamp", ts)
	req.Header.Set("X-Slack-Signature", signBody(secret, ts, body))
	return req
}

func TestVerify_RetryRequestIsIgnored(t *testing.T) {
	rep := &Slack{SigningSecret: testSigningSecret}

	// 署名ヘッダを付けない。署名検証より前にリトライ判定で打ち切られるはず。
	req := httptest.NewRequest(http.MethodPost, "/slack/events", strings.NewReader("irrelevant body"))
	req.Header.Set(headerRetryNum, "1")
	req.Header.Set(headerRetryReason, retryReasonTimeout)

	result, err := rep.Verify(req)
	if err != nil {
		t.Fatalf("Verify() unexpected error: %v", err)
	}
	if result.Challenge != "" || result.Mention != nil {
		t.Errorf("Verify() = %+v, want empty VerifyResult", result)
	}
}

func TestVerify_InvalidSignature(t *testing.T) {
	rep := &Slack{SigningSecret: testSigningSecret}

	req := newSignedRequest(t, "wrong-secret", `{"type":"url_verification","challenge":"abc"}`)

	_, err := rep.Verify(req)
	if err == nil {
		t.Fatal("Verify() error = nil, want an error for invalid signature")
	}
}

func TestVerify_URLVerification(t *testing.T) {
	rep := &Slack{SigningSecret: testSigningSecret}

	body := `{"type":"url_verification","challenge":"abc123","token":"tok"}`
	req := newSignedRequest(t, testSigningSecret, body)

	result, err := rep.Verify(req)
	if err != nil {
		t.Fatalf("Verify() unexpected error: %v", err)
	}
	if result.Challenge != "abc123" {
		t.Errorf("Verify().Challenge = %q, want %q", result.Challenge, "abc123")
	}
	if result.Mention != nil {
		t.Errorf("Verify().Mention = %+v, want nil", result.Mention)
	}
}

func TestVerify_EventCallback(t *testing.T) {
	// 実運用ではSubscribe to bot eventsをapp_mentionのみに絞っているため他のevent.typeは
	// 送られてこないはずだが、念のため無視されることも確認しておく。
	tests := []struct {
		name        string
		eventType   string
		wantMention bool
	}{
		{name: "app_mention", eventType: "app_mention", wantMention: true},
		{name: "app_mention以外は無視される", eventType: "message", wantMention: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep := &Slack{SigningSecret: testSigningSecret}

			body := fmt.Sprintf(`{
				"type": "event_callback",
				"token": "tok",
				"team_id": "T123",
				"api_app_id": "A123",
				"event": {
					"type": %q,
					"user": "U123",
					"text": "<@BOT> hello",
					"ts": "1234567890.000100",
					"channel": "C123",
					"thread_ts": "1234567890.000000"
				}
			}`, tt.eventType)
			req := newSignedRequest(t, testSigningSecret, body)

			result, err := rep.Verify(req)
			if err != nil {
				t.Fatalf("Verify() unexpected error: %v", err)
			}
			if result.Challenge != "" {
				t.Errorf("Verify().Challenge = %q, want empty", result.Challenge)
			}

			if !tt.wantMention {
				if result.Mention != nil {
					t.Errorf("Verify().Mention = %+v, want nil", result.Mention)
				}
				return
			}

			if result.Mention == nil {
				t.Fatal("Verify().Mention = nil, want non-nil")
			}
			got := result.Mention
			if got.Channel != "C123" || got.UserID != "U123" || got.Timestamp != "1234567890.000100" ||
				got.ThreadTimestamp != "1234567890.000000" || got.Text != "<@BOT> hello" {
				t.Errorf("Verify().Mention = %+v", got)
			}
		})
	}
}
