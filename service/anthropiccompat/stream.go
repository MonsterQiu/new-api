package anthropiccompat

import (
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
)

type ResponsesToClaudeStreamState struct {
	ResponseID          string
	Model               string
	CreatedAt           int
	EstimateInputTokens int

	started       bool
	done          bool
	activeKind    string
	activeIndex   int
	nextIndex     int
	sawText       bool
	sawToolCall   bool
	outputText    strings.Builder
	reasoningText strings.Builder

	callIDByItemID map[string]string
	toolNameByID   map[string]string
	toolArgsByID   map[string]string
	toolIndexByID  map[string]int
	openToolByID   map[string]bool
	toolOrder      []string
	bufferedArgs   map[string]string
}

func NewResponsesToClaudeStreamState(responseID string, model string, estimateInputTokens int) *ResponsesToClaudeStreamState {
	if responseID == "" {
		responseID = "msg_" + common.GetUUID()
	}
	return &ResponsesToClaudeStreamState{
		ResponseID:          responseID,
		Model:               model,
		CreatedAt:           int(time.Now().Unix()),
		EstimateInputTokens: estimateInputTokens,
		callIDByItemID:      make(map[string]string),
		toolNameByID:        make(map[string]string),
		toolArgsByID:        make(map[string]string),
		toolIndexByID:       make(map[string]int),
		openToolByID:        make(map[string]bool),
		bufferedArgs:        make(map[string]string),
	}
}

func (s *ResponsesToClaudeStreamState) OutputText() string {
	if s == nil {
		return ""
	}
	if text := s.outputText.String(); text != "" {
		return text
	}
	return s.reasoningText.String()
}

func (s *ResponsesToClaudeStreamState) IsDone() bool {
	return s != nil && s.done
}

func (s *ResponsesToClaudeStreamState) EventsForStreamResponse(streamResp dto.ResponsesStreamResponse) ([]dto.ClaudeResponse, *dto.Usage) {
	if s == nil || s.done {
		return nil, nil
	}

	switch streamResp.Type {
	case "response.created":
		if streamResp.Response != nil {
			s.mergeResponse(streamResp.Response)
		}
		return s.ensureStarted(), nil
	case "response.output_text.delta":
		return s.textDeltaEvents("text", streamResp.Delta), nil
	case "response.reasoning_summary_text.delta":
		return s.textDeltaEvents("thinking", streamResp.Delta), nil
	case dto.ResponsesOutputTypeItemAdded, dto.ResponsesOutputTypeItemDone:
		if streamResp.Item == nil || streamResp.Item.Type != "function_call" {
			return nil, nil
		}
		return s.toolItemEvents(streamResp.Item), nil
	case "response.function_call_arguments.delta":
		return s.toolArgsDeltaEvents(streamResp.ItemID, streamResp.Delta), nil
	case "response.completed":
		return s.completedEvents(streamResp.Response)
	default:
		return nil, nil
	}
}

func (s *ResponsesToClaudeStreamState) mergeResponse(resp *dto.OpenAIResponsesResponse) {
	if resp == nil {
		return
	}
	if resp.ID != "" {
		s.ResponseID = resp.ID
	}
	if resp.Model != "" {
		s.Model = resp.Model
	}
	if resp.CreatedAt != 0 {
		s.CreatedAt = resp.CreatedAt
	}
}

func (s *ResponsesToClaudeStreamState) ensureStarted() []dto.ClaudeResponse {
	if s.started {
		return nil
	}
	s.started = true
	msg := &dto.ClaudeMediaMessage{
		Id:    s.ResponseID,
		Model: s.Model,
		Type:  "message",
		Role:  "assistant",
		Usage: &dto.ClaudeUsage{
			InputTokens:  s.EstimateInputTokens,
			OutputTokens: 0,
		},
	}
	msg.SetContent(make([]any, 0))
	return []dto.ClaudeResponse{
		{
			Type:    "message_start",
			Message: msg,
		},
	}
}

