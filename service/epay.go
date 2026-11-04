package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Calcium-Ion/go-epay/epay"
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/shopspring/decimal"
)

func GetCallbackAddress() string {
	if operation_setting.CustomCallbackAddress == "" {
		return system_setting.ServerAddress
	}
	return operation_setting.CustomCallbackAddress
}

// ConfirmEpayPayment requires an independently queried, paid gateway order.
// A correctly signed callback alone is insufficient to authorize a credit.
func ConfirmEpayPayment(ctx context.Context, callback *epay.VerifyRes, merchantID string, expectedMoney float64, expectedMethod string) error {
	if callback == nil || callback.ServiceTradeNo == "" || callback.TradeNo == "" ||
		callback.TradeStatus != epay.StatusTradeSuccess || !callback.VerifyStatus {
		return errors.New("invalid payment confirmation")
	}
	if merchantID == "" || merchantID != operation_setting.EpayId {
		return errors.New("payment merchant mismatch")
	}
	if callback.Type != expectedMethod || expectedMethod == "" {
		return errors.New("payment method mismatch")
	}
	if expectedMoney <= 0 || math.IsNaN(expectedMoney) || math.IsInf(expectedMoney, 0) {
		return errors.New("invalid expected payment amount")
	}
	expected, err := decimal.NewFromString(strconv.FormatFloat(expectedMoney, 'f', 2, 64))
	if err != nil || !expected.IsPositive() {
		return errors.New("invalid expected payment amount")
	}
	callbackAmount, err := decimal.NewFromString(callback.Money)
	if err != nil || !callbackAmount.Equal(expected) {
		return errors.New("callback payment amount mismatch")
	}
	endpoint, err := url.Parse(operation_setting.PayAddress)
	if err != nil || endpoint.Host == "" || (endpoint.Scheme != "https" && endpoint.Scheme != "http") ||
		endpoint.User != nil || operation_setting.EpayKey == "" {
		return errors.New("invalid payment gateway configuration")
	}
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/api.php"
	endpoint.RawPath, endpoint.RawQuery, endpoint.Fragment = "", "", ""
	query := url.Values{
		"act":          {"order"},
		"pid":          {operation_setting.EpayId},
		"key":          {operation_setting.EpayKey},
		"out_trade_no": {callback.ServiceTradeNo},
	}
	endpoint.RawQuery = query.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return errors.New("could not prepare payment verification")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "NewAPI-payment-verification/1.0")
	client := &http.Client{
		Timeout: 8 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	response, err := client.Do(request)
	if err != nil {
		// Transport errors include the URL, which contains the merchant key.
		return errors.New("payment gateway verification unavailable")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("payment gateway verification HTTP status %d", response.StatusCode)
	}
	const maxResponseBytes = 64 * 1024
	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil || len(body) > maxResponseBytes {
		return errors.New("invalid payment gateway verification response")
	}
	var order struct {
		Code       json.Number `json:"code"`
		Status     json.Number `json:"status"`
		MerchantID json.Number `json:"pid"`
		TradeNo    string      `json:"trade_no"`
		OutTradeNo string      `json:"out_trade_no"`
		Money      string      `json:"money"`
		Type       string      `json:"type"`
	}
	if err := common.Unmarshal(body, &order); err != nil {
		return errors.New("invalid payment gateway verification response")
	}
	if order.Code.String() != "1" || order.Status.String() != "1" {
		return errors.New("payment gateway has not confirmed a paid order")
	}
	if order.MerchantID.String() != merchantID || order.OutTradeNo != callback.ServiceTradeNo ||
		order.TradeNo != callback.TradeNo || order.Type != expectedMethod {
		return errors.New("payment gateway order details mismatch")
	}
	paidAmount, err := decimal.NewFromString(order.Money)
	if err != nil || !paidAmount.Equal(expected) {
		return errors.New("payment gateway amount mismatch")
	}
	return nil
}
