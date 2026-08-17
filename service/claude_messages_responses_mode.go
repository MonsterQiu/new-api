package service

import "github.com/QuantumNous/new-api/setting/model_setting"

type ClaudeMessagesToResponsesDecision struct {
	UseResponses        bool
	ForceUpstreamStream bool
}

func ResolveClaudeMessagesToResponsesPolicy(policy model_setting.ClaudeMessagesToResponsesPolicy, channelID int, channelType int, model string, usingGroup string, isStream bool) ClaudeMessagesToResponsesDecision {
	if !policy.IsChannelEnabled(channelID, channelType) || !matchAnyModelPattern(policy.ModelPatterns, model) {
		return ClaudeMessagesToResponsesDecision{}
	}
	if len(policy.Groups) > 0 && !matchAnyExact(policy.Groups, usingGroup) {
		return ClaudeMessagesToResponsesDecision{}
	}

	forceUpstreamStream := false
	if policy.ForceUpstreamStream && !isStream {
		forceUpstreamStream = !policy.OnlyNonStream || !isStream
	}
	return ClaudeMessagesToResponsesDecision{
		UseResponses:        true,
		ForceUpstreamStream: forceUpstreamStream,
	}
}

func ResolveClaudeMessagesToResponsesGlobal(channelID int, channelType int, model string, usingGroup string, isStream bool) ClaudeMessagesToResponsesDecision {
	return ResolveClaudeMessagesToResponsesPolicy(model_setting.GetGlobalSettings().ClaudeMessagesToResponsesPolicy, channelID, channelType, model, usingGroup, isStream)
}

func ShouldClaudeMessagesUseResponsesGlobal(channelID int, channelType int, model string, usingGroup string, isStream bool) bool {
	return ResolveClaudeMessagesToResponsesGlobal(channelID, channelType, model, usingGroup, isStream).UseResponses
}
