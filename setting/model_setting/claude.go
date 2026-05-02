package model_setting

import (
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/setting/config"
)

//var claudeHeadersSettings = map[string][]string{}
//
//var ClaudeThinkingAdapterEnabled = true
//var ClaudeThinkingAdapterMaxTokens = 8192
//var ClaudeThinkingAdapterBudgetTokensPercentage = 0.8

// ClaudeSettings 定义Claude模型的配置
type ClaudeSettings struct {
	HeadersSettings                       map[string]map[string][]string `json:"model_headers_settings"`
	DefaultMaxTokens                      map[string]int                 `json:"default_max_tokens"`
	ThinkingAdapterEnabled                bool                           `json:"thinking_adapter_enabled"`
	ThinkingAdapterBudgetTokensPercentage float64                        `json:"thinking_adapter_budget_tokens_percentage"`
	ExcludedSubscriptionPlanIDs           []int                          `json:"excluded_subscription_plan_ids"`
}

// 默认配置
var defaultClaudeSettings = ClaudeSettings{
	HeadersSettings:        map[string]map[string][]string{},
	ThinkingAdapterEnabled: true,
	DefaultMaxTokens: map[string]int{
		"default": 8192,
	},
	ThinkingAdapterBudgetTokensPercentage: 0.8,
	ExcludedSubscriptionPlanIDs:           []int{},
}

// 全局实例
var claudeSettings = defaultClaudeSettings

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("claude", &claudeSettings)
}

// GetClaudeSettings 获取Claude配置
func GetClaudeSettings() *ClaudeSettings {
	// check default max tokens must have default key
	if _, ok := claudeSettings.DefaultMaxTokens["default"]; !ok {
		claudeSettings.DefaultMaxTokens["default"] = 8192
	}
	claudeSettings.ExcludedSubscriptionPlanIDs = normalizeSubscriptionPlanIDs(claudeSettings.ExcludedSubscriptionPlanIDs)
	return &claudeSettings
}

func IsClaudeModelName(modelName string) bool {
	modelName = strings.ToLower(strings.TrimSpace(modelName))
	if modelName == "" {
		return false
	}
	modelName = strings.TrimSuffix(modelName, "-thinking")
	if strings.HasPrefix(modelName, "claude-") {
		return true
	}
	if strings.HasPrefix(modelName, "anthropic/claude-") {
		return true
	}
	if strings.HasPrefix(modelName, "anthropic.claude-") {
		return true
	}
	return false
}

func GetClaudeExcludedSubscriptionPlanIDsForModel(modelName string) []int {
	if !IsClaudeModelName(modelName) {
		return nil
	}
	return GetClaudeExcludedSubscriptionPlanIDs()
}

func GetClaudeExcludedSubscriptionPlanIDs() []int {
	return copySubscriptionPlanIDs(GetClaudeSettings().ExcludedSubscriptionPlanIDs)
}

func copySubscriptionPlanIDs(ids []int) []int {
	if len(ids) == 0 {
		return nil
	}
	copied := make([]int, len(ids))
	copy(copied, ids)
	return copied
}

func normalizeSubscriptionPlanIDs(ids []int) []int {
	if len(ids) == 0 {
		return []int{}
	}
	seen := make(map[int]struct{}, len(ids))
	normalized := make([]int, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}
	if len(normalized) == 0 {
		return []int{}
	}
	return normalized
}

func (c *ClaudeSettings) WriteHeaders(originModel string, httpHeader *http.Header) {
	if headers, ok := c.HeadersSettings[originModel]; ok {
		for headerKey, headerValues := range headers {
			mergedValues := normalizeHeaderListValues(
				append(append([]string(nil), httpHeader.Values(headerKey)...), headerValues...),
			)
			if len(mergedValues) == 0 {
				continue
			}
			httpHeader.Set(headerKey, strings.Join(mergedValues, ","))
		}
	}
}

func normalizeHeaderListValues(values []string) []string {
	normalizedValues := make([]string, 0, len(values))
	seenValues := make(map[string]struct{}, len(values))
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			normalizedItem := strings.TrimSpace(item)
			if normalizedItem == "" {
				continue
			}
			if _, exists := seenValues[normalizedItem]; exists {
				continue
			}
			seenValues[normalizedItem] = struct{}{}
			normalizedValues = append(normalizedValues, normalizedItem)
		}
	}
	return normalizedValues
}

func (c *ClaudeSettings) GetDefaultMaxTokens(model string) int {
	if maxTokens, ok := c.DefaultMaxTokens[model]; ok {
		return maxTokens
	}
	return c.DefaultMaxTokens["default"]
}
