package relay

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/relay/channel"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/service/anthropiccompat"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

func claudeMessagesViaResponses(c *gin.Context, info *relaycommon.RelayInfo, adaptor channel.Adaptor, request *dto.ClaudeRequest, forceUpstreamStream bool) (*dto.Usage, *types.NewAPIError) {
	clientStream := info != nil && info.IsStream

	claudeJSON, err := common.Marshal(request)
	if err != nil {
		return nil, types.NewError(err, types.ErrorCodeConvertRequestFailed, types.ErrOptionWithSkipRetry())
	}

	claudeJSON, err = relaycommon.RemoveDisabledFields(claudeJSON, info.ChannelOtherSettings, info.ChannelSetting.PassThroughBodyEnabled)
	if err != nil {
		return nil, types.NewError(err, types.ErrorCodeConvertRequestFailed, types.ErrOptionWithSkipRetry())
	}

	if len(info.ParamOverride) > 0 {
		claudeJSON, err = relaycommon.ApplyParamOverrideWithRelayInfo(claudeJSON, info)
		if err != nil {
			return nil, newAPIErrorFromParamOverride(err)
		}
	}

	var overriddenClaudeReq dto.ClaudeRequest
	if err := common.Unmarshal(claudeJSON, &overriddenClaudeReq); err != nil {
		return nil, types.NewError(err, types.ErrorCodeChannelParamOverrideInvalid, types.ErrOptionWithSkipRetry())
	}

	responsesReq, err := anthropiccompat.ClaudeRequestToResponsesRequest(&overriddenClaudeReq)
	if err != nil {
		return nil, types.NewErrorWithStatusCode(err, types.ErrorCodeInvalidRequest, http.StatusBadRequest, types.ErrOptionWithSkipRetry())
	}
	if forceUpstreamStream {
		responsesReq.Stream = common.GetPointer(true)
	}
	injectClaudeMessagesPromptCacheKey(c, info, &overriddenClaudeReq, responsesReq)
	info.AppendRequestConversion(types.RelayFormatOpenAIResponses)

	savedRelayMode := info.RelayMode
	savedRequestURLPath := info.RequestURLPath
	savedIsStream := info.IsStream
	defer func() {
		info.RelayMode = savedRelayMode
		info.RequestURLPath = savedRequestURLPath
		info.IsStream = savedIsStream
	}()

	info.RelayMode = relayconstant.RelayModeResponses
	info.RequestURLPath = "/v1/responses"
	info.IsStream = clientStream || forceUpstreamStream

	convertedRequest, err := adaptor.ConvertOpenAIResponsesRequest(c, info, *responsesReq)
	if err != nil {
		return nil, types.NewError(err, types.ErrorCodeConvertRequestFailed, types.ErrOptionWithSkipRetry())
	}
	relaycommon.AppendRequestConversionFromRequest(info, convertedRequest)

	jsonData, err := common.Marshal(convertedRequest)
	if err != nil {
		return nil, types.NewError(err, types.ErrorCodeConvertRequestFailed, types.ErrOptionWithSkipRetry())
	}

	jsonData, err = relaycommon.RemoveDisabledFields(jsonData, info.ChannelOtherSettings, info.ChannelSetting.PassThroughBodyEnabled)
	if err != nil {
		return nil, types.NewError(err, types.ErrorCodeConvertRequestFailed, types.ErrOptionWithSkipRetry())
	}

	resp, err := adaptor.DoRequest(c, info, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeDoRequestFailed, http.StatusInternalServerError)
	}
	if resp == nil {
		return nil, types.NewOpenAIError(nil, types.ErrorCodeBadResponse, http.StatusInternalServerError)
	}

	httpResp := resp.(*http.Response)
	upstreamIsStream := detectResponsesEventStream(httpResp)
	statusCodeMappingStr := c.GetString("status_code_mapping")
	if httpResp.StatusCode != http.StatusOK {
		newApiErr := service.RelayErrorHandler(c.Request.Context(), httpResp, false)
		service.ResetStatusCode(newApiErr, statusCodeMappingStr)
		return nil, newApiErr
	}

	info.IsStream = clientStream
	if clientStream {
		usage, newApiErr := responsesStreamToClaudeHandler(c, info, httpResp)
		if newApiErr != nil {
			service.ResetStatusCode(newApiErr, statusCodeMappingStr)
			return nil, newApiErr
		}
		return usage, nil
	}
	if upstreamIsStream {
		usage, newApiErr := responsesStreamToClaudeResponseHandler(c, info, httpResp)
		if newApiErr != nil {
			service.ResetStatusCode(newApiErr, statusCodeMappingStr)
			return nil, newApiErr
		}
		return usage, nil
	}

	usage, newApiErr := responsesToClaudeHandler(c, info, httpResp)
	if newApiErr != nil {
		service.ResetStatusCode(newApiErr, statusCodeMappingStr)
		return nil, newApiErr
	}
	return usage, nil
}

