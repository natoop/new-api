package controller

// 路由注册需求（请在 router/api-router.go 中由负责路由的同学添加，本文件不直接改 router）：
//
//	在 subscriptionRoute 分组（router/api-router.go ~L155，apiRouter.Group("/subscription")，
//	鉴权 middleware.UserAuth()，与 POST /subscription/balance/pay 同组）新增一行：
//
//	    subscriptionRoute.POST("/promo/validate", middleware.CriticalRateLimit(), controller.ValidateSubscriptionPromo)
//
// 即 POST /api/subscription/promo/validate
// 请求 body: {"code": "<促销码>", "plan_id": <套餐ID>}
// 响应 data: {"code", "plan_id", "plan_title", "currency", "discount_bps",
//             "original_amount", "final_amount"}

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

type SubscriptionPromoValidateRequest struct {
	Code   string `json:"code"`
	PlanId int    `json:"plan_id"`
}

// ValidateSubscriptionPromo 用户侧校验促销码并返回折后价。
func ValidateSubscriptionPromo(c *gin.Context) {
	if !requirePaymentCompliance(c) {
		return
	}

	var req SubscriptionPromoValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PlanId <= 0 || strings.TrimSpace(req.Code) == "" {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	common.RandomSleep()
	common.ApiErrorMsg(c, "优惠码功能已禁用")
}
