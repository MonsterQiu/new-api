package model

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertUserForPaymentGuardTest(t *testing.T, id int, quota int) {
	t.Helper()
	user := &User{
		Id:       id,
		Username: "payment_guard_user",
		Status:   common.UserStatusEnabled,
		Quota:    quota,
	}
	require.NoError(t, DB.Create(user).Error)
}

func insertPaymentGuardUserWithName(t *testing.T, id int, username string, inviterID int) {
	t.Helper()
	user := &User{
		Id:        id,
		Username:  username,
		Password:  "pass12345",
		Status:    common.UserStatusEnabled,
		Role:      common.RoleCommonUser,
		Group:     "default",
		AffCode:   username + "_aff",
		InviterId: inviterID,
	}
	require.NoError(t, DB.Create(user).Error)
}

func insertSubscriptionPlanForPaymentGuardTest(t *testing.T, id int) *SubscriptionPlan {
	t.Helper()
	plan := &SubscriptionPlan{
		Id:            id,
		Title:         "Guard Plan",
		PriceAmount:   9.99,
		Currency:      "USD",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   1000,
	}
	require.NoError(t, DB.Create(plan).Error)
	return plan
}

func insertSubscriptionOrderForPaymentGuardTest(t *testing.T, tradeNo string, userID int, planID int, paymentMethod string) {
	t.Helper()
	order := &SubscriptionOrder{
		UserId:        userID,
		PlanId:        planID,
		Money:         9.99,
		TradeNo:       tradeNo,
		PaymentMethod: paymentMethod,
		Status:        common.TopUpStatusPending,
		CreateTime:    time.Now().Unix(),
	}
	require.NoError(t, order.Insert())
}

func insertTopUpForPaymentGuardTest(t *testing.T, tradeNo string, userID int, paymentMethod string) {
	t.Helper()
	topUp := &TopUp{
		UserId:        userID,
		Amount:        2,
		Money:         9.99,
		TradeNo:       tradeNo,
		PaymentMethod: paymentMethod,
		Status:        common.TopUpStatusPending,
		CreateTime:    time.Now().Unix(),
	}
	require.NoError(t, topUp.Insert())
}

func getTopUpStatusForPaymentGuardTest(t *testing.T, tradeNo string) string {
	t.Helper()
	topUp := GetTopUpByTradeNo(tradeNo)
	require.NotNil(t, topUp)
	return topUp.Status
}

func countUserSubscriptionsForPaymentGuardTest(t *testing.T, userID int) int64 {
	t.Helper()
	var count int64
	require.NoError(t, DB.Model(&UserSubscription{}).Where("user_id = ?", userID).Count(&count).Error)
	return count
}

func getUserQuotaForPaymentGuardTest(t *testing.T, userID int) int {
	t.Helper()
	var user User
	require.NoError(t, DB.Select("quota").Where("id = ?", userID).First(&user).Error)
	return user.Quota
}

func TestRechargeWaffoPancake_RejectsMismatchedPaymentMethod(t *testing.T) {
	truncateTables(t)

	insertUserForPaymentGuardTest(t, 101, 0)
	insertTopUpForPaymentGuardTest(t, "waffo-pancake-guard", 101, PaymentMethodStripe)

	err := RechargeWaffoPancake("waffo-pancake-guard")
	require.Error(t, err)

	topUp := GetTopUpByTradeNo("waffo-pancake-guard")
	require.NotNil(t, topUp)
	assert.Equal(t, common.TopUpStatusPending, topUp.Status)
	assert.Equal(t, 0, getUserQuotaForPaymentGuardTest(t, 101))
}

func TestUpdatePendingTopUpStatus_RejectsMismatchedPaymentMethod(t *testing.T) {
	testCases := []struct {
		name                  string
		tradeNo               string
		storedPaymentMethod   string
		expectedPaymentMethod string
		targetStatus          string
	}{
		{
			name:                  "stripe expire",
			tradeNo:               "stripe-expire-guard",
			storedPaymentMethod:   PaymentMethodCreem,
			expectedPaymentMethod: PaymentMethodStripe,
			targetStatus:          common.TopUpStatusExpired,
		},
		{
			name:                  "waffo failed",
			tradeNo:               "waffo-failed-guard",
			storedPaymentMethod:   PaymentMethodStripe,
			expectedPaymentMethod: PaymentMethodWaffo,
			targetStatus:          common.TopUpStatusFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			truncateTables(t)
			insertUserForPaymentGuardTest(t, 150, 0)
			insertTopUpForPaymentGuardTest(t, tc.tradeNo, 150, tc.storedPaymentMethod)

			err := UpdatePendingTopUpStatus(tc.tradeNo, tc.expectedPaymentMethod, tc.targetStatus)
			require.ErrorIs(t, err, ErrPaymentMethodMismatch)
			assert.Equal(t, common.TopUpStatusPending, getTopUpStatusForPaymentGuardTest(t, tc.tradeNo))
		})
	}
}

