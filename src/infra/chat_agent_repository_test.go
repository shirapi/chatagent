package infra

import (
	"io"
	"regexp"
	"strings"
	"testing"
)

var runtimeSessionIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-_]*$`)

func TestToRuntimeSessionID(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
	}{
		{name: "slack thread timestamp", sessionID: "1234567890.123456"},
		{name: "short string", sessionID: "x"},
		{name: "empty string", sessionID: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toRuntimeSessionID(tt.sessionID)

			if len(got) < minRuntimeSessionIDLength {
				t.Errorf("toRuntimeSessionID(%q) = %q, length %d, want >= %d", tt.sessionID, got, len(got), minRuntimeSessionIDLength)
			}
			if !runtimeSessionIDPattern.MatchString(got) {
				t.Errorf("toRuntimeSessionID(%q) = %q, does not match required pattern", tt.sessionID, got)
			}
			if strings.Contains(got, ".") {
				t.Errorf("toRuntimeSessionID(%q) = %q, contains a dot", tt.sessionID, got)
			}
		})
	}
}

func TestExtractText(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "concatenates multiple contentBlockDelta events",
			body: "data: {\"event\":{\"contentBlockDelta\":{\"delta\":{\"text\":\"Hello\"}}}}\n" +
				"data: {\"event\":{\"contentBlockDelta\":{\"delta\":{\"text\":\" World\"}}}}\n",
			want: "Hello World",
		},
		{
			name: "ignores lines without data prefix",
			body: "event: message\n" +
				"data: {\"event\":{\"contentBlockDelta\":{\"delta\":{\"text\":\"Hi\"}}}}\n",
			want: "Hi",
		},
		{
			name: "ignores malformed JSON lines",
			body: "data: not-json\n" +
				"data: {\"event\":{\"contentBlockDelta\":{\"delta\":{\"text\":\"Hi\"}}}}\n",
			want: "Hi",
		},
		{
			name: "no matching events returns empty string",
			body: "data: {\"event\":{\"messageStart\":{}}}\n",
			want: "",
		},
		{
			name: "empty stream returns empty string",
			body: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractText(io.NopCloser(strings.NewReader(tt.body)))
			if err != nil {
				t.Fatalf("extractText() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("extractText() = %q, want %q", got, tt.want)
			}
		})
	}
}
