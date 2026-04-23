package service

import (
	"github.com/QuantumNous/new-api/service/openaicompat"
	"github.com/QuantumNous/new-api/setting/model_setting"
)

type ChatCompletionsToResponsesDecision = openaicompat.ChatCompletionsToResponsesDecision

func ResolveChatCompletionsToResponsesPolicy(policy model_setting.ChatCompletionsToResponsesPolicy, channelID int, channelType int, model string, usingGroup string, isStream bool) ChatCompletionsToResponsesDecision {
	return openaicompat.ResolveChatCompletionsToResponsesPolicy(policy, channelID, channelType, model, usingGroup, isStream)
}

func ShouldChatCompletionsUseResponsesPolicy(policy model_setting.ChatCompletionsToResponsesPolicy, channelID int, channelType int, model string, usingGroup string, isStream bool) bool {
	return openaicompat.ShouldChatCompletionsUseResponsesPolicy(policy, channelID, channelType, model, usingGroup, isStream)
}

func ResolveChatCompletionsToResponsesGlobal(channelID int, channelType int, model string, usingGroup string, isStream bool) ChatCompletionsToResponsesDecision {
	return openaicompat.ResolveChatCompletionsToResponsesGlobal(channelID, channelType, model, usingGroup, isStream)
}

func ShouldChatCompletionsUseResponsesGlobal(channelID int, channelType int, model string, usingGroup string, isStream bool) bool {
	return openaicompat.ShouldChatCompletionsUseResponsesGlobal(channelID, channelType, model, usingGroup, isStream)
}
