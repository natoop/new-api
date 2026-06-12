package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func couponTruncate(t *testing.T) {
	t.Helper()
	truncate(t)
	t.Cleanup(func() {
		model.DB.Exec("DELETE FROM redemptions")
		model.DB.Exec("DELETE FROM p3_coupons")
		model.DB.Exec("DELETE FROM p3_balance_ledgers")
	})
}

func seedCouponAgent(t *testing.T, userID int, balance float64) *model.DistributionAgent {
	t.Helper()
	user := &model.User{
		Id:       userID,
		Username: fmt.Sprintf("coupon_user_%d", userID),
		AffCode:  fmt.Sprintf("coupon_aff_%d", userID),
		Status:   common.UserStatusEnabled,
	}
	require.NoError(t, model.DB.Create(user).Error)
	now := time.Now().Unix()
	agent := &model.DistributionAgent{
		UserId:    userID,
		Name:      fmt.Sprintf("coupon_agent_%d", userID),
		Status:    DistributionStatusEnabled,
		Balance:   balance,
		Level:     DistributionAgentLevelSecondary,
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, model.DB.Create(agent).Error)
	return agent
}

func couponAgentBalance(t *testing.T, agentID int) float64 {
	t.Helper()
	var agent model.DistributionAgent
	require.NoError(t, model.DB.Where("id = ?", agentID).First(&agent).Error)
	return agent.Balance
}

func TestApplyAgentCoupon_DeductsBalanceAndCreatesLedger(t *testing.T) {
	couponTruncate(t)
	agent := seedCouponAgent(t, 2000, 10)

	coupon, err := ApplyAgentCoupon(2000, 1)
	require.NoError(t, err)
	require.NotNil(t, coupon)

	assert.Equal(t, agent.Id, coupon.AgentId)
	assert.Equal(t, DistributionCouponSourceSelf, coupon.Source)
	assert.Equal(t, DistributionCouponStatusActive, coupon.Status)
	assert.Equal(t, 1.0, coupon.Amount)
	expectedQuota := int(common.QuotaPerUnit)
	assert.Equal(t, expectedQuota, coupon.Quota)
	assert.Greater(t, coupon.ExpiresAt, time.Now().Unix())

	// 余额 1:1 扣除
	assert.InDelta(t, 9, couponAgentBalance(t, agent.Id), 1e-9)

	// 同步创建了绑定的原生兑换码
	var redemption model.Redemption
	require.NoError(t, model.DB.Where("id = ?", coupon.RedemptionId).First(&redemption).Error)
	assert.Equal(t, coupon.Code, redemption.Key)
	assert.Equal(t, expectedQuota, redemption.Quota)
	assert.Equal(t, coupon.ExpiresAt, redemption.ExpiredTime)
	assert.Equal(t, common.RedemptionCodeStatusEnabled, redemption.Status)

	// debit 流水
	var ledger model.DistributionBalanceLedger
	require.NoError(t, model.DB.Where("agent_id = ? AND source_type = ?", agent.Id, DistributionSourceCouponApply).First(&ledger).Error)
	assert.Equal(t, DistributionLedgerEntryDebit, ledger.EntryType)
	assert.InDelta(t, -1, ledger.Delta, 1e-9)
	assert.Equal(t, coupon.Code, ledger.SourceNo)
}

func TestApplyAgentCoupon_InsufficientBalance(t *testing.T) {
	couponTruncate(t)
	agent := seedCouponAgent(t, 2100, 0.5)

	_, err := ApplyAgentCoupon(2100, 1)
	require.Error(t, err)

	// 余额不变，无券、无码、无流水
	assert.InDelta(t, 0.5, couponAgentBalance(t, agent.Id), 1e-9)
	var couponCount, redemptionCount, ledgerCount int64
	require.NoError(t, model.DB.Model(&model.DistributionCoupon{}).Count(&couponCount).Error)
	require.NoError(t, model.DB.Model(&model.Redemption{}).Count(&redemptionCount).Error)
	require.NoError(t, model.DB.Model(&model.DistributionBalanceLedger{}).Count(&ledgerCount).Error)
	assert.Zero(t, couponCount)
	assert.Zero(t, redemptionCount)
	assert.Zero(t, ledgerCount)
}

