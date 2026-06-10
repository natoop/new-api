package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertUserForRedeemTest(t *testing.T, id int, quota int) {
	t.Helper()
	user := &User{
		Id:       id,
		Username: "redeem_user",
		Status:   common.UserStatusEnabled,
		Quota:    quota,
	}
	require.NoError(t, DB.Create(user).Error)
}

func insertRedemptionForRedeemTest(t *testing.T, redemption *Redemption) *Redemption {
	t.Helper()
	if redemption.Status == 0 {
		redemption.Status = common.RedemptionCodeStatusEnabled
	}
	if redemption.CreatedTime == 0 {
		redemption.CreatedTime = common.GetTimestamp()
	}
	require.NoError(t, DB.Create(redemption).Error)
	return redemption
}

func TestRedeem_BalanceTypeBackwardCompatible(t *testing.T) {
	truncateTables(t)

	insertUserForRedeemTest(t, 501, 100)
	// Type 为空，模拟存量余额码
	insertRedemptionForRedeemTest(t, &Redemption{
		Key:   "11111111111111111111111111111111",
		Name:  "legacy balance",
		Quota: 5000,
	})

	result, err := Redeem("11111111111111111111111111111111", 501)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, RedemptionTypeBalance, result.Type)
	assert.Equal(t, 5000, result.Quota)
	assert.Empty(t, result.PlanTitle)

	var user User
	require.NoError(t, DB.Where("id = ?", 501).First(&user).Error)
	assert.Equal(t, 5100, user.Quota)

	var redemption Redemption
	require.NoError(t, DB.Where("used_user_id = ?", 501).First(&redemption).Error)
	assert.Equal(t, common.RedemptionCodeStatusUsed, redemption.Status)
	assert.Equal(t, 501, redemption.UsedUserId)

	// 二次兑换必须失败（防重放）
	_, err = Redeem("11111111111111111111111111111111", 501)
	require.ErrorIs(t, err, ErrRedeemFailed)
}

func TestRedeem_PlanTypeCreatesSubscription(t *testing.T) {
	truncateTables(t)

	insertUserForRedeemTest(t, 502, 0)
	plan := &SubscriptionPlan{
		Id:            601,
		Title:         "Redeem Plan",
		PriceAmount:   19.9,
		Currency:      "USD",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   2000,
	}
	require.NoError(t, DB.Create(plan).Error)
	insertRedemptionForRedeemTest(t, &Redemption{
		Key:    "22222222222222222222222222222222",
		Name:   "plan code",
		Type:   RedemptionTypePlan,
		PlanId: plan.Id,
	})

	result, err := Redeem("22222222222222222222222222222222", 502)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, RedemptionTypePlan, result.Type)
	assert.Equal(t, plan.Id, result.PlanId)
	assert.Equal(t, "Redeem Plan", result.PlanTitle)
	assert.Zero(t, result.Quota)

	var subs []UserSubscription
	require.NoError(t, DB.Where("user_id = ?", 502).Find(&subs).Error)
	require.Len(t, subs, 1)
	assert.Equal(t, plan.Id, subs[0].PlanId)
	assert.Equal(t, "active", subs[0].Status)
	assert.Equal(t, int64(2000), subs[0].AmountTotal)
	assert.Equal(t, "redemption", subs[0].Source)

	// 余额不应变化
	var user User
	require.NoError(t, DB.Where("id = ?", 502).First(&user).Error)
	assert.Equal(t, 0, user.Quota)

	// 码已标记使用
	var redemption Redemption
	require.NoError(t, DB.Where("used_user_id = ?", 502).First(&redemption).Error)
	assert.Equal(t, common.RedemptionCodeStatusUsed, redemption.Status)
}

func TestRedeem_PromoTypeRejected(t *testing.T) {
	truncateTables(t)

	insertUserForRedeemTest(t, 503, 0)
	insertRedemptionForRedeemTest(t, &Redemption{
		Key:         "33333333333333333333333333333333",
		Name:        "promo code",
		Type:        RedemptionTypePromo,
		DiscountBps: 2000,
		MaxUses:     10,
	})

	result, err := Redeem("33333333333333333333333333333333", 503)
	require.ErrorIs(t, err, ErrPromoCodeNotRedeemable)
	assert.Nil(t, result)

	// 促销码不应被标记使用，也不应改余额
	var redemption Redemption
	require.NoError(t, DB.Where("type = ?", RedemptionTypePromo).First(&redemption).Error)
	assert.Equal(t, common.RedemptionCodeStatusEnabled, redemption.Status)
	assert.Equal(t, 0, redemption.UsedCount)

	var user User
	require.NoError(t, DB.Where("id = ?", 503).First(&user).Error)
	assert.Equal(t, 0, user.Quota)
}

func TestApplyPromoDiscount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		amount float64
		bps    int
		want   float64
	}{
		{name: "zero bps keeps amount", amount: 9.99, bps: 0, want: 9.99},
		{name: "negative bps keeps amount", amount: 9.99, bps: -100, want: 9.99},
		{name: "full discount", amount: 9.99, bps: 10000, want: 0},
		{name: "over full discount", amount: 9.99, bps: 12000, want: 0},
		{name: "20 percent off", amount: 9.99, bps: 2000, want: 7.99}, // 7.992 -> 7.99
		{name: "cent precision", amount: 0.03, bps: 3333, want: 0.02}, // 0.020001 -> 0.02
		{name: "float artifact", amount: 19.9, bps: 1000, want: 17.91},
		{name: "zero amount", amount: 0, bps: 5000, want: 0},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.InDelta(t, tt.want, ApplyPromoDiscount(tt.amount, tt.bps), 1e-9)
		})
	}
}

func TestConsumePromoCodeById_MaxUsesGuard(t *testing.T) {
	truncateTables(t)

	redemption := insertRedemptionForRedeemTest(t, &Redemption{
		Key:         "44444444444444444444444444444444",
		Name:        "limited promo",
		Type:        RedemptionTypePromo,
		DiscountBps: 1000,
		MaxUses:     2,
	})

	require.NoError(t, ConsumePromoCodeById(nil, redemption.Id))
	require.NoError(t, ConsumePromoCodeById(nil, redemption.Id))
	require.Error(t, ConsumePromoCodeById(nil, redemption.Id))

	var got Redemption
	require.NoError(t, DB.Where("id = ?", redemption.Id).First(&got).Error)
	assert.Equal(t, 2, got.UsedCount)
}
