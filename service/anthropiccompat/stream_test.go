package anthropiccompat

import (
	"testing"

	"github.com/QuantumNous/new-api/dto"
)

func TestResponsesStreamToClaudeEventsConvertsFunctionCallArgumentDeltas(t *testing.T) {
	state := NewResponsesToClaudeStreamState("msg_1", "gpt-5.4", 12)

	events, usage := state.EventsForStreamResponse(dto.ResponsesStreamResponse{
		Type: "response.output_item.added",
		Item: &dto.ResponsesOutput{
			Type:   "function_call",
			ID:     "item_1",
			CallId: "call_1",
			Name:   "read_file",
		},
	})
	if usage != nil {
		t.Fatalf("did not expect usage before completion")
	}
	if len(events) != 2 {
		t.Fatalf("expected message_start and content_block_start, got %d: %#v", len(events), events)
	}
	if events[0].Type != "message_start" {
		t.Fatalf("expected message_start, got %s", events[0].Type)
	}
	if events[1].Type != "content_block_start" || events[1].ContentBlock == nil || events[1].ContentBlock.Type != "tool_use" {
		t.Fatalf("unexpected tool start event: %#v", events[1])
	}

	events, usage = state.EventsForStreamResponse(dto.ResponsesStreamResponse{
		Type:   "response.function_call_arguments.delta",
		ItemID: "item_1",
		Delta:  `{"path"`,
	})
	if usage != nil {
		t.Fatalf("did not expect usage before completion")
	}
	if len(events) != 1 || events[0].Delta == nil || events[0].Delta.Type != "input_json_delta" {
		t.Fatalf("expected input_json_delta event, got %#v", events)
	}
	if events[0].Delta.PartialJson == nil || *events[0].Delta.PartialJson != `{"path"` {
		t.Fatalf("unexpected partial json: %#v", events[0].Delta.PartialJson)
	}

	events, usage = state.EventsForStreamResponse(dto.ResponsesStreamResponse{
		Type: "response.completed",
		Response: &dto.OpenAIResponsesResponse{
			ID:    "resp_1",
			Model: "gpt-5.4",
			Usage: &dto.Usage{
				InputTokens:  20,
				OutputTokens: 5,
				InputTokensDetails: &dto.InputTokenDetails{
					CachedTokens: 8,
				},
			},
		},
	})
	if usage == nil || usage.PromptTokens != 12 || usage.UsageSemantic != UsageSemanticAnthropic {
		t.Fatalf("unexpected usage: %#v", usage)
	}
	if len(events) != 3 {
		t.Fatalf("expected stop, message_delta, message_stop, got %d: %#v", len(events), events)
	}
	if events[0].Type != "content_block_stop" || events[1].Type != "message_delta" || events[2].Type != "message_stop" {
		t.Fatalf("unexpected completion events: %#v", events)
	}
	if events[1].Delta == nil || events[1].Delta.StopReason == nil || *events[1].Delta.StopReason != "tool_use" {
		t.Fatalf("expected tool_use stop reason, got %#v", events[1].Delta)
	}
}
