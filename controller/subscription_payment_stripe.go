package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/coupon"
	"github.com/thanhpk/randstr"
)

type SubscriptionStripePayRequest struct {
	PlanId    int    `json:"plan_id"`
	PromoCode string `json:"promo_code"` // 可选促销码：通过 Stripe 一次性 Coupon 实现首期折扣
}

func SubscriptionRequestStripePay(c *gin.Context) {
	if !requirePaymentCompliance(c) {
		return
	}

	var req SubscriptionStripePayRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PlanId <= 0 {
		common.ApiErrorMsg(c, "参数错误")
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
	if plan.StripePriceId == "" {
		common.ApiErrorMsg(c, "该套餐未配置 StripePriceId")
		return
	}
	if !strings.HasPrefix(setting.StripeApiSecret, "sk_") && !strings.HasPrefix(setting.StripeApiSecret, "rk_") {
		common.ApiErrorMsg(c, "Stripe 未配置或密钥无效")
		return
	}
	if setting.StripeWebhookSecret == "" {
		common.ApiErrorMsg(c, "Stripe Webhook 未配置")
		return
	}

	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if user == nil {
		common.ApiErrorMsg(c, "用户不存在")
		return
	}

	if plan.MaxPurchasePerUser > 0 {
		count, err := model.CountUserSubscriptionsByPlan(userId, plan.Id)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		if count >= int64(plan.MaxPurchasePerUser) {
			common.ApiErrorMsg(c, "已达到该套餐购买上限")
			return
		}
	}

	promoCode := strings.TrimSpace(req.PromoCode)

	reference := fmt.Sprintf("sub-stripe-ref-%d-%d-%s", user.Id, time.Now().UnixMilli(), randstr.String(4))
	referenceId := "sub_ref_" + common.Sha1([]byte(reference))

	order := &model.SubscriptionOrder{
		UserId:          userId,
		PlanId:          plan.Id,
		TradeNo:         referenceId,
		PaymentMethod:   model.PaymentMethodStripe,
		PaymentProvider: model.PaymentProviderStripe,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
		PromoCode:       promoCode,
	}
	// 同一事务内创建订单并预占促销码次数（用尽/失效则拒单），order.Money 为折后金额
	promo, err := model.CreateSubscriptionOrderWithPromoReserve(order, plan.PriceAmount, 0)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": err.Error()})
		return
	}
	promoDiscountBps := 0
	if promo != nil {
		promoDiscountBps = promo.DiscountBps
	}

	payLink, err := genStripeSubscriptionLink(referenceId, user.StripeCustomer, user.Email, plan.StripePriceId, promoCode, promoDiscountBps)
	if err != nil {
		// 拉起失败：订单置过期并回补预占的促销码次数
		_ = model.ExpireSubscriptionOrder(referenceId, model.PaymentProviderStripe)
		logger.LogError(c.Request.Context(), fmt.Sprintf("Stripe 订阅支付链接创建失败 trade_no=%s plan_id=%d error=%q", referenceId, plan.Id, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"pay_link": payLink,
		},
	})
}

// getOrCreateStripePromoCoupon 按折扣万分比复用确定性 ID 的 Stripe Coupon（promo-bps-<bps>），
// 避免每次折扣下单都动态创建一次性 Coupon 在 Stripe 后台堆积垃圾对象。
// 流程：先按 ID Get；404（resource_missing）才 Create；并发 Create 冲突（resource_already_exists）回退 Get。
func getOrCreateStripePromoCoupon(promoDiscountBps int) (*stripe.Coupon, error) {
	couponID := fmt.Sprintf("promo-bps-%d", promoDiscountBps)
	existing, err := coupon.Get(couponID, nil)
	if err == nil {
		return existing, nil
	}
	if !stripeErrorHasCode(err, stripe.ErrorCodeResourceMissing) {
		return nil, err
	}
	created, err := coupon.New(&stripe.CouponParams{
		ID:         stripe.String(couponID),
		PercentOff: stripe.Float64(float64(promoDiscountBps) / 100.0),
		Duration:   stripe.String(string(stripe.CouponDurationOnce)),
		Name:       stripe.String(fmt.Sprintf("PROMO %g%% OFF", float64(promoDiscountBps)/100.0)),
	})
	if err == nil {
		return created, nil
	}
	if stripeErrorHasCode(err, stripe.ErrorCodeResourceAlreadyExists) {
		return coupon.Get(couponID, nil)
	}
	return nil, err
}

func stripeErrorHasCode(err error, code stripe.ErrorCode) bool {
	var stripeErr *stripe.Error
	return errors.As(err, &stripeErr) && stripeErr.Code == code
}

func genStripeSubscriptionLink(referenceId string, customerId string, email string, priceId string, promoCode string, promoDiscountBps int) (string, error) {
	stripe.Key = setting.StripeApiSecret

	params := &stripe.CheckoutSessionParams{
		ClientReferenceID: stripe.String(referenceId),
		SuccessURL:        stripe.String(paymentReturnPath("/console/topup")),
		CancelURL:         stripe.String(paymentReturnPath("/console/topup")),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceId),
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
	}

	// 促销码：Stripe 订阅按 PriceId 计价，折扣通过确定性 Coupon（按折扣 bps 复用）实现，仅首期生效
	if promoCode != "" && promoDiscountBps > 0 && promoDiscountBps < 10000 {
		stripeCoupon, err := getOrCreateStripePromoCoupon(promoDiscountBps)
		if err != nil {
			return "", err
		}
		params.Discounts = []*stripe.CheckoutSessionDiscountParams{
			{Coupon: stripe.String(stripeCoupon.ID)},
		}
	}

	if "" == customerId {
		if "" != email {
			params.CustomerEmail = stripe.String(email)
		}
		params.CustomerCreation = stripe.String(string(stripe.CheckoutSessionCustomerCreationAlways))
	} else {
		params.Customer = stripe.String(customerId)
	}

	result, err := session.New(params)
	if err != nil {
		return "", err
	}
	return result.URL, nil
}
