package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DistributionPromoCodeSaveInput struct {
	Id             int    `json:"id"`
	PackageId      int    `json:"package_id"`
	Code           string `json:"code"`
	Status         string `json:"status"`
	DiscountType   string `json:"discount_type"`
	DiscountValue  int    `json:"discount_value"`
	MaxRedemptions int    `json:"max_redemptions"`
	StartsAt       int64  `json:"starts_at"`
	ExpiresAt      int64  `json:"expires_at"`
}

type DistributionPromoCodeListInput struct {
	TimeFilter  string
	UsageFilter string
	StartIdx    int
	PageSize    int
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
	input.DiscountType = DistributionDiscountTypeAmount
	if input.Status == "" {
		input.Status = DistributionStatusEnabled
	}
	if input.PackageId <= 0 {
		return input, fmt.Errorf("package_id must be greater than 0")
	}
	if err := ValidateDistributionStatus(input.Status); err != nil {
		return input, err
	}
	if input.DiscountValue < 0 || input.DiscountValue > 2000 {
		return input, fmt.Errorf("discount_value must be between 0 and 2000")
	}
	if input.MaxRedemptions < 0 {
		return input, fmt.Errorf("max_redemptions cannot be negative")
	}
	if err := ValidateDistributionTimeWindow(input.StartsAt, input.ExpiresAt); err != nil {
		return input, err
	}
	return input, nil
}

func buildDistributionPromoCode(tx *gorm.DB) (string, error) {
	for range 10 {
		code := strings.ReplaceAll(uuid.NewString(), "-", "")[:8]
		var count int64
		if err := tx.Model(&model.DistributionPromoCode{}).Where("code = ?", code).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return code, nil
		}
	}
	return "", fmt.Errorf("failed to generate promo code")
}

func hydrateDistributionPromoCodePackages(codes []model.DistributionPromoCode) error {
	if len(codes) == 0 {
		return nil
	}
	packageIDs := make([]int, 0, len(codes))
	seen := map[int]struct{}{}
	for _, code := range codes {
		if code.PackageId <= 0 {
			continue
		}
		if _, ok := seen[code.PackageId]; ok {
			continue
		}
		seen[code.PackageId] = struct{}{}
		packageIDs = append(packageIDs, code.PackageId)
	}
	if len(packageIDs) == 0 {
		return nil
	}
	var packages []model.DistributionPackage
	if err := model.DB.Select("id, name").Where("id IN ?", packageIDs).Find(&packages).Error; err != nil {
		return err
	}
	packageMap := make(map[int]string, len(packages))
	for _, distributionPackage := range packages {
		packageMap[distributionPackage.Id] = distributionPackage.Name
	}
	for i := range codes {
		codes[i].PackageName = packageMap[codes[i].PackageId]
	}
	return nil
}

func ListDistributionPromoCodes(userID int, input DistributionPromoCodeListInput) ([]model.DistributionPromoCode, int64, error) {
	agent, err := GetEnabledDistributionAgentByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	input.TimeFilter = strings.TrimSpace(input.TimeFilter)
	input.UsageFilter = strings.TrimSpace(input.UsageFilter)
	var codes []model.DistributionPromoCode
	query := model.DB.Model(&model.DistributionPromoCode{}).Where("agent_id = ?", agent.Id)
	now := time.Now().Unix()
	switch input.TimeFilter {
	case "active":
		query = query.Where("expires_at = ? OR expires_at >= ?", 0, now)
	case "expired":
		query = query.Where("expires_at > ? AND expires_at < ?", 0, now)
	}
	switch input.UsageFilter {
	case "used":
		query = query.Where("used_count > ?", 0)
	case "unused":
		query = query.Where("used_count = ?", 0)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id desc").Limit(input.PageSize).Offset(input.StartIdx).Find(&codes).Error; err != nil {
		return nil, 0, err
	}
	if err := hydrateDistributionPromoCodePackages(codes); err != nil {
		return nil, 0, err
	}
	return codes, total, nil
}

func ConsumeDistributionPromoCode(tx *gorm.DB, code string, packageID int, now int64) (*model.DistributionPromoCode, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, fmt.Errorf("promo code cannot be empty")
	}
	if packageID <= 0 {
		return nil, fmt.Errorf("package_id must be greater than 0")
	}
	if now <= 0 {
		now = time.Now().Unix()
	}
	if tx == nil {
		tx = model.DB
	}
	var promoCode model.DistributionPromoCode
	if err := distributionLock(tx).
		Where("code = ? AND package_id = ? AND status = ?", code, packageID, DistributionStatusEnabled).
		First(&promoCode).Error; err != nil {
		return nil, err
	}
	if promoCode.StartsAt > 0 && promoCode.StartsAt > now {
		return nil, fmt.Errorf("promo code is not active")
	}
	if promoCode.ExpiresAt > 0 && promoCode.ExpiresAt < now {
		return nil, fmt.Errorf("promo code is expired")
	}
	if promoCode.MaxRedemptions > 0 && promoCode.UsedCount >= promoCode.MaxRedemptions {
		return nil, fmt.Errorf("promo code usage limit exceeded")
	}
	res := tx.Model(&model.DistributionPromoCode{}).
		Where("id = ?", promoCode.Id).
		Where("max_redemptions = ? OR used_count < max_redemptions", 0).
		Updates(map[string]any{"used_count": gorm.Expr("used_count + ?", 1), "updated_at": now})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("promo code usage limit exceeded")
	}
	promoCode.UsedCount++
	promoCode.UpdatedAt = now
	return &promoCode, nil
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
		var packageCount int64
		if err := tx.Model(&model.DistributionInventory{}).
			Where("agent_id = ? AND package_id = ?", agent.Id, input.PackageId).
			Where("status <> ? AND status <> ?", DistributionInventoryStatusRefunded, DistributionInventoryStatusVoided).
			Count(&packageCount).Error; err != nil {
			return err
		}
		if packageCount == 0 {
			return fmt.Errorf("package is not in agent inventory")
		}
		if input.Code == "" {
			code, err := buildDistributionPromoCode(tx)
			if err != nil {
				return err
			}
			input.Code = code
		}
		if input.Id > 0 {
			if err := tx.Where("id = ? AND agent_id = ?", input.Id, agent.Id).First(&saved).Error; err != nil {
				return err
			}
			if input.MaxRedemptions > 0 && input.MaxRedemptions < saved.UsedCount {
				return fmt.Errorf("max_redemptions cannot be less than used_count")
			}
		}
		saved.AgentId = agent.Id
		saved.PackageId = input.PackageId
		saved.Code = input.Code
		saved.Status = input.Status
		saved.DiscountType = input.DiscountType
		saved.DiscountValue = input.DiscountValue
		saved.MaxRedemptions = input.MaxRedemptions
		saved.StartsAt = input.StartsAt
		if saved.StartsAt == 0 {
			saved.StartsAt = now
		}
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
