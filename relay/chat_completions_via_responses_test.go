package relay

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDetectResponsesEventStreamFromBodyPrefix(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader("event: response.created\ndata: {\"type\":\"response.created\"}\n")),
	}

	if !detectResponsesEventStream(resp) {
		t.Fatalf("expected SSE body detection to succeed")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read preserved response body: %v", err)
	}
	if !strings.HasPrefix(string(body), "event: response.created") {
		t.Fatalf("expected response body to remain readable, got %q", string(body))
	}
}

func TestDetectResponsesEventStreamReturnsFalseForJSONBody(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader("{\"id\":\"resp_1\"}")),
	}

	if detectResponsesEventStream(resp) {
		t.Fatalf("expected plain JSON body to stay non-stream")
	}
}

func TestIsResponsesEventStreamContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{name: "plain", contentType: "text/event-stream", want: true},
		{name: "mixed case with charset", contentType: "Text/Event-Stream; charset=utf-8", want: true},
		{name: "json", contentType: "application/json", want: false},
		{name: "empty", contentType: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isResponsesEventStreamContentType(tt.contentType); got != tt.want {
				t.Fatalf("isResponsesEventStreamContentType(%q) = %v, want %v", tt.contentType, got, tt.want)
			}
		})
	}
}
