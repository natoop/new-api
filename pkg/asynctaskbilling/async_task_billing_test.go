package asynctaskbilling

import (
	"math"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func outputRule() Rule {
	return Rule{
		Version:         "output-v1",
		Terms:           []Term{{Name: OutputSeconds, MaxValue: 3600}},
		AllowedRounding: []string{RoundingNone},
	}
}

func TestCalculateOutputSecondsWithGroupRatio(t *testing.T) {
	result, err := Calculate(outputRule(), Config{
		Version:  ConfigVersion,
		Rounding: RoundingNone,
		Terms:    map[string]float64{OutputSeconds: 0.12},
	}, map[string]float64{OutputSeconds: 5}, 1.5, 500000)

	require.NoError(t, err)
	assert.Equal(t, 450000, result.Quota)
	assert.Equal(t, "output_seconds * 0.12", result.Snapshot.Expression)
	assert.Equal(t, "5", result.Snapshot.Metrics[OutputSeconds])
}

func TestCalculateTwoTermsPreservesDecimalMetrics(t *testing.T) {
	rule := Rule{
		Version: "output-reference-v1",
		Terms: []Term{
			{Name: OutputSeconds, MaxValue: 3600},
			{Name: ReferenceVideo, MaxValue: 3600},
		},
		AllowedRounding: []string{RoundingNone},
	}
	result, err := Calculate(rule, Config{
		Version:  ConfigVersion,
		Rounding: RoundingNone,
		Terms: map[string]float64{
			OutputSeconds:  0.12,
			ReferenceVideo: 0.04,
		},
	}, map[string]float64{
		OutputSeconds:  5.5,
		ReferenceVideo: 2.25,
	}, 1, 500000)

	require.NoError(t, err)
	assert.Equal(t, 375000, result.Quota)
	assert.Equal(t, "5.5", result.Snapshot.Metrics[OutputSeconds])
	assert.Equal(t, "2.25", result.Snapshot.Metrics[ReferenceVideo])
}

func TestCalculateRejectsInvalidConfigurationAndMetrics(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		metrics map[string]float64
	}{
		{
			name:    "missing required term",
			config:  Config{Version: ConfigVersion, Rounding: RoundingNone, Terms: map[string]float64{}},
			metrics: map[string]float64{OutputSeconds: 5},
		},
		{
			name:    "unknown term",
			config:  Config{Version: ConfigVersion, Rounding: RoundingNone, Terms: map[string]float64{"unknown": 1}},
			metrics: map[string]float64{OutputSeconds: 5},
		},
		{
			name:    "non positive price",
			config:  Config{Version: ConfigVersion, Rounding: RoundingNone, Terms: map[string]float64{OutputSeconds: 0}},
			metrics: map[string]float64{OutputSeconds: 5},
		},
		{
			name:    "non finite price",
			config:  Config{Version: ConfigVersion, Rounding: RoundingNone, Terms: map[string]float64{OutputSeconds: math.Inf(1)}},
			metrics: map[string]float64{OutputSeconds: 5},
		},
		{
			name:    "oversized metric",
			config:  Config{Version: ConfigVersion, Rounding: RoundingNone, Terms: map[string]float64{OutputSeconds: 1}},
			metrics: map[string]float64{OutputSeconds: 3601},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Calculate(outputRule(), tt.config, tt.metrics, 1, 500000)
			require.Error(t, err)
		})
	}
}

func TestCalculateReturnsClampAndFailsClosed(t *testing.T) {
	result, err := Calculate(outputRule(), Config{
		Version:  ConfigVersion,
		Rounding: RoundingNone,
		Terms:    map[string]float64{OutputSeconds: math.MaxFloat64},
	}, map[string]float64{OutputSeconds: 3600}, 1, 500000)

	require.Error(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Clamp)
	assert.Equal(t, common.QuotaClampOverflow, result.Clamp.Kind)
}
