package service

import (
	"github.com/QuantumNous/new-api/service/openaicompat"
	"github.com/QuantumNous/new-api/setting/model_setting"
)

type ClaudeMessagesToResponsesDecision = openaicompat.ClaudeMessagesToResponsesDecision

func ResolveClaudeMessagesToResponsesPolicy(policy model_setting.ClaudeMessagesToResponsesPolicy, channelID int, channelType int, model string, usingGroup string, isStream bool) ClaudeMessagesToResponsesDecision {
	return openaicompat.ResolveClaudeMessagesToResponsesPolicy(policy, channelID, channelType, model, usingGroup, isStream)
}

func ResolveClaudeMessagesToResponsesGlobal(channelID int, channelType int, model string, usingGroup string, isStream bool) ClaudeMessagesToResponsesDecision {
	return openaicompat.ResolveClaudeMessagesToResponsesGlobal(channelID, channelType, model, usingGroup, isStream)
}

func ShouldClaudeMessagesUseResponsesGlobal(channelID int, channelType int, model string, usingGroup string, isStream bool) bool {
	return openaicompat.ShouldClaudeMessagesUseResponsesGlobal(channelID, channelType, model, usingGroup, isStream)
}
