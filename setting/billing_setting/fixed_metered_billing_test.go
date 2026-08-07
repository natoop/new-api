package billing_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateFixedMeteredBillingJSON(t *testing.T) {
	valid := `{
		"example-video": {
			"version": 1,
			"unit_price": 0.12,
			"usage_mode": "duration_plus_reference_video_seconds",
			"rounding": "ceil_total_units"
		}
	}`

	assert.NoError(t, ValidateFixedMeteredBillingJSON(valid))
	assert.Error(t, ValidateFixedMeteredBillingJSON(`{
		"example-video": {
			"version": 1,
			"unit_price": 0.12,
			"usage_mode": "not_supported",
			"rounding": "none"
		}
	}`))
	assert.Error(t, ValidateFixedMeteredBillingJSON(`{
		"example-video": {
			"version": 1,
			"unit_price": -1,
			"usage_mode": "per_request",
			"rounding": "none"
		}
	}`))
}

func TestValidateBillingModeJSONRejectsRetiredAsyncTaskExpression(t *testing.T) {
	assert.NoError(t, ValidateBillingModeJSON(`{"example-video":"fixed_metered"}`))
	assert.Error(t, ValidateBillingModeJSON(`{"example-video":"async_task_expr"}`))
}
