package openaicompat

import (
	"strings"

	"github.com/QuantumNous/new-api/setting/model_setting"
)

type ChatCompletionsToResponsesDecision struct {
	UseResponses        bool
	ForceUpstreamStream bool
}

func ResolveChatCompletionsToResponsesPolicy(
	policy model_setting.ChatCompletionsToResponsesPolicy,
	channelID int,
	channelType int,
	model string,
	usingGroup string,
	isStream bool,
) ChatCompletionsToResponsesDecision {
	if !policy.IsChannelEnabled(channelID, channelType) {
		return ChatCompletionsToResponsesDecision{}
	}
	if policy.OnlyNonStream && isStream {
		return ChatCompletionsToResponsesDecision{}
	}
	if len(policy.Groups) > 0 && !matchAnyExact(policy.Groups, usingGroup) {
		return ChatCompletionsToResponsesDecision{}
	}
	if !matchAnyRegex(policy.ModelPatterns, model) {
		return ChatCompletionsToResponsesDecision{}
	}

	return ChatCompletionsToResponsesDecision{
		UseResponses:        true,
		ForceUpstreamStream: policy.ForceUpstreamStream && !isStream,
	}
}

func ShouldChatCompletionsUseResponsesPolicy(
	policy model_setting.ChatCompletionsToResponsesPolicy,
	channelID int,
	channelType int,
	model string,
	usingGroup string,
	isStream bool,
) bool {
	return ResolveChatCompletionsToResponsesPolicy(policy, channelID, channelType, model, usingGroup, isStream).UseResponses
}

func ResolveChatCompletionsToResponsesGlobal(
	channelID int,
	channelType int,
	model string,
	usingGroup string,
	isStream bool,
) ChatCompletionsToResponsesDecision {
	return ResolveChatCompletionsToResponsesPolicy(
		model_setting.GetGlobalSettings().ChatCompletionsToResponsesPolicy,
		channelID,
		channelType,
		model,
		usingGroup,
		isStream,
	)
}

func ShouldChatCompletionsUseResponsesGlobal(
	channelID int,
	channelType int,
	model string,
	usingGroup string,
	isStream bool,
) bool {
	return ResolveChatCompletionsToResponsesGlobal(channelID, channelType, model, usingGroup, isStream).UseResponses
}

func matchAnyExact(values []string, candidate string) bool {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" || len(values) == 0 {
		return false
	}
	for _, value := range values {
		if strings.TrimSpace(value) == candidate {
			return true
		}
	}
	return false
}
