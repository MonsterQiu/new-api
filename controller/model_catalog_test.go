package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
)

func TestOpenAIModelCatalogIncludesGPT55(t *testing.T) {
	if _, ok := openAIModelsMap["gpt-5.5"]; !ok {
		t.Fatal("expected gpt-5.5 to be exported in the OpenAI model catalog")
	}
}

func TestCodexChannelModelsIncludeGPT55AndCompactVariant(t *testing.T) {
	codexModels, ok := channelId2Models[constant.ChannelTypeCodex]
	if !ok {
		t.Fatal("expected codex channel model catalog to be initialized")
	}

	wantModels := []string{
		"gpt-5.5",
		ratio_setting.WithCompactModelSuffix("gpt-5.5"),
	}
	for _, want := range wantModels {
		found := false
		for _, got := range codexModels {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected codex channel model catalog to include %s", want)
		}
	}
}
