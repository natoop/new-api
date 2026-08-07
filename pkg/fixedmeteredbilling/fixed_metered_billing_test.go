package fixedmeteredbilling

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateDurationWithReferenceRoundsTotalUnitsOnce(t *testing.T) {
	result, err := Calculate(Config{
		Version:   ConfigVersion,
		UnitPrice: 0.12,
		UsageMode: UsageModeDurationWithRef,
		Rounding:  RoundingCeilTotalUnits,
	}, Metrics{
		OutputSeconds:         5,
		ReferenceVideoSeconds: 2.25,
	}, 1.5, 500000)

	require.NoError(t, err)
	assert.Equal(t, 720000, result.Quota)
	assert.Equal(t, "8", result.Snapshot.BillableUnits)
	assert.Equal(t, "0.12", result.Snapshot.UnitPrice)
}

func TestCalculatePerRequestDoesNotRequireDuration(t *testing.T) {
	result, err := Calculate(Config{
		Version:   ConfigVersion,
		UnitPrice: 0.4,
		UsageMode: UsageModePerRequest,
		Rounding:  RoundingNone,
	}, Metrics{}, 0.5, 500000)

	require.NoError(t, err)
	assert.Equal(t, 100000, result.Quota)
	assert.Equal(t, "1", result.Snapshot.BillableUnits)
}

func TestCalculateRejectsMissingOrInvalidDurationMetric(t *testing.T) {
	config := Config{
		Version:   ConfigVersion,
		UnitPrice: 0.12,
		UsageMode: UsageModeDurationSeconds,
		Rounding:  RoundingNone,
	}

	_, err := Calculate(config, Metrics{}, 1, 500000)
	require.Error(t, err)

	_, err = Calculate(config, Metrics{OutputSeconds: MaxDurationSeconds + 1}, 1, 500000)
	require.Error(t, err)
}

func TestValidateConfigRejectsPerRequestCeiling(t *testing.T) {
	err := ValidateConfig(Config{
		Version:   ConfigVersion,
		UnitPrice: 0.1,
		UsageMode: UsageModePerRequest,
		Rounding:  RoundingCeilTotalUnits,
	})

	require.Error(t, err)
}
