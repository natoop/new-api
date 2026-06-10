package system_setting

import "github.com/QuantumNous/new-api/setting/config"

type LegalSettings struct {
	UserAgreement           string `json:"user_agreement"`
	PrivacyPolicy           string `json:"privacy_policy"`
	UserAgreementVersion    string `json:"user_agreement_version"`
	ConsoleAgreementEnabled bool   `json:"console_agreement_enabled"`
}

var defaultLegalSettings = LegalSettings{
	UserAgreement:           "",
	PrivacyPolicy:           "",
	UserAgreementVersion:    "",
	ConsoleAgreementEnabled: false,
}

func init() {
	config.GlobalConfig.Register("legal", &defaultLegalSettings)
}

func GetLegalSettings() *LegalSettings {
	return &defaultLegalSettings
}
