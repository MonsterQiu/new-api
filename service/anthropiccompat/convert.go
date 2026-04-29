package anthropiccompat

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/types"
)

const UsageSemanticAnthropic = "anthropic"

func ClaudeRequestToResponsesRequest(req *dto.ClaudeRequest) (*dto.OpenAIResponsesRequest, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	if strings.TrimSpace(req.Model) == "" {
		return nil, errors.New("model is required")
	}

	inputItems := make([]map[string]any, 0, len(req.Messages))
	for _, message := range req.Messages {
		items, err := claudeMessageToResponsesItems(message)
		if err != nil {
			return nil, err
		}
		inputItems = append(inputItems, items...)
	}

	inputRaw, err := common.Marshal(inputItems)
	if err != nil {
		return nil, err
	}

	out := &dto.OpenAIResponsesRequest{
		Model:             req.Model,
		Input:             inputRaw,
		Instructions:      claudeSystemToResponsesInstructions(req),
		MaxOutputTokens:   req.MaxTokens,
		ContextManagement: req.ContextManagement,
		ServiceTier:       req.ServiceTier,
		Stream:            req.Stream,
		Temperature:       req.Temperature,
		TopP:              req.TopP,
		Tools:             claudeToolsToResponsesTools(req.Tools),
	}

	toolChoice, parallelToolCalls := claudeToolChoiceToResponses(req.ToolChoice)
	out.ToolChoice = toolChoice
	out.ParallelToolCalls = parallelToolCalls

	if effort := req.GetEfforts(); effort != "" {
		out.Reasoning = &dto.Reasoning{Effort: effort, Summary: "detailed"}
	} else if req.Thinking != nil && strings.TrimSpace(req.Thinking.Type) != "" {
		out.Reasoning = &dto.Reasoning{Summary: "detailed"}
	}

	return out, nil
}

func claudeMessageToResponsesItems(message dto.ClaudeMessage) ([]map[string]any, error) {
	role := strings.TrimSpace(message.Role)
	if role == "" {
		role = "user"
	}
	if message.IsStringContent() {
		text := message.GetStringContent()
		if text == "" {
			return []map[string]any{buildResponsesMessageItem(role, []map[string]any{})}, nil
		}
		return []map[string]any{
			buildResponsesMessageItem(role, []map[string]any{
				buildResponsesTextPart(role, text),
			}),
		}, nil
	}

	content, err := message.ParseContent()
	if err != nil {
		return nil, err
	}

	items := make([]map[string]any, 0, len(content))
	pendingParts := make([]map[string]any, 0)
	flushMessage := func() {
		if len(pendingParts) == 0 {
			return
		}
		items = append(items, buildResponsesMessageItem(role, pendingParts))
		pendingParts = make([]map[string]any, 0)
	}

	for _, part := range content {
		switch part.Type {
		case "text", "input_text":
			pendingParts = append(pendingParts, buildResponsesTextPart(role, part.GetText()))
		case "image":
			imageURL := claudeImageURL(part)
			if imageURL != "" {
				pendingParts = append(pendingParts, map[string]any{
					"type":      "input_image",
					"image_url": imageURL,
				})
			}
		case "tool_use":
			flushMessage()
			callID := strings.TrimSpace(part.Id)
			if callID == "" {
				callID = strings.TrimSpace(part.ToolUseId)
			}
			if callID == "" || strings.TrimSpace(part.Name) == "" {
				continue
			}
			items = append(items, map[string]any{
				"type":      "function_call",
				"call_id":   callID,
				"name":      part.Name,
				"arguments": jsonArgumentString(part.Input),
			})
		case "tool_result":
			flushMessage()
			callID := strings.TrimSpace(part.ToolUseId)
			if callID == "" {
				callID = strings.TrimSpace(part.Id)
			}
			if callID == "" {
				continue
			}
			items = append(items, map[string]any{
				"type":    "function_call_output",
				"call_id": callID,
				"output":  claudeToolResultOutput(part),
			})
		default:
			if raw, err := common.Marshal(part); err == nil {
				pendingParts = append(pendingParts, buildResponsesTextPart(role, string(raw)))
			}
		}
	}
	flushMessage()

	if len(items) == 0 {
		items = append(items, buildResponsesMessageItem(role, []map[string]any{}))
	}
	return items, nil
}