func TestAdminIssueCoupons_CountValidationAndIssue(t *testing.T) {
	couponTruncate(t)
	agent := seedCouponAgent(t, 2200, 3)

	// 单次总张数超限
	_, err := AdminIssueCoupons(agent.Id, 1, []DistributionCouponIssueItem{
		{Count: 60, Amount: 1, ValidityDays: 7},
		{Count: 41, Amount: 2, ValidityDays: 3},
	}, "")
	require.Error(t, err)

	// 非法明细
	_, err = AdminIssueCoupons(agent.Id, 1, []DistributionCouponIssueItem{{Count: 1, Amount: 0, ValidityDays: 7}}, "")
	require.Error(t, err)
	_, err = AdminIssueCoupons(agent.Id, 1, []DistributionCouponIssueItem{{Count: 1, Amount: 1, ValidityDays: 0}}, "")
	require.Error(t, err)

	// 多档正常发放
	issued, err := AdminIssueCoupons(agent.Id, 1, []DistributionCouponIssueItem{
		{Count: 2, Amount: 1, ValidityDays: 7},
		{Count: 3, Amount: 0.5, ValidityDays: 3},
	}, "活动赠券")
	require.NoError(t, err)
	assert.Equal(t, 5, issued)

	var coupons []model.DistributionCoupon
	require.NoError(t, model.DB.Where("agent_id = ?", agent.Id).Find(&coupons).Error)
	require.Len(t, coupons, 5)
	for _, coupon := range coupons {
		assert.Equal(t, DistributionCouponSourceAdmin, coupon.Source)
		assert.Equal(t, DistributionCouponStatusActive, coupon.Status)
		assert.Equal(t, 1, coupon.IssuedBy)
		assert.Equal(t, "活动赠券", coupon.Remark)
	}

	// 发放不动代理余额、不产生流水
	assert.InDelta(t, 3, couponAgentBalance(t, agent.Id), 1e-9)
	var ledgerCount int64
	require.NoError(t, model.DB.Model(&model.DistributionBalanceLedger{}).Count(&ledgerCount).Error)
	assert.Zero(t, ledgerCount)
}

func TestSweepExpiredCoupons_SelfRefundsAdminDestroys(t *testing.T) {
	couponTruncate(t)
	selfAgent := seedCouponAgent(t, 2300, 5)
	adminAgent := seedCouponAgent(t, 2301, 0)

	selfCoupon, err := ApplyAgentCoupon(2300, 2)
	require.NoError(t, err)
	issued, err := AdminIssueCoupons(adminAgent.Id, 1, []DistributionCouponIssueItem{{Count: 1, Amount: 1, ValidityDays: 7}}, "")
	require.NoError(t, err)
	require.Equal(t, 1, issued)

	// 手工把两张券都改为已过期
	past := time.Now().Unix() - 60
	require.NoError(t, model.DB.Model(&model.DistributionCoupon{}).Where("1 = 1").Update("expires_at", past).Error)

	swept, err := SweepExpiredCoupons()
	require.NoError(t, err)
	assert.Equal(t, 2, swept)

	// 券和兑换码均被物理销毁
	var couponCount int64
	require.NoError(t, model.DB.Model(&model.DistributionCoupon{}).Count(&couponCount).Error)
	assert.Zero(t, couponCount)
	var redemptionCount int64
	require.NoError(t, model.DB.Model(&model.Redemption{}).Count(&redemptionCount).Error)
	assert.Zero(t, redemptionCount)

	// self 来源退回余额并记 credit 流水
	assert.InDelta(t, 5, couponAgentBalance(t, selfAgent.Id), 1e-9)
	var refundLedger model.DistributionBalanceLedger
	require.NoError(t, model.DB.Where("agent_id = ? AND source_type = ?", selfAgent.Id, DistributionSourceCouponRefund).First(&refundLedger).Error)
	assert.Equal(t, DistributionLedgerEntryCredit, refundLedger.EntryType)
	assert.InDelta(t, 2, refundLedger.Delta, 1e-9)
	assert.Equal(t, selfCoupon.Code, refundLedger.SourceNo)

	// admin 来源仅销毁，不退款
	assert.InDelta(t, 0, couponAgentBalance(t, adminAgent.Id), 1e-9)
	var adminLedgerCount int64
	require.NoError(t, model.DB.Model(&model.DistributionBalanceLedger{}).
		Where("agent_id = ? AND source_type = ?", adminAgent.Id, DistributionSourceCouponRefund).
		Count(&adminLedgerCount).Error)
	assert.Zero(t, adminLedgerCount)
}

