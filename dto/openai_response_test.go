package dto

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
)

func TestResponsesStreamResponseAcceptsObjectArguments(t *testing.T) {
	var streamResp ResponsesStreamResponse
	err := common.UnmarshalJsonStr(`{"type":"response.output_item.added","item":{"type":"function_call","id":"item_1","call_id":"call_1","name":"search","arguments":{"q":"hello"}}}`, &streamResp)
	if err != nil {
		t.Fatalf("unmarshal responses stream response: %v", err)
	}
	if streamResp.Item == nil {
		t.Fatalf("expected response item")
	}
	if got := streamResp.Item.ArgumentsString(); got != `{"q":"hello"}` {
		t.Fatalf("unexpected arguments: %q", got)
	}
}

func TestResponsesStreamResponseAcceptsStringArguments(t *testing.T) {
	var streamResp ResponsesStreamResponse
	err := common.UnmarshalJsonStr(`{"type":"response.output_item.added","item":{"type":"function_call","id":"item_1","call_id":"call_1","name":"search","arguments":"{\"q\":\"hello\"}"}}`, &streamResp)
	if err != nil {
		t.Fatalf("unmarshal responses stream response: %v", err)
	}
	if streamResp.Item == nil {
		t.Fatalf("expected response item")
	}
	if got := streamResp.Item.ArgumentsString(); got != `{"q":"hello"}` {
		t.Fatalf("unexpected arguments: %q", got)
	}
}
