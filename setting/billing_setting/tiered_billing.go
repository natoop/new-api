package billing_setting

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/pkg/asynctaskbilling"
	"github.com/QuantumNous/new-api/pkg/billingexpr"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/samber/lo"
)

const (
	BillingModeRatio         = "ratio"
	BillingModeTieredExpr    = "tiered_expr"
	BillingModeAsyncTaskExpr = asynctaskbilling.BillingMode
	BillingModeField         = "billing_mode"
	BillingExprField         = "billing_expr"
	AsyncTaskBillingField    = "async_task_billing"
)

// BillingSetting is managed by config.GlobalConfig.Register.
// DB keys: billing_setting.billing_mode, billing_setting.billing_expr,
// billing_setting.async_task_billing.
type BillingSetting struct {
	BillingMode      map[string]string                  `json:"billing_mode"`
	BillingExpr      map[string]string                  `json:"billing_expr"`
	AsyncTaskBilling map[string]asynctaskbilling.Config `json:"async_task_billing"`
}

var billingSetting = BillingSetting{
	BillingMode:      make(map[string]string),
	BillingExpr:      make(map[string]string),
	AsyncTaskBilling: make(map[string]asynctaskbilling.Config),
}

func init() {
	config.GlobalConfig.Register("billing_setting", &billingSetting)
}

// ---------------------------------------------------------------------------
// Read accessors (hot path, must be fast)
// ---------------------------------------------------------------------------

func GetBillingMode(model string) string {
	if mode, ok := billingSetting.BillingMode[model]; ok {
		return mode
	}
	return BillingModeRatio
}

func GetBillingExpr(model string) (string, bool) {
	expr, ok := billingSetting.BillingExpr[model]
	return expr, ok
}

func GetBillingModeCopy() map[string]string {
	return lo.Assign(billingSetting.BillingMode)
}

func GetBillingExprCopy() map[string]string {
	return lo.Assign(billingSetting.BillingExpr)
}

func GetAsyncTaskBilling(model string) (asynctaskbilling.Config, bool) {
	config, ok := billingSetting.AsyncTaskBilling[model]
	if !ok {
		return asynctaskbilling.Config{}, false
	}
	config.Terms = lo.Assign(config.Terms)
	return config, true
}

func GetAsyncTaskBillingCopy() map[string]asynctaskbilling.Config {
	configs := make(map[string]asynctaskbilling.Config, len(billingSetting.AsyncTaskBilling))
	for model, config := range billingSetting.AsyncTaskBilling {
		config.Terms = lo.Assign(config.Terms)
		configs[model] = config
	}
	return configs
}

func GetPricingSyncData(base map[string]any) map[string]any {
	extra := make(map[string]any, 3)
	if modes := GetBillingModeCopy(); len(modes) > 0 {
		extra[BillingModeField] = modes
	}
	if exprs := GetBillingExprCopy(); len(exprs) > 0 {
		extra[BillingExprField] = exprs
	}
	if configs := GetAsyncTaskBillingCopy(); len(configs) > 0 {
		extra[AsyncTaskBillingField] = configs
	}
	return lo.Assign(base, extra)
}

// ValidateAsyncTaskBillingJSON checks generic configuration safety before an
// option is stored. The provider profile validates allowed terms at submit.
func ValidateAsyncTaskBillingJSON(value string) error {
	configs := make(map[string]asynctaskbilling.Config)
	if err := common.UnmarshalJsonStr(value, &configs); err != nil {
		return fmt.Errorf("invalid async task billing JSON: %w", err)
	}
	for model, config := range configs {
		if strings.TrimSpace(model) == "" {
			return fmt.Errorf("async task billing model name is required")
		}
		if err := asynctaskbilling.ValidateConfig(config); err != nil {
			return fmt.Errorf("async task billing for %s: %w", model, err)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Smoke test (called externally for validation before save)
// ---------------------------------------------------------------------------

func SmokeTestExpr(exprStr string) error {
	return smokeTestExpr(exprStr)
}

func smokeTestExpr(exprStr string) error {
	vectors := []billingexpr.TokenParams{
		{P: 0, C: 0, Len: 0},
		{P: 1000, C: 1000, Len: 1000},
		{P: 100000, C: 100000, Len: 100000},
		{P: 1000000, C: 1000000, Len: 1000000},
	}
	requests := []billingexpr.RequestInput{
		{},
		{
			Headers: map[string]string{
				"anthropic-beta": "fast-mode-2026-02-01",
			},
			Body: []byte(`{"service_tier":"fast","stream_options":{"include_usage":true},"messages":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21]}`),
		},
	}

	for _, v := range vectors {
		for _, request := range requests {
			result, _, err := billingexpr.RunExprWithRequest(exprStr, v, request)
			if err != nil {
				return fmt.Errorf("vector {p=%g, c=%g}: run failed: %w", v.P, v.C, err)
			}
			if result < 0 {
				return fmt.Errorf("vector {p=%g, c=%g}: result %f < 0", v.P, v.C, result)
			}
		}
	}
	return nil
}
