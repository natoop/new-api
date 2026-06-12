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
// 分销订单同步守卫（库存码兑换订单不生成 original 订单）
// ---------------------------------------------------------------------------

func insertAgentInventorySubscriptionOrder(t *testing.T, userID int, tradeNo string) *SubscriptionOrder {
	t.Helper()
	order := &SubscriptionOrder{
		UserId:          userID,
		PlanId:          1,
		Money:           10,
		TradeNo:         tradeNo,
		PaymentMethod:   SubscriptionPaymentMethodAgentInventory,
		PaymentProvider: SubscriptionPaymentMethodAgentInventory,
		Status:          common.TopUpStatusSuccess,
	}
	require.NoError(t, DB.Create(order).Error)
	return order
}

func TestSyncSubscriptionDistributionOrder_SkipsAgentInventoryOrder(t *testing.T) {
	truncateTables(t)

	order := insertAgentInventorySubscriptionOrder(t, 703, "SUBINV_SYNC_GUARD_1")

	// 库存码兑换的订阅订单由 redeem 类型订单表示，同步守卫直接跳过
	require.NoError(t, SyncSubscriptionDistributionOrderTx(DB, order))

	var count int64
	require.NoError(t, DB.Model(&DistributionOrder{}).Count(&count).Error)
	assert.Zero(t, count)
}

func TestBackfillDistributionOrders_CleansAgentInventoryOriginalOrders(t *testing.T) {
	truncateTables(t)

	order := insertAgentInventorySubscriptionOrder(t, 704, "SUBINV_BACKFILL_GUARD_1")

	// 历史脏数据：库存码兑换曾被同步生成的 original 订单
	dirty := &DistributionOrder{
		OrderNo:             "DIRTY_ORIGINAL_1",
		OrderType:           DistributionOrderTypeOriginal,
		SubscriptionOrderId: order.Id,
		IdempotencyKey:      "dirty_original_1",
		AgentId:             1,
		UserId:              order.UserId,
		PackageId:           1,
		Status:              DistributionOrderStatusFulfilled,
	}
	require.NoError(t, DB.Create(dirty).Error)
	// 兑换本身产生的 redeem 订单必须保留
	keep := &DistributionOrder{
		OrderNo:             "KEEP_REDEEM_1",
		OrderType:           DistributionOrderTypeRedeem,
		SubscriptionOrderId: order.Id,
		IdempotencyKey:      "keep_redeem_1",
		AgentId:             1,
		UserId:              order.UserId,
		PackageId:           1,
		Status:              DistributionOrderStatusFulfilled,
	}
	require.NoError(t, DB.Create(keep).Error)

	require.NoError(t, BackfillDistributionOrders())

	// original 脏数据被清掉，redeem 订单保留，且不会被重新同步出来
	var types []string
	require.NoError(t, DB.Model(&DistributionOrder{}).Pluck("order_type", &types).Error)
	require.Equal(t, []string{DistributionOrderTypeRedeem}, types)
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
