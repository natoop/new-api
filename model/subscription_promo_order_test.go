package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// 促销码预占模式（下单预占 / 用尽拒单 / 过期回补 / 完成不双扣）
// ---------------------------------------------------------------------------

func insertPlanForPromoTest(t *testing.T, id int, price float64) *SubscriptionPlan {
	t.Helper()
	plan := &SubscriptionPlan{
		Id:            id,
		Title:         "Promo Plan",
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

func promoUsedCount(t *testing.T, key string) int {
	t.Helper()
	var redemption Redemption
	require.NoError(t, DB.Where(redemptionKeyCol()+" = ?", key).First(&redemption).Error)
	return redemption.UsedCount
}

func TestCreateSubscriptionOrderWithPromoReserve_ReservesAtomically(t *testing.T) {
	truncateTables(t)

	insertUserForRedeemTest(t, 701, 0)
	plan := insertPlanForPromoTest(t, 701, 10.0)
	promoKey := "aaaa1111aaaa1111aaaa1111aaaa1111"
	insertRedemptionForRedeemTest(t, &Redemption{
		Key:         promoKey,
		Name:        "promo reserve",
		Type:        RedemptionTypePromo,
		DiscountBps: 2000,
		MaxUses:     1,
	})

	order := &SubscriptionOrder{
		UserId:          701,
		PlanId:          plan.Id,
		TradeNo:         "PROMO_RESERVE_NO_1",
		PaymentMethod:   "alipay",
		PaymentProvider: PaymentProviderEpay,
		Status:          common.TopUpStatusPending,
		PromoCode:       promoKey,
	}
	promo, err := CreateSubscriptionOrderWithPromoReserve(order, plan.PriceAmount, 0.01)
	require.NoError(t, err)
	require.NotNil(t, promo)
	assert.Equal(t, 2000, promo.DiscountBps)
	// 折后金额已在事务内重算
	assert.InDelta(t, 8.0, order.Money, 1e-9)
	// 下单即预占
	assert.Equal(t, 1, promoUsedCount(t, promoKey))

	// 用尽后再下单：拒单且不创建订单
	order2 := &SubscriptionOrder{
		UserId:          701,
		PlanId:          plan.Id,
		TradeNo:         "PROMO_RESERVE_NO_2",
		PaymentMethod:   "alipay",
		PaymentProvider: PaymentProviderEpay,
		Status:          common.TopUpStatusPending,
		PromoCode:       promoKey,
	}
	_, err = CreateSubscriptionOrderWithPromoReserve(order2, plan.PriceAmount, 0.01)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "促销码已被用尽")

	var orderCount int64
	require.NoError(t, DB.Model(&SubscriptionOrder{}).Count(&orderCount).Error)
	assert.Equal(t, int64(1), orderCount)
	assert.Equal(t, 1, promoUsedCount(t, promoKey))
}

func TestCreateSubscriptionOrderWithPromoReserve_MinMoneyRollsBack(t *testing.T) {
	truncateTables(t)

	insertUserForRedeemTest(t, 702, 0)
	plan := insertPlanForPromoTest(t, 702, 0.01)
	promoKey := "bbbb2222bbbb2222bbbb2222bbbb2222"
	insertRedemptionForRedeemTest(t, &Redemption{
		Key:         promoKey,
		Name:        "deep discount",
		Type:        RedemptionTypePromo,
		DiscountBps: 9000,
		MaxUses:     5,
	})

	order := &SubscriptionOrder{
		UserId:          702,
		PlanId:          plan.Id,
		TradeNo:         "PROMO_MIN_NO_1",
		PaymentMethod:   "alipay",
		PaymentProvider: PaymentProviderEpay,
		Status:          common.TopUpStatusPending,
		PromoCode:       promoKey,
	}
	_, err := CreateSubscriptionOrderWithPromoReserve(order, plan.PriceAmount, 0.01)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "折后金额过低")

	// 整体回滚：预占已释放、订单未创建
	assert.Equal(t, 0, promoUsedCount(t, promoKey))
	var orderCount int64
	require.NoError(t, DB.Model(&SubscriptionOrder{}).Count(&orderCount).Error)
	assert.Equal(t, int64(0), orderCount)
}

