package model

import (
	"fmt"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTopUpRechargeTest(t *testing.T) {
	t.Helper()

	require.NoError(t, DB.AutoMigrate(&TopUp{}))

	t.Cleanup(func() {
		DB.Exec("DELETE FROM invite_rebates")
		DB.Exec("DELETE FROM top_ups")
		DB.Exec("DELETE FROM users")
		DB.Exec("DELETE FROM logs")
	})
}

func createRechargeTestUser(t *testing.T) *User {
	return createRechargeTestUserWithSuffix(t, "user", 0)
}

func createRechargeTestUserWithSuffix(t *testing.T, suffix string, inviterID int) *User {
	t.Helper()

	name := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
	user := &User{
		Username:    fmt.Sprintf("%s_%s", name, suffix),
		Password:    "pass12345",
		DisplayName: "tester",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		Email:       fmt.Sprintf("%s_%s@example.com", name, suffix),
		Group:       "default",
		AffCode:     fmt.Sprintf("%s_%s_aff", name, suffix),
		InviterId:   inviterID,
	}
	require.NoError(t, DB.Create(user).Error)
	return user
}

func createRechargeTestTopUp(t *testing.T, userID int, tradeNo string, method string, money float64) *TopUp {
	t.Helper()

	topUp := &TopUp{
		UserId:        userID,
		Amount:        10,
		Money:         money,
		TradeNo:       tradeNo,
		PaymentMethod: method,
		CreateTime:    1,
		Status:        common.TopUpStatusPending,
	}
	require.NoError(t, DB.Create(topUp).Error)
	return topUp
}

func TestInviteRebateAutoMigrateIsRepeatable(t *testing.T) {
	require.NoError(t, DB.AutoMigrate(&InviteRebate{}))
	require.NoError(t, DB.AutoMigrate(&InviteRebate{}))
}

func TestRechargeRejectsPaymentMethodMismatch(t *testing.T) {
	setupTopUpRechargeTest(t)

	user := createRechargeTestUser(t)
	topUp := createRechargeTestTopUp(t, user.Id, "USR95NOFAKE", "alipay", 10)

	err := Recharge(topUp.TradeNo, "cus_fake", "127.0.0.1")
	require.Error(t, err)

	var reloadedTopUp TopUp
	require.NoError(t, DB.First(&reloadedTopUp, topUp.Id).Error)
	assert.Equal(t, common.TopUpStatusPending, reloadedTopUp.Status)
	assert.Zero(t, reloadedTopUp.CompleteTime)

	var reloadedUser User
	require.NoError(t, DB.First(&reloadedUser, user.Id).Error)
	assert.Zero(t, reloadedUser.Quota)
	assert.Equal(t, "", reloadedUser.StripeCustomer)
}

func TestRechargeSucceedsForExpectedStripeMethod(t *testing.T) {
	setupTopUpRechargeTest(t)

	user := createRechargeTestUser(t)
	topUp := createRechargeTestTopUp(t, user.Id, "ref_valid_stripe_topup", "stripe", 2)

	err := Recharge(topUp.TradeNo, "cus_live", "127.0.0.1")
	require.NoError(t, err)

	var reloadedTopUp TopUp
	require.NoError(t, DB.First(&reloadedTopUp, topUp.Id).Error)
	assert.Equal(t, common.TopUpStatusSuccess, reloadedTopUp.Status)
	assert.NotZero(t, reloadedTopUp.CompleteTime)

	var reloadedUser User
	require.NoError(t, DB.First(&reloadedUser, user.Id).Error)
	assert.Equal(t, int(2*common.QuotaPerUnit), reloadedUser.Quota)
	assert.Equal(t, "cus_live", reloadedUser.StripeCustomer)
}

func TestRechargeGrantsInviteRebateOnce(t *testing.T) {
	setupTopUpRechargeTest(t)

	originalEnabled := common.InviteRebateEnabled
	originalRatio := common.InviteRebateRatio
	originalQuotaPerUnit := common.QuotaPerUnit
	common.InviteRebateEnabled = true
	common.InviteRebateRatio = 0.2
	common.QuotaPerUnit = 500000
	t.Cleanup(func() {
		common.InviteRebateEnabled = originalEnabled
		common.InviteRebateRatio = originalRatio
		common.QuotaPerUnit = originalQuotaPerUnit
	})

	inviter := createRechargeTestUserWithSuffix(t, "inviter", 0)
	invitee := createRechargeTestUserWithSuffix(t, "invitee", inviter.Id)
	topUp := createRechargeTestTopUp(t, invitee.Id, "ref_invite_rebate_stripe", PaymentMethodStripe, 10)

	require.NoError(t, Recharge(topUp.TradeNo, "cus_rebate", "127.0.0.1"))

	expectedBaseQuota := int(10 * common.QuotaPerUnit)
	expectedRebateQuota := int(float64(expectedBaseQuota) * common.InviteRebateRatio)

	var reloadedInviter User
	require.NoError(t, DB.First(&reloadedInviter, inviter.Id).Error)
	assert.Equal(t, expectedRebateQuota, reloadedInviter.AffQuota)
	assert.Equal(t, expectedRebateQuota, reloadedInviter.AffHistoryQuota)

	var rebate InviteRebate
	require.NoError(t, DB.Where("source_type = ? AND source_id = ?", InviteRebateSourceTopUp, topUp.TradeNo).First(&rebate).Error)
	assert.Equal(t, inviter.Id, rebate.InviterId)
	assert.Equal(t, invitee.Id, rebate.InviteeId)
	assert.Equal(t, expectedBaseQuota, rebate.BaseQuota)
	assert.Equal(t, expectedRebateQuota, rebate.RebateQuota)

	_, err := GrantInviteRebate(InviteRebateGrantParams{
		InviteeId:     invitee.Id,
		SourceType:    InviteRebateSourceTopUp,
		SourceId:      topUp.TradeNo,
		PaymentMethod: topUp.PaymentMethod,
		BaseQuota:     expectedBaseQuota,
		BaseAmount:    topUp.Money,
		Currency:      "USD",
	})
	require.NoError(t, err)

	require.NoError(t, DB.First(&reloadedInviter, inviter.Id).Error)
	assert.Equal(t, expectedRebateQuota, reloadedInviter.AffQuota)
	assert.Equal(t, expectedRebateQuota, reloadedInviter.AffHistoryQuota)

	var rebateCount int64
	require.NoError(t, DB.Model(&InviteRebate{}).Where("source_type = ? AND source_id = ?", InviteRebateSourceTopUp, topUp.TradeNo).Count(&rebateCount).Error)
	assert.EqualValues(t, 1, rebateCount)
}

func TestGetUserInviteRebatesReturnsSanitizedRows(t *testing.T) {
	setupTopUpRechargeTest(t)

	originalEnabled := common.InviteRebateEnabled
	originalRatio := common.InviteRebateRatio
	originalQuotaPerUnit := common.QuotaPerUnit
	common.InviteRebateEnabled = true
	common.InviteRebateRatio = 0.2
	common.QuotaPerUnit = 500000
	t.Cleanup(func() {
		common.InviteRebateEnabled = originalEnabled
		common.InviteRebateRatio = originalRatio
		common.QuotaPerUnit = originalQuotaPerUnit
	})

	inviter := createRechargeTestUserWithSuffix(t, "inviter_view", 0)
	invitee := createRechargeTestUserWithSuffix(t, "invitee_view", inviter.Id)
	topUp := createRechargeTestTopUp(t, invitee.Id, "ref_invite_rebate_view_stripe", PaymentMethodStripe, 10)

	require.NoError(t, Recharge(topUp.TradeNo, "cus_rebate_view", "127.0.0.1"))

	rebates, total, err := GetUserInviteRebates(inviter.Id, &common.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, rebates, 1)
	assert.Equal(t, InviteRebateSourceTopUp, rebates[0].SourceType)
	assert.Equal(t, "ref_...ripe", rebates[0].SourceId)
	assert.Equal(t, PaymentMethodStripe, rebates[0].PaymentMethod)
	assert.Equal(t, int(10*common.QuotaPerUnit), rebates[0].BaseQuota)
	assert.Equal(t, int(float64(10*common.QuotaPerUnit)*common.InviteRebateRatio), rebates[0].RebateQuota)
}

func TestInviteRebateRatioOptionRejectsInvalidValues(t *testing.T) {
	require.NoError(t, validateOptionValue("InviteRebateRatio", "0"))
	require.NoError(t, validateOptionValue("InviteRebateRatio", "0.2"))
	require.NoError(t, validateOptionValue("InviteRebateRatio", "1"))
	require.Error(t, validateOptionValue("InviteRebateRatio", "-0.01"))
	require.Error(t, validateOptionValue("InviteRebateRatio", "1.01"))
	require.Error(t, validateOptionValue("InviteRebateRatio", "NaN"))
}
