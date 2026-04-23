package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
)

func TestGetRandomSatisfiedChannelWithExcludedSkipsUsedChannels(t *testing.T) {
	origMemoryCacheEnabled := common.MemoryCacheEnabled
	origGroup2Model2Channels := group2model2channels
	origChannelsIDM := channelsIDM
	t.Cleanup(func() {
		common.MemoryCacheEnabled = origMemoryCacheEnabled
		group2model2channels = origGroup2Model2Channels
		channelsIDM = origChannelsIDM
	})

	common.MemoryCacheEnabled = true
	group2model2channels = map[string]map[string][]int{
		"default": {
			"gpt-5": {1, 2, 3},
		},
	}
	group2model2channels["default"]["gpt-5"] = []int{1, 2, 3}

	priority := int64(0)
	weight := uint(0)
	channelsIDM = map[int]*Channel{
		1: {Id: 1, Priority: &priority, Weight: &weight},
		2: {Id: 2, Priority: &priority, Weight: &weight},
		3: {Id: 3, Priority: &priority, Weight: &weight},
	}

	channel, err := GetRandomSatisfiedChannelWithExcluded("default", "gpt-5", 0, map[int]struct{}{
		1: {},
		2: {},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if channel == nil {
		t.Fatalf("expected channel, got nil")
	}
	if channel.Id != 3 {
		t.Fatalf("expected channel 3, got %d", channel.Id)
	}
}

func TestGetRandomSatisfiedChannelWithExcludedReturnsNilWhenExhausted(t *testing.T) {
	origMemoryCacheEnabled := common.MemoryCacheEnabled
	origGroup2Model2Channels := group2model2channels
	origChannelsIDM := channelsIDM
	t.Cleanup(func() {
		common.MemoryCacheEnabled = origMemoryCacheEnabled
		group2model2channels = origGroup2Model2Channels
		channelsIDM = origChannelsIDM
	})

	common.MemoryCacheEnabled = true
	group2model2channels = map[string]map[string][]int{
		"default": {
			"gpt-5": {1},
		},
	}

	priority := int64(0)
	weight := uint(0)
	channelsIDM = map[int]*Channel{
		1: {Id: 1, Priority: &priority, Weight: &weight},
	}

	channel, err := GetRandomSatisfiedChannelWithExcluded("default", "gpt-5", 0, map[int]struct{}{
		1: {},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if channel != nil {
		t.Fatalf("expected nil channel when all channels are excluded, got %d", channel.Id)
	}
}
