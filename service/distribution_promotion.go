package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

// distributionNaturalPromotionThreshold 自然晋升门槛：
// 邀请注册人数 ≥ 3 且其中实际付费（订阅或充值成功）的去重人数 ≥ 3。
const distributionNaturalPromotionThreshold = 3

// MaybePromoteInviterToAgent 在某用户付费成功后检查其邀请人是否满足自然晋升代理条件，
// 满足则将邀请人晋升为二级代理（复用 ensureDistributionAgentForUser）。
// 设计为钩子回调：内部吞错并记日志，不向调用方传播。
func MaybePromoteInviterToAgent(buyerUserID int) {
	if buyerUserID <= 0 {
		return
	}
	var buyer model.User
	if err := model.DB.Select("id, inviter_id").Where("id = ?", buyerUserID).First(&buyer).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			common.SysError("distribution promotion: failed to load buyer: " + err.Error())
		}
		return
	}
	inviterID := buyer.InviterId
	if inviterID <= 0 {
		return
	}

	eligible, err := inviterEligibleForNaturalPromotion(inviterID)
	if err != nil {
		common.SysError("distribution promotion: eligibility check failed: " + err.Error())
		return
	}
	if !eligible {
		return
	}

	err = model.DB.Transaction(func(tx *gorm.DB) error {
		var inviter model.User
		if err := tx.Where("id = ?", inviterID).First(&inviter).Error; err != nil {
			return err
		}
		_, err := ensureDistributionAgentForUser(tx, &inviter, DistributionAgentLevelSecondary, 0)
		return err
	})
	if err != nil {
		// 并发下唯一键冲突视为已晋升，非错误
		if isDuplicateKeyError(err) {
			return
		}
		common.SysError("distribution promotion: failed to promote inviter: " + err.Error())
		return
	}
	common.SysLog(fmt.Sprintf("distribution promotion: user %d promoted to agent (natural promotion, triggered by buyer %d)", inviterID, buyerUserID))
}

// inviterEligibleForNaturalPromotion 判断邀请人是否满足自然晋升条件。
func inviterEligibleForNaturalPromotion(inviterID int) (bool, error) {
	// 已有代理记录 → 跳过
	var agentCount int64
	if err := model.DB.Model(&model.DistributionAgent{}).
		Where("user_id = ?", inviterID).Count(&agentCount).Error; err != nil {
		return false, err
	}
	if agentCount > 0 {
		return false, nil
	}

	// 角色已 >= 代理（含管理员/Root）→ 跳过
	var inviter model.User
	if err := model.DB.Select("id, role").Where("id = ?", inviterID).First(&inviter).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	if inviter.Role >= common.RoleAgentUser {
		return false, nil
	}

	// 邀请注册人数 ≥ 3 且实际付费的被邀请人去重数 ≥ 3
	invitedCount, paidCount, err := CountInviterPromotionStats(inviterID)
	if err != nil {
		return false, err
	}
	return invitedCount >= distributionNaturalPromotionThreshold &&
		paidCount >= distributionNaturalPromotionThreshold, nil
}

// DistributionPromotionThreshold 暴露自然晋升门槛（邀请数与付费邀请数共用同一阈值），供展示层使用。
func DistributionPromotionThreshold() int {
	return distributionNaturalPromotionThreshold
}

// CountInviterPromotionStats 统计某邀请人的邀请注册人数，以及其中实际付费
//（subscription_orders ∪ top_ups，status=success）的去重人数。
func CountInviterPromotionStats(inviterID int) (invitedCount int, paidCount int, err error) {
	var inviteeIDs []int
	if err = model.DB.Model(&model.User{}).
		Where("inviter_id = ?", inviterID).Pluck("id", &inviteeIDs).Error; err != nil {
		return 0, 0, err
	}
	invitedCount = len(inviteeIDs)
	if invitedCount == 0 {
		return 0, 0, nil
	}

	paidInvitees := make(map[int]struct{})
	var paidIDs []int
	if err = model.DB.Model(&model.SubscriptionOrder{}).
		Distinct("user_id").
		Where("user_id IN ? AND status = ?", inviteeIDs, common.TopUpStatusSuccess).
		Pluck("user_id", &paidIDs).Error; err != nil {
		return 0, 0, err
	}
	for _, id := range paidIDs {
		paidInvitees[id] = struct{}{}
	}
	paidIDs = paidIDs[:0]
	if err = model.DB.Model(&model.TopUp{}).
		Distinct("user_id").
		Where("user_id IN ? AND status = ?", inviteeIDs, common.TopUpStatusSuccess).
		Pluck("user_id", &paidIDs).Error; err != nil {
		return 0, 0, err
	}
	for _, id := range paidIDs {
		paidInvitees[id] = struct{}{}
	}
	return invitedCount, len(paidInvitees), nil
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique constraint") || strings.Contains(msg, "unique failed") || strings.Contains(msg, "constraint failed")
}
