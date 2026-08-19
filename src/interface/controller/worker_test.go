package controller

import (
	"chatagent/domain"
	"context"
	"errors"
	"testing"
)

func TestWorkerExec(t *testing.T) {
	mention := domain.MentionEvent{Channel: "C123", UserID: "U123", Timestamp: "1.1", Text: "hi"}
	payload := []byte(`{"Channel":"C123","UserID":"U123","Timestamp":"1.1","Text":"hi"}`)

	tests := []struct {
		name    string
		payload []byte
		execErr error
		wantErr bool
	}{
		{name: "usecase succeeds", payload: payload, execErr: nil, wantErr: false},
		{name: "usecase fails", payload: payload, execErr: errors.New("boom"), wantErr: true},
		{name: "invalid payload", payload: []byte("not json"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockUsecase{ExecErr: tt.execErr}
			c := NewWorker(uc)

			err := c.Exec(context.Background(), tt.payload)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Exec() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.name == "invalid payload" {
				if len(uc.execCalls) != 0 {
					t.Errorf("Usecase.Exec should not be called, but was called %d times", len(uc.execCalls))
				}
				return
			}
			if len(uc.execCalls) != 1 || uc.execCalls[0] != mention {
				t.Errorf("Usecase.Exec calls = %+v, want [%+v]", uc.execCalls, mention)
			}
		})
	}
}
