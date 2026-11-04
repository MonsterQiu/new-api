package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/Calcium-Ion/go-epay/epay"
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestEpayGatewayPaymentConfirmation(t *testing.T) {
	confirmPaymentComplianceForTest(t)
	previousDB, previousLogDB := model.DB, model.LOG_DB
	mainType, logType := common.MainDatabaseType(), common.LogDatabaseType()
	previousRedis, previousRebate := common.RedisEnabled, common.InviteRebateEnabled
	previousAddress, previousID, previousKey := operation_setting.PayAddress, operation_setting.EpayId, operation_setting.EpayKey
	previousMethods := operation_setting.PayMethods
	t.Cleanup(func() {
		model.DB, model.LOG_DB = previousDB, previousLogDB
		common.SetDatabaseTypes(mainType, logType)
		common.RedisEnabled, common.InviteRebateEnabled = previousRedis, previousRebate
		operation_setting.PayAddress, operation_setting.EpayId, operation_setting.EpayKey = previousAddress, previousID, previousKey
		operation_setting.PayMethods = previousMethods
	})
	common.RedisEnabled, common.InviteRebateEnabled = false, false
	operation_setting.EpayId, operation_setting.EpayKey = "12345", "payment-regression-test-key"
	operation_setting.PayMethods = []map[string]string{{"type": "alipay"}}

	var dialector gorm.Dialector = sqlite.Open(t.TempDir() + "/payments.db")
	databaseType := common.DatabaseTypeSQLite
	if dsn := os.Getenv("PAYMENT_TEST_MYSQL_DSN"); dsn != "" {
		dialector, databaseType = mysql.Open(dsn), common.DatabaseTypeMySQL
	}
	if dsn := os.Getenv("PAYMENT_TEST_POSTGRES_DSN"); dsn != "" {
		dialector, databaseType = postgres.Open(dsn), common.DatabaseTypePostgreSQL
	}
	db, err := gorm.Open(dialector, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	model.DB, model.LOG_DB = db, db
	common.SetDatabaseTypes(databaseType, databaseType)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.TopUp{}, &model.Log{}, &model.SubscriptionPlan{}, &model.SubscriptionOrder{}, &model.UserSubscription{}))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(4)
	t.Cleanup(func() { _ = sqlDB.Close() })
	versionQuery := "SELECT version()"
	if databaseType == common.DatabaseTypeSQLite {
		versionQuery = "SELECT sqlite_version()"
	}
	var databaseVersion string
	require.NoError(t, db.Raw(versionQuery).Scan(&databaseVersion).Error)
	t.Logf("database=%s version=%s", db.Dialector.Name(), databaseVersion)

	cases := []struct {
		name     string
		reply    func(map[string]any)
		callback func(map[string]string)
		httpCode int
		rawBody  string
		paid     bool
		replay   bool
		paused   bool
		badSign  bool
		canceled bool
	}{
		{name: "paid", paid: true},
		{name: "paid_string_merchant", paid: true, reply: func(m map[string]any) { m["pid"] = "12345" }},
		{name: "duplicate_paid_callback", paid: true, replay: true},
		{name: "missing_gateway_order", reply: func(m map[string]any) { m["code"] = 0 }},
		{name: "gateway_unpaid", reply: func(m map[string]any) { m["status"] = 0 }},
		{name: "gateway_wrong_amount", reply: func(m map[string]any) { m["money"] = "0.01" }},
		{name: "callback_wrong_amount", callback: func(m map[string]string) { m["money"] = "0.01" }},
		{name: "gateway_wrong_merchant", reply: func(m map[string]any) { m["pid"] = 54321 }},
		{name: "callback_wrong_merchant", callback: func(m map[string]string) { m["pid"] = "54321" }},
		{name: "gateway_wrong_order", reply: func(m map[string]any) { m["out_trade_no"] = "another-order" }},
		{name: "gateway_wrong_transaction", reply: func(m map[string]any) { m["trade_no"] = "another-transaction" }},
		{name: "gateway_wrong_method", reply: func(m map[string]any) { m["type"] = "wxpay" }},
		{name: "callback_wrong_method", callback: func(m map[string]string) { m["type"] = "wxpay" }},
		{name: "gateway_missing_merchant", reply: func(m map[string]any) { delete(m, "pid") }},
		{name: "gateway_invalid_money", reply: func(m map[string]any) { m["money"] = "NaN" }},
		{name: "gateway_http_error", httpCode: http.StatusBadGateway},
		{name: "gateway_invalid_json", rawBody: "not-json"},
		{name: "gateway_response_too_large", rawBody: strings.Repeat("x", 65537)},
		{name: "gateway_redirect", httpCode: http.StatusFound},
		{name: "invalid_signature", badSign: true},
		{name: "canceled_gateway_request", canceled: true},
		{name: "payment_disabled", paused: true},
	}
	entries := []struct {
		name         string
		path         string
		handler      gin.HandlerFunc
		subscription bool
		redirect     bool
	}{
		{"wallet", "/api/user/epay/notify", EpayNotify, false, false},
		{"subscription_notify", "/api/subscription/epay/notify", SubscriptionEpayNotify, true, false},
		{"subscription_return", "/api/subscription/epay/return", SubscriptionEpayReturn, true, true},
	}
	counter := 0
	for _, entry := range entries {
		t.Run(entry.name, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					counter++
					user := model.User{Username: fmt.Sprintf("payment-user-%d", counter), AffCode: fmt.Sprintf("payment-test-%d", counter), Role: 1, Status: 1, Group: "default", Quota: 100}
					require.NoError(t, db.Create(&user).Error)
					tradeNo := fmt.Sprintf("PAYMENT-REGRESSION-%d", counter)
					if entry.subscription {
						plan := model.SubscriptionPlan{Title: tradeNo, PriceAmount: 10, Currency: "USD", DurationUnit: model.SubscriptionDurationMonth, DurationValue: 1, TotalAmount: 5000000, Enabled: true}
						require.NoError(t, db.Create(&plan).Error)
						model.InvalidateSubscriptionPlanCache(plan.Id)
						order := model.SubscriptionOrder{UserId: user.Id, PlanId: plan.Id, Money: 10, TradeNo: tradeNo, PaymentMethod: "alipay", PaymentProvider: model.PaymentProviderEpay, Status: common.TopUpStatusPending}
						require.NoError(t, db.Create(&order).Error)
					} else {
						order := model.TopUp{UserId: user.Id, Amount: 10, Money: 10, TradeNo: tradeNo, PaymentMethod: "alipay", PaymentProvider: model.PaymentProviderEpay, Status: common.TopUpStatusPending}
						require.NoError(t, db.Create(&order).Error)
					}
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						assert.Equal(t, "/api.php", r.URL.Path)
						assert.Equal(t, "order", r.URL.Query().Get("act"))
						assert.Equal(t, tradeNo, r.URL.Query().Get("out_trade_no"))
						assert.Equal(t, "12345", r.URL.Query().Get("pid"))
						assert.Equal(t, "payment-regression-test-key", r.URL.Query().Get("key"))
						if tc.httpCode != 0 {
							w.Header().Set("Location", "/must-not-follow")
							w.WriteHeader(tc.httpCode)
						}
						if tc.rawBody != "" {
							_, _ = w.Write([]byte(tc.rawBody))
							return
						}
						reply := map[string]any{"code": 1, "status": 1, "pid": 12345, "trade_no": "gateway-transaction", "out_trade_no": tradeNo, "money": "10.00", "type": "alipay"}
						if tc.reply != nil {
							tc.reply(reply)
						}
						body, marshalErr := common.Marshal(reply)
						require.NoError(t, marshalErr)
						_, _ = w.Write(body)
					}))
					defer server.Close()
					operation_setting.PayAddress = server.URL
					operation_setting.PayMethods = []map[string]string{{"type": "alipay"}}
					if tc.paused {
						operation_setting.PayMethods = nil
					}
					callback := map[string]string{"pid": "12345", "type": "alipay", "out_trade_no": tradeNo, "trade_no": "gateway-transaction", "name": "payment test", "money": "10.00", "trade_status": epay.StatusTradeSuccess}
					if tc.callback != nil {
						tc.callback(callback)
					}
					callback = epay.GenerateParams(callback, operation_setting.EpayKey)
					if tc.badSign {
						callback["sign"] = "invalid-signature"
					}
					form := url.Values{}
					for key, value := range callback {
						form.Set(key, value)
					}
					attempts := 1
					if tc.replay {
						attempts = 2
					}
					for attempt := range attempts {
						method := http.MethodPost
						path := entry.path
						request := httptest.NewRequest(method, path, strings.NewReader(form.Encode()))
						request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
						if tc.canceled {
							canceled, cancel := context.WithCancel(request.Context())
							cancel()
							request = request.WithContext(canceled)
						}
						// Exercise GET parsing on the second idempotency request.
						if attempt == 1 {
							request = httptest.NewRequest(http.MethodGet, path+"?"+form.Encode(), nil)
						}
						response := httptest.NewRecorder()
						context, _ := gin.CreateTestContext(response)
						context.Request = request
						entry.handler(context)
						if entry.redirect {
							if tc.paid {
								assert.Contains(t, response.Header().Get("Location"), "pay=success")
							} else {
								assert.Contains(t, response.Header().Get("Location"), "pay=fail")
							}
						} else if tc.paid {
							assert.Equal(t, "success", response.Body.String())
						} else {
							assert.Equal(t, "fail", response.Body.String())
						}
					}
					expectedStatus := common.TopUpStatusPending
					if tc.paid {
						expectedStatus = common.TopUpStatusSuccess
					}
					require.NoError(t, db.First(&user, user.Id).Error)
					if entry.subscription {
						order := model.GetSubscriptionOrderByTradeNo(tradeNo)
						require.NotNil(t, order)
						assert.Equal(t, expectedStatus, order.Status)
						var count int64
						require.NoError(t, db.Model(&model.UserSubscription{}).Where("user_id = ?", user.Id).Count(&count).Error)
						expectedCount := int64(0)
						if tc.paid {
							expectedCount = 1
						}
						assert.Equal(t, expectedCount, count)
						assert.Equal(t, 100, user.Quota)
					} else {
						order := model.GetTopUpByTradeNo(tradeNo)
						require.NotNil(t, order)
						assert.Equal(t, expectedStatus, order.Status)
						expectedQuota := 100
						if tc.paid {
							expectedQuota += 5000000
						}
						assert.Equal(t, expectedQuota, user.Quota)
					}
				})
			}
		})
	}
}
