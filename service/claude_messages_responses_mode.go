package service

import (
	"github.com/QuantumNous/new-api/service/relayconvert"
	"github.com/QuantumNous/new-api/setting/model_setting"
)

type ClaudeMessagesToResponsesDecision = relayconvert.ClaudeMessagesToResponsesDecision

func ResolveClaudeMessagesToResponsesPolicy(policy model_setting.ClaudeMessagesToResponsesPolicy, channelID int, channelType int, model string, usingGroup string, isStream bool) ClaudeMessagesToResponsesDecision {
	return relayconvert.ResolveClaudeMessagesToResponsesPolicy(policy, channelID, channelType, model, usingGroup, isStream)
}

func ResolveClaudeMessagesToResponsesGlobal(channelID int, channelType int, model string, usingGroup string, isStream bool) ClaudeMessagesToResponsesDecision {
	return relayconvert.ResolveClaudeMessagesToResponsesGlobal(channelID, channelType, model, usingGroup, isStream)
}

func ShouldClaudeMessagesUseResponsesGlobal(channelID int, channelType int, model string, usingGroup string, isStream bool) bool {
	return relayconvert.ShouldClaudeMessagesUseResponsesGlobal(channelID, channelType, model, usingGroup, isStream)
}
