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
		DB.Exec("DELETE FROM top_ups")
		DB.Exec("DELETE FROM users")
		DB.Exec("DELETE FROM logs")
	})
}

func createRechargeTestUser(t *testing.T) *User {
	t.Helper()

	name := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
	user := &User{
		Username:    fmt.Sprintf("%s_user", name),
		Password:    "pass12345",
		DisplayName: "tester",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		Email:       fmt.Sprintf("%s@example.com", name),
		Group:       "default",
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
