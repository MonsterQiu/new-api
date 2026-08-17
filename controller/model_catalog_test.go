package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
)

func TestOpenAIModelCatalogIncludesGPT55(t *testing.T) {
	if _, ok := openAIModelsMap["gpt-5.5"]; !ok {
		t.Fatal("expected gpt-5.5 to be exported in the OpenAI model catalog")
	}
}

func TestCodexChannelModelsIncludeGPT55(t *testing.T) {
	codexModels, ok := channelId2Models[constant.ChannelTypeCodex]
	if !ok {
		t.Fatal("expected codex channel model catalog to be initialized")
	}

	for _, got := range codexModels {
		if got == "gpt-5.5" {
			return
		}
	}
	t.Fatal("expected codex channel model catalog to include gpt-5.5")
}
