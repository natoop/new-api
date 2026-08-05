package common

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/stretchr/testify/require"
)

func TestZZDHEndpointTypesSeparateVideoAndImageTracks(t *testing.T) {
	for _, model := range []string{
		"qwen-image-2.0",
		"qwen-image-2.0-pro",
		"qwen-image-edit-max",
		"qwen-image-max",
	} {
		require.Equal(t,
			[]constant.EndpointType{constant.EndpointTypeOpenAI},
			GetEndpointTypesByChannelType(constant.ChannelTypeZZDH, model),
			model,
		)
	}
	require.Equal(t,
		[]constant.EndpointType{constant.EndpointTypeOpenAIVideo},
		GetEndpointTypesByChannelType(constant.ChannelTypeZZDH, "wan2.7-image"),
	)
	require.Equal(t,
		[]constant.EndpointType{constant.EndpointTypeOpenAIVideo},
		GetEndpointTypesByChannelType(constant.ChannelTypeZZDH, "doubao-seedance-2-480p"),
	)
	require.Equal(t,
		[]constant.EndpointType{constant.EndpointTypeOpenAIVideo},
		GetEndpointTypesByChannelType(constant.ChannelTypeZZDH, "zzdh-Minimax-h3-1080p"),
	)
}

func TestZZDHMapsToOpenAIAPITypeForImageRelay(t *testing.T) {
	apiType, ok := ChannelType2APIType(constant.ChannelTypeZZDH)
	require.True(t, ok)
	require.Equal(t, constant.APITypeOpenAI, apiType)
}
