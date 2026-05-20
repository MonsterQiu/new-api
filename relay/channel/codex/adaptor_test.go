package codex

import (
	"testing"

	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
)

func TestConvertOpenAIResponsesRequestStripsCodexUnsupportedParams(t *testing.T) {
	temp := 0.7
	topP := 0.95
	maxOutputTokens := uint(1024)

	converted, err := (&Adaptor{}).ConvertOpenAIResponsesRequest(nil, &relaycommon.RelayInfo{
		RelayMode: relayconstant.RelayModeResponses,
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelSetting: dto.ChannelSettings{},
		},
	}, dto.OpenAIResponsesRequest{
		Model:           "gpt-5.5",
		MaxOutputTokens: &maxOutputTokens,
		Temperature:     &temp,
		TopP:            &topP,
	})
	if err != nil {
		t.Fatalf("ConvertOpenAIResponsesRequest returned error: %v", err)
	}

	request, ok := converted.(dto.OpenAIResponsesRequest)
	if !ok {
		t.Fatalf("converted request type = %T, want dto.OpenAIResponsesRequest", converted)
	}
	if request.MaxOutputTokens != nil {
		t.Fatalf("expected max_output_tokens to be stripped")
	}
	if request.Temperature != nil {
		t.Fatalf("expected temperature to be stripped")
	}
	if request.TopP != nil {
		t.Fatalf("expected top_p to be stripped")
	}
	if string(request.Store) != "false" {
		t.Fatalf("store = %s, want false", string(request.Store))
	}
	if len(request.Instructions) == 0 {
		t.Fatalf("expected instructions to be defaulted for codex")
	}
}
