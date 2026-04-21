package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
)

func TestShouldChannelTestUseStream(t *testing.T) {
	tests := []struct {
		name         string
		channel      *model.Channel
		modelName    string
		endpointType string
		want         bool
	}{
		{
			name:    "codex channel defaults to stream",
			channel: &model.Channel{Type: constant.ChannelTypeCodex},
			want:    true,
		},
		{
			name:         "codex compact endpoint stays non-stream",
			channel:      &model.Channel{Type: constant.ChannelTypeCodex},
			endpointType: string(constant.EndpointTypeOpenAIResponseCompact),
			want:         false,
		},
		{
			name:         "image generation stays non-stream",
			channel:      &model.Channel{Type: constant.ChannelTypeCodex},
			endpointType: string(constant.EndpointTypeImageGeneration),
			want:         false,
		},
		{
			name:      "regular openai channel stays non-stream",
			channel:   &model.Channel{Type: constant.ChannelTypeOpenAI},
			modelName: "gpt-4o-mini",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldChannelTestUseStream(tt.channel, tt.modelName, tt.endpointType)
			if got != tt.want {
				t.Fatalf("shouldChannelTestUseStream() = %v, want %v", got, tt.want)
			}
		})
	}
}
