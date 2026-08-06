package asynctaskbilling

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/QuantumNous/new-api/common"
	"github.com/shopspring/decimal"
)

const (
	BillingMode    = "async_task_expr"
	ConfigVersion  = 1
	RoundingNone   = "none"
	OutputSeconds  = "output_seconds"
	ReferenceVideo = "reference_video_seconds"
	InputVideo     = "input_video_seconds"
)

var knownTerms = map[string]struct{}{
	OutputSeconds:  {},
	ReferenceVideo: {},
	InputVideo:     {},
}

// Config is the operator-owned part of task pricing. The profile-owned Rule
// determines which terms a model is allowed to use.
type Config struct {
	Version  int                `json:"version"`
	Rounding string             `json:"rounding"`
	Terms    map[string]float64 `json:"terms"`
}

// Term is a provider-neutral bounded input into a fixed task-price formula.
type Term struct {
	Name     string  `json:"name"`
	MaxValue float64 `json:"max_value"`
}

// Rule belongs to the provider profile. Administrators may provide prices for
// its terms but may not add a term or change the formula order.
type Rule struct {
	Version         string   `json:"version"`
	Terms           []Term   `json:"terms"`
	AllowedRounding []string `json:"allowed_rounding"`
}

// Profile is the sanitized rule metadata exposed to the configuration UI.
type Profile struct {
	Version         string   `json:"version"`
	Terms           []Term   `json:"terms"`
	AllowedRounding []string `json:"allowed_rounding"`
}

// Snapshot is persisted before the upstream task submission. It remains the
// source of truth if settings change while the task is queued or retried.
type Snapshot struct {
	BillingMode     string             `json:"billing_mode"`
	ConfigVersion   int                `json:"config_version"`
	FormulaVersion  string             `json:"formula_version"`
	Terms           map[string]string  `json:"terms"`
	Metrics         map[string]string  `json:"metrics"`
	Rounding        string             `json:"rounding"`
	Expression      string             `json:"expression"`
	GroupRatio      float64            `json:"group_ratio"`
	QuotaPerUnit    float64            `json:"quota_per_unit"`
	ReservedQuota   int                `json:"reserved_quota"`
	QuotaSaturation *common.QuotaClamp `json:"quota_saturation,omitempty"`
}

type Result struct {
	Quota    int
	Snapshot Snapshot
	Clamp    *common.QuotaClamp
}

var (
	profileRegistryMu sync.RWMutex
	profileRegistry   = make(map[string]Profile)
)

// RegisterProfile makes profile-owned pricing metadata available to the
// administration UI without exposing provider protocol or price data.
func RegisterProfile(model string, rule Rule) {
	if strings.TrimSpace(model) == "" || validateRule(rule) != nil {
		return
	}
	profileRegistryMu.Lock()
	defer profileRegistryMu.Unlock()
	profileRegistry[model] = Profile{
		Version:         rule.Version,
		Terms:           append([]Term(nil), rule.Terms...),
		AllowedRounding: normalizedAllowedRounding(rule),
	}
}

// RegisteredProfiles returns a defensive copy for configuration display.
func RegisteredProfiles() map[string]Profile {
	profileRegistryMu.RLock()
	defer profileRegistryMu.RUnlock()
	profiles := make(map[string]Profile, len(profileRegistry))
	for model, profile := range profileRegistry {
		profiles[model] = Profile{
			Version:         profile.Version,
			Terms:           append([]Term(nil), profile.Terms...),
			AllowedRounding: append([]string(nil), profile.AllowedRounding...),
		}
	}
	return profiles
}

func ValidateConfig(config Config) error {
	if config.Version != ConfigVersion {
		return fmt.Errorf("async task billing version must be %d", ConfigVersion)
	}
	if config.Rounding != RoundingNone {
		return fmt.Errorf("unsupported async task billing rounding %q", config.Rounding)
	}
	if len(config.Terms) == 0 {
		return fmt.Errorf("async task billing terms are required")
	}
	for name, price := range config.Terms {
		if _, ok := knownTerms[name]; !ok {
			return fmt.Errorf("unknown async task billing term %q", name)
		}
		if !isFinitePositive(price) {
			return fmt.Errorf("async task billing term %q must have a finite positive price", name)
		}
	}
	return nil
}

// ValidateConfigForRule checks a saved configuration against the profile
// before request-specific metrics are calculated.
func ValidateConfigForRule(rule Rule, config Config) error {
	if err := validateRule(rule); err != nil {
		return err
	}
	if err := ValidateConfig(config); err != nil {
		return err
	}
	if !roundingAllowed(config.Rounding, rule) {
		return fmt.Errorf("rounding %q is not allowed by async task billing profile", config.Rounding)
	}
	ruleTerms := make(map[string]struct{}, len(rule.Terms))
	for _, term := range rule.Terms {
		ruleTerms[term.Name] = struct{}{}
	}
	for name := range config.Terms {
		if _, ok := ruleTerms[name]; !ok {
			return fmt.Errorf("term %q is not allowed by async task billing profile", name)
		}
	}
	for _, term := range rule.Terms {
		if _, ok := config.Terms[term.Name]; !ok {
			return fmt.Errorf("required async task billing term %q is missing", term.Name)
		}
	}
	return nil
}