func TestSweepExpiredCoupons_RedeemedCouponMarkedUsedNotRefunded(t *testing.T) {
	couponTruncate(t)
	agent := seedCouponAgent(t, 2400, 5)

	coupon, err := ApplyAgentCoupon(2400, 1)
	require.NoError(t, err)

	// 模拟历史残留：兑换码已用、券仍 active 且已过期
	require.NoError(t, model.DB.Model(&model.Redemption{}).Where("id = ?", coupon.RedemptionId).
		Updates(map[string]any{
			"status":        common.RedemptionCodeStatusUsed,
			"used_user_id":  9999,
			"redeemed_time": time.Now().Unix(),
		}).Error)
	require.NoError(t, model.DB.Model(&model.DistributionCoupon{}).Where("id = ?", coupon.Id).
		Update("expires_at", time.Now().Unix()-60).Error)

	swept, err := SweepExpiredCoupons()
	require.NoError(t, err)
	assert.Equal(t, 1, swept)

	// 券被补记 used 而不是销毁，且不退款
	var saved model.DistributionCoupon
	require.NoError(t, model.DB.Where("id = ?", coupon.Id).First(&saved).Error)
	assert.Equal(t, DistributionCouponStatusUsed, saved.Status)
	assert.Equal(t, 9999, saved.UsedUserId)
	assert.InDelta(t, 4, couponAgentBalance(t, agent.Id), 1e-9)
}

func seedCouponRedeemer(t *testing.T, userID int, quota int) {
	t.Helper()
	user := &model.User{
		Id:       userID,
		Username: fmt.Sprintf("coupon_redeemer_%d", userID),
		AffCode:  fmt.Sprintf("coupon_raff_%d", userID),
		Status:   common.UserStatusEnabled,
		Quota:    quota,
	}
	require.NoError(t, model.DB.Create(user).Error)
}

func TestRedeemCoupon_AtomicSuccess(t *testing.T) {
	couponTruncate(t)
	seedCouponAgent(t, 2500, 5)
	seedCouponRedeemer(t, 2501, 100)

	coupon, err := ApplyAgentCoupon(2500, 1)
	require.NoError(t, err)

	quota, err := RedeemCoupon(coupon, 2501)
	require.NoError(t, err)
	assert.Equal(t, int(common.QuotaPerUnit), quota)

	// 用户额度到账
	var user model.User
	require.NoError(t, model.DB.Where("id = ?", 2501).First(&user).Error)
	assert.Equal(t, 100+quota, user.Quota)

	// 兑换码同事务标记已使用
	var redemption model.Redemption
	require.NoError(t, model.DB.Where("id = ?", coupon.RedemptionId).First(&redemption).Error)
	assert.Equal(t, common.RedemptionCodeStatusUsed, redemption.Status)
	assert.Equal(t, 2501, redemption.UsedUserId)
	assert.NotZero(t, redemption.RedeemedTime)

	// 券同事务原子核销
	var saved model.DistributionCoupon
	require.NoError(t, model.DB.Where("id = ?", coupon.Id).First(&saved).Error)
	assert.Equal(t, DistributionCouponStatusUsed, saved.Status)
	assert.Equal(t, 2501, saved.UsedUserId)
	assert.NotZero(t, saved.UsedAt)
}

func TestRedeemCoupon_SecondRedeemReturnsAlreadyUsed(t *testing.T) {
	couponTruncate(t)
	seedCouponAgent(t, 2510, 5)
	seedCouponRedeemer(t, 2511, 0)
	seedCouponRedeemer(t, 2512, 0)

	coupon, err := ApplyAgentCoupon(2510, 1)
	require.NoError(t, err)

	_, err = RedeemCoupon(coupon, 2511)
	require.NoError(t, err)

	// 重复兑换必须失败（条件核销防并发重复兑换）
	_, err = RedeemCoupon(coupon, 2512)
	require.ErrorIs(t, err, ErrCouponAlreadyUsed)

	// 第二个用户额度不变，核销人仍是第一个用户
	var user model.User
	require.NoError(t, model.DB.Where("id = ?", 2512).First(&user).Error)
	assert.Zero(t, user.Quota)
	var saved model.DistributionCoupon
	require.NoError(t, model.DB.Where("id = ?", coupon.Id).First(&saved).Error)
	assert.Equal(t, 2511, saved.UsedUserId)
}

func TestRedeemCoupon_ExpiredReturnsExpired(t *testing.T) {
	couponTruncate(t)
	seedCouponAgent(t, 2520, 5)
	seedCouponRedeemer(t, 2521, 0)

	coupon, err := ApplyAgentCoupon(2520, 1)
	require.NoError(t, err)

	// 手工把绑定的兑换码改为已过期
	require.NoError(t, model.DB.Model(&model.Redemption{}).Where("id = ?", coupon.RedemptionId).
		Update("expired_time", time.Now().Unix()-60).Error)

	_, err = RedeemCoupon(coupon, 2521)
	require.ErrorIs(t, err, ErrCouponExpired)

	// 额度未到账，券仍 active 留给到期清扫处理
	var user model.User
	require.NoError(t, model.DB.Where("id = ?", 2521).First(&user).Error)
	assert.Zero(t, user.Quota)
	var saved model.DistributionCoupon
	require.NoError(t, model.DB.Where("id = ?", coupon.Id).First(&saved).Error)
	assert.Equal(t, DistributionCouponStatusActive, saved.Status)
}
