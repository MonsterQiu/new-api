package openaicompat

import (
	"testing"

	"github.com/QuantumNous/new-api/setting/model_setting"
)

func TestResolveChatCompletionsToResponsesPolicyHonorsGroupAndStreamScope(t *testing.T) {
	policy := model_setting.ChatCompletionsToResponsesPolicy{
		Enabled:             true,
		AllChannels:         true,
		ModelPatterns:       []string{`^gpt-5\.4.*$`},
		Groups:              []string{"legacy_nonstream"},
		OnlyNonStream:       true,
		ForceUpstreamStream: true,
	}

	decision := ResolveChatCompletionsToResponsesPolicy(policy, 1, 1, "gpt-5.4", "legacy_nonstream", false)
	if !decision.UseResponses {
		t.Fatalf("expected policy to enable responses compatibility")
	}
	if !decision.ForceUpstreamStream {
		t.Fatalf("expected policy to force upstream stream for non-stream client")
	}

	streamDecision := ResolveChatCompletionsToResponsesPolicy(policy, 1, 1, "gpt-5.4", "legacy_nonstream", true)
	if !streamDecision.UseResponses {
		t.Fatalf("expected streaming requests to keep using responses compatibility")
	}
	if streamDecision.ForceUpstreamStream {
		t.Fatalf("expected streaming requests to skip force-upstream-stream")
	}

	groupMiss := ResolveChatCompletionsToResponsesPolicy(policy, 1, 1, "gpt-5.4", "default", false)
	if !groupMiss.UseResponses {
		t.Fatalf("expected group mismatch to keep base responses compatibility")
	}
	if groupMiss.ForceUpstreamStream {
		t.Fatalf("expected group mismatch to skip force-upstream-stream")
	}
}

func TestResolveChatCompletionsToResponsesPolicyPreservesLegacyBehavior(t *testing.T) {
	policy := model_setting.ChatCompletionsToResponsesPolicy{
		Enabled:       true,
		AllChannels:   true,
		ModelPatterns: []string{`^gpt-5.*$`},
	}

	decision := ResolveChatCompletionsToResponsesPolicy(policy, 7, 1, "gpt-5.4-mini", "any-group", true)
	if !decision.UseResponses {
		t.Fatalf("expected legacy policy without group restrictions to continue matching")
	}
	if decision.ForceUpstreamStream {
		t.Fatalf("expected upstream stream forcing to stay disabled by default")
	}
}

func TestResolveClaudeMessagesToResponsesPolicyScopesUseResponsesByGroup(t *testing.T) {
	policy := model_setting.ClaudeMessagesToResponsesPolicy{
		Enabled:             true,
		AllChannels:         true,
		ModelPatterns:       []string{`^claude-opus-4\.7$`, `^claude-sonnet-4\.6$`},
		Groups:              []string{"claude_code"},
		OnlyNonStream:       true,
		ForceUpstreamStream: true,
	}

	decision := ResolveClaudeMessagesToResponsesPolicy(policy, 1, 1, "claude-opus-4.7", "claude_code", false)
	if !decision.UseResponses {
		t.Fatalf("expected Claude Messages policy to enable responses compatibility")
	}
	if !decision.ForceUpstreamStream {
		t.Fatalf("expected non-stream Claude Code request to force upstream stream")
	}

	streamDecision := ResolveClaudeMessagesToResponsesPolicy(policy, 1, 1, "claude-sonnet-4.6", "claude_code", true)
	if !streamDecision.UseResponses {
		t.Fatalf("expected streaming Claude Code request to keep responses compatibility")
	}
	if streamDecision.ForceUpstreamStream {
		t.Fatalf("expected streaming request to skip force-upstream-stream")
	}

	groupMiss := ResolveClaudeMessagesToResponsesPolicy(policy, 1, 1, "claude-opus-4.7", "default", false)
	if groupMiss.UseResponses {
		t.Fatalf("expected group mismatch to skip Claude Messages responses compatibility")
	}

	modelMiss := ResolveClaudeMessagesToResponsesPolicy(policy, 1, 1, "gpt-5.5", "claude_code", false)
	if modelMiss.UseResponses {
		t.Fatalf("expected model mismatch to skip Claude Messages responses compatibility")
	}
}
