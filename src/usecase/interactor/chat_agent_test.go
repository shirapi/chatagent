package interactor

import (
	"chatagent/domain"
	"context"
	"errors"
	"net/http"
	"testing"
)

type mockNotificationRepo struct {
	AddReactionErr error
	PostReplyErr   error

	addReactionCalls []domain.NotificationTarget
	postReplyCalls   []struct {
		Target  domain.NotificationTarget
		Message string
	}
}

func (m *mockNotificationRepo) Verify(r *http.Request) (domain.VerifyResult, error) {
	return domain.VerifyResult{}, nil
}

func (m *mockNotificationRepo) AddReaction(ctx context.Context, target domain.NotificationTarget, reaction string) error {
	m.addReactionCalls = append(m.addReactionCalls, target)
	return m.AddReactionErr
}

func (m *mockNotificationRepo) PostReply(ctx context.Context, target domain.NotificationTarget, message string) error {
	m.postReplyCalls = append(m.postReplyCalls, struct {
		Target  domain.NotificationTarget
		Message string
	}{target, message})
	return m.PostReplyErr
}

type mockAgentRepo struct {
	Reply string
	Err   error

	callSessionIDs []string
}

func (m *mockAgentRepo) Call(ctx context.Context, sessionID, message string) (string, error) {
	m.callSessionIDs = append(m.callSessionIDs, sessionID)
	return m.Reply, m.Err
}

type mockWorkerDispatcher struct {
	Err error

	dispatchCalls []domain.MentionEvent
}

func (m *mockWorkerDispatcher) Dispatch(ctx context.Context, mention domain.MentionEvent) error {
	m.dispatchCalls = append(m.dispatchCalls, mention)
	return m.Err
}

func TestAccept(t *testing.T) {
	mention := domain.MentionEvent{Channel: "C_ALLOWED", Timestamp: "1.1"}

	tests := []struct {
		name          string
		dispatcherErr error
		wantErr       bool
	}{
		{name: "dispatch succeeds", dispatcherErr: nil, wantErr: false},
		{name: "dispatch fails", dispatcherErr: errors.New("dispatch failed"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dispatcher := &mockWorkerDispatcher{Err: tt.dispatcherErr}
			u := &ChatAgent{
				WorkerDispatcher: dispatcher,
			}

			err := u.Accept(context.Background(), mention)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Accept() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(dispatcher.dispatchCalls) != 1 || dispatcher.dispatchCalls[0] != mention {
				t.Errorf("WorkerDispatcher.Dispatch calls = %+v, want [%+v]", dispatcher.dispatchCalls, mention)
			}
		})
	}
}

