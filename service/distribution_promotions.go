package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

type DistributionPromoCodeSaveInput struct {
	Id             int    `json:"id"`
	Code           string `json:"code"`
	Status         string `json:"status"`
	DiscountType   string `json:"discount_type"`
	DiscountValue  int    `json:"discount_value"`
	MaxRedemptions int    `json:"max_redemptions"`
	StartsAt       int64  `json:"starts_at"`
	ExpiresAt      int64  `json:"expires_at"`
}

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

func validateDistributionPromoCodeInput(input DistributionPromoCodeSaveInput) (DistributionPromoCodeSaveInput, error) {
	input.Code = strings.TrimSpace(input.Code)
	input.Status = strings.TrimSpace(input.Status)
	input.DiscountType = strings.TrimSpace(input.DiscountType)
	if input.Status == "" {
		input.Status = DistributionStatusEnabled
	}
	if input.Code == "" {
		return input, fmt.Errorf("code cannot be empty")
	}
	if err := ValidateDistributionStatus(input.Status); err != nil {
		return input, err
	}
	if err := ValidateDistributionPromoDiscount(input.DiscountType, input.DiscountValue); err != nil {
		return input, err
	}
	if input.MaxRedemptions < 0 {
		return input, fmt.Errorf("max_redemptions cannot be negative")
	}
	if err := ValidateDistributionTimeWindow(input.StartsAt, input.ExpiresAt); err != nil {
		return input, err
	}
	return input, nil
}

func ListDistributionPromoCodes(userID int) ([]model.DistributionPromoCode, error) {
	agent, err := GetEnabledDistributionAgentByUserID(userID)
	if err != nil {
		return nil, err
	}
	var codes []model.DistributionPromoCode
	err = model.DB.Where("agent_id = ?", agent.Id).Order("id desc").Find(&codes).Error
	return codes, err
}

func SaveDistributionPromoCode(userID int, input DistributionPromoCodeSaveInput) (*model.DistributionPromoCode, error) {
	input, err := validateDistributionPromoCodeInput(input)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var saved model.DistributionPromoCode
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		var agent model.DistributionAgent
		if err := distributionLock(tx).Where("user_id = ? AND status = ?", userID, DistributionStatusEnabled).First(&agent).Error; err != nil {
			return err
		}
		if input.Id > 0 {
			if err := tx.Where("id = ? AND agent_id = ?", input.Id, agent.Id).First(&saved).Error; err != nil {
				return err
			}
		}
		saved.AgentId = agent.Id
		saved.Code = input.Code
		saved.Status = input.Status
		saved.DiscountType = input.DiscountType
		saved.DiscountValue = input.DiscountValue
		saved.MaxRedemptions = input.MaxRedemptions
		saved.StartsAt = input.StartsAt
		saved.ExpiresAt = input.ExpiresAt
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

func UpdateDistributionPromoCodeStatus(userID int, promoCodeID int, status string) (*model.DistributionPromoCode, error) {
	if promoCodeID <= 0 {
		return nil, fmt.Errorf("invalid promo code id")
	}
	if err := ValidateDistributionStatus(status); err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var saved model.DistributionPromoCode
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var agent model.DistributionAgent
		if err := distributionLock(tx).Where("user_id = ? AND status = ?", userID, DistributionStatusEnabled).First(&agent).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ? AND agent_id = ?", promoCodeID, agent.Id).First(&saved).Error; err != nil {
			return err
		}
		saved.Status = status
		saved.UpdatedAt = now
		return tx.Save(&saved).Error
	})
	if err != nil {
		return nil, err
	}
	return &saved, nil
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
