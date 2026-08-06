package zzdh

import (
	"sort"
	"strings"

	"github.com/QuantumNous/new-api/pkg/asynctaskbilling"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

type protocol string

const (
	protocolV1 protocol = "v1"
	protocolV8 protocol = "v8"
)

type billingRule string

const (
	billingOutputSeconds           billingRule = "output_seconds"
	billingOutputPlusReferenceSecs billingRule = "output_plus_reference_seconds"
)

// modelProfile is the single source of truth for ZZDH video model capability.
// Numeric prices deliberately do not live here: they remain model-price
// configuration in new-api so operators can change them without a deploy.
type modelProfile struct {
	Name               string
	Protocol           protocol
	Family             string
	MinDuration        int
	MaxDuration        int
	FixedFPS           int
	MaxReferenceImages int
	MaxReferenceVideos int
	MaxReferenceAudios int
	AllowDataReference bool
	RequiresReference  bool
	RejectsReference   bool
	RequiresImage      bool
	RequiresVideo      bool
	RejectsVideo       bool
	RequiresAudio      bool
	RejectsAudio       bool
	DefaultResolution  string
	ResolutionTier     string
	BillingRule        billingRule
	PricingRuleVersion string
}

var modelProfiles = buildModelProfiles()

func init() {
	for name, profile := range modelProfiles {
		asynctaskbilling.RegisterProfile(name, profile.AsyncTaskBillingRule())
	}
}

func buildModelProfiles() map[string]modelProfile {
	profiles := make(map[string]modelProfile, 59)
	add := func(profile modelProfile, names ...string) {
		for _, name := range names {
			profile.Name = name
			profiles[name] = profile
		}
	}
	addSeedance := func(name string) {
		profile := modelProfile{
			Name:               name,
			Protocol:           protocolV1,
			Family:             "seedance_v1",
			BillingRule:        billingOutputSeconds,
			PricingRuleVersion: "seedance-v1-seconds-v1",
			ResolutionTier:     resolutionTier(name),
		}
		profile.DefaultResolution = profile.ResolutionTier
		if strings.Contains(name, "-video-") {
			profile.Family = "seedance_v1_reference"
			profile.RequiresVideo = true
			profile.BillingRule = billingOutputPlusReferenceSecs
			profile.PricingRuleVersion = "seedance-v1-output-plus-reference-seconds-v1"
		} else {
			profile.RejectsVideo = true
		}
		profiles[name] = profile
	}

	for _, name := range []string{
		"doubao-seedance-2-0-fast-480p",
		"doubao-seedance-2-0-fast-720p",
		"doubao-seedance-2-0-fast-video-480p",
		"doubao-seedance-2-0-fast-video-720p",
		"doubao-seedance-2-0-mini-480p",
		"doubao-seedance-2-0-mini-720p",
		"doubao-seedance-2-0-mini-video-480p",
		"doubao-seedance-2-0-mini-video-720p",
		"doubao-seedance-2-1080p",
		"doubao-seedance-2-480p",
		"doubao-seedance-2-4k",
		"doubao-seedance-2-720p",
		"doubao-seedance-2-video-1080p",
		"doubao-seedance-2-video-480p",
		"doubao-seedance-2-video-4k",
		"doubao-seedance-2-video-720p",
	} {
		addSeedance(name)
	}

	add(modelProfile{
		Protocol: protocolV8, Family: "kling_v8", ResolutionTier: "720p", DefaultResolution: "1280x720",
		RejectsReference: true, RequiresAudio: true, BillingRule: billingOutputSeconds, PricingRuleVersion: "kling-v3-omni-v1",
	}, "kling-3.0-omni-720p-noref-audio")
	add(modelProfile{
		Protocol: protocolV8, Family: "kling_v8", ResolutionTier: "720p", DefaultResolution: "1280x720",
		RejectsReference: true, RejectsAudio: true, BillingRule: billingOutputSeconds, PricingRuleVersion: "kling-v3-omni-v1",
	}, "kling-3.0-omni-720p-noref-mute")
	add(modelProfile{
		Protocol: protocolV8, Family: "kling_v8", ResolutionTier: "720p", DefaultResolution: "1280x720",
		RequiresReference: true, RequiresAudio: true, BillingRule: billingOutputSeconds, PricingRuleVersion: "kling-v3-omni-v1",
	}, "kling-3.0-omni-720p-ref-audio")
	add(modelProfile{
		Protocol: protocolV8, Family: "kling_v8", ResolutionTier: "720p", DefaultResolution: "1280x720",
		RequiresReference: true, RejectsAudio: true, BillingRule: billingOutputSeconds, PricingRuleVersion: "kling-v3-omni-v1",
	}, "kling-3.0-omni-720p-ref-mute")
	add(modelProfile{
		Protocol: protocolV8, Family: "kling_v8", ResolutionTier: "1080p", DefaultResolution: "1920x1080",
		RejectsReference: true, RequiresAudio: true, BillingRule: billingOutputSeconds, PricingRuleVersion: "kling-v3-omni-v1",
	}, "kling-3.0-omni-1080p-noref-audio")
	add(modelProfile{
		Protocol: protocolV8, Family: "kling_v8", ResolutionTier: "1080p", DefaultResolution: "1920x1080",
		RejectsReference: true, RejectsAudio: true, BillingRule: billingOutputSeconds, PricingRuleVersion: "kling-v3-omni-v1",
	}, "kling-3.0-omni-1080p-noref-mute")
	add(modelProfile{
		Protocol: protocolV8, Family: "kling_v8", ResolutionTier: "1080p", DefaultResolution: "1920x1080",
		RequiresReference: true, RequiresAudio: true, BillingRule: billingOutputSeconds, PricingRuleVersion: "kling-v3-omni-v1",
	}, "kling-3.0-omni-1080p-ref-audio")
	add(modelProfile{
		Protocol: protocolV8, Family: "kling_v8", ResolutionTier: "1080p", DefaultResolution: "1920x1080",
		RequiresReference: true, RejectsAudio: true, BillingRule: billingOutputSeconds, PricingRuleVersion: "kling-v3-omni-v1",
	}, "kling-3.0-omni-1080p-ref-mute")
	add(modelProfile{
		Protocol: protocolV8, Family: "kling_v8", BillingRule: billingOutputSeconds, PricingRuleVersion: "kling-v3-omni-v1",
	}, "kling-v3-omni")

	for _, name := range []string{
		"happyhorse-1.0-i2v-1080p", "happyhorse-1.0-i2v-720p",
		"happyhorse-1.0-r2v-1080p", "happyhorse-1.0-r2v-720p",
		"happyhorse-1.0-t2v-1080p", "happyhorse-1.0-t2v-720p",
		"happyhorse-1.0-video-edit-1080p", "happyhorse-1.0-video-edit-720p",
	} {
		profile := modelProfile{
			Name:     name,
			Protocol: protocolV8, Family: "happyhorse_v8", ResolutionTier: resolutionTier(name),
			DefaultResolution: resolutionDefault(resolutionTier(name)), BillingRule: billingOutputSeconds,
			PricingRuleVersion: "happyhorse-v8-seconds-v1",
		}
		switch {
		case strings.Contains(name, "-i2v-") || strings.Contains(name, "-r2v-"):
			profile.RequiresReference = true
			profile.RequiresImage = true
		case strings.Contains(name, "-t2v-"):
			profile.RejectsReference = true
		case strings.Contains(name, "-video-edit-"):
			profile.RequiresReference = true
		}
		profiles[name] = profile
	}

	for _, name := range []string{
		"wan2.6-image", "wan2.6-i2v", "wan2.6-i2v-flash", "wan2.6-r2v", "wan2.6-t2i",
		"wan2.7-image", "wan2.7-i2v", "wan2.7-r2v", "wan2.7-t2v", "wan2.7-videoedit",
	} {
		profile := modelProfile{
			Name:     name,
			Protocol: protocolV8, Family: "wan_v8", BillingRule: billingOutputSeconds,
			PricingRuleVersion: "wan-v8-seconds-v1",
		}
		switch {
		case strings.Contains(name, "-i2v"):
			profile.RequiresReference = true
			profile.RequiresImage = true
		case strings.Contains(name, "-r2v") || strings.Contains(name, "videoedit"):
			profile.RequiresReference = true
		case strings.Contains(name, "-t2v"):
			profile.RejectsReference = true
		}
		profiles[name] = profile
	}

	for _, name := range []string{
		"vidu-q3-pro-540p", "vidu-q3-pro-540p-offpeak",
		"vidu-q3-pro-720p", "vidu-q3-pro-720p-offpeak",
		"vidu-q3-pro-1080p", "vidu-q3-pro-1080p-offpeak",
		"vidu-q3-turbo-540p", "vidu-q3-turbo-540p-offpeak",
		"vidu-q3-turbo-720p", "vidu-q3-turbo-720p-offpeak",
		"vidu-q3-turbo-1080p", "vidu-q3-turbo-1080p-offpeak",
	} {
		profiles[name] = modelProfile{
			Name:               name,
			Protocol:           protocolV8,
			Family:             "vidu_q3_v8",
			RejectsReference:   true,
			BillingRule:        billingOutputSeconds,
			PricingRuleVersion: "vidu-q3-v8-seconds-v1",
		}
	}

	for _, name := range []string{
		"zzdh-Minimax-h3-480p",
		"zzdh-Minimax-h3-720p",
		"zzdh-Minimax-h3-1080p",
		"zzdh-Minimax-h3-2k",
	} {
		profiles[name] = modelProfile{
			Name:               name,
			Protocol:           protocolV8,
			Family:             "minimax_h3_v8",
			ResolutionTier:     resolutionTier(name),
			DefaultResolution:  resolutionDefault(resolutionTier(name)),
			MinDuration:        5,
			MaxDuration:        15,
			FixedFPS:           24,
			MaxReferenceImages: 9,
			MaxReferenceVideos: 3,
			MaxReferenceAudios: 3,
			AllowDataReference: true,
			BillingRule:        billingOutputSeconds,
			PricingRuleVersion: "minimax-h3-v8-seconds-v1",
		}
	}
	return profiles
}

func profileForModel(name string) (modelProfile, bool) {
	profile, ok := modelProfiles[name]
	return profile, ok
}

// AsyncTaskBillingRule keeps approved metric admission with the validated
// profile. Only numeric coefficients are configured outside provider code.
func (profile modelProfile) AsyncTaskBillingRule() asynctaskbilling.Rule {
	maxDuration := float64(relaycommon.MaxTaskDurationSeconds)
	if profile.MaxDuration > 0 && profile.MaxDuration < relaycommon.MaxTaskDurationSeconds {
		maxDuration = float64(profile.MaxDuration)
	}
	rule := asynctaskbilling.Rule{
		Version: profile.PricingRuleVersion,
		Terms: []asynctaskbilling.Term{
			{Name: asynctaskbilling.OutputSeconds, MaxValue: maxDuration},
		},
		AllowedRounding: []string{asynctaskbilling.RoundingNone},
	}
	if profile.BillingRule == billingOutputPlusReferenceSecs {
		rule.Terms = append(rule.Terms, asynctaskbilling.Term{
			Name:     asynctaskbilling.ReferenceVideo,
			MaxValue: float64(relaycommon.MaxTaskDurationSeconds),
		})
	}
	return rule
}

func modelList() []string {
	models := make([]string, 0, len(modelProfiles))
	for name := range modelProfiles {
		models = append(models, name)
	}
	sort.Strings(models)
	return models
}

func resolutionTier(name string) string {
	name = strings.ToLower(name)
	switch {
	case strings.Contains(name, "480p"):
		return "480p"
	case strings.Contains(name, "720p"):
		return "720p"
	case strings.Contains(name, "1080p"):
		return "1080p"
	case strings.Contains(name, "4k"):
		return "4k"
	case strings.Contains(name, "2k"):
		return "2k"
	default:
		return ""
	}
}

func resolutionDefault(tier string) string {
	switch tier {
	case "480p":
		return "854x480"
	case "720p":
		return "1280x720"
	case "1080p":
		return "1920x1080"
	case "4k":
		return "3840x2160"
	case "2k":
		return "2560x1440"
	default:
		return ""
	}
}