func injectClaudeMessagesPromptCacheKey(c *gin.Context, info *relaycommon.RelayInfo, claudeReq *dto.ClaudeRequest, request *dto.OpenAIResponsesRequest) {
	if request == nil || len(request.PromptCacheKey) > 0 {
		return
	}
	if value := claudeMessagesPromptCacheKeyFromHeaders(c); value != "" {
		raw, err := common.Marshal(value)
		if err != nil {
			return
		}
		request.PromptCacheKey = raw
		return
	}

	if !shouldAutoInjectClaudeMessagesPromptCacheKey(request.Model) {
		return
	}

	key := deriveClaudeMessagesPromptCacheKey(info, claudeReq, request.Model)
	if key == "" {
		return
	}
	raw, err := common.Marshal(key)
	if err != nil {
		return
	}
	request.PromptCacheKey = raw
}

func claudeMessagesPromptCacheKeyFromHeaders(c *gin.Context) string {
	if c == nil {
		return ""
	}
	for _, header := range []string{
		"prompt_cache_key",
		"prompt-cache-key",
		"session_id",
		"Session_ID",
		"conversation_id",
		"Conversation_ID",
		"x-session-id",
	} {
		if value := strings.TrimSpace(c.GetHeader(header)); value != "" {
			return value
		}
	}
	return ""
}

func shouldAutoInjectClaudeMessagesPromptCacheKey(model string) bool {
	normalized := strings.ToLower(strings.TrimSpace(model))
	return strings.Contains(normalized, "gpt-5") || strings.Contains(normalized, "codex")
}

func deriveClaudeMessagesPromptCacheKey(info *relaycommon.RelayInfo, req *dto.ClaudeRequest, model string) string {
	if req == nil {
		return ""
	}

	seedParts := []string{"model=" + strings.ToLower(strings.TrimSpace(model))}
	if info != nil {
		if info.UserId > 0 {
			seedParts = append(seedParts, fmt.Sprintf("user_id=%d", info.UserId))
		} else if info.TokenId > 0 {
			seedParts = append(seedParts, fmt.Sprintf("token_id=%d", info.TokenId))
		}
	}
	if system := normalizeClaudeMessagesCacheSeedValue(req.System); system != "" {
		seedParts = append(seedParts, "system="+system)
	}
	if tools := normalizeClaudeMessagesCacheSeedValue(req.Tools); tools != "" {
		seedParts = append(seedParts, "tools="+tools)
	}
	if toolChoice := normalizeClaudeMessagesCacheSeedValue(req.ToolChoice); toolChoice != "" {
		seedParts = append(seedParts, "tool_choice="+toolChoice)
	}
	if req.OutputConfig != nil {
		if outputConfig := normalizeClaudeMessagesCacheSeedRaw(req.OutputConfig); outputConfig != "" {
			seedParts = append(seedParts, "output_config="+outputConfig)
		}
	}
	if req.Thinking != nil {
		if thinking := normalizeClaudeMessagesCacheSeedValue(req.Thinking); thinking != "" {
			seedParts = append(seedParts, "thinking="+thinking)
		}
	}
	if firstUser := firstClaudeMessagesUserCacheSeed(req.Messages); firstUser != "" {
		seedParts = append(seedParts, "first_user="+firstUser)
	}

	return "claude_cc_" + common.Sha1(common.StringToByteSlice(strings.Join(seedParts, "|")))
}

func firstClaudeMessagesUserCacheSeed(messages []dto.ClaudeMessage) string {
	for _, message := range messages {
		role := strings.TrimSpace(message.Role)
		if role != "" && role != "user" {
			continue
		}
		return normalizeClaudeMessagesCacheSeedValue(message.Content)
	}
	return ""
}

func normalizeClaudeMessagesCacheSeedValue(value any) string {
	if value == nil {
		return ""
	}
	raw, err := common.Marshal(value)
	if err != nil {
		return common.Interface2String(value)
	}
	return normalizeClaudeMessagesCacheSeedRaw(raw)
}

func normalizeClaudeMessagesCacheSeedRaw(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	var value any
	if err := common.Unmarshal(raw, &value); err != nil {
		return string(raw)
	}
	normalized, err := common.Marshal(value)
	if err != nil {
		return string(raw)
	}
	return string(normalized)
}

func responsesToClaudeHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.NewAPIError) {
	defer service.CloseResponseBodyGracefully(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeReadResponseBodyFailed, http.StatusInternalServerError)
	}

	var responsesResp dto.OpenAIResponsesResponse
	if err := common.Unmarshal(body, &responsesResp); err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}
	if oaiError := responsesResp.GetOpenAIError(); oaiError != nil && oaiError.Type != "" {
		return nil, types.WithOpenAIError(*oaiError, resp.StatusCode)
	}

	claudeResp, usage, err := anthropiccompat.ResponsesResponseToClaudeResponse(&responsesResp, info.UpstreamModelName)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}
	if usage == nil || usage.TotalTokens == 0 {
		usage = anthropicFallbackUsage(c, anthropiccompatResponseText(&responsesResp), info)
		claudeResp.Usage = &dto.ClaudeUsage{
			InputTokens:  usage.PromptTokens,
			OutputTokens: usage.CompletionTokens,
		}
	}

	responseBody, err := common.Marshal(claudeResp)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeJsonMarshalFailed, http.StatusInternalServerError)
	}
	resp.Header.Set("Content-Type", "application/json")
	service.IOCopyBytesGracefully(c, resp, responseBody)
	return usage, nil
}

func responsesStreamToClaudeResponseHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.NewAPIError) {
	defer service.CloseResponseBodyGracefully(resp)

	responsesResp, fallbackText, newApiErr := anthropiccompat.CollectResponsesStreamResponse(resp.Body, info.UpstreamModelName)
	if newApiErr != nil {
		return nil, newApiErr
	}
	if responsesResp == nil {
		return nil, types.NewOpenAIError(fmt.Errorf("responses stream ended without payload"), types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}
	if oaiError := responsesResp.GetOpenAIError(); oaiError != nil && oaiError.Type != "" {
		return nil, types.WithOpenAIError(*oaiError, resp.StatusCode)
	}

	claudeResp, usage, err := anthropiccompat.ResponsesResponseToClaudeResponse(responsesResp, info.UpstreamModelName)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}
	if usage == nil || usage.TotalTokens == 0 {
		text := anthropiccompatResponseText(responsesResp)
		if strings.TrimSpace(text) == "" {
			text = fallbackText
		}
		usage = anthropicFallbackUsage(c, text, info)
		claudeResp.Usage = &dto.ClaudeUsage{
			InputTokens:  usage.PromptTokens,
			OutputTokens: usage.CompletionTokens,
		}
	}

	responseBody, err := common.Marshal(claudeResp)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeJsonMarshalFailed, http.StatusInternalServerError)
	}
	resp.Header.Set("Content-Type", "application/json")
	service.IOCopyBytesGracefully(c, resp, responseBody)
	return usage, nil
}

func responsesStreamToClaudeHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.NewAPIError) {
	if resp == nil || resp.Body == nil {
		return nil, types.NewOpenAIError(fmt.Errorf("invalid response"), types.ErrorCodeBadResponse, http.StatusInternalServerError)
	}
	defer service.CloseResponseBodyGracefully(resp)

	state := anthropiccompat.NewResponsesToClaudeStreamState(helper.GetResponseID(c), info.UpstreamModelName, info.GetEstimatePromptTokens())
	var usage *dto.Usage
	var streamErr *types.NewAPIError

	helper.StreamScannerHandler(c, resp, info, func(data string, sr *helper.StreamResult) {
		var streamResp dto.ResponsesStreamResponse
		if err := common.UnmarshalJsonStr(data, &streamResp); err != nil {
			logger.LogError(c, "failed to unmarshal responses stream event: "+err.Error())
			sr.Error(err)
			return
		}

		switch streamResp.Type {
		case "response.error", "response.failed":
			if streamResp.Response != nil {
				if oaiErr := streamResp.Response.GetOpenAIError(); oaiErr != nil && oaiErr.Type != "" {
					streamErr = types.WithOpenAIError(*oaiErr, http.StatusInternalServerError)
					sr.Stop(streamErr)
					return
				}
			}
			streamErr = types.NewOpenAIError(fmt.Errorf("responses stream error: %s", streamResp.Type), types.ErrorCodeBadResponse, http.StatusInternalServerError)
			sr.Stop(streamErr)
			return
		}

		events, eventUsage := state.EventsForStreamResponse(streamResp)
		if eventUsage != nil {
			usage = eventUsage
		}
		for _, event := range events {
			if err := helper.ClaudeData(c, event); err != nil {
				streamErr = types.NewOpenAIError(err, types.ErrorCodeBadResponse, http.StatusInternalServerError)
				sr.Stop(streamErr)
				return
			}
		}
	})

	if streamErr != nil {
		return nil, streamErr
	}
	if usage == nil || usage.TotalTokens == 0 {
		usage = anthropicFallbackUsage(c, state.OutputText(), info)
	}
	if usage != nil {
		usage.UsageSemantic = anthropiccompat.UsageSemanticAnthropic
	}
	return usage, nil
}

func anthropiccompatResponseText(resp *dto.OpenAIResponsesResponse) string {
	if resp == nil {
		return ""
	}
	var sb strings.Builder
	for _, output := range resp.Output {
		if output.Type != "message" {
			continue
		}
		for _, part := range output.Content {
			if part.Text != "" {
				sb.WriteString(part.Text)
			}
		}
	}
	return sb.String()
}

func anthropicFallbackUsage(c *gin.Context, text string, info *relaycommon.RelayInfo) *dto.Usage {
	usage := service.ResponseText2Usage(c, text, info.UpstreamModelName, info.GetEstimatePromptTokens())
	usage.UsageSemantic = anthropiccompat.UsageSemanticAnthropic
	usage.InputTokens = usage.PromptTokens
	usage.OutputTokens = usage.CompletionTokens
	return usage
}
