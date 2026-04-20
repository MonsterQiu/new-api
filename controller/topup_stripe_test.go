package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"
)

func TestStripeWebhookRejectsWhenSecretMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalSecret := setting.StripeWebhookSecret
	setting.StripeWebhookSecret = ""
	t.Cleanup(func() {
		setting.StripeWebhookSecret = originalSecret
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/stripe/webhook", strings.NewReader(`{"type":"checkout.session.completed"}`))

	StripeWebhook(ctx)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected missing Stripe webhook secret to return 404, got %d", recorder.Code)
	}
}
