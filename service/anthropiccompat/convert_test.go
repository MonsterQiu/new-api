package anthropiccompat

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
)

func TestClaudeRequestToResponsesRequestConvertsToolsAndToolMessages(t *testing.T) {
	toolInput := map[string]any{"path": "main.go"}
	toolUse := dto.ClaudeMediaMessage{
		Type:  "tool_use",
		Id:    "toolu_1",
		Name:  "read_file",
		Input: toolInput,
	}
	toolResult := dto.ClaudeMediaMessage{
		Type:      "tool_result",
		ToolUseId: "toolu_1",
		Content:   "package main",
	}
	req := &dto.ClaudeRequest{
		Model:  "gpt-5.4",
		System: "You are concise.",
		Tools: []map[string]any{
			{
				"name":        "read_file",
				"description": "Read a file",
				"input_schema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{"type": "string"},
					},
				},
			},
		},
		Messages: []dto.ClaudeMessage{
			{Role: "user", Content: "Inspect the file."},
			{Role: "assistant", Content: []dto.ClaudeMediaMessage{toolUse}},
			{Role: "user", Content: []dto.ClaudeMediaMessage{toolResult}},
		},
	}

	responsesReq, err := ClaudeRequestToResponsesRequest(req)
	if err != nil {
		t.Fatalf("ClaudeRequestToResponsesRequest returned error: %v", err)
	}

	var instructions string
	if err := common.Unmarshal(responsesReq.Instructions, &instructions); err != nil {
		t.Fatalf("unmarshal instructions: %v", err)
	}
	if instructions != "You are concise." {
		t.Fatalf("unexpected instructions: %q", instructions)
	}

	var tools []map[string]any
	if err := common.Unmarshal(responsesReq.Tools, &tools); err != nil {
		t.Fatalf("unmarshal tools: %v", err)
	}
	if len(tools) != 1 || tools[0]["type"] != "function" || tools[0]["name"] != "read_file" {
		t.Fatalf("unexpected tools: %#v", tools)
	}

	var input []map[string]any
	if err := common.Unmarshal(responsesReq.Input, &input); err != nil {
		t.Fatalf("unmarshal input: %v", err)
	}
	if len(input) != 3 {
		t.Fatalf("expected 3 input items, got %d: %#v", len(input), input)
	}
	if input[1]["type"] != "function_call" || input[1]["call_id"] != "toolu_1" || input[1]["name"] != "read_file" {
		t.Fatalf("unexpected function_call item: %#v", input[1])
	}
	if input[1]["arguments"] != `{"path":"main.go"}` {
		t.Fatalf("unexpected function_call arguments: %#v", input[1]["arguments"])
	}
	if input[2]["type"] != "function_call_output" || input[2]["call_id"] != "toolu_1" || input[2]["output"] != "package main" {
		t.Fatalf("unexpected function_call_output item: %#v", input[2])
	}
}

func TestResponsesResponseToClaudeResponseConvertsFunctionCallAndCacheUsage(t *testing.T) {
	resp := &dto.OpenAIResponsesResponse{
		ID:    "resp_1",
		Model: "gpt-5.5",
		Output: []dto.ResponsesOutput{
			{
				Type:      "function_call",
				ID:        "fc_1",
				CallId:    "call_1",
				Name:      "edit_file",
				Arguments: `{"path":"main.go"}`,
			},
		},
		Usage: &dto.Usage{
			InputTokens:  100,
			OutputTokens: 25,
			InputTokensDetails: &dto.InputTokenDetails{
				CachedTokens:         30,
				CachedCreationTokens: 10,
			},
		},
	}

	claudeResp, usage, err := ResponsesResponseToClaudeResponse(resp, "")
	if err != nil {
		t.Fatalf("ResponsesResponseToClaudeResponse returned error: %v", err)
	}
	if claudeResp.StopReason != "tool_use" {
		t.Fatalf("expected tool_use stop reason, got %q", claudeResp.StopReason)
	}
	if len(claudeResp.Content) != 1 || claudeResp.Content[0].Type != "tool_use" {
		t.Fatalf("unexpected Claude content: %#v", claudeResp.Content)
	}
	if claudeResp.Usage.InputTokens != 60 {
		t.Fatalf("expected non-cached input_tokens=60, got %d", claudeResp.Usage.InputTokens)
	}
	if claudeResp.Usage.CacheReadInputTokens != 30 {
		t.Fatalf("expected cache_read_input_tokens=30, got %d", claudeResp.Usage.CacheReadInputTokens)
	}
	if claudeResp.Usage.CacheCreationInputTokens != 10 {
		t.Fatalf("expected cache_creation_input_tokens=10, got %d", claudeResp.Usage.CacheCreationInputTokens)
	}
	if usage == nil || usage.UsageSemantic != UsageSemanticAnthropic || usage.PromptTokens != 60 {
		t.Fatalf("unexpected billing usage: %#v", usage)
	}
}
