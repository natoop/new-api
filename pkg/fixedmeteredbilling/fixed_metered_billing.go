package fixedmeteredbilling

import (
	"fmt"
	"math"

	"github.com/QuantumNous/new-api/common"
	"github.com/shopspring/decimal"
)

const (
	BillingMode              = "fixed_metered"
	ConfigVersion            = 1
	UsageModePerRequest      = "per_request"
	UsageModeDurationSeconds = "duration_seconds"
	UsageModeDurationWithRef = "duration_plus_reference_video_seconds"
	RoundingNone             = "none"
	RoundingCeilTotalUnits   = "ceil_total_units"
	MaxDurationSeconds       = 3600.0
)

// Config contains the complete operator-owned fixed-metered rule for one
// public model. The price intentionally lives here rather than in ModelPrice.
type Config struct {
	Version   int     `json:"version"`
	UnitPrice float64 `json:"unit_price"`
	UsageMode string  `json:"usage_mode"`
	Rounding  string  `json:"rounding"`
}

// Metrics are provider-normalized quantities. Providers must bound their
// values before calling Calculate; Calculate validates them again.
type Metrics struct {
	OutputSeconds         float64 `json:"output_seconds,omitempty"`
	ReferenceVideoSeconds float64 `json:"reference_video_seconds,omitempty"`
}

// Snapshot is persisted before a request is dispatched so a later settings
// change cannot alter the charge for an accepted task.
type Snapshot struct {
	BillingMode           string             `json:"billing_mode"`
	ConfigVersion         int                `json:"config_version"`
	UnitPrice             string             `json:"unit_price"`
	UsageMode             string             `json:"usage_mode"`
	Rounding              string             `json:"rounding"`
	OutputSeconds         string             `json:"output_seconds,omitempty"`
	ReferenceVideoSeconds string             `json:"reference_video_seconds,omitempty"`
	BillableUnits         string             `json:"billable_units"`
	GroupRatio            float64            `json:"group_ratio"`
	QuotaPerUnit          float64            `json:"quota_per_unit"`
	ReservedQuota         int                `json:"reserved_quota"`
	QuotaSaturation       *common.QuotaClamp `json:"quota_saturation,omitempty"`
}

type Result struct {
	Quota    int
	Snapshot Snapshot
	Clamp    *common.QuotaClamp
}

func ValidateConfig(config Config) error {
	if config.Version != ConfigVersion {
		return fmt.Errorf("fixed metered billing version must be %d", ConfigVersion)
	}
	if !isFiniteNonNegative(config.UnitPrice) {
		return fmt.Errorf("fixed metered billing unit price must be finite and non-negative")
	}
	switch config.UsageMode {
	case UsageModePerRequest, UsageModeDurationSeconds, UsageModeDurationWithRef:
	default:
		return fmt.Errorf("unsupported fixed metered usage mode %q", config.UsageMode)
	}
	switch config.Rounding {
	case RoundingNone, RoundingCeilTotalUnits:
	default:
		return fmt.Errorf("unsupported fixed metered rounding %q", config.Rounding)
	}
	if config.UsageMode == UsageModePerRequest && config.Rounding != RoundingNone {
		return fmt.Errorf("per-request fixed metered billing only supports %q rounding", RoundingNone)
	}
	return nil
}

func Calculate(config Config, metrics Metrics, groupRatio, quotaPerUnit float64) (*Result, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}
	if !isFiniteNonNegative(groupRatio) {
		return nil, fmt.Errorf("group ratio must be finite and non-negative")
	}
	if !isFinitePositive(quotaPerUnit) {
		return nil, fmt.Errorf("quota per unit must be finite and positive")
	}

	units, err := billableUnits(config, metrics)
	if err != nil {
		return nil, err
	}
	if config.Rounding == RoundingCeilTotalUnits {
		units = units.Ceil()
	}

	quotaDecimal := units.
		Mul(decimal.NewFromFloat(config.UnitPrice)).
		Mul(decimal.NewFromFloat(groupRatio)).
		Mul(decimal.NewFromFloat(quotaPerUnit))
	quota, clamp := common.QuotaFromDecimalChecked(quotaDecimal)
	result := &Result{
		Quota: quota,
		Snapshot: Snapshot{
			BillingMode:           BillingMode,
			ConfigVersion:         config.Version,
			UnitPrice:             decimal.NewFromFloat(config.UnitPrice).String(),
			UsageMode:             config.UsageMode,
			Rounding:              config.Rounding,
			OutputSeconds:         decimal.NewFromFloat(metrics.OutputSeconds).String(),
			ReferenceVideoSeconds: decimal.NewFromFloat(metrics.ReferenceVideoSeconds).String(),
			BillableUnits:         units.String(),
			GroupRatio:            groupRatio,
			QuotaPerUnit:          quotaPerUnit,
			ReservedQuota:         quota,
			QuotaSaturation:       clamp,
		},
		Clamp: clamp,
	}
	if clamp != nil {
		return result, clamp
	}
	return result, nil
}

func billableUnits(config Config, metrics Metrics) (decimal.Decimal, error) {
	switch config.UsageMode {
	case UsageModePerRequest:
		return decimal.NewFromInt(1), nil
	case UsageModeDurationSeconds:
		if !validDuration(metrics.OutputSeconds) {
			return decimal.Zero, fmt.Errorf("fixed metered output seconds are outside the supported range")
		}
		return decimal.NewFromFloat(metrics.OutputSeconds), nil
	case UsageModeDurationWithRef:
		if !validDuration(metrics.OutputSeconds) {
			return decimal.Zero, fmt.Errorf("fixed metered output seconds are outside the supported range")
		}
		if !validOptionalDuration(metrics.ReferenceVideoSeconds) {
			return decimal.Zero, fmt.Errorf("fixed metered reference video seconds are outside the supported range")
		}
		return decimal.NewFromFloat(metrics.OutputSeconds).Add(decimal.NewFromFloat(metrics.ReferenceVideoSeconds)), nil
	default:
		return decimal.Zero, fmt.Errorf("unsupported fixed metered usage mode %q", config.UsageMode)
	}
}

func CloneSnapshot(snapshot *Snapshot) *Snapshot {
	if snapshot == nil {
		return nil
	}
	clone := *snapshot
	if snapshot.QuotaSaturation != nil {
		clamp := *snapshot.QuotaSaturation
		clone.QuotaSaturation = &clamp
	}
	return &clone
}

func validDuration(value float64) bool {
	return isFinitePositive(value) && value <= MaxDurationSeconds
}

func validOptionalDuration(value float64) bool {
	return isFiniteNonNegative(value) && value <= MaxDurationSeconds
}

func isFinitePositive(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func isFiniteNonNegative(value float64) bool {
	return value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}
