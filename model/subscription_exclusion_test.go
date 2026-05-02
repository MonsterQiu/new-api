package model

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
)

func seedSubscriptionPlanForExclusionTest(t *testing.T, id int, title string) {
	t.Helper()
	plan := &SubscriptionPlan{
		Id:            id,
		Title:         title,
		PriceAmount:   1,
		Currency:      "USD",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   1000,
	}
	require.NoError(t, DB.Create(plan).Error)
}

func seedUserSubscriptionForExclusionTest(t *testing.T, id int, userID int, planID int, endTime int64) {
	t.Helper()
	sub := &UserSubscription{
		Id:          id,
		UserId:      userID,
		PlanId:      planID,
		AmountTotal: 1000,
		AmountUsed:  0,
		StartTime:   endTime - 3600,
		EndTime:     endTime,
		Status:      "active",
		Source:      "order",
	}
	require.NoError(t, DB.Create(sub).Error)
}

func TestPreConsumeUserSubscriptionWithExcludedPlanIDsSkipsExcludedPlans(t *testing.T) {
	truncateTables(t)

	userID := 701
	now := common.GetTimestamp()
	require.NoError(t, DB.Create(&User{Id: userID, Username: "sub_exclusion_user", Status: common.UserStatusEnabled}).Error)
	seedSubscriptionPlanForExclusionTest(t, 1, "monthly")
	seedSubscriptionPlanForExclusionTest(t, 9, "wallet-pack")
	seedUserSubscriptionForExclusionTest(t, 11, userID, 1, now+86400)
	seedUserSubscriptionForExclusionTest(t, 19, userID, 9, now+86400)

	res, err := PreConsumeUserSubscriptionWithExcludedPlanIDs("req-exclude-uses-allowed", userID, "claude-3-5-sonnet", 0, 100, []int{1})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 19, res.UserSubscriptionId)
	require.Equal(t, int64(100), res.AmountUsedAfter)
}

func TestPreConsumeUserSubscriptionWithExcludedPlanIDsReturnsNoActiveWhenOnlyExcludedPlansExist(t *testing.T) {
	truncateTables(t)

	userID := 702
	now := common.GetTimestamp()
	require.NoError(t, DB.Create(&User{Id: userID, Username: "sub_exclusion_only_user", Status: common.UserStatusEnabled}).Error)
	seedSubscriptionPlanForExclusionTest(t, 1, "monthly")
	seedUserSubscriptionForExclusionTest(t, 11, userID, 1, now+86400)

	_, err := PreConsumeUserSubscriptionWithExcludedPlanIDs("req-exclude-all", userID, "claude-3-5-sonnet", 0, 100, []int{1})
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "no active subscription"), "unexpected error: %v", err)
}
