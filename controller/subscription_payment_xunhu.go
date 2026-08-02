package controller

import (
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

type SubscriptionXunhuPayRequest struct {
	PlanId    int    `json:"plan_id"`
	PayType   string `json:"pay_type"`   // "wechat" | "alipay"
	PromoCode string `json:"promo_code"` // compatibility only; ordinary redemption codes are not promo discounts
}

func normalizeXunhuPayType(payType string) (string, bool) {
	payType = strings.ToLower(strings.TrimSpace(payType))
	if payType == service.XunhuPayTypeWechat || payType == service.XunhuPayTypeAlipay {
		return payType, true
	}
	return "", false
}

// parseXunhuNotifyForm extracts the POST form params from a XunhuPay async notify.
func parseXunhuNotifyForm(c *gin.Context) map[string]string {
	if err := c.Request.ParseForm(); err != nil {
		return nil
	}
	return lo.Reduce(lo.Keys(c.Request.PostForm), func(r map[string]string, t string, i int) map[string]string {
		r[t] = c.Request.PostForm.Get(t)
		return r
	}, map[string]string{})
}

func SubscriptionRequestXunhuPay(c *gin.Context) {
	if !requirePaymentCompliance(c) {
		return
	}
	if !isXunhuTopUpEnabled() {
		common.ApiErrorMsg(c, "虎皮椒支付未启用")
		return
	}

	var req SubscriptionXunhuPayRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PlanId <= 0 {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	payType, ok := normalizeXunhuPayType(req.PayType)
	if !ok {
		common.ApiErrorMsg(c, "支付方式不存在")
		return
	}
	if _, _, err := service.XunhuCredentials(payType); err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}

	plan, err := model.GetSubscriptionPlanById(req.PlanId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if !plan.Enabled {
		common.ApiErrorMsg(c, "套餐未启用")
		return
	}
	if plan.PriceAmount < 0.01 {
		common.ApiErrorMsg(c, "套餐金额过低")
		return
	}

	userId := c.GetInt("id")
	if !ensurePlanPurchasable(c, userId, plan) {
		return
	}

	callBackAddress := service.GetCallbackAddress()
	notifyUrl := callBackAddress + "/api/subscription/xunhu/notify"
	returnUrl := paymentReturnPath("/console/topup?pay=success")

	tradeNo := fmt.Sprintf("%s%d", common.GetRandomString(6), time.Now().Unix())
	tradeNo = fmt.Sprintf("SUBUSR%dNO%s", userId, tradeNo)

	order := &model.SubscriptionOrder{
		UserId:          userId,
		PlanId:          plan.Id,
		TradeNo:         tradeNo,
		PaymentMethod:   service.XunhuPaymentMethod(payType),
		PaymentProvider: model.PaymentProviderXunhu,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	// 兼容历史方法名：普通兑换码不再作为促销码折扣使用。
	if _, err := model.CreateSubscriptionOrderWithPromoReserve(order, plan.PriceAmount, 0.01); err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}

	payUrl, qrUrl, err := service.XunhuCreatePayment(payType, tradeNo, order.Money, fmt.Sprintf("SUB:%s", plan.Title), notifyUrl, returnUrl)
	if err != nil {
		_ = model.ExpireSubscriptionOrder(tradeNo, model.PaymentProviderXunhu)
		logger.LogError(c.Request.Context(), fmt.Sprintf("虎皮椒 订阅拉起支付失败 user_id=%d trade_no=%s pay_type=%s error=%q", userId, tradeNo, payType, err.Error()))
		common.ApiErrorMsg(c, "拉起支付失败")
		return
	}
	common.ApiSuccess(c, gin.H{
		"pay_url":  payUrl,
		"qr_url":   qrUrl,
		"order_no": tradeNo,
	})
}

func SubscriptionXunhuNotify(c *gin.Context) {
	if !isXunhuWebhookEnabled() {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("虎皮椒 订阅 webhook 被拒绝 reason=webhook_disabled client_ip=%s", c.ClientIP()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	params := parseXunhuNotifyForm(c)
	if len(params) == 0 {
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	tradeNo := params["trade_order_id"]
	if tradeNo == "" {
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	order := model.GetSubscriptionOrderByTradeNo(tradeNo)
	if order == nil || order.PaymentProvider != model.PaymentProviderXunhu {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("虎皮椒 订阅回调订单不存在或网关不匹配 trade_no=%s client_ip=%s", tradeNo, c.ClientIP()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	appSecret, err := service.XunhuSecretByPaymentMethod(order.PaymentMethod)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("虎皮椒 订阅回调密钥解析失败 trade_no=%s payment_method=%s error=%q", tradeNo, order.PaymentMethod, err.Error()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	if !service.XunhuVerifyNotify(params, appSecret) {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("虎皮椒 订阅回调验签失败 trade_no=%s client_ip=%s", tradeNo, c.ClientIP()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	// 幂等：已完成的订单直接确认，阻止网关重试
	if order.Status == common.TopUpStatusSuccess {
		_, _ = c.Writer.Write([]byte("success"))
		return
	}

	if !service.XunhuAmountMatches(params["total_fee"], order.Money) {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("虎皮椒 订阅回调金额不匹配 trade_no=%s notify_total_fee=%s order_money=%.2f client_ip=%s", tradeNo, params["total_fee"], order.Money, c.ClientIP()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	if params["status"] != service.XunhuNotifyStatusPaid {
		logger.LogInfo(c.Request.Context(), fmt.Sprintf("虎皮椒 订阅回调忽略事件 trade_no=%s status=%s client_ip=%s", tradeNo, params["status"], c.ClientIP()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	LockOrder(tradeNo)
	defer UnlockOrder(tradeNo)

	if err := model.CompleteSubscriptionOrder(tradeNo, common.GetJsonString(params), model.PaymentProviderXunhu, order.PaymentMethod); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("虎皮椒 订阅订单完成失败 trade_no=%s error=%q", tradeNo, err.Error()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	logger.LogInfo(c.Request.Context(), fmt.Sprintf("虎皮椒 订阅支付成功 trade_no=%s user_id=%d money=%.2f transaction_id=%s", tradeNo, order.UserId, order.Money, params["transaction_id"]))
	_, _ = c.Writer.Write([]byte("success"))
}
