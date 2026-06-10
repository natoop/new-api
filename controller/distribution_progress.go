package controller

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

// GetAgentProgress 返回当前用户的代理自然晋升进度（UserAuth）。
// 用于「代理计划」说明页：邀请数 / 付费邀请数 / 晋升门槛 / 是否已是代理 / 邀请码。
func GetAgentProgress(c *gin.Context) {
	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, true)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// 邀请码为空时补发一个（与 GetAffCode 行为一致），保证前端能直接拼邀请链接
	if user.AffCode == "" {
		user.AffCode = common.GetRandomString(4)
		if err := user.Update(false); err != nil {
			common.ApiError(c, err)
			return
		}
	}

	// 已是代理：角色恰为代理，或已有分销代理记录
	// （管理员/Root 不再被标记 is_agent，避免前端把无 DistributionAgent 记录的管理员引导到 403 的 /agent 页）
	isAgent := user.Role == common.RoleAgentUser
	if !isAgent {
		var agentCount int64
		if err := model.DB.Model(&model.DistributionAgent{}).
			Where("user_id = ?", userId).Count(&agentCount).Error; err != nil {
			common.ApiError(c, err)
			return
		}
		isAgent = agentCount > 0
	}

	threshold := service.DistributionPromotionThreshold()
	invitedCount, paidCount := 0, 0
	if !isAgent {
		invitedCount, paidCount, err = service.CountInviterPromotionStats(userId)
		if err != nil {
			common.ApiError(c, err)
			return
		}
	}

	common.ApiSuccess(c, gin.H{
		"invited_count":    invitedCount,
		"paid_count":       paidCount,
		"required_invites": threshold,
		"required_paid":    threshold,
		"is_agent":         isAgent,
		"aff_code":         user.AffCode,
	})
}
