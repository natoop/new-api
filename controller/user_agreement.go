package controller

import (
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
)

// userAgreementVersionMaxLen 协议版本号最大长度（与 UserAgreementConsent.Version varchar(32) 对齐）
const userAgreementVersionMaxLen = 32

// GetUserAgreementStatus 获取当前用户的协议签署状态（需登录）
func GetUserAgreementStatus(c *gin.Context) {
	legalSetting := system_setting.GetLegalSettings()
	version := legalSetting.UserAgreementVersion

	// 逃生门：未启用、未设置版本号、或协议内容为空时不强制签约（防"空协议阻断"）
	if !legalSetting.ConsoleAgreementEnabled || version == "" || strings.TrimSpace(legalSetting.UserAgreement) == "" {
		common.ApiSuccess(c, gin.H{
			"required": false,
			"version":  version,
			"agreed":   false,
		})
		return
	}

	userId := c.GetInt("id")
	agreed, err := model.HasUserConsented(userId, version)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{
		"required": !agreed,
		"version":  version,
		"agreed":   agreed,
	})
}

// ConsentUserAgreement 记录当前用户签署协议（需登录）
func ConsentUserAgreement(c *gin.Context) {
	var req struct {
		Version string `json:"version"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	if len(req.Version) > userAgreementVersionMaxLen {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "协议版本号过长",
		})
		return
	}

	currentVersion := system_setting.GetLegalSettings().UserAgreementVersion
	if currentVersion == "" || req.Version != currentVersion {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "协议版本不匹配，请刷新后重试",
		})
		return
	}

	userId := c.GetInt("id")
	if err := model.RecordUserConsent(userId, currentVersion, c.ClientIP()); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{
		"version": currentVersion,
		"agreed":  true,
	})
}
