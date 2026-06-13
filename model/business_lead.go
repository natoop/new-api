package model

import (
	"errors"

	"github.com/QuantumNous/new-api/common"
)

// BusinessLead 商务线索（合作意向收集）
type BusinessLead struct {
	Id              int    `json:"id" gorm:"primaryKey;autoIncrement"`
	CompanyName     string `json:"company_name" gorm:"type:varchar(191);not null"`
	ContactName     string `json:"contact_name" gorm:"type:varchar(191);not null"`
	ContactInfo     string `json:"contact_info" gorm:"type:varchar(191);not null"`
	CooperationType string `json:"cooperation_type" gorm:"type:varchar(32);not null"`
	Requirements    string `json:"requirements" gorm:"type:text"`
	Status          string `json:"status" gorm:"type:varchar(32);not null;default:'pending';index"`
	CreatedAt       int64  `json:"created_at" gorm:"bigint;index"`
}

func (BusinessLead) TableName() string {
	return "business_leads"
}

// 合作类型枚举
const (
	CooperationTypeApiWholesale = "api_wholesale"
	CooperationTypeReseller     = "reseller"
	CooperationTypeIntegration  = "integration"
	CooperationTypeOther        = "other"
)

// 状态枚举
const (
	BusinessLeadStatusPending   = "pending"
	BusinessLeadStatusContacted = "contacted"
	BusinessLeadStatusArchived  = "archived"
)

// IsValidCooperationType 校验合作类型是否在枚举内
func IsValidCooperationType(t string) bool {
	switch t {
	case CooperationTypeApiWholesale, CooperationTypeReseller, CooperationTypeIntegration, CooperationTypeOther:
		return true
	default:
		return false
	}
}

// IsValidBusinessLeadStatus 校验状态是否在枚举内
func IsValidBusinessLeadStatus(s string) bool {
	switch s {
	case BusinessLeadStatusPending, BusinessLeadStatusContacted, BusinessLeadStatusArchived:
		return true
	default:
		return false
	}
}

// CreateBusinessLead 创建商务线索，默认状态 pending
func CreateBusinessLead(lead *BusinessLead) error {
	if lead == nil {
		return errors.New("线索不能为空")
	}
	if lead.Status == "" {
		lead.Status = BusinessLeadStatusPending
	}
	if lead.CreatedAt == 0 {
		lead.CreatedAt = common.GetTimestamp()
	}
	return DB.Create(lead).Error
}

// GetAllBusinessLeads 分页查询商务线索，支持 status 精确过滤与 keyword 模糊匹配
// keyword 同时匹配 company_name / contact_name（LIKE）
func GetAllBusinessLeads(status string, keyword string, startIdx int, num int) (leads []*BusinessLead, total int64, err error) {
	tx := DB.Model(&BusinessLead{})
	if status != "" {
		tx = tx.Where("status = ?", status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		tx = tx.Where("company_name LIKE ? OR contact_name LIKE ?", like, like)
	}
	if err = tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err = tx.Order("created_at desc, id desc").Limit(num).Offset(startIdx).Find(&leads).Error
	if err != nil {
		return nil, 0, err
	}
	return leads, total, nil
}

// UpdateBusinessLeadStatus 更新指定线索状态
func UpdateBusinessLeadStatus(id int, status string) error {
	if !IsValidBusinessLeadStatus(status) {
		return errors.New("无效的状态")
	}
	result := DB.Model(&BusinessLead{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("线索不存在")
	}
	return nil
}

// DeleteBusinessLead 删除指定线索
func DeleteBusinessLead(id int) error {
	result := DB.Delete(&BusinessLead{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("线索不存在")
	}
	return nil
}
