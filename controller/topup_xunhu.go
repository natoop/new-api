package controller

import (
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type XunhuTopupRequest struct {
	Amount  int64  `json:"amount"`
	PayType string `json:"pay_type"` // "wechat" | "alipay"
}

func RequestXunhuTopup(c *gin.Context) {
	if !requirePaymentCompliance(c) {
		return
	}
	if !isXunhuTopUpEnabled() {
		common.ApiErrorMsg(c, "虎皮椒支付未启用")
		return
	}

	var req XunhuTopupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
	if req.Amount < getMinTopup() {
		common.ApiErrorMsg(c, fmt.Sprintf("充值数量不能小于 %d", getMinTopup()))
		return
	}

	userId := c.GetInt("id")
	group, err := model.GetUserGroup(userId, true)
	if err != nil {
		common.ApiErrorMsg(c, "获取用户分组失败")
		return
	}
	payMoney := getPayMoney(req.Amount, group)
	if payMoney < 0.01 {
		common.ApiErrorMsg(c, "充值金额过低")
		return
	}

	callBackAddress := service.GetCallbackAddress()
	notifyUrl := callBackAddress + "/api/user/xunhu/notify"
	returnUrl := paymentReturnPath("/console/topup?pay=success")

	tradeNo := fmt.Sprintf("%s%d", common.GetRandomString(6), time.Now().Unix())
	tradeNo = fmt.Sprintf("USR%dNO%s", userId, tradeNo)

	amount := req.Amount
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dAmount := decimal.NewFromInt(amount)
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		amount = dAmount.Div(dQuotaPerUnit).IntPart()
	}

	topUp := &model.TopUp{
		UserId:          userId,
		Amount:          amount,
		Money:           payMoney,
		TradeNo:         tradeNo,
		PaymentMethod:   service.XunhuPaymentMethod(payType),
		PaymentProvider: model.PaymentProviderXunhu,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err := topUp.Insert(); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("虎皮椒 创建充值订单失败 user_id=%d trade_no=%s pay_type=%s amount=%d error=%q", userId, tradeNo, payType, req.Amount, err.Error()))
		common.ApiErrorMsg(c, "创建订单失败")
		return
	}

	payUrl, qrUrl, err := service.XunhuCreatePayment(payType, tradeNo, payMoney, fmt.Sprintf("TUC%d", req.Amount), notifyUrl, returnUrl)
	if err != nil {
		_ = model.UpdatePendingTopUpStatus(tradeNo, model.PaymentProviderXunhu, common.TopUpStatusExpired)
		logger.LogError(c.Request.Context(), fmt.Sprintf("虎皮椒 拉起支付失败 user_id=%d trade_no=%s pay_type=%s amount=%d error=%q", userId, tradeNo, payType, req.Amount, err.Error()))
		common.ApiErrorMsg(c, "拉起支付失败")
		return
	}

	logger.LogInfo(c.Request.Context(), fmt.Sprintf("虎皮椒 充值订单创建成功 user_id=%d trade_no=%s pay_type=%s amount=%d money=%.2f", userId, tradeNo, payType, req.Amount, payMoney))
	common.ApiSuccess(c, gin.H{
		"pay_url":  payUrl,
		"qr_url":   qrUrl,
		"order_no": tradeNo,
	})
}

func XunhuTopupNotify(c *gin.Context) {
	if !isXunhuWebhookEnabled() {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("虎皮椒 充值 webhook 被拒绝 reason=webhook_disabled client_ip=%s", c.ClientIP()))
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

	LockOrder(tradeNo)
	defer UnlockOrder(tradeNo)

	topUp := model.GetTopUpByTradeNo(tradeNo)
	if topUp == nil || topUp.PaymentProvider != model.PaymentProviderXunhu {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("虎皮椒 充值回调订单不存在或网关不匹配 trade_no=%s client_ip=%s", tradeNo, c.ClientIP()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	appSecret, err := service.XunhuSecretByPaymentMethod(topUp.PaymentMethod)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("虎皮椒 充值回调密钥解析失败 trade_no=%s payment_method=%s error=%q", tradeNo, topUp.PaymentMethod, err.Error()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	if !service.XunhuVerifyNotify(params, appSecret) {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("虎皮椒 充值回调验签失败 trade_no=%s client_ip=%s", tradeNo, c.ClientIP()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	// 幂等：已完成的订单直接确认，阻止网关重试
	if topUp.Status == common.TopUpStatusSuccess {
		_, _ = c.Writer.Write([]byte("success"))
		return
	}
	if topUp.Status != common.TopUpStatusPending {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("虎皮椒 充值回调订单状态异常 trade_no=%s status=%s client_ip=%s", tradeNo, topUp.Status, c.ClientIP()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	if !service.XunhuAmountMatches(params["total_fee"], topUp.Money) {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("虎皮椒 充值回调金额不匹配 trade_no=%s notify_total_fee=%s order_money=%.2f client_ip=%s", tradeNo, params["total_fee"], topUp.Money, c.ClientIP()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	if params["status"] != service.XunhuNotifyStatusPaid {
		logger.LogInfo(c.Request.Context(), fmt.Sprintf("虎皮椒 充值回调忽略事件 trade_no=%s status=%s client_ip=%s", tradeNo, params["status"], c.ClientIP()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	// 订单翻 success + 加额度在同一事务内原子完成（含日志与晋升钩子），
	// 失败整体回滚并返回 fail 让网关重试
	quotaToAdd, err := model.RechargeXunhu(tradeNo, c.ClientIP())
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("虎皮椒 充值入账失败 trade_no=%s user_id=%d error=%q", tradeNo, topUp.UserId, err.Error()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	logger.LogInfo(c.Request.Context(), fmt.Sprintf("虎皮椒 充值成功 trade_no=%s user_id=%d quota_to_add=%d money=%.2f transaction_id=%s", tradeNo, topUp.UserId, quotaToAdd, topUp.Money, params["transaction_id"]))
	_, _ = c.Writer.Write([]byte("success"))
}