func TestExpireSubscriptionOrder_RefundsPromoReservation(t *testing.T) {
	truncateTables(t)

	insertUserForRedeemTest(t, 703, 0)
	plan := insertPlanForPromoTest(t, 703, 10.0)
	promoKey := "cccc3333cccc3333cccc3333cccc3333"
	insertRedemptionForRedeemTest(t, &Redemption{
		Key:         promoKey,
		Name:        "promo refund",
		Type:        RedemptionTypePromo,
		DiscountBps: 1000,
		MaxUses:     1,
	})

	order := &SubscriptionOrder{
		UserId:          703,
		PlanId:          plan.Id,
		TradeNo:         "PROMO_EXPIRE_NO_1",
		PaymentMethod:   "alipay",
		PaymentProvider: PaymentProviderEpay,
		Status:          common.TopUpStatusPending,
		PromoCode:       promoKey,
	}
	_, err := CreateSubscriptionOrderWithPromoReserve(order, plan.PriceAmount, 0.01)
	require.NoError(t, err)
	require.Equal(t, 1, promoUsedCount(t, promoKey))

	// 过期回补
	require.NoError(t, ExpireSubscriptionOrder(order.TradeNo, PaymentProviderEpay))
	assert.Equal(t, 0, promoUsedCount(t, promoKey))

	got := GetSubscriptionOrderByTradeNo(order.TradeNo)
	require.NotNil(t, got)
	assert.Equal(t, common.TopUpStatusExpired, got.Status)

	// 再次过期为幂等：不会重复回补（used_count > 0 防负数）
	require.NoError(t, ExpireSubscriptionOrder(order.TradeNo, PaymentProviderEpay))
	assert.Equal(t, 0, promoUsedCount(t, promoKey))
}

func TestCompleteSubscriptionOrder_DoesNotDoubleConsumePromo(t *testing.T) {
	truncateTables(t)

	insertUserForRedeemTest(t, 704, 0)
	plan := insertPlanForPromoTest(t, 704, 10.0)
	promoKey := "dddd4444dddd4444dddd4444dddd4444"
	insertRedemptionForRedeemTest(t, &Redemption{
		Key:         promoKey,
		Name:        "promo complete",
		Type:        RedemptionTypePromo,
		DiscountBps: 2000,
		MaxUses:     3,
	})

	order := &SubscriptionOrder{
		UserId:          704,
		PlanId:          plan.Id,
		TradeNo:         "PROMO_COMPLETE_NO_1",
		PaymentMethod:   "alipay",
		PaymentProvider: PaymentProviderEpay,
		Status:          common.TopUpStatusPending,
		PromoCode:       promoKey,
	}
	_, err := CreateSubscriptionOrderWithPromoReserve(order, plan.PriceAmount, 0.01)
	require.NoError(t, err)
	require.Equal(t, 1, promoUsedCount(t, promoKey))

	require.NoError(t, CompleteSubscriptionOrder(order.TradeNo, "payload", PaymentProviderEpay, "alipay"))

	// 完成订单不再二次消耗促销码
	assert.Equal(t, 1, promoUsedCount(t, promoKey))

	got := GetSubscriptionOrderByTradeNo(order.TradeNo)
	require.NotNil(t, got)
	assert.Equal(t, common.TopUpStatusSuccess, got.Status)

	var subs []UserSubscription
	require.NoError(t, DB.Where("user_id = ?", 704).Find(&subs).Error)
	require.Len(t, subs, 1)
	assert.Equal(t, plan.Id, subs[0].PlanId)

	// 回调重放（幂等）：仍不消耗
	require.NoError(t, CompleteSubscriptionOrder(order.TradeNo, "payload", PaymentProviderEpay, "alipay"))
	assert.Equal(t, 1, promoUsedCount(t, promoKey))
}

// ---------------------------------------------------------------------------
// Redeem：停用套餐的 plan 码兑换被拒并整体回滚
// ---------------------------------------------------------------------------

func TestRedeem_PlanDisabledRejected(t *testing.T) {
	truncateTables(t)

	insertUserForRedeemTest(t, 705, 0)
	plan := &SubscriptionPlan{
		Id:            705,
		Title:         "Disabled Plan",
		PriceAmount:   9.9,
		Currency:      "USD",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		TotalAmount:   1000,
	}
	require.NoError(t, DB.Create(plan).Error)
	// Enabled 带 gorm default:true，零值 false 在 Create 时会被列默认值覆盖，需显式置停用
	require.NoError(t, DB.Model(&SubscriptionPlan{}).Where("id = ?", plan.Id).Update("enabled", false).Error)
	InvalidateSubscriptionPlanCache(plan.Id)

	key := "eeee5555eeee5555eeee5555eeee5555"
	insertRedemptionForRedeemTest(t, &Redemption{
		Key:    key,
		Name:   "disabled plan code",
		Type:   RedemptionTypePlan,
		PlanId: plan.Id,
	})

	_, err := Redeem(key, 705)
	require.ErrorIs(t, err, ErrRedeemFailed)

	// 整体回滚：码仍可用、未创建订阅
	var redemption Redemption
	require.NoError(t, DB.Where(redemptionKeyCol()+" = ?", key).First(&redemption).Error)
	assert.Equal(t, common.RedemptionCodeStatusEnabled, redemption.Status)
	assert.Zero(t, redemption.UsedUserId)

	var subCount int64
	require.NoError(t, DB.Model(&UserSubscription{}).Where("user_id = ?", 705).Count(&subCount).Error)
	assert.Equal(t, int64(0), subCount)
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
