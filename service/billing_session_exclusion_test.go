package service

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/model_setting"
	"github.com/stretchr/testify/require"
)

func TestExcludedSubscriptionPlanIDsForRelayUsesClaudeModelSetting(t *testing.T) {
	settings := model_setting.GetClaudeSettings()
	previous := append([]int(nil), settings.ExcludedSubscriptionPlanIDs...)
	t.Cleanup(func() {
		settings.ExcludedSubscriptionPlanIDs = previous
	})
	settings.ExcludedSubscriptionPlanIDs = []int{1, 2, 7}

	ids := excludedSubscriptionPlanIDsForRelay(&relaycommon.RelayInfo{
		OriginModelName: "anthropic/claude-3.7-sonnet",
	})

	require.Equal(t, []int{1, 2, 7}, ids)
}

func TestExcludedSubscriptionPlanIDsForRelayAppliesToAnthropicChannel(t *testing.T) {
	settings := model_setting.GetClaudeSettings()
	previous := append([]int(nil), settings.ExcludedSubscriptionPlanIDs...)
	t.Cleanup(func() {
		settings.ExcludedSubscriptionPlanIDs = previous
	})
	settings.ExcludedSubscriptionPlanIDs = []int{1, 2, 7}

	ids := excludedSubscriptionPlanIDsForRelay(&relaycommon.RelayInfo{
		OriginModelName: "custom-alias",
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelType: constant.ChannelTypeAnthropic,
		},
	})

	require.Equal(t, []int{1, 2, 7}, ids)
}

func TestExcludedSubscriptionPlanIDsForRelaySkipsNonClaudeModels(t *testing.T) {
	settings := model_setting.GetClaudeSettings()
	previous := append([]int(nil), settings.ExcludedSubscriptionPlanIDs...)
	t.Cleanup(func() {
		settings.ExcludedSubscriptionPlanIDs = previous
	})
	settings.ExcludedSubscriptionPlanIDs = []int{1, 2, 7}

	ids := excludedSubscriptionPlanIDsForRelay(&relaycommon.RelayInfo{
		OriginModelName: "gpt-4o",
	})

	require.Empty(t, ids)
}
