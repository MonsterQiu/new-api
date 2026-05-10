package openaicompat

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
)

func TestChatCompletionsRequestToResponsesRequestNormalizesStringMessages(t *testing.T) {
	stream := false
	req := &dto.GeneralOpenAIRequest{
		Model:  "gpt-5.4",
		Stream: &stream,
		Messages: []dto.Message{
			{Role: "system", Content: "you are helpful"},
			{Role: "user", Content: "hello"},
			{Role: "assistant", Content: "hi there"},
		},
	}

	respReq, err := ChatCompletionsRequestToResponsesRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var instructions string
	if err := common.Unmarshal(respReq.Instructions, &instructions); err != nil {
		t.Fatalf("failed to decode instructions: %v", err)
	}
	if instructions != "you are helpful" {
		t.Fatalf("instructions = %q, want %q", instructions, "you are helpful")
	}

	var inputItems []map[string]any
	if err := common.Unmarshal(respReq.Input, &inputItems); err != nil {
		t.Fatalf("failed to decode input items: %v", err)
	}
	if len(inputItems) != 2 {
		t.Fatalf("input item count = %d, want 2", len(inputItems))
	}

	assertMessagePart := func(item map[string]any, wantRole string, wantPartType string, wantText string) {
		if got := item["type"]; got != "message" {
			t.Fatalf("message type = %v, want %q", got, "message")
		}
		if got := item["role"]; got != wantRole {
			t.Fatalf("message role = %v, want %q", got, wantRole)
		}

		content, ok := item["content"].([]any)
		if !ok {
			t.Fatalf("message content type = %T, want []any", item["content"])
		}
		if len(content) != 1 {
			t.Fatalf("content part count = %d, want 1", len(content))
		}

		part, ok := content[0].(map[string]any)
		if !ok {
			t.Fatalf("content part type = %T, want map[string]any", content[0])
		}
		if got := part["type"]; got != wantPartType {
			t.Fatalf("content part type = %v, want %q", got, wantPartType)
		}
		if got := part["text"]; got != wantText {
			t.Fatalf("content part text = %v, want %q", got, wantText)
		}
	}

	assertMessagePart(inputItems[0], "user", "input_text", "hello")
	assertMessagePart(inputItems[1], "assistant", "output_text", "hi there")
}

func TestChatCompletionsRequestToResponsesRequestKeepsTypedFallbackForToolOutputWithoutCallID(t *testing.T) {
	req := &dto.GeneralOpenAIRequest{
		Model: "gpt-5.4",
		Messages: []dto.Message{
			{Role: "tool", Content: "tool said hi"},
		},
	}

	respReq, err := ChatCompletionsRequestToResponsesRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var inputItems []map[string]any
	if err := common.Unmarshal(respReq.Input, &inputItems); err != nil {
		t.Fatalf("failed to decode input items: %v", err)
	}
	if len(inputItems) != 1 {
		t.Fatalf("input item count = %d, want 1", len(inputItems))
	}

	item := inputItems[0]
	if got := item["type"]; got != "message" {
		t.Fatalf("message type = %v, want %q", got, "message")
	}
	if got := item["role"]; got != "user" {
		t.Fatalf("message role = %v, want %q", got, "user")
	}

	content, ok := item["content"].([]any)
	if !ok || len(content) != 1 {
		t.Fatalf("message content = %#v, want single typed part", item["content"])
	}
	part, ok := content[0].(map[string]any)
	if !ok {
		t.Fatalf("content part type = %T, want map[string]any", content[0])
	}
	if got := part["type"]; got != "input_text" {
		t.Fatalf("content part type = %v, want %q", got, "input_text")
	}
	if got := part["text"]; got != "[tool_output_missing_call_id] tool said hi" {
		t.Fatalf("content part text = %v, want fallback tool output text", got)
	}
}

func TestChatCompletionsRequestToResponsesRequestPreservesPromptCacheFields(t *testing.T) {
	req := &dto.GeneralOpenAIRequest{
		Model:                "gpt-5.4",
		PromptCacheKey:       "cline-session-123",
		PromptCacheRetention: []byte(`"24h"`),
		Messages: []dto.Message{
			{Role: "user", Content: "hello"},
		},
	}

	respReq, err := ChatCompletionsRequestToResponsesRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var cacheKey string
	if err := common.Unmarshal(respReq.PromptCacheKey, &cacheKey); err != nil {
		t.Fatalf("failed to decode prompt_cache_key: %v", err)
	}
	if cacheKey != "cline-session-123" {
		t.Fatalf("prompt_cache_key = %q, want %q", cacheKey, "cline-session-123")
	}

	var retention string
	if err := common.Unmarshal(respReq.PromptCacheRetention, &retention); err != nil {
		t.Fatalf("failed to decode prompt_cache_retention: %v", err)
	}
	if retention != "24h" {
		t.Fatalf("prompt_cache_retention = %q, want %q", retention, "24h")
	}
}
