package model_setting

import (
	"net/http"
	"testing"
)

func TestClaudeSettingsWriteHeadersMergesConfiguredValuesIntoSingleHeader(t *testing.T) {
	settings := &ClaudeSettings{
		HeadersSettings: map[string]map[string][]string{
			"claude-3-7-sonnet-20250219-thinking": {
				"anthropic-beta": {
					"token-efficient-tools-2025-02-19",
				},
			},
		},
	}

	headers := http.Header{}
	headers.Set("anthropic-beta", "output-128k-2025-02-19")

	settings.WriteHeaders("claude-3-7-sonnet-20250219-thinking", &headers)

	got := headers.Values("anthropic-beta")
	if len(got) != 1 {
		t.Fatalf("expected a single merged header value, got %v", got)
	}
	expected := "output-128k-2025-02-19,token-efficient-tools-2025-02-19"
	if got[0] != expected {
		t.Fatalf("expected merged header %q, got %q", expected, got[0])
	}
}

func TestClaudeSettingsWriteHeadersDeduplicatesAcrossCommaSeparatedAndRepeatedValues(t *testing.T) {
	settings := &ClaudeSettings{
		HeadersSettings: map[string]map[string][]string{
			"claude-3-7-sonnet-20250219-thinking": {
				"anthropic-beta": {
					"token-efficient-tools-2025-02-19",
					"computer-use-2025-01-24",
				},
			},
		},
	}

	headers := http.Header{}
	headers.Add("anthropic-beta", "output-128k-2025-02-19, token-efficient-tools-2025-02-19")
	headers.Add("anthropic-beta", "token-efficient-tools-2025-02-19")

	settings.WriteHeaders("claude-3-7-sonnet-20250219-thinking", &headers)

	got := headers.Values("anthropic-beta")
	if len(got) != 1 {
		t.Fatalf("expected duplicate values to collapse into one header, got %v", got)
	}
	expected := "output-128k-2025-02-19,token-efficient-tools-2025-02-19,computer-use-2025-01-24"
	if got[0] != expected {
		t.Fatalf("expected deduplicated merged header %q, got %q", expected, got[0])
	}
}

func TestIsClaudeModelNameMatchesCommonClaudePrefixes(t *testing.T) {
	models := []string{
		"claude-3-5-sonnet-20240620",
		"claude-sonnet-4-20250514-thinking",
		"anthropic/claude-3.7-sonnet",
		"anthropic.claude-3-sonnet-20240229-v1:0",
	}
	for _, model := range models {
		if !IsClaudeModelName(model) {
			t.Fatalf("expected %q to be detected as Claude", model)
		}
	}
}

func TestClaudeExcludedSubscriptionPlanIDsForModelNormalizesConfiguredIDs(t *testing.T) {
	previous := claudeSettings
	t.Cleanup(func() {
		claudeSettings = previous
	})
	claudeSettings.ExcludedSubscriptionPlanIDs = []int{1, 0, 2, 2, -3, 7}

	got := GetClaudeExcludedSubscriptionPlanIDsForModel("claude-3-5-sonnet-20240620")

	expected := []int{1, 2, 7}
	if len(got) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	}
}

func TestClaudeExcludedSubscriptionPlanIDsForModelSkipsNonClaudeModels(t *testing.T) {
	previous := claudeSettings
	t.Cleanup(func() {
		claudeSettings = previous
	})
	claudeSettings.ExcludedSubscriptionPlanIDs = []int{1, 2, 7}

	if got := GetClaudeExcludedSubscriptionPlanIDsForModel("gpt-4o"); len(got) != 0 {
		t.Fatalf("expected non-Claude model to have no excluded plan IDs, got %v", got)
	}
}
