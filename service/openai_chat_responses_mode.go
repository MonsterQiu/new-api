package service

import (
	"regexp"
	"sync"

	"github.com/QuantumNous/new-api/setting/model_setting"
)

type ChatCompletionsToResponsesDecision struct {
	UseResponses        bool
	ForceUpstreamStream bool
}

var chatResponsesRegexCache sync.Map

func matchAnyModelPattern(patterns []string, model string) bool {
	if len(patterns) == 0 || model == "" {
		return false
	}
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}
		re, ok := chatResponsesRegexCache.Load(pattern)
		if !ok {
			compiled, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}
			re = compiled
			chatResponsesRegexCache.Store(pattern, re)
		}
		if re.(*regexp.Regexp).MatchString(model) {
			return true
		}
	}
	return false
}

func matchAnyExact(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func ResolveChatCompletionsToResponsesPolicy(policy model_setting.ChatCompletionsToResponsesPolicy, channelID int, channelType int, model string, usingGroup string, isStream bool) ChatCompletionsToResponsesDecision {
	if !policy.IsChannelEnabled(channelID, channelType) || !matchAnyModelPattern(policy.ModelPatterns, model) {
		return ChatCompletionsToResponsesDecision{}
	}

	forceUpstreamStream := false
	if policy.ForceUpstreamStream && !isStream {
		groupMatched := len(policy.Groups) == 0 || matchAnyExact(policy.Groups, usingGroup)
		nonStreamMatched := !policy.OnlyNonStream || !isStream
		forceUpstreamStream = groupMatched && nonStreamMatched
	}
	return ChatCompletionsToResponsesDecision{
		UseResponses:        true,
		ForceUpstreamStream: forceUpstreamStream,
	}
}

func ShouldChatCompletionsUseResponsesPolicy(policy model_setting.ChatCompletionsToResponsesPolicy, channelID int, channelType int, model string, usingGroup string, isStream bool) bool {
	return ResolveChatCompletionsToResponsesPolicy(policy, channelID, channelType, model, usingGroup, isStream).UseResponses
}

func ResolveChatCompletionsToResponsesGlobal(channelID int, channelType int, model string, usingGroup string, isStream bool) ChatCompletionsToResponsesDecision {
	return ResolveChatCompletionsToResponsesPolicy(model_setting.GetGlobalSettings().ChatCompletionsToResponsesPolicy, channelID, channelType, model, usingGroup, isStream)
}

func ShouldChatCompletionsUseResponsesGlobal(channelID int, channelType int, model string, usingGroup string, isStream bool) bool {
	return ResolveChatCompletionsToResponsesGlobal(channelID, channelType, model, usingGroup, isStream).UseResponses
}
