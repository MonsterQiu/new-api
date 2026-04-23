package openai

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

func OaiResponsesStreamToChatHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.NewAPIError) {
	if resp == nil || resp.Body == nil {
		return nil, types.NewOpenAIError(fmt.Errorf("invalid response"), types.ErrorCodeBadResponse, http.StatusInternalServerError)
	}

	defer service.CloseResponseBodyGracefully(resp)

	responsesResp, fallbackText, newAPIErr := collectResponsesStreamResponse(resp.Body, info.UpstreamModelName)
	if newAPIErr != nil {
		return nil, newAPIErr
	}

	if responsesResp == nil {
		return nil, types.NewOpenAIError(fmt.Errorf("responses stream ended without payload"), types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}
	if oaiError := responsesResp.GetOpenAIError(); oaiError != nil && oaiError.Type != "" {
		return nil, types.WithOpenAIError(*oaiError, resp.StatusCode)
	}

	chatID := helper.GetResponseID(c)
	chatResp, usage, err := service.ResponsesResponseToChatCompletionsResponse(responsesResp, chatID)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}

	if usage == nil || usage.TotalTokens == 0 {
		text := service.ExtractOutputTextFromResponses(responsesResp)
		if strings.TrimSpace(text) == "" {
			text = fallbackText
		}
		usage = service.ResponseText2Usage(c, text, info.UpstreamModelName, info.GetEstimatePromptTokens())
		chatResp.Usage = *usage
	}

	var responseBody []byte
	switch info.RelayFormat {
	case types.RelayFormatClaude:
		claudeResp := service.ResponseOpenAI2Claude(chatResp, info)
		responseBody, err = common.Marshal(claudeResp)
	case types.RelayFormatGemini:
		geminiResp := service.ResponseOpenAI2Gemini(chatResp, info)
		responseBody, err = common.Marshal(geminiResp)
	default:
		responseBody, err = common.Marshal(chatResp)
	}
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeJsonMarshalFailed, http.StatusInternalServerError)
	}

	service.IOCopyBytesGracefully(c, resp, responseBody)
	return usage, nil
}

func collectResponsesStreamResponse(body io.Reader, fallbackModel string) (*dto.OpenAIResponsesResponse, string, *types.NewAPIError) {
	if body == nil {
		return nil, "", types.NewOpenAIError(fmt.Errorf("invalid response body"), types.ErrorCodeBadResponse, http.StatusInternalServerError)
	}

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, helper.InitialScannerBufferSize), helper.DefaultMaxScannerBufferSize)
	scanner.Split(bufio.ScanLines)

	response := &dto.OpenAIResponsesResponse{
		Object:    "response",
		Model:     fallbackModel,
		CreatedAt: int(time.Now().Unix()),
	}
	var outputText strings.Builder

	toolCallCanonicalIDByItemID := make(map[string]string)
	toolCallNameByID := make(map[string]string)
	toolCallArgsByID := make(map[string]string)
	toolCallOrder := make([]string, 0)
	toolCallSeen := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 6 {
			continue
		}
		if line[:5] != "data:" && line[:6] != "[DONE]" {
			continue
		}

		line = strings.TrimSpace(line[5:])
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[DONE]") {
			break
		}

		var streamResp dto.ResponsesStreamResponse
		if err := common.UnmarshalJsonStr(line, &streamResp); err != nil {
			return nil, "", types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
		}

		switch streamResp.Type {
		case "response.created":
			if streamResp.Response != nil {
				if streamResp.Response.ID != "" {
					response.ID = streamResp.Response.ID
				}
				if streamResp.Response.Model != "" {
					response.Model = streamResp.Response.Model
				}
				if streamResp.Response.CreatedAt != 0 {
					response.CreatedAt = streamResp.Response.CreatedAt
				}
			}

		case "response.output_text.delta":
			if streamResp.Delta != "" {
				outputText.WriteString(streamResp.Delta)
			}

		case "response.output_item.added", "response.output_item.done":
			if streamResp.Item == nil || streamResp.Item.Type != "function_call" {
				break
			}
			itemID := strings.TrimSpace(streamResp.Item.ID)
			callID := strings.TrimSpace(streamResp.Item.CallId)
			if callID == "" {
				callID = itemID
			}
			if callID == "" {
				break
			}
			if itemID != "" {
				toolCallCanonicalIDByItemID[itemID] = callID
			}
			if !toolCallSeen[callID] {
				toolCallSeen[callID] = true
				toolCallOrder = append(toolCallOrder, callID)
			}
			if name := strings.TrimSpace(streamResp.Item.Name); name != "" {
				toolCallNameByID[callID] = name
			}
			if streamResp.Item.Arguments != "" {
				toolCallArgsByID[callID] = streamResp.Item.Arguments
			}

		case "response.function_call_arguments.delta":
			itemID := strings.TrimSpace(streamResp.ItemID)
			callID := toolCallCanonicalIDByItemID[itemID]
			if callID == "" {
				callID = itemID
			}
			if callID == "" {
				break
			}
			if !toolCallSeen[callID] {
				toolCallSeen[callID] = true
				toolCallOrder = append(toolCallOrder, callID)
			}
			toolCallArgsByID[callID] += streamResp.Delta

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

	if response == nil {
		return nil, outputText.String(), nil
	}
	if strings.TrimSpace(response.Model) == "" {
		response.Model = fallbackModel
	}
	if response.CreatedAt == 0 {
		response.CreatedAt = int(time.Now().Unix())
	}
	if len(response.Output) == 0 {
		if outputText.Len() > 0 {
			response.Output = []dto.ResponsesOutput{
				{
					Type:   "message",
					Role:   "assistant",
					Status: "completed",
					Content: []dto.ResponsesOutputContent{
						{
							Type: "output_text",
							Text: outputText.String(),
						},
					},
				},
			}
		} else if len(toolCallOrder) > 0 {
			response.Output = make([]dto.ResponsesOutput, 0, len(toolCallOrder))
			for _, callID := range toolCallOrder {
				response.Output = append(response.Output, dto.ResponsesOutput{
					Type:      "function_call",
					ID:        callID,
					CallId:    callID,
					Name:      toolCallNameByID[callID],
					Arguments: toolCallArgsByID[callID],
					Status:    "completed",
				})
			}
		}
	}

	return response, outputText.String(), nil
}