func buildResponsesMessageItem(role string, content any) map[string]any {
	return map[string]any{
		"type":    "message",
		"role":    role,
		"content": content,
	}
}

func buildResponsesTextPart(role string, text string) map[string]any {
	textType := "input_text"
	if role == "assistant" {
		textType = "output_text"
	}
	return map[string]any{
		"type": textType,
		"text": text,
	}
}

func claudeImageURL(part dto.ClaudeMediaMessage) string {
	if part.Source == nil {
		return ""
	}
	if strings.TrimSpace(part.Source.Url) != "" {
		return part.Source.Url
	}
	data := common.Interface2String(part.Source.Data)
	if data == "" {
		return ""
	}
	mediaType := strings.TrimSpace(part.Source.MediaType)
	if mediaType == "" {
		mediaType = "image/png"
	}
	return fmt.Sprintf("data:%s;base64,%s", mediaType, data)
}

func claudeSystemToResponsesInstructions(req *dto.ClaudeRequest) []byte {
	if req == nil || req.System == nil {
		return nil
	}
	if req.IsStringSystem() {
		text := strings.TrimSpace(req.GetStringSystem())
		if text == "" {
			return nil
		}
		raw, _ := common.Marshal(text)
		return raw
	}

	var parts []string
	for _, item := range req.ParseSystem() {
		if item.Type == "text" || item.Type == "input_text" {
			if text := strings.TrimSpace(item.GetText()); text != "" {
				parts = append(parts, text)
			}
		}
	}
	if len(parts) == 0 {
		return nil
	}
	raw, _ := common.Marshal(strings.Join(parts, "\n\n"))
	return raw
}

func claudeToolsToResponsesTools(tools any) []byte {
	if tools == nil {
		return nil
	}
	var rawTools []map[string]any
	if b, err := common.Marshal(tools); err == nil {
		_ = common.Unmarshal(b, &rawTools)
	}
	if len(rawTools) == 0 {
		return nil
	}

	responsesTools := make([]map[string]any, 0, len(rawTools))
	for _, tool := range rawTools {
		name := strings.TrimSpace(common.Interface2String(tool["name"]))
		if name == "" {
			continue
		}
		if toolType := strings.TrimSpace(common.Interface2String(tool["type"])); strings.Contains(toolType, "web_search") {
			responsesTools = append(responsesTools, map[string]any{"type": "web_search_preview"})
			continue
		}
		responsesTool := map[string]any{
			"type": "function",
			"name": name,
		}
		if desc := strings.TrimSpace(common.Interface2String(tool["description"])); desc != "" {
			responsesTool["description"] = desc
		}
		if schema, ok := tool["input_schema"]; ok && schema != nil {
			responsesTool["parameters"] = schema
		}
		responsesTools = append(responsesTools, responsesTool)
	}
	if len(responsesTools) == 0 {
		return nil
	}
	raw, _ := common.Marshal(responsesTools)
	return raw
}

func claudeToolChoiceToResponses(toolChoice any) ([]byte, []byte) {
	if toolChoice == nil {
		return nil, nil
	}
	var choice map[string]any
	if b, err := common.Marshal(toolChoice); err == nil {
		_ = common.Unmarshal(b, &choice)
	}
	if len(choice) == 0 {
		raw, _ := common.Marshal(toolChoice)
		return raw, nil
	}

	var parallelToolCalls []byte
	if disabled, ok := choice["disable_parallel_tool_use"].(bool); ok && disabled {
		parallelToolCalls, _ = common.Marshal(false)
	}

	switch strings.TrimSpace(common.Interface2String(choice["type"])) {
	case "auto":
		raw, _ := common.Marshal("auto")
		return raw, parallelToolCalls
	case "any":
		raw, _ := common.Marshal("required")
		return raw, parallelToolCalls
	case "none":
		raw, _ := common.Marshal("none")
		return raw, parallelToolCalls
	case "tool":
		name := strings.TrimSpace(common.Interface2String(choice["name"]))
		if name != "" {
			raw, _ := common.Marshal(map[string]any{
				"type": "function",
				"name": name,
			})
			return raw, parallelToolCalls
		}
	}

	raw, _ := common.Marshal(toolChoice)
	return raw, parallelToolCalls
}

