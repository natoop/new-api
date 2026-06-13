package setting

// XunhuPay (虎皮椒, xunhupay.com) hosted checkout configuration.
// WeChat and Alipay channels use two independent appid/appsecret pairs,
// opened separately in the XunhuPay merchant console. A channel is usable
// once XunhuEnabled is true and its appid + appsecret are populated.
var (
	XunhuEnabled         = false
	XunhuWechatAppId     string
	XunhuWechatAppSecret string
	XunhuAlipayAppId     string
	XunhuAlipayAppSecret string
	XunhuGatewayUrl      = "https://api.xunhupay.com/payment/do.html"

	// 资金展示配置（钱包页应付金额用）
	XunhuFundType     = "CNY"
	XunhuFundSymbol   = "¥"
	XunhuExchangeRate = 1.0
)
