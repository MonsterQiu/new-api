package openai

import (
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

const defaultResponsesImageGenerationModel = "gpt-image-2"

func recordResponsesToolUsage(info *relaycommon.RelayInfo, resp *dto.OpenAIResponsesResponse) {
	if info == nil || resp == nil || resp.ToolUsage == nil || resp.ToolUsage.ImageGen == nil {
		return
	}
	if info.ResponsesUsageInfo == nil {
		info.ResponsesUsageInfo = &relaycommon.ResponsesUsageInfo{}
	}

	modelName := resp.GetImageGenerationToolModel()
	if modelName == "" {
		modelName = defaultResponsesImageGenerationModel
	}

	info.ResponsesUsageInfo.ImageGeneration = &relaycommon.ResponsesImageGenerationUsageInfo{
		ModelName: modelName,
		Usage:     resp.ToolUsage.ImageGen.ToUsage(),
	}
}
