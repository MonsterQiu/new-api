package openai

import (
	"testing"

	"github.com/QuantumNous/new-api/dto"
)

func TestConvertOpenAIResponsesRequestStripsGPT5UnsupportedSamplingParams(t *testing.T) {
	temp := 0.7
	topP := 0.95

	converted, err := (&Adaptor{}).ConvertOpenAIResponsesRequest(nil, nil, dto.OpenAIResponsesRequest{
		Model:       "gpt-5.5-high",
		Temperature: &temp,
		TopP:        &topP,
	})
	if err != nil {
		t.Fatalf("ConvertOpenAIResponsesRequest returned error: %v", err)
	}

	request, ok := converted.(dto.OpenAIResponsesRequest)
	if !ok {
		t.Fatalf("converted request type = %T, want dto.OpenAIResponsesRequest", converted)
	}
	if request.Model != "gpt-5.5" {
		t.Fatalf("model = %q, want gpt-5.5", request.Model)
	}
	if request.Reasoning == nil || request.Reasoning.Effort != "high" {
		t.Fatalf("reasoning = %+v, want high effort", request.Reasoning)
	}
	if request.Temperature != nil {
		t.Fatalf("expected temperature to be stripped")
	}
	if request.TopP != nil {
		t.Fatalf("expected top_p to be stripped")
	}
}