// Calculate evaluates a fixed profile rule from validated metrics. A saturated
// quota returns an error so a clamped charge can never be submitted upstream.
func Calculate(rule Rule, config Config, metrics map[string]float64, groupRatio, quotaPerUnit float64) (*Result, error) {
	if err := ValidateConfigForRule(rule, config); err != nil {
		return nil, err
	}
	if !isFiniteNonNegative(groupRatio) {
		return nil, fmt.Errorf("group ratio must be finite and non-negative")
	}
	if !isFinitePositive(quotaPerUnit) {
		return nil, fmt.Errorf("quota per unit must be finite and positive")
	}

	ruleTerms := make(map[string]Term, len(rule.Terms))
	for _, term := range rule.Terms {
		ruleTerms[term.Name] = term
	}
	for name := range metrics {
		if _, ok := ruleTerms[name]; !ok {
			return nil, fmt.Errorf("metric %q is not allowed by async task billing profile", name)
		}
	}

	total := decimal.Zero
	snapshotTerms := make(map[string]string, len(rule.Terms))
	snapshotMetrics := make(map[string]string, len(rule.Terms))
	expressionParts := make([]string, 0, len(rule.Terms))
	for _, term := range rule.Terms {
		metric, ok := metrics[term.Name]
		if !ok {
			return nil, fmt.Errorf("required async task billing metric %q is missing", term.Name)
		}
		if !isFinitePositive(metric) || metric > term.MaxValue {
			return nil, fmt.Errorf("async task billing metric %q is outside the supported range", term.Name)
		}

		priceDecimal := decimal.NewFromFloat(config.Terms[term.Name])
		metricDecimal := decimal.NewFromFloat(metric)
		total = total.Add(metricDecimal.Mul(priceDecimal))
		snapshotTerms[term.Name] = priceDecimal.String()
		snapshotMetrics[term.Name] = metricDecimal.String()
		expressionParts = append(expressionParts, fmt.Sprintf("%s * %s", term.Name, priceDecimal.String()))
	}

	quotaDecimal := total.Mul(decimal.NewFromFloat(quotaPerUnit)).Mul(decimal.NewFromFloat(groupRatio))
	quota, clamp := common.QuotaFromDecimalChecked(quotaDecimal)
	result := &Result{
		Quota: quota,
		Snapshot: Snapshot{
			BillingMode:     BillingMode,
			ConfigVersion:   config.Version,
			FormulaVersion:  rule.Version,
			Terms:           snapshotTerms,
			Metrics:         snapshotMetrics,
			Rounding:        config.Rounding,
			Expression:      strings.Join(expressionParts, " + "),
			GroupRatio:      groupRatio,
			QuotaPerUnit:    quotaPerUnit,
			ReservedQuota:   quota,
			QuotaSaturation: clamp,
		},
		Clamp: clamp,
	}
	if clamp != nil {
		return result, clamp
	}
	return result, nil
}

func CloneSnapshot(snapshot *Snapshot) *Snapshot {
	if snapshot == nil {
		return nil
	}
	clone := *snapshot
	clone.Terms = cloneStringMap(snapshot.Terms)
	clone.Metrics = cloneStringMap(snapshot.Metrics)
	if snapshot.QuotaSaturation != nil {
		clamp := *snapshot.QuotaSaturation
		clone.QuotaSaturation = &clamp
	}
	return &clone
}

func validateRule(rule Rule) error {
	if strings.TrimSpace(rule.Version) == "" {
		return fmt.Errorf("async task billing profile version is required")
	}
	if len(rule.Terms) == 0 {
		return fmt.Errorf("async task billing profile terms are required")
	}
	seen := make(map[string]struct{}, len(rule.Terms))
	for _, term := range rule.Terms {
		if _, ok := knownTerms[term.Name]; !ok {
			return fmt.Errorf("unknown async task billing profile term %q", term.Name)
		}
		if _, ok := seen[term.Name]; ok {
			return fmt.Errorf("duplicate async task billing profile term %q", term.Name)
		}
		if !isFinitePositive(term.MaxValue) {
			return fmt.Errorf("async task billing profile term %q must have a finite positive maximum", term.Name)
		}
		seen[term.Name] = struct{}{}
	}
	return nil
}

func normalizedAllowedRounding(rule Rule) []string {
	if len(rule.AllowedRounding) == 0 {
		return []string{RoundingNone}
	}
	values := append([]string(nil), rule.AllowedRounding...)
	sort.Strings(values)
	return values
}

func roundingAllowed(rounding string, rule Rule) bool {
	for _, allowed := range normalizedAllowedRounding(rule) {
		if rounding == allowed {
			return true
		}
	}
	return false
}

func cloneStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	clone := make(map[string]string, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
}

func isFinitePositive(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func isFiniteNonNegative(value float64) bool {
	return value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}
