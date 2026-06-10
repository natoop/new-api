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

func seedPromotionUser(t *testing.T, id int, inviterID int, role int) {
	t.Helper()
	user := &model.User{
		Id:        id,
		Username:  fmt.Sprintf("promo_user_%d", id),
		AffCode:   fmt.Sprintf("aff_%d", id),
		Role:      role,
		Status:    common.UserStatusEnabled,
		InviterId: inviterID,
	}
	require.NoError(t, model.DB.Create(user).Error)
}

func seedSuccessSubscriptionOrder(t *testing.T, userID int, tradeNo string) {
	t.Helper()
	order := &model.SubscriptionOrder{
		UserId:          userID,
		PlanId:          1,
		Money:           9.9,
		TradeNo:         tradeNo,
		PaymentMethod:   "epay",
		PaymentProvider: model.PaymentProviderEpay,
		Status:          common.TopUpStatusSuccess,
		CreateTime:      time.Now().Unix(),
	}
	require.NoError(t, model.DB.Create(order).Error)
}

func seedSuccessTopUp(t *testing.T, userID int, tradeNo string) {
	t.Helper()
	topUp := &model.TopUp{
		UserId:          userID,
		Amount:          5,
		Money:           5,
		TradeNo:         tradeNo,
		PaymentMethod:   "stripe",
		PaymentProvider: model.PaymentProviderStripe,
		Status:          common.TopUpStatusSuccess,
		CreateTime:      time.Now().Unix(),
	}
	require.NoError(t, model.DB.Create(topUp).Error)
}

func countAgentsForUser(t *testing.T, userID int) int64 {
	t.Helper()
	var count int64
	require.NoError(t, model.DB.Model(&model.DistributionAgent{}).Where("user_id = ?", userID).Count(&count).Error)
	return count
}

func TestMaybePromoteInviterToAgent_TwoInviteesNotEnough(t *testing.T) {
	truncate(t)

	inviter := 1000
	seedPromotionUser(t, inviter, 0, common.RoleCommonUser)
	seedPromotionUser(t, 1001, inviter, common.RoleCommonUser)
	seedPromotionUser(t, 1002, inviter, common.RoleCommonUser)
	seedSuccessSubscriptionOrder(t, 1001, "promo-order-1001")
	seedSuccessSubscriptionOrder(t, 1002, "promo-order-1002")

	MaybePromoteInviterToAgent(1001)

	assert.Zero(t, countAgentsForUser(t, inviter))
}

func TestMaybePromoteInviterToAgent_ThreeInviteesWithoutEnoughPayers(t *testing.T) {
	truncate(t)

	inviter := 1100
	seedPromotionUser(t, inviter, 0, common.RoleCommonUser)
	seedPromotionUser(t, 1101, inviter, common.RoleCommonUser)
	seedPromotionUser(t, 1102, inviter, common.RoleCommonUser)
	seedPromotionUser(t, 1103, inviter, common.RoleCommonUser)
	// 只有两个付费（其中一人重复付费不应被重复计数）
	seedSuccessSubscriptionOrder(t, 1101, "promo-order-1101")
	seedSuccessSubscriptionOrder(t, 1101, "promo-order-1101b")
	seedSuccessTopUp(t, 1102, "promo-topup-1102")

	MaybePromoteInviterToAgent(1101)

	assert.Zero(t, countAgentsForUser(t, inviter))
}

func TestMaybePromoteInviterToAgent_ThreePaidInviteesPromotes(t *testing.T) {
	truncate(t)

	inviter := 1200
	seedPromotionUser(t, inviter, 0, common.RoleCommonUser)
	seedPromotionUser(t, 1201, inviter, common.RoleCommonUser)
	seedPromotionUser(t, 1202, inviter, common.RoleCommonUser)
	seedPromotionUser(t, 1203, inviter, common.RoleCommonUser)
	// 订阅 ∪ 充值 两个来源都算实际消费
	seedSuccessSubscriptionOrder(t, 1201, "promo-order-1201")
	seedSuccessTopUp(t, 1202, "promo-topup-1202")
	seedSuccessSubscriptionOrder(t, 1203, "promo-order-1203")

	MaybePromoteInviterToAgent(1203)

	require.EqualValues(t, 1, countAgentsForUser(t, inviter))
	var agent model.DistributionAgent
	require.NoError(t, model.DB.Where("user_id = ?", inviter).First(&agent).Error)
	assert.Equal(t, DistributionAgentLevelSecondary, agent.Level)
	assert.Equal(t, DistributionStatusEnabled, agent.Status)
	assert.Equal(t, 0, agent.ParentAgentId)

	var inviterUser model.User
	require.NoError(t, model.DB.Where("id = ?", inviter).First(&inviterUser).Error)
	assert.Equal(t, common.RoleAgentUser, inviterUser.Role)

	// 历史邀请客户被同步绑定
	var ownershipCount int64
	require.NoError(t, model.DB.Model(&model.DistributionCustomerOwnership{}).Where("agent_id = ?", agent.Id).Count(&ownershipCount).Error)
	assert.EqualValues(t, 3, ownershipCount)

	// 再次触发应幂等（已是代理 → 跳过，不重复创建）
	MaybePromoteInviterToAgent(1203)
	assert.EqualValues(t, 1, countAgentsForUser(t, inviter))
}

func TestMaybePromoteInviterToAgent_NoInviterNoPanic(t *testing.T) {
	truncate(t)

	seedPromotionUser(t, 1300, 0, common.RoleCommonUser)
	seedSuccessSubscriptionOrder(t, 1300, "promo-order-1300")

	// 无邀请人 / 用户不存在 都不应 panic
	MaybePromoteInviterToAgent(1300)
	MaybePromoteInviterToAgent(99999)
}
