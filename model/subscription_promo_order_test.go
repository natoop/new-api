package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// 订阅下单（促销码折扣已下线，订单按套餐原价创建）
// ---------------------------------------------------------------------------

func insertPlanForOrderTest(t *testing.T, id int, price float64) *SubscriptionPlan {
	t.Helper()
	plan := &SubscriptionPlan{
		Id:            id,
		Title:         "Order Plan",
		PriceAmount:   price,
		Currency:      "USD",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   1000,
	}
	require.NoError(t, DB.Create(plan).Error)
	// 清理可能残留的套餐缓存，确保读到本测试的数据
	InvalidateSubscriptionPlanCache(id)
	return plan
}

func TestCreateSubscriptionOrderWithPromoReserve_CreatesAtPlanPrice(t *testing.T) {
	truncateTables(t)

	insertUserForRedeemTest(t, 701, 0)
	plan := insertPlanForOrderTest(t, 701, 10.0)

	order := &SubscriptionOrder{
		UserId:          701,
		PlanId:          plan.Id,
		TradeNo:         "ORDER_PLAN_PRICE_NO_1",
		PaymentMethod:   "alipay",
		PaymentProvider: PaymentProviderEpay,
		Status:          common.TopUpStatusPending,
	}
	promo, err := CreateSubscriptionOrderWithPromoReserve(order, plan.PriceAmount, 0.01)
	require.NoError(t, err)
	// 促销码折扣已下线：不再返回促销码，订单金额即套餐原价
	assert.Nil(t, promo)
	assert.InDelta(t, 10.0, order.Money, 1e-9)

	got := GetSubscriptionOrderByTradeNo(order.TradeNo)
	require.NotNil(t, got)
	assert.Equal(t, common.TopUpStatusPending, got.Status)
}

func TestCreateSubscriptionOrderWithPromoReserve_MinMoneyRejected(t *testing.T) {
	truncateTables(t)

	insertUserForRedeemTest(t, 702, 0)
	plan := insertPlanForOrderTest(t, 702, 0.005)

	order := &SubscriptionOrder{
		UserId:          702,
		PlanId:          plan.Id,
		TradeNo:         "ORDER_MIN_NO_1",
		PaymentMethod:   "alipay",
		PaymentProvider: PaymentProviderEpay,
		Status:          common.TopUpStatusPending,
	}
	_, err := CreateSubscriptionOrderWithPromoReserve(order, plan.PriceAmount, 0.01)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "订单金额过低")

	var orderCount int64
	require.NoError(t, DB.Model(&SubscriptionOrder{}).Count(&orderCount).Error)
	assert.Equal(t, int64(0), orderCount)
}

// ---------------------------------------------------------------------------
// RechargeXunhu：订单翻 success 与加额度同事务，幂等不双加
// ---------------------------------------------------------------------------

func TestRechargeXunhu_AtomicAndIdempotent(t *testing.T) {
	truncateTables(t)

	insertUserForRedeemTest(t, 706, 100)
	topUp := &TopUp{
		UserId:          706,
		Amount:          2,
		Money:           2.0,
		TradeNo:         "XUNHU_TOPUP_NO_1",
		PaymentMethod:   "wxpay",
		PaymentProvider: PaymentProviderXunhu,
		CreateTime:      common.GetTimestamp(),
		Status:          common.TopUpStatusPending,
	}
	require.NoError(t, DB.Create(topUp).Error)

	expectedQuota := int(2 * common.QuotaPerUnit)

	quotaToAdd, err := RechargeXunhu(topUp.TradeNo, "127.0.0.1")
	require.NoError(t, err)
	assert.Equal(t, expectedQuota, quotaToAdd)

	var user User
	require.NoError(t, DB.Where("id = ?", 706).First(&user).Error)
	assert.Equal(t, 100+expectedQuota, user.Quota)

	got := GetTopUpByTradeNo(topUp.TradeNo)
	require.NotNil(t, got)
	assert.Equal(t, common.TopUpStatusSuccess, got.Status)

	// 幂等：重放返回 0，不重复加额度
	quotaToAdd, err = RechargeXunhu(topUp.TradeNo, "127.0.0.1")
	require.NoError(t, err)
	assert.Zero(t, quotaToAdd)
	require.NoError(t, DB.Where("id = ?", 706).First(&user).Error)
	assert.Equal(t, 100+expectedQuota, user.Quota)

	// 网关不匹配防护
	_, err = RechargeXunhu("NOT_EXIST_NO", "127.0.0.1")
	require.Error(t, err)
}