func (s *ResponsesToClaudeStreamState) textDeltaEvents(kind string, delta string) []dto.ClaudeResponse {
	if delta == "" {
		return nil
	}
	events := s.ensureStarted()
	if s.activeKind != kind {
		events = append(events, s.closeActiveBlocks()...)
		idx := s.nextIndex
		contentBlock := &dto.ClaudeMediaMessage{Type: kind}
		if kind == "thinking" {
			contentBlock.Thinking = common.GetPointer("")
		} else {
			contentBlock.Text = common.GetPointer("")
		}
		events = append(events, dto.ClaudeResponse{
			Type:         "content_block_start",
			Index:        &idx,
			ContentBlock: contentBlock,
		})
		s.activeKind = kind
		s.activeIndex = idx
	}
	idx := s.activeIndex
	deltaBlock := &dto.ClaudeMediaMessage{}
	if kind == "thinking" {
		s.reasoningText.WriteString(delta)
		deltaBlock.Type = "thinking_delta"
		deltaBlock.Thinking = &delta
	} else {
		s.sawText = true
		s.outputText.WriteString(delta)
		deltaBlock.Type = "text_delta"
		deltaBlock.Text = &delta
	}
	events = append(events, dto.ClaudeResponse{
		Type:  "content_block_delta",
		Index: &idx,
		Delta: deltaBlock,
	})
	return events
}

func (s *ResponsesToClaudeStreamState) toolItemEvents(item *dto.ResponsesOutput) []dto.ClaudeResponse {
	if item == nil {
		return nil
	}
	callID := s.rememberCall(item.ID, item.CallId)
	if callID == "" {
		return nil
	}
	if name := strings.TrimSpace(item.Name); name != "" {
		s.toolNameByID[callID] = name
	}
	events := s.ensureToolBlock(callID)
	if item.Arguments != "" {
		prev := s.toolArgsByID[callID]
		delta := item.Arguments
		if prev != "" && strings.HasPrefix(item.Arguments, prev) {
			delta = item.Arguments[len(prev):]
		}
		s.toolArgsByID[callID] = item.Arguments
		events = append(events, s.toolArgsDeltaForCall(callID, delta)...)
	}
	return events
}

func (s *ResponsesToClaudeStreamState) toolArgsDeltaEvents(itemID string, delta string) []dto.ClaudeResponse {
	if delta == "" {
		return nil
	}
	callID := s.callIDByItemID[strings.TrimSpace(itemID)]
	if callID == "" {
		callID = s.rememberCall(itemID, itemID)
	}
	if callID == "" {
		return nil
	}
	s.toolArgsByID[callID] += delta
	events := s.ensureToolBlock(callID)
	events = append(events, s.toolArgsDeltaForCall(callID, delta)...)
	return events
}

func (s *ResponsesToClaudeStreamState) ensureToolBlock(callID string) []dto.ClaudeResponse {
	if callID == "" {
		return nil
	}
	events := s.ensureStarted()
	if s.activeKind != "tools" {
		events = append(events, s.closeActiveBlocks()...)
		s.activeKind = "tools"
	}
	if _, ok := s.toolIndexByID[callID]; !ok {
		s.toolIndexByID[callID] = s.nextIndex + len(s.toolOrder)
		s.toolOrder = append(s.toolOrder, callID)
	}
	if s.openToolByID[callID] {
		return events
	}
	name := strings.TrimSpace(s.toolNameByID[callID])
	if name == "" {
		return events
	}
	idx := s.toolIndexByID[callID]
	events = append(events, dto.ClaudeResponse{
		Type:  "content_block_start",
		Index: &idx,
		ContentBlock: &dto.ClaudeMediaMessage{
			Type:  "tool_use",
			Id:    callID,
			Name:  name,
			Input: map[string]any{},
		},
	})
	s.openToolByID[callID] = true
	s.sawToolCall = true
	if buffered := s.bufferedArgs[callID]; buffered != "" {
		delete(s.bufferedArgs, callID)
		events = append(events, s.toolArgsDeltaForCall(callID, buffered)...)
	}
	return events
}

func (s *ResponsesToClaudeStreamState) toolArgsDeltaForCall(callID string, delta string) []dto.ClaudeResponse {
	if delta == "" {
		return nil
	}
	if !s.openToolByID[callID] {
		s.bufferedArgs[callID] += delta
		return nil
	}
	idx := s.toolIndexByID[callID]
	return []dto.ClaudeResponse{
		{
			Type:  "content_block_delta",
			Index: &idx,
			Delta: &dto.ClaudeMediaMessage{
				Type:        "input_json_delta",
				PartialJson: &delta,
			},
		},
	}
}