func jsonArgumentString(v any) string {
	if v == nil {
		return "{}"
	}
	if s, ok := v.(string); ok {
		s = strings.TrimSpace(s)
		if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
			return s
		}
	}
	raw, err := common.Marshal(v)
	if err != nil || len(raw) == 0 {
		return "{}"
	}
	return string(raw)
}

func claudeToolResultOutput(part dto.ClaudeMediaMessage) any {
	if part.Content == nil {
		return ""
	}
	if part.IsStringContent() {
		return part.GetStringContent()
	}
	content := part.ParseMediaContent()
	if len(content) > 0 {
		var text strings.Builder
		for _, item := range content {
			if item.Type == "text" || item.Type == "input_text" {
				text.WriteString(item.GetText())
			}
		}
		if text.Len() > 0 {
			return text.String()
		}
	}
	raw, err := common.Marshal(part.Content)
	if err != nil {
		return fmt.Sprintf("%v", part.Content)
	}
	return string(raw)
}

func ResponsesResponseToClaudeResponse(resp *dto.OpenAIResponsesResponse, fallbackModel string) (*dto.ClaudeResponse, *dto.Usage, error) {
	if resp == nil {
		return nil, nil, errors.New("response is nil")
	}

	model := strings.TrimSpace(resp.Model)
	if model == "" {
		model = fallbackModel
	}
	claudeResp := &dto.ClaudeResponse{
		Id:         resp.ID,
		Type:       "message",
		Role:       "assistant",
		Model:      model,
		StopReason: responseStopReason(resp),
		Content:    responsesOutputsToClaudeContent(resp.Output),
		Usage:      ClaudeUsageFromResponsesUsage(resp.Usage),
	}

	usage := UsageFromResponsesUsage(resp.Usage)
	return claudeResp, &usage, nil
}

func responsesOutputsToClaudeContent(outputs []dto.ResponsesOutput) []dto.ClaudeMediaMessage {
	content := make([]dto.ClaudeMediaMessage, 0, len(outputs))
	for _, output := range outputs {
		switch output.Type {
		case "message":
			if output.Role != "" && output.Role != "assistant" {
				continue
			}
			for _, part := range output.Content {
				if part.Text == "" {
					continue
				}
				block := dto.ClaudeMediaMessage{Type: "text"}
				block.SetText(part.Text)
				content = append(content, block)
			}
		case "function_call":
			callID := strings.TrimSpace(output.CallId)
			if callID == "" {
				callID = strings.TrimSpace(output.ID)
			}
			if callID == "" || strings.TrimSpace(output.Name) == "" {
				continue
			}
			content = append(content, dto.ClaudeMediaMessage{
				Type:  "tool_use",
				Id:    callID,
				Name:  output.Name,
				Input: parseToolArguments(output.Arguments),
			})
		}
	}
	return content
}

func parseToolArguments(args string) any {
	args = strings.TrimSpace(args)
	if args == "" {
		return map[string]any{}
	}
	var obj map[string]any
	if err := common.Unmarshal(common.StringToByteSlice(args), &obj); err == nil {
		return obj
	}
	var arr []any
	if err := common.Unmarshal(common.StringToByteSlice(args), &arr); err == nil {
		return arr
	}
	return args
}

func responseStopReason(resp *dto.OpenAIResponsesResponse) string {
	for _, output := range resp.Output {
		if output.Type == "function_call" {
			return "tool_use"
		}
	}
	status := strings.Trim(string(resp.Status), `"`)
	if strings.EqualFold(status, "incomplete") {
		return "max_tokens"
	}
	return "end_turn"
}

