package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

// SetUserChannelPinRequest is the payload for creating or overwriting a user channel pin.
type SetUserChannelPinRequest struct {
	UserId     int `json:"user_id"`
	ChannelId  int `json:"channel_id"`
	TtlSeconds int `json:"ttl_seconds"`
}

// SetUserChannelPin pins a user to a specific channel for a limited time.
func SetUserChannelPin(c *gin.Context) {
	var req SetUserChannelPinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	channel, errMsg := validateUserChannelPinRequest(&req)
	if errMsg != "" {
		common.ApiErrorMsg(c, errMsg)
		return
	}
	if err := service.SetUserPin(req.UserId, req.ChannelId, req.TtlSeconds); err != nil {
		common.ApiError(c, err)
		return
	}
	message := ""
	if !channelServesAnyGroupModel(channel) {
		message = "警告：该渠道当前未服务任何分组模型，被 pin 的用户请求可能失败"
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data": gin.H{
			"user_id":     req.UserId,
			"channel_id":  req.ChannelId,
			"ttl_seconds": req.TtlSeconds,
		},
	})
}

// validateUserChannelPinRequest validates params and returns the target channel on success.
func validateUserChannelPinRequest(req *SetUserChannelPinRequest) (*model.Channel, string) {
	if req.UserId <= 0 || req.ChannelId <= 0 || req.TtlSeconds <= 0 {
		return nil, "user_id、channel_id、ttl_seconds 必须为正整数"
	}
	if req.TtlSeconds > service.UserChannelPinMaxTTLSeconds {
		return nil, "ttl_seconds 超出上限（最长 7 天）"
	}
	channel, err := model.CacheGetChannel(req.ChannelId)
	if err != nil {
		return nil, "渠道不存在：" + err.Error()
	}
	if channel.Status != common.ChannelStatusEnabled {
		return nil, "渠道未启用，禁止 pin"
	}
	if _, err := model.GetUserById(req.UserId, false); err != nil {
		return nil, "用户不存在：" + err.Error()
	}
	return channel, ""
}

// channelServesAnyGroupModel reports whether the channel currently serves at least one
// (group, model) pair. It is a soft check used for warning only, because the caller
// side has already done candidate filtering before pinning.
func channelServesAnyGroupModel(channel *model.Channel) bool {
	if channel == nil {
		return false
	}
	groups := channel.GetGroups()
	for _, name := range channel.GetModels() {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if model.IsChannelEnabledForAnyGroupModel(groups, name, channel.Id) {
			return true
		}
	}
	return false
}

// ClearUserChannelPin removes the channel pin of the given user.
func ClearUserChannelPin(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil || userId <= 0 {
		common.ApiErrorMsg(c, "user_id 必须为正整数")
		return
	}
	service.ClearUserPin(userId)
	common.ApiSuccess(c, gin.H{"user_id": userId})
}

// ListUserChannelPins returns all active user channel pins.
func ListUserChannelPins(c *gin.Context) {
	pins, err := service.ListUserPins()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, pins)
}
