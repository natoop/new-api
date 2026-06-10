package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

type DistributionOpsAuthorizationInput struct {
	UserId int    `json:"user_id"`
	Remark string `json:"remark"`
}

func AdminListDistributionOpsAuthorizations(startIdx int, pageSize int) ([]model.DistributionOpsDashboardAuthorization, int64, error) {
	var authorizations []model.DistributionOpsDashboardAuthorization
	query := model.DB.Model(&model.DistributionOpsDashboardAuthorization{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id desc").Limit(pageSize).Offset(startIdx).Find(&authorizations).Error
	return authorizations, total, err
}

func AdminGrantDistributionOpsAuthorization(operatorUserID int, input DistributionOpsAuthorizationInput) (*model.DistributionOpsDashboardAuthorization, error) {
	if input.UserId <= 0 {
		return nil, fmt.Errorf("user_id must be greater than 0")
	}
	now := time.Now().Unix()
	remark := strings.TrimSpace(input.Remark)
	var authorization model.DistributionOpsDashboardAuthorization
	err := model.DB.Where("user_id = ?", input.UserId).First(&authorization).Error
	if err == nil {
		authorization.Status = "granted"
		authorization.GrantedByUserId = operatorUserID
		authorization.GrantedAt = now
		authorization.RevokedByUserId = 0
		authorization.RevokedAt = 0
		authorization.Remark = remark
		authorization.UpdatedAt = now
		if err := model.DB.Save(&authorization).Error; err != nil {
			return nil, err
		}
		return &authorization, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	authorization = model.DistributionOpsDashboardAuthorization{
		UserId:          input.UserId,
		Status:          "granted",
		GrantedByUserId: operatorUserID,
		GrantedAt:       now,
		Remark:          remark,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := model.DB.Create(&authorization).Error; err != nil {
		return nil, err
	}
	return &authorization, nil
}

func AdminRevokeDistributionOpsAuthorization(operatorUserID int, userID int) (*model.DistributionOpsDashboardAuthorization, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	now := time.Now().Unix()
	var authorization model.DistributionOpsDashboardAuthorization
	if err := model.DB.Where("user_id = ?", userID).First(&authorization).Error; err != nil {
		return nil, err
	}
	authorization.Status = "revoked"
	authorization.RevokedByUserId = operatorUserID
	authorization.RevokedAt = now
	authorization.UpdatedAt = now
	if err := model.DB.Save(&authorization).Error; err != nil {
		return nil, err
	}
	return &authorization, nil
}

func HasDistributionOpsAuthorization(userID int) bool {
	if userID <= 0 {
		return false
	}
	var count int64
	if err := model.DB.Model(&model.DistributionOpsDashboardAuthorization{}).
		Where("user_id = ? AND status = ?", userID, "granted").
		Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}
