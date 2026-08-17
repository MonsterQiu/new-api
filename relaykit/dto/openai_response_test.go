package dto

import (
	"testing"

	kitutil "github.com/QuantumNous/new-api/relaykit/relayconvert/kitutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponsesStreamResponseAcceptsObjectArguments(t *testing.T) {
	var streamResp ResponsesStreamResponse
	err := kitutil.UnmarshalJsonStr(`{"type":"response.output_item.added","item":{"type":"function_call","id":"item_1","call_id":"call_1","name":"search","arguments":{"q":"hello"}}}`, &streamResp)
	require.NoError(t, err)
	require.NotNil(t, streamResp.Item)
	assert.Equal(t, `{"q":"hello"}`, streamResp.Item.ArgumentsString())
}

func TestResponsesStreamResponseAcceptsStringArguments(t *testing.T) {
	var streamResp ResponsesStreamResponse
	err := kitutil.UnmarshalJsonStr(`{"type":"response.output_item.added","item":{"type":"function_call","id":"item_1","call_id":"call_1","name":"search","arguments":"{\"q\":\"hello\"}"}}`, &streamResp)
	require.NoError(t, err)
	require.NotNil(t, streamResp.Item)
	assert.Equal(t, `{"q":"hello"}`, streamResp.Item.ArgumentsString())
}
