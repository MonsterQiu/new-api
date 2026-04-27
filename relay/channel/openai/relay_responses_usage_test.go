package openai

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"

	"github.com/gin-gonic/gin"
)

func newResponsesUsageTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	return ctx, recorder
}

func TestOaiResponsesHandlerMapsImageTokenDetails(t *testing.T) {
	ctx, _ := newResponsesUsageTestContext()
	body := `{
		"id":"resp_1",
		"object":"response",
		"created_at":1,
		"output":[{"type":"image_generation_call","quality":"high","size":"1024x1024"}],
		"tools":[{"type":"image_generation","model":"gpt-image-2","quality":"auto","size":"1024x1024"}],
		"tool_usage":{
			"image_gen":{
				"input_tokens":31,
				"input_tokens_details":{"image_tokens":0,"text_tokens":31},
				"output_tokens":439,
				"output_tokens_details":{"image_tokens":439,"text_tokens":0},
				"total_tokens":470
			}
		},
		"usage":{
			"input_tokens":1200,
			"output_tokens":0,
			"total_tokens":4200,
			"input_tokens_details":{
				"cached_tokens":100,
				"cached_creation_tokens":20,
				"text_tokens":800,
				"audio_tokens":30,
				"image_tokens":250
			},
			"output_tokens_details":{
				"text_tokens":10,
				"audio_tokens":20,
				"image_tokens":2900,
				"reasoning_tokens":70
			}
		}
	}`
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}

	info := &relaycommon.RelayInfo{
		ResponsesUsageInfo: &relaycommon.ResponsesUsageInfo{
			BuiltInTools: map[string]*relaycommon.BuildInToolInfo{},
		},
	}
	usage, newAPIErr := OaiResponsesHandler(ctx, info, resp)
	if newAPIErr != nil {
		t.Fatalf("OaiResponsesHandler returned error: %v", newAPIErr)
	}

	if usage.PromptTokens != 1200 || usage.CompletionTokens != 3000 || usage.TotalTokens != 4200 {
		t.Fatalf("usage totals = %+v", usage)
	}
	if usage.PromptTokensDetails.CachedTokens != 100 ||
		usage.PromptTokensDetails.CachedCreationTokens != 20 ||
		usage.PromptTokensDetails.TextTokens != 800 ||
		usage.PromptTokensDetails.AudioTokens != 30 ||
		usage.PromptTokensDetails.ImageTokens != 250 {
		t.Fatalf("prompt token details = %+v", usage.PromptTokensDetails)
	}
	if usage.CompletionTokenDetails.TextTokens != 10 ||
		usage.CompletionTokenDetails.AudioTokens != 20 ||
		usage.CompletionTokenDetails.ImageTokens != 2900 ||
		usage.CompletionTokenDetails.ReasoningTokens != 70 {
		t.Fatalf("completion token details = %+v", usage.CompletionTokenDetails)
	}
	if !ctx.GetBool("image_generation_call") {
		t.Fatal("expected image generation call context flag")
	}
	if info.ResponsesUsageInfo.ImageGeneration == nil {
		t.Fatal("expected image generation tool usage")
	}
	if info.ResponsesUsageInfo.ImageGeneration.ModelName != "gpt-image-2" {
		t.Fatalf("image generation model = %s", info.ResponsesUsageInfo.ImageGeneration.ModelName)
	}
	if info.ResponsesUsageInfo.ImageGeneration.Usage.CompletionTokenDetails.ImageTokens != 439 {
		t.Fatalf("image generation usage = %+v", info.ResponsesUsageInfo.ImageGeneration.Usage)
	}
}

func TestOaiResponsesStreamHandlerMapsImageTokenDetails(t *testing.T) {
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })

	ctx, _ := newResponsesUsageTestContext()
	body := `data: {"type":"response.completed","response":{"id":"resp_1","object":"response","created_at":1,"output":[{"type":"image_generation_call","quality":"medium","size":"1024x1536"}],"tools":[{"type":"image_generation","model":"gpt-image-2","quality":"auto","size":"1024x1024"}],"tool_usage":{"image_gen":{"input_tokens":31,"input_tokens_details":{"image_tokens":0,"text_tokens":31},"output_tokens":439,"output_tokens_details":{"image_tokens":439,"text_tokens":0},"total_tokens":470}},"usage":{"input_tokens":900,"output_tokens":1800,"total_tokens":2700,"input_tokens_details":{"cached_tokens":80,"text_tokens":700,"audio_tokens":20,"image_tokens":100},"output_tokens_details":{"text_tokens":0,"audio_tokens":0,"image_tokens":1800,"reasoning_tokens":0}}}}` + "\n\n" +
		"data: [DONE]\n\n"
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{UpstreamModelName: "gpt-5.5"},
		ResponsesUsageInfo: &relaycommon.ResponsesUsageInfo{
			BuiltInTools: map[string]*relaycommon.BuildInToolInfo{},
		},
	}

	usage, newAPIErr := OaiResponsesStreamHandler(ctx, info, resp)
	if newAPIErr != nil {
		t.Fatalf("OaiResponsesStreamHandler returned error: %v", newAPIErr)
	}

	if usage.PromptTokens != 900 || usage.CompletionTokens != 1800 || usage.TotalTokens != 2700 {
		t.Fatalf("usage totals = %+v", usage)
	}
	if usage.PromptTokensDetails.CachedTokens != 80 ||
		usage.PromptTokensDetails.TextTokens != 700 ||
		usage.PromptTokensDetails.AudioTokens != 20 ||
		usage.PromptTokensDetails.ImageTokens != 100 {
		t.Fatalf("prompt token details = %+v", usage.PromptTokensDetails)
	}
	if usage.CompletionTokenDetails.ImageTokens != 1800 {
		t.Fatalf("completion token details = %+v", usage.CompletionTokenDetails)
	}
	if !ctx.GetBool("image_generation_call") {
		t.Fatal("expected image generation call context flag")
	}
	if info.ResponsesUsageInfo.ImageGeneration == nil {
		t.Fatal("expected image generation tool usage")
	}
	if info.ResponsesUsageInfo.ImageGeneration.Usage.CompletionTokenDetails.ImageTokens != 439 {
		t.Fatalf("image generation usage = %+v", info.ResponsesUsageInfo.ImageGeneration.Usage)
	}
}