func (s *ResponsesToClaudeStreamState) completedEvents(resp *dto.OpenAIResponsesResponse) ([]dto.ClaudeResponse, *dto.Usage) {
	if resp != nil {
		s.mergeResponse(resp)
	}
	events := s.ensureStarted()
	if resp != nil && !s.sawText && !s.sawToolCall {
		events = append(events, s.outputEventsFromCompletedResponse(resp)...)
	}
	events = append(events, s.closeActiveBlocks()...)

	var usage dto.Usage
	var claudeUsage *dto.ClaudeUsage
	if resp != nil && resp.Usage != nil {
		usage = UsageFromResponsesUsage(resp.Usage)
		claudeUsage = ClaudeUsageFromResponsesUsage(resp.Usage)
	} else {
		usage = dto.Usage{
			PromptTokens:     s.EstimateInputTokens,
			TotalTokens:      s.EstimateInputTokens,
			UsageSemantic:    UsageSemanticAnthropic,
			InputTokens:      s.EstimateInputTokens,
			CompletionTokens: 0,
		}
		claudeUsage = &dto.ClaudeUsage{InputTokens: s.EstimateInputTokens}
	}

	stopReason := "end_turn"
	if s.sawToolCall && !s.sawText {
		stopReason = "tool_use"
	}
	events = append(events, dto.ClaudeResponse{
		Type:  "message_delta",
		Usage: claudeUsage,
		Delta: &dto.ClaudeMediaMessage{
			StopReason: &stopReason,
		},
	})
	events = append(events, dto.ClaudeResponse{Type: "message_stop"})
	s.done = true
	return events, &usage
}

func (s *ResponsesToClaudeStreamState) outputEventsFromCompletedResponse(resp *dto.OpenAIResponsesResponse) []dto.ClaudeResponse {
	if resp == nil {
		return nil
	}
	var events []dto.ClaudeResponse
	for _, output := range resp.Output {
		switch output.Type {
		case "message":
			for _, part := range output.Content {
				if part.Text == "" {
					continue
				}
				events = append(events, s.textDeltaEvents("text", part.Text)...)
			}
		case "function_call":
			callID := s.rememberCall(output.ID, output.CallId)
			if callID == "" {
				continue
			}
			if output.Name != "" {
				s.toolNameByID[callID] = output.Name
			}
			events = append(events, s.ensureToolBlock(callID)...)
			events = append(events, s.toolArgsDeltaForCall(callID, output.Arguments)...)
		}
	}
	return events
}

func (s *ResponsesToClaudeStreamState) closeActiveBlocks() []dto.ClaudeResponse {
	switch s.activeKind {
	case "text", "thinking":
		idx := s.activeIndex
		s.activeKind = ""
		s.nextIndex = idx + 1
		return []dto.ClaudeResponse{{Type: "content_block_stop", Index: &idx}}
	case "tools":
		events := make([]dto.ClaudeResponse, 0, len(s.toolOrder))
		maxIndex := s.nextIndex - 1
		for _, callID := range s.toolOrder {
			if !s.openToolByID[callID] {
				continue
			}
			idx := s.toolIndexByID[callID]
			if idx > maxIndex {
				maxIndex = idx
			}
			events = append(events, dto.ClaudeResponse{Type: "content_block_stop", Index: &idx})
		}
		s.activeKind = ""
		s.nextIndex = maxIndex + 1
		s.toolOrder = nil
		s.openToolByID = make(map[string]bool)
		return events
	default:
		return nil
	}
}

func (s *ResponsesToClaudeStreamState) rememberCall(itemID string, callID string) string {
	itemID = strings.TrimSpace(itemID)
	callID = strings.TrimSpace(callID)
	if callID == "" {
		callID = itemID
	}
	if callID == "" {
		return ""
	}
	if itemID != "" {
		s.callIDByItemID[itemID] = callID
	}
	return callID
}
