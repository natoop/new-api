package service

import (
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
)

// XunhuPay (虎皮椒) v2 hosted checkout integration.
// Docs: https://www.xunhupay.com — endpoint POST /payment/do.html.
// WeChat and Alipay are two independent appid/appsecret pairs.

const (
	XunhuPayTypeWechat = "wechat"
	XunhuPayTypeAlipay = "alipay"

	// XunhuNotifyStatusPaid is the order status value in async notify meaning paid.
	XunhuNotifyStatusPaid = "OD"

	xunhuRequestTimeout = 15 * time.Second
)

// XunhuPaymentMethod returns the order PaymentMethod string recorded for a pay type,
// e.g. "xunhu-wechat" / "xunhu-alipay".
func XunhuPaymentMethod(payType string) string {
	return "xunhu-" + payType
}

// XunhuCredentials returns the appid/appsecret pair for a pay type ("wechat"|"alipay").
func XunhuCredentials(payType string) (appId string, appSecret string, err error) {
	switch payType {
	case XunhuPayTypeWechat:
		appId, appSecret = setting.XunhuWechatAppId, setting.XunhuWechatAppSecret
	case XunhuPayTypeAlipay:
		appId, appSecret = setting.XunhuAlipayAppId, setting.XunhuAlipayAppSecret
	default:
		return "", "", fmt.Errorf("不支持的虎皮椒支付类型: %s", payType)
	}
	if strings.TrimSpace(appId) == "" || strings.TrimSpace(appSecret) == "" {
		return "", "", fmt.Errorf("虎皮椒 %s 通道未配置", payType)
	}
	return appId, appSecret, nil
}

// XunhuSecretByPaymentMethod resolves the appsecret from an order PaymentMethod
// ("xunhu-wechat" / "xunhu-alipay"), used by notify verification.
func XunhuSecretByPaymentMethod(paymentMethod string) (string, error) {
	payType, ok := strings.CutPrefix(paymentMethod, "xunhu-")
	if !ok {
		return "", fmt.Errorf("非虎皮椒支付方式: %s", paymentMethod)
	}
	_, secret, err := XunhuCredentials(payType)
	return secret, err
}

// XunhuFormatAmount formats a CNY amount as a two-decimal string (元), e.g. 1 -> "1.00".
func XunhuFormatAmount(amountYuan float64) string {
	return strconv.FormatFloat(amountYuan, 'f', 2, 64)
}

// XunhuSign computes hash = MD5(sorted "k=v&..." joined by ASCII key order + appsecret).
// Empty-valued params and the "hash" param itself are excluded from signing,
// per XunhuPay v2 signature rules.
func XunhuSign(params map[string]string, appSecret string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k == "hash" || v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+params[k])
	}
	raw := strings.Join(pairs, "&") + appSecret
	sum := md5.Sum([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// XunhuVerifyNotify verifies the hash of an async notify form against appsecret.
func XunhuVerifyNotify(form map[string]string, appSecret string) bool {
	got := strings.ToLower(strings.TrimSpace(form["hash"]))
	if got == "" {
		return false
	}
	want := XunhuSign(form, appSecret)
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

type xunhuDoResponse struct {
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	URL       string `json:"url"`
	URLQrcode string `json:"url_qrcode"`
	Hash      string `json:"hash"`
}

var xunhuHTTPClient = &http.Client{Timeout: xunhuRequestTimeout}

// XunhuCreatePayment creates a XunhuPay order and returns the cashier URL and QR-code URL.
// payType is "wechat" or "alipay"; amountYuan is in CNY and is sent with two decimals.
func XunhuCreatePayment(payType string, tradeNo string, amountYuan float64, title string, notifyUrl string, returnUrl string) (payUrl string, qrUrl string, err error) {
	appId, appSecret, err := XunhuCredentials(payType)
	if err != nil {
		return "", "", err
	}
	gateway := strings.TrimSpace(setting.XunhuGatewayUrl)
	if gateway == "" {
		return "", "", errors.New("虎皮椒网关地址未配置")
	}

	params := map[string]string{
		"version":        "1.1",
		"appid":          appId,
		"trade_order_id": tradeNo,
		"total_fee":      XunhuFormatAmount(amountYuan),
		"title":          title,
		"notify_url":     notifyUrl,
		"return_url":     returnUrl,
		"nonce_str":      common.GetRandomString(16),
		"time":           strconv.FormatInt(time.Now().Unix(), 10),
	}
	params["hash"] = XunhuSign(params, appSecret)

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	resp, err := xunhuHTTPClient.PostForm(gateway, form)
	if err != nil {
		return "", "", fmt.Errorf("请求虎皮椒网关失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("虎皮椒网关返回状态码 %d", resp.StatusCode)
	}

	var result xunhuDoResponse
	if err := common.DecodeJson(resp.Body, &result); err != nil {
		return "", "", fmt.Errorf("解析虎皮椒响应失败: %w", err)
	}
	if result.ErrCode != 0 {
		return "", "", fmt.Errorf("虎皮椒下单失败: errcode=%d errmsg=%s", result.ErrCode, result.ErrMsg)
	}
	if result.URL == "" && result.URLQrcode == "" {
		return "", "", errors.New("虎皮椒未返回支付链接")
	}
	return result.URL, result.URLQrcode, nil
}

// XunhuAmountMatches reports whether the notify total_fee matches the order amount
// within one-fen tolerance (both values in CNY yuan).
func XunhuAmountMatches(notifyTotalFee string, orderAmountYuan float64) bool {
	notifyAmount, err := strconv.ParseFloat(strings.TrimSpace(notifyTotalFee), 64)
	if err != nil {
		return false
	}
	diff := notifyAmount - orderAmountYuan
	if diff < 0 {
		diff = -diff
	}
	return diff < 0.01
}
