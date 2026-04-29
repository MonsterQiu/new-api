package relay

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"

	"github.com/gin-gonic/gin"
)

func TestInjectClaudeMessagesPromptCacheKeyDerivesStableSessionKey(t *testing.T) {
	req1 := claudePromptCacheTestRequest("inspect this repo", "tool result A")
	req2 := claudePromptCacheTestRequest("inspect this repo", "tool result B")

	responsesReq1 := &dto.OpenAIResponsesRequest{Model: "gpt-5.4"}
	responsesReq2 := &dto.OpenAIResponsesRequest{Model: "gpt-5.4"}
	info := &relaycommon.RelayInfo{UserId: 42}

	injectClaudeMessagesPromptCacheKey(nil, info, req1, responsesReq1)
	injectClaudeMessagesPromptCacheKey(nil, info, req2, responsesReq2)

	key1 := promptCacheKeyString(t, responsesReq1)
	key2 := promptCacheKeyString(t, responsesReq2)
	if key1 == "" || !strings.HasPrefix(key1, "claude_cc_") {
		t.Fatalf("expected derived claude_cc prompt_cache_key, got %q", key1)
	}
	if key1 != key2 {
		t.Fatalf("expected later dynamic messages not to change prompt_cache_key, got %q and %q", key1, key2)
	}

	req3 := claudePromptCacheTestRequest("summarize this repo", "tool result A")
	responsesReq3 := &dto.OpenAIResponsesRequest{Model: "gpt-5.4"}
	injectClaudeMessagesPromptCacheKey(nil, info, req3, responsesReq3)
	key3 := promptCacheKeyString(t, responsesReq3)
	if key3 == key1 {
		t.Fatalf("expected different first user prompt to change prompt_cache_key, got %q", key3)
	}
}

func TestInjectClaudeMessagesPromptCacheKeyPreservesClientHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	c.Request.Header.Set("session_id", "client-session-123")

	responsesReq := &dto.OpenAIResponsesRequest{Model: "gpt-5.4"}
	injectClaudeMessagesPromptCacheKey(c, &relaycommon.RelayInfo{UserId: 42}, claudePromptCacheTestRequest("inspect", "result"), responsesReq)

	if got := promptCacheKeyString(t, responsesReq); got != "client-session-123" {
		t.Fatalf("expected client header to win, got %q", got)
	}
}

func TestInjectClaudeMessagesPromptCacheKeySkipsNonCodexFamilyModels(t *testing.T) {
	responsesReq := &dto.OpenAIResponsesRequest{Model: "gpt-4o"}
	injectClaudeMessagesPromptCacheKey(nil, &relaycommon.RelayInfo{UserId: 42}, claudePromptCacheTestRequest("inspect", "result"), responsesReq)

	if len(responsesReq.PromptCacheKey) != 0 {
		t.Fatalf("expected no auto prompt_cache_key for non gpt-5/codex model, got %s", string(responsesReq.PromptCacheKey))
	}
}

func claudePromptCacheTestRequest(firstUser string, laterToolResult string) *dto.ClaudeRequest {
	return &dto.ClaudeRequest{
		Model:  "gpt-5.4",
		System: "You are a coding assistant.",
		Tools: []map[string]any{
			{
				"name":        "read_file",
				"description": "Read a file",
				"input_schema": map[string]any{
					"type":       "object",
					"properties": map[string]any{"path": map[string]any{"type": "string"}},
				},
			},
		},
		ToolChoice: map[string]any{"type": "auto"},
		Messages: []dto.ClaudeMessage{
			{Role: "user", Content: firstUser},
			{Role: "assistant", Content: "I will inspect the project."},
			{Role: "user", Content: laterToolResult},
		},
	}
}

func promptCacheKeyString(t *testing.T, request *dto.OpenAIResponsesRequest) string {
	t.Helper()
	var key string
	if err := common.Unmarshal(request.PromptCacheKey, &key); err != nil {
		t.Fatalf("unmarshal prompt_cache_key: %v", err)
	}
	return key
}
