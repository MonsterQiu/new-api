package service

import (
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/types"
)

func TestShouldDisableChannelQuotaExhausted429ByMessage(t *testing.T) {
	origAutomaticDisableChannelEnabled := common.AutomaticDisableChannelEnabled
	t.Cleanup(func() {
		common.AutomaticDisableChannelEnabled = origAutomaticDisableChannelEnabled
	})

	common.AutomaticDisableChannelEnabled = true

	err := types.WithOpenAIError(types.OpenAIError{
		Message: "The usage limit has been reached for this account.",
		Type:    "rate_limit_error",
		Code:    "rate_limit_error",
	}, http.StatusTooManyRequests)

	if !ShouldDisableChannel(0, err) {
		t.Fatal("expected quota-exhausted 429 to auto-disable channel")
	}
}

func TestShouldDisableChannelQuotaExhausted429ByCode(t *testing.T) {
	origAutomaticDisableChannelEnabled := common.AutomaticDisableChannelEnabled
	t.Cleanup(func() {
		common.AutomaticDisableChannelEnabled = origAutomaticDisableChannelEnabled
	})

	common.AutomaticDisableChannelEnabled = true

	err := types.WithOpenAIError(types.OpenAIError{
		Message: "Quota exceeded.",
		Type:    "rate_limit_error",
		Code:    "insufficient_quota",
	}, http.StatusTooManyRequests)

	if !ShouldDisableChannel(0, err) {
		t.Fatal("expected insufficient_quota 429 to auto-disable channel")
	}
}

func TestShouldDisableChannelTransient429DoesNotDisable(t *testing.T) {
	origAutomaticDisableChannelEnabled := common.AutomaticDisableChannelEnabled
	t.Cleanup(func() {
		common.AutomaticDisableChannelEnabled = origAutomaticDisableChannelEnabled
	})

	common.AutomaticDisableChannelEnabled = true

	err := types.WithOpenAIError(types.OpenAIError{
		Message: "Rate limit exceeded, retry after 2 seconds.",
		Type:    "rate_limit_error",
		Code:    "rate_limit_error",
	}, http.StatusTooManyRequests)

	if ShouldDisableChannel(0, err) {
		t.Fatal("expected transient 429 to stay retry-only and not auto-disable channel")
	}
}
