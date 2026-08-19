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

func TestExec_ChannelNotAllowed(t *testing.T) {
	notiRepo := &mockNotificationRepo{}
	agentRepo := &mockAgentRepo{}
	u := &ChatAgent{
		AgentRepo:        agentRepo,
		NotificationRepo: notiRepo,
		AllowedChannel:   "C_ALLOWED",
		ReactionName:     "eyes",
	}

	err := u.Exec(context.Background(), domain.MentionEvent{Channel: "C_OTHER", Timestamp: "1.1"})
	if err != nil {
		t.Fatalf("Exec() unexpected error: %v", err)
	}

	if len(notiRepo.postReplyCalls) != 1 {
		t.Fatalf("PostReply call count = %d, want 1", len(notiRepo.postReplyCalls))
	}
	if got := notiRepo.postReplyCalls[0].Message; got != domain.MessageInvalidChannel {
		t.Errorf("PostReply message = %q, want %q", got, domain.MessageInvalidChannel)
	}
	if len(notiRepo.addReactionCalls) != 0 {
		t.Errorf("AddReaction should not be called, but was called %d times", len(notiRepo.addReactionCalls))
	}
	if len(agentRepo.callSessionIDs) != 0 {
		t.Errorf("AgentRepo.Call should not be called, but was called %d times", len(agentRepo.callSessionIDs))
	}
}

func TestExec_AddReactionFails(t *testing.T) {
	wantErr := errors.New("add reaction failed")
	notiRepo := &mockNotificationRepo{AddReactionErr: wantErr}
	agentRepo := &mockAgentRepo{}
	u := &ChatAgent{
		AgentRepo:        agentRepo,
		NotificationRepo: notiRepo,
		AllowedChannel:   "C_ALLOWED",
		ReactionName:     "eyes",
	}

	err := u.Exec(context.Background(), domain.MentionEvent{Channel: "C_ALLOWED", Timestamp: "1.1"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Exec() error = %v, want %v", err, wantErr)
	}
	if len(agentRepo.callSessionIDs) != 0 {
		t.Errorf("AgentRepo.Call should not be called, but was called %d times", len(agentRepo.callSessionIDs))
	}
}

func TestExec_AgentCallFails(t *testing.T) {
	wantErr := errors.New("agent call failed")
	notiRepo := &mockNotificationRepo{}
	agentRepo := &mockAgentRepo{Err: wantErr}
	u := &ChatAgent{
		AgentRepo:        agentRepo,
		NotificationRepo: notiRepo,
		AllowedChannel:   "C_ALLOWED",
		ReactionName:     "eyes",
	}

	err := u.Exec(context.Background(), domain.MentionEvent{Channel: "C_ALLOWED", Timestamp: "1.1"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Exec() error = %v, want %v", err, wantErr)
	}
	// AddReaction should still have been called before the failure.
	if len(notiRepo.addReactionCalls) != 1 {
		t.Errorf("AddReaction call count = %d, want 1", len(notiRepo.addReactionCalls))
	}
	// PostReply should not be called since Call failed.
	if len(notiRepo.postReplyCalls) != 0 {
		t.Errorf("PostReply should not be called, but was called %d times", len(notiRepo.postReplyCalls))
	}
}

func TestExec_Success(t *testing.T) {
	tests := []struct {
		name            string
		threadTimestamp string
		wantSessionID   string
	}{
		{name: "uses ThreadTimestamp when present", threadTimestamp: "1.0", wantSessionID: "1.0"},
		{name: "falls back to Timestamp when ThreadTimestamp is empty", threadTimestamp: "", wantSessionID: "1.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notiRepo := &mockNotificationRepo{}
			agentRepo := &mockAgentRepo{Reply: "hello!"}
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
			if got := notiRepo.postReplyCalls[0].Message; got != "hello!" {
				t.Errorf("PostReply message = %q, want %q", got, "hello!")
			}
		})
	}
}
