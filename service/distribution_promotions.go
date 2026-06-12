package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

type DistributionGiftRuleSaveInput struct {
	Id              int    `json:"id"`
	Name            string `json:"name"`
	PackageId       int    `json:"package_id"`
	GiftPackageId   int    `json:"gift_package_id"`
	TriggerQuantity int    `json:"trigger_quantity"`
	GiftQuantity    int    `json:"gift_quantity"`
	StartsAt        int64  `json:"starts_at"`
	ExpiresAt       int64  `json:"expires_at"`
	Status          string `json:"status"`
}

func validateDistributionGiftRuleInput(input DistributionGiftRuleSaveInput) (DistributionGiftRuleSaveInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Status = strings.TrimSpace(input.Status)
	if input.Status == "" {
		input.Status = DistributionStatusEnabled
	}
	if input.Name == "" {
		return input, fmt.Errorf("name cannot be empty")
	}
	if input.PackageId <= 0 || input.GiftPackageId <= 0 {
		return input, fmt.Errorf("package ids must be greater than 0")
	}
	if input.TriggerQuantity <= 0 || input.GiftQuantity <= 0 {
		return input, fmt.Errorf("quantities must be greater than 0")
	}
	if err := ValidateDistributionStatus(input.Status); err != nil {
		return input, err
	}
	if err := ValidateDistributionTimeWindow(input.StartsAt, input.ExpiresAt); err != nil {
		return input, err
	}
	return input, nil
}

func AdminListDistributionGiftRules(startIdx int, pageSize int) ([]model.DistributionGiftRule, int64, error) {
	var rules []model.DistributionGiftRule
	query := model.DB.Model(&model.DistributionGiftRule{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id desc").Limit(pageSize).Offset(startIdx).Find(&rules).Error
	return rules, total, err
}

func AdminSaveDistributionGiftRule(input DistributionGiftRuleSaveInput) (*model.DistributionGiftRule, error) {
	input, err := validateDistributionGiftRuleInput(input)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var saved model.DistributionGiftRule
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		if input.Id > 0 {
			if err := tx.Where("id = ?", input.Id).First(&saved).Error; err != nil {
				return err
			}
		}
		saved.Name = input.Name
		saved.PackageId = input.PackageId
		saved.GiftPackageId = input.GiftPackageId
		saved.TriggerQuantity = input.TriggerQuantity
		saved.GiftQuantity = input.GiftQuantity
		saved.StartsAt = input.StartsAt
		saved.ExpiresAt = input.ExpiresAt
		saved.Status = input.Status
		saved.UpdatedAt = now
		if input.Id > 0 {
			return tx.Save(&saved).Error
		}
		saved.CreatedAt = now
		return tx.Create(&saved).Error
	})
	if err != nil {
		return nil, err
	}
	return &saved, nil
}

func AdminUpdateDistributionGiftRuleStatus(ruleID int, status string) (*model.DistributionGiftRule, error) {
	if ruleID <= 0 {
		return nil, fmt.Errorf("invalid gift rule id")
	}
	if err := ValidateDistributionStatus(status); err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var rule model.DistributionGiftRule
	if err := model.DB.Where("id = ?", ruleID).First(&rule).Error; err != nil {
		return nil, err
	}
	rule.Status = status
	rule.UpdatedAt = now
	if err := model.DB.Save(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}
