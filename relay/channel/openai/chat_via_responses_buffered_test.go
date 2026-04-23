package openai

import (
	"strings"
	"testing"
)

func TestCollectResponsesStreamResponseUsesCompletedPayload(t *testing.T) {
	body := strings.NewReader(strings.Join([]string{
		`data: {"type":"response.created","response":{"id":"resp_1","model":"gpt-5.4","created_at":123}}`,
		`data: {"type":"response.output_text.delta","delta":"Hello "}`,
		`data: {"type":"response.output_text.delta","delta":"world"}`,
		`data: {"type":"response.completed","response":{"id":"resp_1","object":"response","created_at":123,"model":"gpt-5.4","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"Hello world"}]}],"usage":{"input_tokens":10,"output_tokens":5,"total_tokens":15,"prompt_tokens_details":{},"completion_tokens_details":{},"input_tokens_details":{"cached_tokens":0,"text_tokens":0,"audio_tokens":0,"image_tokens":0}}}}`,
		`data: [DONE]`,
	}, "\n"))

	resp, fallbackText, err := collectResponsesStreamResponse(body, "fallback-model")
	if err != nil {
		t.Fatalf("collectResponsesStreamResponse returned error: %v", err)
	}
	if resp == nil {
		t.Fatalf("expected response payload")
	}
	if resp.Model != "gpt-5.4" {
		t.Fatalf("unexpected model: %s", resp.Model)
	}
	if fallbackText != "Hello world" {
		t.Fatalf("unexpected fallback text: %q", fallbackText)
	}
	if len(resp.Output) != 1 || len(resp.Output[0].Content) != 1 || resp.Output[0].Content[0].Text != "Hello world" {
		t.Fatalf("unexpected output payload: %#v", resp.Output)
	}
}

func TestCollectResponsesStreamResponseSynthesizesToolCallsWhenCompletedPayloadIsMissing(t *testing.T) {
	body := strings.NewReader(strings.Join([]string{
		`data: {"type":"response.created","response":{"id":"resp_2","model":"gpt-5.4","created_at":321}}`,
		`data: {"type":"response.output_item.added","item":{"type":"function_call","id":"item_1","call_id":"call_1","name":"search","arguments":"{\"q\":\"he"}}`,
		`data: {"type":"response.function_call_arguments.delta","item_id":"item_1","delta":"llo\"}"}`,
		`data: [DONE]`,
	}, "\n"))

	resp, fallbackText, err := collectResponsesStreamResponse(body, "fallback-model")
	if err != nil {
		t.Fatalf("collectResponsesStreamResponse returned error: %v", err)
	}
	if resp == nil {
		t.Fatalf("expected synthesized response payload")
	}
	if fallbackText != "" {
		t.Fatalf("expected empty fallback text, got %q", fallbackText)
	}
	if len(resp.Output) != 1 {
		t.Fatalf("expected one synthesized tool call, got %#v", resp.Output)
	}
	if resp.Output[0].Type != "function_call" || resp.Output[0].Name != "search" {
		t.Fatalf("unexpected synthesized tool call: %#v", resp.Output[0])
	}
	if resp.Output[0].Arguments != "{\"q\":\"hello\"}" {
		t.Fatalf("unexpected synthesized arguments: %q", resp.Output[0].Arguments)
	}
}
