package billing_setting

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/pkg/billingexpr"
	"github.com/QuantumNous/new-api/pkg/fixedmeteredbilling"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/samber/lo"
)

const (
	BillingModeRatio         = "ratio"
	BillingModeTieredExpr    = "tiered_expr"
	BillingModeFixedMetered  = fixedmeteredbilling.BillingMode
	BillingModeField         = "billing_mode"
	BillingExprField         = "billing_expr"
	FixedMeteredBillingField = "fixed_metered_billing"
)

// BillingSetting is managed by config.GlobalConfig.Register.
// DB keys: billing_setting.billing_mode, billing_setting.billing_expr,
// billing_setting.fixed_metered_billing.
type BillingSetting struct {
	BillingMode         map[string]string                     `json:"billing_mode"`
	BillingExpr         map[string]string                     `json:"billing_expr"`
	FixedMeteredBilling map[string]fixedmeteredbilling.Config `json:"fixed_metered_billing"`
}

var billingSetting = BillingSetting{
	BillingMode:         make(map[string]string),
	BillingExpr:         make(map[string]string),
	FixedMeteredBilling: make(map[string]fixedmeteredbilling.Config),
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

func ValidateBillingModeJSON(value string) error {
	modes := make(map[string]string)
	if err := common.UnmarshalJsonStr(value, &modes); err != nil {
		return fmt.Errorf("invalid billing mode JSON: %w", err)
	}
	for model, mode := range modes {
		if strings.TrimSpace(model) == "" {
			return fmt.Errorf("billing mode model name is required")
		}
		if !isSupportedBillingMode(mode) {
			return fmt.Errorf("unsupported billing mode %q for %s", mode, model)
		}
	}
	return nil
}

func IsSupportedBillingMode(mode string) bool {
	return isSupportedBillingMode(mode)
}

func isSupportedBillingMode(mode string) bool {
	switch mode {
	case BillingModeRatio, BillingModeTieredExpr, BillingModeFixedMetered:
		return true
	default:
		return false
	}
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

func GetFixedMeteredBilling(model string) (fixedmeteredbilling.Config, bool) {
	config, ok := billingSetting.FixedMeteredBilling[model]
	if !ok {
		return fixedmeteredbilling.Config{}, false
	}
	return config, true
}

func GetFixedMeteredBillingCopy() map[string]fixedmeteredbilling.Config {
	return lo.Assign(billingSetting.FixedMeteredBilling)
}

func GetPricingSyncData(base map[string]any) map[string]any {
	extra := make(map[string]any, 3)
	if modes := GetBillingModeCopy(); len(modes) > 0 {
		extra[BillingModeField] = modes
	}
	if exprs := GetBillingExprCopy(); len(exprs) > 0 {
		extra[BillingExprField] = exprs
	}
	if configs := GetFixedMeteredBillingCopy(); len(configs) > 0 {
		extra[FixedMeteredBillingField] = configs
	}
	return lo.Assign(base, extra)
}

func ValidateFixedMeteredBillingJSON(value string) error {
	configs := make(map[string]fixedmeteredbilling.Config)
	if err := common.UnmarshalJsonStr(value, &configs); err != nil {
		return fmt.Errorf("invalid fixed metered billing JSON: %w", err)
	}
	for model, config := range configs {
		if strings.TrimSpace(model) == "" {
			return fmt.Errorf("fixed metered billing model name is required")
		}
		if err := fixedmeteredbilling.ValidateConfig(config); err != nil {
			return fmt.Errorf("fixed metered billing for %s: %w", model, err)
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
