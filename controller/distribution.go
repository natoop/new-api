package controller

import (
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

type distributionStatusRequest struct {
	Status string `json:"status"`
}

type distributionPurchaseRequest struct {
	PackageId      int    `json:"package_id"`
	IdempotencyKey string `json:"idempotency_key"`
}

type distributionAssignInventoryRequest struct {
	InventoryId    int `json:"inventory_id"`
	CustomerUserId int `json:"customer_user_id"`
}

type distributionRefundInventoryRequest struct {
	InventoryId int `json:"inventory_id"`
}

func GetDistributionAgentProfile(c *gin.Context) {
	userID := c.GetInt("id")
	profile, err := service.GetDistributionAgentProfile(userID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, profile)
}

func ListDistributionAgentPackages(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	packages, total, err := service.ListDistributionAgentPackages(pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(packages)
	common.ApiSuccess(c, pageInfo)
}

func PurchaseDistributionPackage(c *gin.Context) {
	userID := c.GetInt("id")
	var req distributionPurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	result, err := service.PurchaseDistributionPackage(userID, req.PackageId, req.IdempotencyKey)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func ListDistributionAgentInventory(c *gin.Context) {
	userID := c.GetInt("id")
	pageInfo := common.GetPageQuery(c)
	inventories, total, err := service.ListDistributionAgentInventory(userID, service.DistributionInventoryListInput{
		Keyword:  c.Query("keyword"),
		Status:   c.Query("status"),
		StartIdx: pageInfo.GetStartIdx(),
		PageSize: pageInfo.GetPageSize(),
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(inventories)
	common.ApiSuccess(c, pageInfo)
}

func ListDistributionAgentInventoryPackageOptions(c *gin.Context) {
	options, err := service.ListDistributionAgentInventoryPackageOptions(c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, options)
}

func AssignDistributionAgentInventory(c *gin.Context) {
	userID := c.GetInt("id")
	var req distributionAssignInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	result, err := service.AssignDistributionAgentInventory(userID, req.InventoryId, req.CustomerUserId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func RefundDistributionAgentInventory(c *gin.Context) {
	userID := c.GetInt("id")
	var req distributionRefundInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	result, err := service.RefundDistributionAgentInventory(userID, req.InventoryId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func ListDistributionAgentLedger(c *gin.Context) {
	userID := c.GetInt("id")
	pageInfo := common.GetPageQuery(c)
	ledgers, total, err := service.ListDistributionAgentLedger(userID, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(ledgers)
	common.ApiSuccess(c, pageInfo)
}

func ListDistributionAgentProfit(c *gin.Context) {
	userID := c.GetInt("id")
	pageInfo := common.GetPageQuery(c)
	profits, total, err := service.ListDistributionAgentProfit(userID, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(profits)
	common.ApiSuccess(c, pageInfo)
}

func ListDistributionCustomers(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	customers, total, err := service.ListDistributionCustomers(c.GetInt("id"), c.Query("keyword"), pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(customers)
	common.ApiSuccess(c, pageInfo)
}

func ListDistributionInvitations(c *gin.Context) {
	invitations, err := service.ListDistributionInvitations(c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, invitations)
}

func CreateDistributionInvitation(c *gin.Context) {
	var req service.DistributionInvitationCreateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	invitation, err := service.CreateDistributionInvitation(c.GetInt("id"), req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, invitation)
}

func AcceptDistributionInvitation(c *gin.Context) {
	var req service.DistributionInvitationAcceptInput
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	invitation, err := service.AcceptDistributionInvitation(c.GetInt("id"), req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, invitation)
}

func ListDistributionPromoCodes(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	codes, total, err := service.ListDistributionPromoCodes(c.GetInt("id"), service.DistributionPromoCodeListInput{
		TimeFilter:  c.Query("time_filter"),
		UsageFilter: c.Query("usage_filter"),
		StartIdx:    pageInfo.GetStartIdx(),
		PageSize:    pageInfo.GetPageSize(),
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(codes)
	common.ApiSuccess(c, pageInfo)
}

func SaveDistributionPromoCode(c *gin.Context) {
	var req service.DistributionPromoCodeSaveInput
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if idParam := c.Param("id"); idParam != "" {
		promoCodeID, err := strconv.Atoi(idParam)
		if err != nil {
			common.ApiErrorMsg(c, "无效的促销码 ID")
			return
		}
		req.Id = promoCodeID
	}
	code, err := service.SaveDistributionPromoCode(c.GetInt("id"), req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, code)
}

func UpdateDistributionPromoCodeStatus(c *gin.Context) {
	promoCodeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorMsg(c, "无效的促销码 ID")
		return
	}
	var req distributionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	code, err := service.UpdateDistributionPromoCodeStatus(c.GetInt("id"), promoCodeID, req.Status)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, code)
}

func AdminListDistributionAgents(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	agents, total, err := service.AdminListDistributionAgents(c.Query("keyword"), pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(agents)
	common.ApiSuccess(c, pageInfo)
}

func AdminSaveDistributionAgent(c *gin.Context) {
	var req service.DistributionAgentSaveInput
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	agent, err := service.AdminSaveDistributionAgent(req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, agent)
}

func AdminUpdateDistributionAgentStatus(c *gin.Context) {
	agentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorMsg(c, "无效的代理 ID")
		return
	}
	var req distributionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	agent, err := service.AdminUpdateDistributionAgentStatus(agentID, req.Status)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, agent)
}

func AdminAdjustDistributionAgentBalance(c *gin.Context) {
	agentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorMsg(c, "无效的代理 ID")
		return
	}
	var req service.DistributionBalanceAdjustmentInput
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	adjustment, err := service.AdminAdjustDistributionAgentBalance(agentID, c.GetInt("id"), req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, adjustment)
}

func AdminListDistributionPackages(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	packages, total, err := service.AdminListDistributionPackages(pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(packages)
	common.ApiSuccess(c, pageInfo)
}

func AdminCreateDistributionPackage(c *gin.Context) {
	var req service.DistributionPackageSaveInput
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	distributionPackage, err := service.AdminCreateDistributionPackage(req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, distributionPackage)
}

func AdminUpdateDistributionPackage(c *gin.Context) {
	packageID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorMsg(c, "无效的套餐 ID")
		return
	}
	var req service.DistributionPackageSaveInput
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	distributionPackage, err := service.AdminUpdateDistributionPackage(packageID, req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, distributionPackage)
}

func AdminUpdateDistributionPackageStatus(c *gin.Context) {
	packageID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorMsg(c, "无效的套餐 ID")
		return
	}
	var req distributionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	distributionPackage, err := service.AdminUpdateDistributionPackageStatus(packageID, req.Status)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, distributionPackage)
}

func AdminListDistributionPriceConfigs(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	configs, total, err := service.AdminListDistributionPriceConfigs(pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(configs)
	common.ApiSuccess(c, pageInfo)
}

func AdminSaveDistributionPriceConfig(c *gin.Context) {
	var req service.DistributionPriceConfigSaveInput
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if idParam := c.Param("id"); idParam != "" {
		configID, err := strconv.Atoi(idParam)
		if err != nil {
			common.ApiErrorMsg(c, "无效的价格配置 ID")
			return
		}
		req.Id = configID
	}
	config, err := service.AdminSaveDistributionPriceConfig(req, c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, config)
}

func AdminUpdateDistributionPriceConfigStatus(c *gin.Context) {
	configID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorMsg(c, "无效的价格配置 ID")
		return
	}
	var req distributionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	config, err := service.AdminUpdateDistributionPriceConfigStatus(configID, req.Status, c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, config)
}

func AdminListDistributionProfit(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	profits, total, err := service.AdminListDistributionProfit(pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(profits)
	common.ApiSuccess(c, pageInfo)
}

func AdminListDistributionAttributionLogs(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	logs, total, err := service.AdminListDistributionAttributionLogs(pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(logs)
	common.ApiSuccess(c, pageInfo)
}

func AdminListDistributionGiftRules(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	rules, total, err := service.AdminListDistributionGiftRules(pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(rules)
	common.ApiSuccess(c, pageInfo)
}

func AdminSaveDistributionGiftRule(c *gin.Context) {
	var req service.DistributionGiftRuleSaveInput
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if idParam := c.Param("id"); idParam != "" {
		ruleID, err := strconv.Atoi(idParam)
		if err != nil {
			common.ApiErrorMsg(c, "无效的赠送规则 ID")
			return
		}
		req.Id = ruleID
	}
	rule, err := service.AdminSaveDistributionGiftRule(req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, rule)
}

func AdminUpdateDistributionGiftRuleStatus(c *gin.Context) {
	ruleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiErrorMsg(c, "无效的赠送规则 ID")
		return
	}
	var req distributionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	rule, err := service.AdminUpdateDistributionGiftRuleStatus(ruleID, req.Status)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, rule)
}

func AdminListDistributionOpsAuthorizations(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	authorizations, total, err := service.AdminListDistributionOpsAuthorizations(pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(authorizations)
	common.ApiSuccess(c, pageInfo)
}

func AdminGrantDistributionOpsAuthorization(c *gin.Context) {
	var req service.DistributionOpsAuthorizationInput
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	authorization, err := service.AdminGrantDistributionOpsAuthorization(c.GetInt("id"), req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, authorization)
}

func AdminRevokeDistributionOpsAuthorization(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		common.ApiErrorMsg(c, "无效的用户 ID")
		return
	}
	authorization, err := service.AdminRevokeDistributionOpsAuthorization(c.GetInt("id"), userID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, authorization)
}
