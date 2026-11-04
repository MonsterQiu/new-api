package ratio_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCompletionRatioInfoGPT55UsesOfficialOutputMultiplier(t *testing.T) {
	info := GetCompletionRatioInfo("gpt-5.5")

	if info.Ratio != 6 {
		t.Fatalf("gpt-5.5 completion ratio = %v, want 6", info.Ratio)
	}
	assert.False(t, info.Locked, "the upstream default permits administrator overrides")
	configured := 4.0
	assert.Equal(t, CompletionRatioInfo{Ratio: configured}, ResolveCompletionRatio("gpt-5.5", &configured))
}

func TestGetCompletionRatioGPT55DatedVariant(t *testing.T) {
	got := GetCompletionRatio("gpt-5.5-2026-04-24")

	if got != 6 {
		t.Fatalf("gpt-5.5 dated variant completion ratio = %v, want 6", got)
	}
}
