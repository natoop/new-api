package service

import (
	"crypto/md5"
	"encoding/hex"
	"testing"
)

func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func TestXunhuSign(t *testing.T) {
	secret := "testsecret"

	t.Run("fixed params sorted by ASCII key order", func(t *testing.T) {
		params := map[string]string{
			"version":        "1.1",
			"appid":          "201906xxxx",
			"trade_order_id": "SUBUSR1NOABC123",
			"total_fee":      "0.01",
			"title":          "SUB:Test Plan",
			"notify_url":     "https://example.com/api/subscription/xunhu/notify",
			"return_url":     "https://example.com/console/topup",
			"nonce_str":      "abcdef1234567890",
			"time":           "1700000000",
		}
		// ASCII order: appid, nonce_str, notify_url, return_url, time, title, total_fee, trade_order_id, version
		expectedRaw := "appid=201906xxxx" +
			"&nonce_str=abcdef1234567890" +
			"&notify_url=https://example.com/api/subscription/xunhu/notify" +
			"&return_url=https://example.com/console/topup" +
			"&time=1700000000" +
			"&title=SUB:Test Plan" +
			"&total_fee=0.01" +
			"&trade_order_id=SUBUSR1NOABC123" +
			"&version=1.1" +
			secret
		want := md5Hex(expectedRaw)
		if got := XunhuSign(params, secret); got != want {
			t.Fatalf("XunhuSign = %s, want %s", got, want)
		}
	})

	t.Run("empty values excluded from signing", func(t *testing.T) {
		withEmpty := map[string]string{
			"appid":          "123",
			"trade_order_id": "NO1",
			"plugins":        "", // 空值不参与签名
			"openid":         "",
		}
		withoutEmpty := map[string]string{
			"appid":          "123",
			"trade_order_id": "NO1",
		}
		if XunhuSign(withEmpty, secret) != XunhuSign(withoutEmpty, secret) {
			t.Fatal("empty-valued params should be excluded from signature")
		}
		want := md5Hex("appid=123&trade_order_id=NO1" + secret)
		if got := XunhuSign(withEmpty, secret); got != want {
			t.Fatalf("XunhuSign = %s, want %s", got, want)
		}
	})

	t.Run("hash param itself excluded from signing", func(t *testing.T) {
		params := map[string]string{
			"appid": "123",
			"hash":  "should-not-matter",
		}
		want := md5Hex("appid=123" + secret)
		if got := XunhuSign(params, secret); got != want {
			t.Fatalf("XunhuSign = %s, want %s", got, want)
		}
	})
}

func TestXunhuVerifyNotify(t *testing.T) {
	secret := "notifysecret"
	form := map[string]string{
		"trade_order_id": "USR1NOXYZ",
		"total_fee":      "10.00",
		"transaction_id": "4200001234202601011234567890",
		"status":         "OD",
		"plugins":        "", // 空值字段，验签时应剔除
	}
	form["hash"] = XunhuSign(form, secret)

	t.Run("valid signature passes", func(t *testing.T) {
		if !XunhuVerifyNotify(form, secret) {
			t.Fatal("expected valid notify to verify")
		}
	})

	t.Run("tampered amount fails", func(t *testing.T) {
		tampered := make(map[string]string, len(form))
		for k, v := range form {
			tampered[k] = v
		}
		tampered["total_fee"] = "0.01"
		if XunhuVerifyNotify(tampered, secret) {
			t.Fatal("expected tampered notify to fail verification")
		}
	})

	t.Run("wrong secret fails", func(t *testing.T) {
		if XunhuVerifyNotify(form, "wrongsecret") {
			t.Fatal("expected wrong secret to fail verification")
		}
	})

	t.Run("missing hash fails", func(t *testing.T) {
		noHash := map[string]string{"trade_order_id": "USR1NOXYZ"}
		if XunhuVerifyNotify(noHash, secret) {
			t.Fatal("expected missing hash to fail verification")
		}
	})
}

func TestXunhuFormatAmount(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0.01, "0.01"},
		{1, "1.00"},
		{10.5, "10.50"},
		{99.99, "99.99"},
		{100, "100.00"},
		{0.015, "0.01"}, // FormatFloat 银行家式就近舍入到两位
	}
	for _, c := range cases {
		if got := XunhuFormatAmount(c.in); got != c.want {
			t.Errorf("XunhuFormatAmount(%v) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestXunhuAmountMatches(t *testing.T) {
	cases := []struct {
		notify string
		order  float64
		want   bool
	}{
		{"10.00", 10.0, true},
		{"10.00", 10.005, true},  // 分位以内容差
		{"10.00", 10.011, false}, // 超过一分
		{"0.01", 0.01, true},
		{"abc", 1.0, false},
		{"", 1.0, false},
	}
	for _, c := range cases {
		if got := XunhuAmountMatches(c.notify, c.order); got != c.want {
			t.Errorf("XunhuAmountMatches(%q, %v) = %v, want %v", c.notify, c.order, got, c.want)
		}
	}
}

func TestXunhuPaymentMethodAndSecretResolution(t *testing.T) {
	if got := XunhuPaymentMethod(XunhuPayTypeWechat); got != "xunhu-wechat" {
		t.Fatalf("XunhuPaymentMethod(wechat) = %s", got)
	}
	if got := XunhuPaymentMethod(XunhuPayTypeAlipay); got != "xunhu-alipay" {
		t.Fatalf("XunhuPaymentMethod(alipay) = %s", got)
	}
	if _, err := XunhuSecretByPaymentMethod("epay"); err == nil {
		t.Fatal("expected error for non-xunhu payment method")
	}
	if _, _, err := XunhuCredentials("paypal"); err == nil {
		t.Fatal("expected error for unsupported pay type")
	}
}
