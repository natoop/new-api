package controller

import (
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

// submitBusinessLeadRequest 匿名提交商务线索的请求体
type submitBusinessLeadRequest struct {
	CompanyName     string `json:"company_name" binding:"required,max=191"`
	ContactName     string `json:"contact_name" binding:"required,max=191"`
	ContactInfo     string `json:"contact_info" binding:"required,max=191"`
	CooperationType string `json:"cooperation_type" binding:"required,max=32"`
	Requirements    string `json:"requirements" binding:"max=5000"`
}

// SubmitBusinessLead 匿名提交商务线索
func SubmitBusinessLead(c *gin.Context) {
	var req submitBusinessLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	// contact_info 可能是手机/微信/邮箱，不强制 email 格式，仅做非空与长度校验（已由 binding 完成）
	if !model.IsValidCooperationType(req.CooperationType) {
		common.ApiErrorMsg(c, "无效的合作类型")
		return
	}

	lead := &model.BusinessLead{
		CompanyName:     strings.TrimSpace(req.CompanyName),
		ContactName:     strings.TrimSpace(req.ContactName),
		ContactInfo:     strings.TrimSpace(req.ContactInfo),
		CooperationType: req.CooperationType,
		Requirements:    req.Requirements,
		Status:          model.BusinessLeadStatusPending,
		CreatedAt:       common.GetTimestamp(),
	}
	if err := model.DB.Create(lead).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"id": lead.Id})
}

// AdminListBusinessLeads 管理员分页查询商务线索
// 返回结构与 controller/log.go 的 GetAllLogs 一致：common.ApiSuccess(c, pageInfo)
func AdminListBusinessLeads(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	// 契约要求 query 使用 page（GetPageQuery 默认读 p），此处显式兼容 page 参数
	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		pageInfo.Page = page
	}

	status := c.Query("status")
	keyword := c.Query("keyword")

	leads, total, err := model.GetAllBusinessLeads(status, keyword, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(leads)
	common.ApiSuccess(c, pageInfo)
}

// AdminUpdateBusinessLeadStatus 管理员更新商务线索状态
func AdminUpdateBusinessLeadStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		common.ApiErrorMsg(c, "无效的 id")
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	if !model.IsValidBusinessLeadStatus(req.Status) {
		common.ApiErrorMsg(c, "无效的状态")
		return
	}
	if err := model.UpdateBusinessLeadStatus(id, req.Status); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"id": id, "status": req.Status})
}

// AdminDeleteBusinessLead 管理员删除商务线索
func AdminDeleteBusinessLead(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		common.ApiErrorMsg(c, "无效的 id")
		return
	}
	if err := model.DeleteBusinessLead(id); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"id": id})
}
