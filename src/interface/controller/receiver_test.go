package controller

import (
	"chatagent/domain"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockUsecase struct {
	VerifyResult domain.VerifyResult
	VerifyErr    error
	AcceptErr    error
	ExecErr      error

	acceptCalls []domain.MentionEvent
	execCalls   []domain.MentionEvent
}

func (m *mockUsecase) Verify(r *http.Request) (domain.VerifyResult, error) {
	return m.VerifyResult, m.VerifyErr
}

func (m *mockUsecase) Accept(ctx context.Context, mention domain.MentionEvent) error {
	m.acceptCalls = append(m.acceptCalls, mention)
	return m.AcceptErr
}

func (m *mockUsecase) Exec(ctx context.Context, mention domain.MentionEvent) error {
	m.execCalls = append(m.execCalls, mention)
	return m.ExecErr
}

func TestExec_VerifyFails(t *testing.T) {
	uc := &mockUsecase{VerifyErr: errors.New("invalid signature")}
	c := NewReceiver("prod", uc)

	res, err := c.Exec(context.Background(), httptest.NewRequest(http.MethodPost, "/", nil))
	if err != nil {
		t.Fatalf("Exec() unexpected error: %v", err)
	}
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", res.StatusCode, http.StatusUnauthorized)
	}
	if len(uc.acceptCalls) != 0 {
		t.Errorf("Usecase.Accept should not be called, but was called %d times", len(uc.acceptCalls))
	}
}

func TestExec_Challenge(t *testing.T) {
	uc := &mockUsecase{VerifyResult: domain.VerifyResult{Challenge: "abc123"}}
	c := NewReceiver("prod", uc)

	res, err := c.Exec(context.Background(), httptest.NewRequest(http.MethodPost, "/", nil))
	if err != nil {
		t.Fatalf("Exec() unexpected error: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", res.StatusCode, http.StatusOK)
	}
	if res.Body != "abc123" {
		t.Errorf("Body = %q, want %q", res.Body, "abc123")
	}
	if len(uc.acceptCalls) != 0 {
		t.Errorf("Usecase.Accept should not be called, but was called %d times", len(uc.acceptCalls))
	}
}

func TestExec_NoMentionNoChallenge(t *testing.T) {
	uc := &mockUsecase{VerifyResult: domain.VerifyResult{}}
	c := NewReceiver("prod", uc)

	res, err := c.Exec(context.Background(), httptest.NewRequest(http.MethodPost, "/", nil))
	if err != nil {
		t.Fatalf("Exec() unexpected error: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", res.StatusCode, http.StatusOK)
	}
	if len(uc.acceptCalls) != 0 {
		t.Errorf("Usecase.Accept should not be called, but was called %d times", len(uc.acceptCalls))
	}
}

func TestExec_Mention(t *testing.T) {
	mention := domain.MentionEvent{Channel: "C123", UserID: "U123", Timestamp: "1.1", Text: "hi"}

	tests := []struct {
		name       string
		acceptErr  error
		wantStatus int
	}{
		{name: "usecase succeeds", acceptErr: nil, wantStatus: http.StatusOK},
		{name: "usecase fails", acceptErr: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockUsecase{
				VerifyResult: domain.VerifyResult{Mention: &mention},
				AcceptErr:    tt.acceptErr,
			}
			c := NewReceiver("prod", uc)

			res, err := c.Exec(context.Background(), httptest.NewRequest(http.MethodPost, "/", nil))
			if err != nil {
				t.Fatalf("Exec() unexpected error: %v", err)
			}
			if res.StatusCode != tt.wantStatus {
				t.Errorf("StatusCode = %d, want %d", res.StatusCode, tt.wantStatus)
			}
			if len(uc.acceptCalls) != 1 || uc.acceptCalls[0] != mention {
				t.Errorf("Usecase.Accept calls = %+v, want [%+v]", uc.acceptCalls, mention)
			}
		})
	}
}