func ClaudeUsageFromResponsesUsage(upstream *dto.Usage) *dto.ClaudeUsage {
	if upstream == nil {
		return nil
	}
	cached, cachedCreation := responseCacheDetails(upstream)
	inputTokens := upstream.InputTokens
	if inputTokens == 0 {
		inputTokens = upstream.PromptTokens
	}
	nonCachedInput := inputTokens - cached - cachedCreation
	if nonCachedInput < 0 {
		nonCachedInput = 0
	}
	outputTokens := upstream.OutputTokens
	if outputTokens == 0 {
		outputTokens = upstream.CompletionTokens
	}
	usage := &dto.ClaudeUsage{
		InputTokens:              nonCachedInput,
		CacheCreationInputTokens: cachedCreation,
		CacheReadInputTokens:     cached,
		OutputTokens:             outputTokens,
	}
	if cachedCreation > 0 {
		usage.CacheCreation = &dto.ClaudeCacheCreationUsage{
			Ephemeral5mInputTokens: cachedCreation,
		}
	}
	return usage
}

func UsageFromResponsesUsage(upstream *dto.Usage) dto.Usage {
	usage := dto.Usage{UsageSemantic: UsageSemanticAnthropic}
	if upstream == nil {
		return usage
	}
	cached, cachedCreation := responseCacheDetails(upstream)
	inputTokens := upstream.InputTokens
	if inputTokens == 0 {
		inputTokens = upstream.PromptTokens
	}
	nonCachedInput := inputTokens - cached - cachedCreation
	if nonCachedInput < 0 {
		nonCachedInput = 0
	}
	outputTokens := upstream.OutputTokens
	if outputTokens == 0 {
		outputTokens = upstream.CompletionTokens
	}

	usage.PromptTokens = nonCachedInput
	usage.CompletionTokens = outputTokens
	usage.TotalTokens = nonCachedInput + outputTokens
	usage.InputTokens = inputTokens
	usage.OutputTokens = outputTokens
	usage.InputTokensDetails = upstream.InputTokensDetails
	usage.OutputTokensDetails = upstream.OutputTokensDetails
	usage.PromptTokensDetails = upstream.PromptTokensDetails
	usage.PromptTokensDetails.CachedTokens = cached
	usage.PromptTokensDetails.CachedCreationTokens = cachedCreation
	if upstream.InputTokensDetails != nil {
		usage.PromptTokensDetails.TextTokens = upstream.InputTokensDetails.TextTokens
		usage.PromptTokensDetails.AudioTokens = upstream.InputTokensDetails.AudioTokens
		usage.PromptTokensDetails.ImageTokens = upstream.InputTokensDetails.ImageTokens
	}
	if upstream.OutputTokensDetails != nil {
		usage.CompletionTokenDetails = *upstream.OutputTokensDetails
	}
	usage.ClaudeCacheCreation5mTokens = cachedCreation
	return usage
}

func responseCacheDetails(upstream *dto.Usage) (cached int, cachedCreation int) {
	if upstream == nil {
		return 0, 0
	}
	if upstream.InputTokensDetails != nil {
		cached = upstream.InputTokensDetails.CachedTokens
		cachedCreation = upstream.InputTokensDetails.CachedCreationTokens
	}
	if cached == 0 {
		cached = upstream.PromptTokensDetails.CachedTokens
	}
	if cachedCreation == 0 {
		cachedCreation = upstream.PromptTokensDetails.CachedCreationTokens
	}
	return cached, cachedCreation
}

