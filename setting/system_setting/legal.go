package system_setting

import "github.com/QuantumNous/new-api/setting/config"

type LegalSettings struct {
	UserAgreement           string `json:"user_agreement"`
	PrivacyPolicy           string `json:"privacy_policy"`
	UserAgreementVersion    string `json:"user_agreement_version"`
	ConsoleAgreementEnabled bool   `json:"console_agreement_enabled"`
}

// defaultUserAgreement 预置的《用户服务协议》默认文案（Markdown）。
// 仅作为默认值：管理员在面板保存过 legal.user_agreement 后，option 表中的值优先生效。
const defaultUserAgreement = `# 用户服务协议

**版本：v1.0｜生效日期：2026-06-11**

欢迎使用本平台（以下称"本服务"）。本服务由 GosWith 运营方（以下称"我们"）提供。请您在使用前仔细阅读并充分理解本协议全部条款；您勾选同意或实际使用本服务，即视为已接受本协议。

## 一、服务性质

1. 本服务为人工智能模型应用程序编程接口（API）的聚合与转发平台，为您提供统一接入、用量计费与账户管理功能。
2. 模型能力由第三方模型服务商提供，其输出内容由模型自动生成，不代表我们的观点；我们不对生成内容的准确性、完整性、适用性作出保证。

## 二、账户与安全

1. 您应提供真实有效的注册信息，并妥善保管账户凭据与 API 密钥；因保管不善造成的损失由您自行承担。
2. 账户仅限您本人（或您所代表的组织）使用，不得出借、转售或共享；我们有权对异常使用行为采取限制、冻结等措施。

## 三、付费、退款与发票

1. 本服务按套餐订阅或按用量预付费计费，具体价格以购买页面公示为准。
2. 充值余额与已开通套餐一经使用即发生消耗；除法律法规另有规定或我们书面同意外，已消耗部分不予退还。未消耗部分的退款申请将在核实后于合理期限内处理。
3. 促销码、兑换码等优惠权益不可兑现、不可转让，解释权在法律允许范围内归我们所有。
4. 如需发票，请通过本协议文末联系方式与我们联系。

## 四、使用规范（禁止性条款）

您承诺不利用本服务从事下列行为，否则我们有权立即中止或终止服务且不退还费用，并保留追责权利：

1. 违反中华人民共和国及您所在司法辖区法律法规的行为；
2. 生成、传播危害国家安全、淫秽色情、暴力恐怖、虚假信息、侵害他人名誉权/隐私权/知识产权的内容；
3. 实施网络攻击、恶意爬取、绕过计费、共享转售接口等破坏服务秩序的行为；
4. 将服务用于医疗诊断、法律意见等高风险场景而不加人工审核。

## 五、数据与隐私

1. 我们遵循最小必要原则收集与处理您的信息（账户信息、用量日志、支付记录），用于提供服务、计费结算与安全审计。
2. 您提交的请求内容将被转发至相应模型服务商以完成调用；我们不会主动将其用于与服务无关的目的。
3. 具体规则详见《隐私政策》；法律法规要求披露的情形除外。

## 六、服务可用性与免责

1. 我们以"现状"提供服务，并尽商业上合理的努力保障可用性，但不承诺服务不中断、无错误。
2. 因第三方模型服务商故障、不可抗力、网络原因、您自身操作导致的损失，我们在法律允许的最大范围内免责。
3. 在任何情况下，我们对您的全部赔偿责任以您过去十二个月内实际支付的费用总额为限。

## 七、协议变更与终止

1. 我们可根据业务调整本协议，更新后将通过站内公告或弹窗提示；您继续使用即视为接受更新版本。
2. 您可随时停止使用并注销账户；注销前请自行处理余额与数据。

## 八、争议解决

本协议适用中华人民共和国法律。因本协议产生的争议，双方应友好协商；协商不成的，任何一方可向运营方所在地有管辖权的人民法院提起诉讼。

## 九、联系我们

如对本协议或服务有任何疑问，请联系：本站公示的联系方式或站内工单。
`

var defaultLegalSettings = LegalSettings{
	UserAgreement:           defaultUserAgreement,
	PrivacyPolicy:           "",
	UserAgreementVersion:    "v1.0",
	ConsoleAgreementEnabled: false,
}

func init() {
	config.GlobalConfig.Register("legal", &defaultLegalSettings)
}

func GetLegalSettings() *LegalSettings {
	return &defaultLegalSettings
}
