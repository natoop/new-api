package billing_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAsyncTaskBillingJSON(t *testing.T) {
	valid := `{
		"zzdh-output": {
			"version": 1,
			"rounding": "none",
			"terms": {"output_seconds": 0.12}
		}
	}`

	assert.NoError(t, ValidateAsyncTaskBillingJSON(valid))
	assert.Error(t, ValidateAsyncTaskBillingJSON(`{
		"zzdh-output": {
			"version": 1,
			"rounding": "none",
			"terms": {"unknown_seconds": 0.12}
		}
	}`))
	assert.Error(t, ValidateAsyncTaskBillingJSON(`{
		"zzdh-output": {
			"version": 1,
			"rounding": "none",
			"terms": {"output_seconds": 0}
		}
	}`))
}