func CollectResponsesStreamResponse(body io.Reader, fallbackModel string) (*dto.OpenAIResponsesResponse, string, *types.NewAPIError) {
	if body == nil {
		return nil, "", types.NewOpenAIError(fmt.Errorf("invalid response body"), types.ErrorCodeBadResponse, http.StatusInternalServerError)
	}

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 64<<10), 64<<20)
	scanner.Split(bufio.ScanLines)

	response := &dto.OpenAIResponsesResponse{
		Object:    "response",
		Model:     fallbackModel,
		CreatedAt: int(time.Now().Unix()),
	}
	var outputText strings.Builder

	callIDByItemID := make(map[string]string)
	nameByCallID := make(map[string]string)
	argsByCallID := make(map[string]string)
	callOrder := make([]string, 0)
	seenCallID := make(map[string]bool)

	rememberCall := func(itemID string, callID string) string {
		itemID = strings.TrimSpace(itemID)
		callID = strings.TrimSpace(callID)
		if callID == "" {
			callID = itemID
		}
		if callID == "" {
			return ""
		}
		if itemID != "" {
			callIDByItemID[itemID] = callID
		}
		if !seenCallID[callID] {
			seenCallID[callID] = true
			callOrder = append(callOrder, callID)
		}
		return callID
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) < 6 {
			continue
		}
		if strings.HasPrefix(line, "[DONE]") {
			break
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if line == "" || strings.HasPrefix(line, "[DONE]") {
			break
		}

		var streamResp dto.ResponsesStreamResponse
		if err := common.UnmarshalJsonStr(line, &streamResp); err != nil {
			return nil, "", types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
		}
		switch streamResp.Type {
		case "response.created":
			if streamResp.Response != nil {
				mergeResponseShell(response, streamResp.Response)
			}
		case "response.output_text.delta", "response.reasoning_summary_text.delta":
			outputText.WriteString(streamResp.Delta)
		case dto.ResponsesOutputTypeItemAdded, dto.ResponsesOutputTypeItemDone:
			if streamResp.Item == nil || streamResp.Item.Type != "function_call" {
				continue
			}
			callID := rememberCall(streamResp.Item.ID, streamResp.Item.CallId)
			if callID == "" {
				continue
			}
			if name := strings.TrimSpace(streamResp.Item.Name); name != "" {
				nameByCallID[callID] = name
			}
			if streamResp.Item.Arguments != "" {
				argsByCallID[callID] = streamResp.Item.Arguments
			}
		case "response.function_call_arguments.delta":
			callID := callIDByItemID[strings.TrimSpace(streamResp.ItemID)]
			if callID == "" {
				callID = strings.TrimSpace(streamResp.ItemID)
			}
			callID = rememberCall(streamResp.ItemID, callID)
			if callID != "" {
				argsByCallID[callID] += streamResp.Delta
			}
		case "response.completed":
			if streamResp.Response != nil {
				response = streamResp.Response
			}
		case "response.error", "response.failed":
			if streamResp.Response != nil {
				if oaiErr := streamResp.Response.GetOpenAIError(); oaiErr != nil && oaiErr.Type != "" {
					return nil, "", types.WithOpenAIError(*oaiErr, http.StatusInternalServerError)
				}
			}
			return nil, "", types.NewOpenAIError(fmt.Errorf("responses stream error: %s", streamResp.Type), types.ErrorCodeBadResponse, http.StatusInternalServerError)
		}
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, "", types.NewOpenAIError(err, types.ErrorCodeReadResponseBodyFailed, http.StatusInternalServerError)
	}

	if response.Model == "" {
		response.Model = fallbackModel
	}
	if response.CreatedAt == 0 {
		response.CreatedAt = int(time.Now().Unix())
	}
	if len(response.Output) == 0 {
		response.Output = synthesizeResponsesOutput(outputText.String(), callOrder, nameByCallID, argsByCallID)
	}
	return response, outputText.String(), nil
}

func mergeResponseShell(dst *dto.OpenAIResponsesResponse, src *dto.OpenAIResponsesResponse) {
	if dst == nil || src == nil {
		return
	}
	if src.ID != "" {
		dst.ID = src.ID
	}
	if src.Model != "" {
		dst.Model = src.Model
	}
	if src.CreatedAt != 0 {
		dst.CreatedAt = src.CreatedAt
	}
}

func synthesizeResponsesOutput(text string, callOrder []string, nameByCallID map[string]string, argsByCallID map[string]string) []dto.ResponsesOutput {
	if text != "" {
		return []dto.ResponsesOutput{
			{
				Type:   "message",
				Role:   "assistant",
				Status: "completed",
				Content: []dto.ResponsesOutputContent{
					{Type: "output_text", Text: text},
				},
			},
		}
	}
	if len(callOrder) == 0 {
		return nil
	}
	outputs := make([]dto.ResponsesOutput, 0, len(callOrder))
	for _, callID := range callOrder {
		outputs = append(outputs, dto.ResponsesOutput{
			Type:      "function_call",
			ID:        callID,
			CallId:    callID,
			Name:      nameByCallID[callID],
			Arguments: argsByCallID[callID],
			Status:    "completed",
		})
	}
	return outputs
}