func TestExec_Failures(t *testing.T) {
	tests := []struct {
		name                 string
		channel              string
		addReactionErr       error
		agentErr             error
		postReplyErr         error
		wantAddReactionCalls int
		wantAgentCallCalls   int
		wantMessage          string
		wantErr              bool
	}{
		{
			// 正常系:チャンネル不一致
			name: "channel not allowed", channel: "C_OTHER",
			wantMessage: domain.MessageInvalidChannel, wantErr: false,
		},
		{
			// 異常系:チャンネル不一致 -> 返信失敗
			name: "channel not allowed and notify fails", channel: "C_OTHER",
			postReplyErr: errors.New("post reply failed"),
			wantErr:      true,
		},
		{
			// 正常系:リアクション失敗
			name: "add reaction fails", channel: "C_ALLOWED",
			addReactionErr:       errors.New("add reaction failed"),
			wantAddReactionCalls: 1, wantMessage: domain.MessageProcessingFailed, wantErr: false,
		},
		{
			// 異常系:リアクション失敗 -> 返信失敗
			name: "add reaction fails and notify also fails", channel: "C_ALLOWED",
			addReactionErr: errors.New("add reaction failed"), postReplyErr: errors.New("post reply failed"),
			wantAddReactionCalls: 1, wantErr: true,
		},
		{
			// 正常系:エージェント呼び出し失敗
			name: "agent call fails", channel: "C_ALLOWED",
			agentErr:             errors.New("agent call failed"),
			wantAddReactionCalls: 1, wantAgentCallCalls: 1, wantMessage: domain.MessageProcessingFailed, wantErr: false,
		},
		{
			// 異常系:エージェント呼び出し失敗 -> 返信失敗
			name: "agent call fails and notify also fails", channel: "C_ALLOWED",
			agentErr: errors.New("agent call failed"), postReplyErr: errors.New("post reply failed"),
			wantAddReactionCalls: 1, wantAgentCallCalls: 1, wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notiRepo := &mockNotificationRepo{
				AddReactionErr: tt.addReactionErr,
				PostReplyErr:   tt.postReplyErr,
			}
			agentRepo := &mockAgentRepo{Err: tt.agentErr}
			u := &ChatAgent{
				AgentRepo:        agentRepo,
				NotificationRepo: notiRepo,
				AllowedChannel:   "C_ALLOWED",
				ReactionName:     "eyes",
			}

			err := u.Exec(context.Background(), domain.MentionEvent{Channel: tt.channel, Timestamp: "1.1"})

			if len(notiRepo.addReactionCalls) != tt.wantAddReactionCalls {
				t.Errorf("AddReaction call count = %d, want %d", len(notiRepo.addReactionCalls), tt.wantAddReactionCalls)
			}
			if len(agentRepo.callSessionIDs) != tt.wantAgentCallCalls {
				t.Errorf("AgentRepo.Call call count = %d, want %d", len(agentRepo.callSessionIDs), tt.wantAgentCallCalls)
			}

			if tt.wantErr {
				if err == nil {
					t.Fatal("Exec() expected an error, got nil")
				}
				if tt.postReplyErr != nil && !errors.Is(err, tt.postReplyErr) {
					t.Errorf("Exec() error = %v, want it to wrap %v", err, tt.postReplyErr)
				}
				cause := tt.addReactionErr
				if cause == nil {
					cause = tt.agentErr
				}
				if cause != nil && !errors.Is(err, cause) {
					t.Errorf("Exec() error = %v, want it to wrap %v", err, cause)
				}
				return
			}

			if err != nil {
				t.Fatalf("Exec() unexpected error: %v", err)
			}
			if len(notiRepo.postReplyCalls) != 1 {
				t.Fatalf("PostReply call count = %d, want 1", len(notiRepo.postReplyCalls))
			}
			if got := notiRepo.postReplyCalls[0].Message; got != tt.wantMessage {
				t.Errorf("PostReply message = %q, want %q", got, tt.wantMessage)
			}
		})
	}
}

func TestExec_Success(t *testing.T) {
	tests := []struct {
		name            string
		threadTimestamp string
		agentReply      string
		wantSessionID   string
		wantMessage     string
	}{
		{
			name: "uses ThreadTimestamp when present", threadTimestamp: "1.0", agentReply: "hello!",
			wantSessionID: "1.0", wantMessage: "hello!",
		},
		{
			name: "falls back to Timestamp when ThreadTimestamp is empty", threadTimestamp: "", agentReply: "hello!",
			wantSessionID: "1.1", wantMessage: "hello!",
		},
		{
			// AgentCoreのストリームからテキストを1件も抽出できなかった場合、空メッセージではなく代替文言を返信する。
			name: "substitutes a message when the agent reply is empty", threadTimestamp: "1.0", agentReply: "",
			wantSessionID: "1.0", wantMessage: domain.MessageNoResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notiRepo := &mockNotificationRepo{}
			agentRepo := &mockAgentRepo{Reply: tt.agentReply}
			u := &ChatAgent{
				AgentRepo:        agentRepo,
				NotificationRepo: notiRepo,
				AllowedChannel:   "C_ALLOWED",
				ReactionName:     "eyes",
			}

			err := u.Exec(context.Background(), domain.MentionEvent{
				Channel:         "C_ALLOWED",
				Timestamp:       "1.1",
				ThreadTimestamp: tt.threadTimestamp,
			})
			if err != nil {
				t.Fatalf("Exec() unexpected error: %v", err)
			}

			if len(agentRepo.callSessionIDs) != 1 || agentRepo.callSessionIDs[0] != tt.wantSessionID {
				t.Errorf("AgentRepo.Call sessionIDs = %v, want [%q]", agentRepo.callSessionIDs, tt.wantSessionID)
			}
			if len(notiRepo.postReplyCalls) != 1 {
				t.Fatalf("PostReply call count = %d, want 1", len(notiRepo.postReplyCalls))
			}
			if got := notiRepo.postReplyCalls[0].Message; got != tt.wantMessage {
				t.Errorf("PostReply message = %q, want %q", got, tt.wantMessage)
			}
		})
	}
}