func TestCompleteSubscriptionOrder_RejectsMismatchedPaymentMethod(t *testing.T) {
	truncateTables(t)

	insertUserForPaymentGuardTest(t, 202, 0)
	plan := insertSubscriptionPlanForPaymentGuardTest(t, 301)
	insertSubscriptionOrderForPaymentGuardTest(t, "sub-guard-order", 202, plan.Id, PaymentMethodStripe)

	err := CompleteSubscriptionOrder("sub-guard-order", `{"provider":"epay"}`, "alipay")
	require.ErrorIs(t, err, ErrPaymentMethodMismatch)

	order := GetSubscriptionOrderByTradeNo("sub-guard-order")
	require.NotNil(t, order)
	assert.Equal(t, common.TopUpStatusPending, order.Status)
	assert.Zero(t, countUserSubscriptionsForPaymentGuardTest(t, 202))

	topUp := GetTopUpByTradeNo("sub-guard-order")
	assert.Nil(t, topUp)
}

func TestExpireSubscriptionOrder_RejectsMismatchedPaymentMethod(t *testing.T) {
	truncateTables(t)

	insertUserForPaymentGuardTest(t, 303, 0)
	plan := insertSubscriptionPlanForPaymentGuardTest(t, 401)
	insertSubscriptionOrderForPaymentGuardTest(t, "sub-expire-guard", 303, plan.Id, PaymentMethodStripe)

	err := ExpireSubscriptionOrder("sub-expire-guard", PaymentMethodCreem)
	require.ErrorIs(t, err, ErrPaymentMethodMismatch)

	order := GetSubscriptionOrderByTradeNo("sub-expire-guard")
	require.NotNil(t, order)
	assert.Equal(t, common.TopUpStatusPending, order.Status)
}

func TestCompleteSubscriptionOrder_GrantsInviteRebateFromPaidAmount(t *testing.T) {
	truncateTables(t)

	originalEnabled := common.InviteRebateEnabled
	originalRatio := common.InviteRebateRatio
	originalQuotaPerUnit := common.QuotaPerUnit
	originalPrice := operation_setting.Price
	common.InviteRebateEnabled = true
	common.InviteRebateRatio = 0.2
	common.QuotaPerUnit = 500000
	operation_setting.Price = 0.5
	t.Cleanup(func() {
		common.InviteRebateEnabled = originalEnabled
		common.InviteRebateRatio = originalRatio
		common.QuotaPerUnit = originalQuotaPerUnit
		operation_setting.Price = originalPrice
	})

	insertPaymentGuardUserWithName(t, 501, "rebate_inviter", 0)
	insertPaymentGuardUserWithName(t, 502, "rebate_invitee", 501)

	plan := &SubscriptionPlan{
		Id:            601,
		Title:         "Rebate Plan",
		PriceAmount:   10,
		Currency:      "CNY",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   0,
	}
	require.NoError(t, DB.Create(plan).Error)

	order := &SubscriptionOrder{
		UserId:        502,
		PlanId:        plan.Id,
		Money:         plan.PriceAmount,
		TradeNo:       "sub-rebate-order",
		PaymentMethod: "alipay",
		Status:        common.TopUpStatusPending,
		CreateTime:    time.Now().Unix(),
	}
	require.NoError(t, order.Insert())

	require.NoError(t, CompleteSubscriptionOrder(order.TradeNo, `{"provider":"epay"}`, "alipay"))

	expectedBaseQuota := int((10 / operation_setting.Price) * common.QuotaPerUnit)
	expectedRebateQuota := int(float64(expectedBaseQuota) * common.InviteRebateRatio)

	var inviter User
	require.NoError(t, DB.First(&inviter, 501).Error)
	assert.Equal(t, expectedRebateQuota, inviter.AffQuota)
	assert.Equal(t, expectedRebateQuota, inviter.AffHistoryQuota)

	var rebate InviteRebate
	require.NoError(t, DB.Where("source_type = ? AND source_id = ?", InviteRebateSourceSubscription, order.TradeNo).First(&rebate).Error)
	assert.Equal(t, 501, rebate.InviterId)
	assert.Equal(t, 502, rebate.InviteeId)
	assert.Equal(t, expectedBaseQuota, rebate.BaseQuota)
	assert.Equal(t, expectedRebateQuota, rebate.RebateQuota)
}
