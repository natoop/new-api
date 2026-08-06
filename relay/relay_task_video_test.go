package relay

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsOpenAIVideoAPIRequestRecognizesBothTaskQueryRoutes(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{path: "/v1/video/generations/task_123", want: true},
		{path: "/v1/videos/task_123", want: true},
		{path: "/v1/video/generations", want: false},
		{path: "/v8/videos/generations/task_123", want: false},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			require.Equal(t, test.want, isOpenAIVideoAPIRequest(test.path))
		})
	}
}
