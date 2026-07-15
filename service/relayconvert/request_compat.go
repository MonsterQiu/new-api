package relayconvert

import (
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	claudemessages "github.com/QuantumNous/new-api/service/relayconvert/internal/claude_messages"
	geminichat "github.com/QuantumNous/new-api/service/relayconvert/internal/gemini_chat"
	oaichat "github.com/QuantumNous/new-api/service/relayconvert/internal/oai_chat"
	oairesponses "github.com/QuantumNous/new-api/service/relayconvert/internal/oai_responses"
	sharedgemini "github.com/QuantumNous/new-api/service/relayconvert/internal/shared/gemini"
	"github.com/QuantumNous/new-api/setting/model_setting"
	"github.com/gin-gonic/gin"
)

func ClaudeMessagesRequestToOpenAIChat(claudeRequest dto.ClaudeRequest, info *relaycommon.RelayInfo) (*dto.GeneralOpenAIRequest, error) {
	return claudemessages.ClaudeMessagesRequestToOpenAIChat(claudeRequest, info)
}

func OpenAIChatRequestToClaudeMessages(c *gin.Context, textRequest dto.GeneralOpenAIRequest) (*dto.ClaudeRequest, error) {
	return oaichat.OpenAIChatRequestToClaudeMessages(c, textRequest)
}

func GeminiGenerateContentRequestToOpenAIChat(geminiRequest *dto.GeminiChatRequest, info *relaycommon.RelayInfo) (*dto.GeneralOpenAIRequest, error) {
	return geminichat.GeminiGenerateContentRequestToOpenAIChat(geminiRequest, info)
}

func OpenAIChatRequestToGeminiGenerateContent(c *gin.Context, textRequest dto.GeneralOpenAIRequest, info *relaycommon.RelayInfo) (*dto.GeminiChatRequest, error) {
	return oaichat.OpenAIChatRequestToGeminiGenerateContent(c, textRequest, info)
}

func ApplyGeminiThinkingConfig(geminiRequest *dto.GeminiChatRequest, info *relaycommon.RelayInfo, oaiRequest ...dto.GeneralOpenAIRequest) {
	sharedgemini.ApplyThinkingConfig(geminiRequest, info, oaiRequest...)
}

func ChatCompletionsRequestToResponsesRequest(req *dto.GeneralOpenAIRequest) (*dto.OpenAIResponsesRequest, error) {
	return oaichat.ChatCompletionsRequestToResponsesRequest(req)
}

func ResponsesRequestToChatCompletionsRequest(req *dto.OpenAIResponsesRequest) (*dto.GeneralOpenAIRequest, error) {
	return oairesponses.ResponsesRequestToChatCompletionsRequest(req)
}

func OpenAIResponsesRequestToClaudeMessages(c *gin.Context, req *dto.OpenAIResponsesRequest) (*dto.ClaudeRequest, error) {
	return oairesponses.OpenAIResponsesRequestToClaudeMessages(c, req)
}

func OpenAIResponsesRequestToGeminiChat(c *gin.Context, req *dto.OpenAIResponsesRequest, info *relaycommon.RelayInfo) (*dto.GeminiChatRequest, error) {
	return oairesponses.OpenAIResponsesRequestToGeminiChat(c, req, info)
}

type ChatCompletionsToResponsesDecision = oaichat.ChatCompletionsToResponsesDecision
type ClaudeMessagesToResponsesDecision = oaichat.ClaudeMessagesToResponsesDecision

func ResolveChatCompletionsToResponsesPolicy(policy model_setting.ChatCompletionsToResponsesPolicy, channelID int, channelType int, model string, usingGroup string, isStream bool) ChatCompletionsToResponsesDecision {
	return oaichat.ResolveChatCompletionsToResponsesPolicy(policy, channelID, channelType, model, usingGroup, isStream)
}

func ShouldChatCompletionsUseResponsesPolicy(policy model_setting.ChatCompletionsToResponsesPolicy, channelID int, channelType int, model string, usingGroup string, isStream bool) bool {
	return oaichat.ShouldChatCompletionsUseResponsesPolicy(policy, channelID, channelType, model, usingGroup, isStream)
}

func ResolveChatCompletionsToResponsesGlobal(channelID int, channelType int, model string, usingGroup string, isStream bool) ChatCompletionsToResponsesDecision {
	return oaichat.ResolveChatCompletionsToResponsesGlobal(channelID, channelType, model, usingGroup, isStream)
}

func ShouldChatCompletionsUseResponsesGlobal(channelID int, channelType int, model string, usingGroup string, isStream bool) bool {
	return oaichat.ShouldChatCompletionsUseResponsesGlobal(channelID, channelType, model, usingGroup, isStream)
}

func ResolveClaudeMessagesToResponsesPolicy(policy model_setting.ClaudeMessagesToResponsesPolicy, channelID int, channelType int, model string, usingGroup string, isStream bool) ClaudeMessagesToResponsesDecision {
	return oaichat.ResolveClaudeMessagesToResponsesPolicy(policy, channelID, channelType, model, usingGroup, isStream)
}

func ResolveClaudeMessagesToResponsesGlobal(channelID int, channelType int, model string, usingGroup string, isStream bool) ClaudeMessagesToResponsesDecision {
	return oaichat.ResolveClaudeMessagesToResponsesGlobal(channelID, channelType, model, usingGroup, isStream)
}

func ShouldClaudeMessagesUseResponsesGlobal(channelID int, channelType int, model string, usingGroup string, isStream bool) bool {
	return oaichat.ShouldClaudeMessagesUseResponsesGlobal(channelID, channelType, model, usingGroup, isStream)
}
