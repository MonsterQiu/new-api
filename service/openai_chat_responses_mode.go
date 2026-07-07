package service

import (
	"github.com/QuantumNous/new-api/service/relayconvert"
	"github.com/QuantumNous/new-api/setting/model_setting"
)

type ChatCompletionsToResponsesDecision = relayconvert.ChatCompletionsToResponsesDecision

func ResolveChatCompletionsToResponsesPolicy(policy model_setting.ChatCompletionsToResponsesPolicy, channelID int, channelType int, model string, usingGroup string, isStream bool) ChatCompletionsToResponsesDecision {
	return relayconvert.ResolveChatCompletionsToResponsesPolicy(policy, channelID, channelType, model, usingGroup, isStream)
}

func ShouldChatCompletionsUseResponsesPolicy(policy model_setting.ChatCompletionsToResponsesPolicy, channelID int, channelType int, model string, usingGroup string, isStream bool) bool {
	return relayconvert.ShouldChatCompletionsUseResponsesPolicy(policy, channelID, channelType, model, usingGroup, isStream)
}

func ResolveChatCompletionsToResponsesGlobal(channelID int, channelType int, model string, usingGroup string, isStream bool) ChatCompletionsToResponsesDecision {
	return relayconvert.ResolveChatCompletionsToResponsesGlobal(channelID, channelType, model, usingGroup, isStream)
}

func ShouldChatCompletionsUseResponsesGlobal(channelID int, channelType int, model string, usingGroup string, isStream bool) bool {
	return relayconvert.ShouldChatCompletionsUseResponsesGlobal(channelID, channelType, model, usingGroup, isStream)
}
