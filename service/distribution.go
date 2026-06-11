package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type DistributionAgentSaveInput struct {
	Id            int    `json:"id"`
	UserId        int    `json:"user_id"`
	Name          string `json:"name"`
	Balance       int    `json:"balance"`
	CommissionBps int    `json:"commission_bps"`
	ParentAgentId int    `json:"parent_agent_id"`
	Level         int    `json:"level"`
	Contact       string `json:"contact"`
	Remark        string `json:"remark"`
}

type DistributionAgentProfile struct {
	Agent              *model.DistributionAgent `json:"agent"`
	AvailableInventory int64                    `json:"available_inventory"`
	AffCode            string                   `json:"aff_code"`
}

type DistributionPackageSaveInput struct {
	SubscriptionPlanId  int    `json:"subscription_plan_id"`
	Name                string `json:"name"`
	Sku                 string `json:"sku"`
	Description         string `json:"description"`
	Status              string `json:"status"`
	AgentPrice          int    `json:"agent_price"`
	RetailPrice         int    `json:"retail_price"`
	SecondaryAgentPrice int    `json:"secondary_agent_price"`
	CreditAmount        int    `json:"credit_amount"`
	SortOrder           int    `json:"sort_order"`
}

type DistributionPriceConfigSaveInput struct {
	Id             int    `json:"id"`
	PackageId      int    `json:"package_id"`
	TargetType     string `json:"target_type"`
	CustomerUserId int    `json:"customer_user_id"`
	AgentLevel     int    `json:"agent_level"`
	PriceType      string `json:"price_type"`
	PriceValue     int    `json:"price_value"`
	Status         string `json:"status"`
	Remark         string `json:"remark"`
}

const (
	errDistributionPackageSubscriptionExists = "distribution package subscription plan already exists"
	errDistributionPackageTierPriceOrder     = "tier 1 agent price must be less than or equal to tier 2 agent price"
	errDistributionPackagePriceTooHigh       = "agent prices must be less than or equal to subscription plan price"
)

func validateDistributionAgentInput(input DistributionAgentSaveInput) (DistributionAgentSaveInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Contact = strings.TrimSpace(input.Contact)
	input.Remark = strings.TrimSpace(input.Remark)
	if input.UserId <= 0 {
		return input, fmt.Errorf("user_id must be greater than 0")
	}
	if input.Name == "" {
		return input, fmt.Errorf("name cannot be empty")
	}
	if input.Balance < 0 {
		return input, fmt.Errorf("balance cannot be negative")
	}
	if input.CommissionBps < 0 || input.CommissionBps > DistributionMaxBPS {
		return input, ErrDistributionInvalidBPS
	}
	if input.ParentAgentId < 0 {
		return input, fmt.Errorf("parent_agent_id cannot be negative")
	}
	if input.Level == 0 {
		input.Level = DistributionAgentLevelSecondary
	}
	if input.Level != DistributionAgentLevelPrimary && input.Level != DistributionAgentLevelSecondary {
		return input, fmt.Errorf("level must be 1 or 2")
	}
	if input.Level == DistributionAgentLevelPrimary {
		input.ParentAgentId = 0
	}
	if input.Id > 0 && input.ParentAgentId == input.Id {
		return input, fmt.Errorf("parent_agent_id cannot point to the same agent")
	}
	return input, nil
}

func normalizeDistributionAgentHierarchy(tx *gorm.DB, currentAgentID int, level int, parentAgentID int) (int, int, error) {
	if level == 0 {
		level = DistributionAgentLevelSecondary
	}
	if level != DistributionAgentLevelPrimary && level != DistributionAgentLevelSecondary {
		return level, parentAgentID, fmt.Errorf("level must be 1 or 2")
	}
	if level == DistributionAgentLevelPrimary {
		return level, 0, nil
	}
	if parentAgentID <= 0 {
		return level, 0, nil
	}
	if currentAgentID > 0 && parentAgentID == currentAgentID {
		return level, parentAgentID, fmt.Errorf("parent_agent_id cannot point to the same agent")
	}
	var parent model.DistributionAgent
	if err := tx.Select("id, level").Where("id = ?", parentAgentID).First(&parent).Error; err != nil {
		return level, parentAgentID, fmt.Errorf("parent agent not found")
	}
	if parent.Level != DistributionAgentLevelPrimary {
		return level, parentAgentID, fmt.Errorf("parent agent must be level 1")
	}
	return level, parentAgentID, nil
}

func distributionPromotionParentAgentID(tx *gorm.DB, inviterUserID int) (int, error) {
	if inviterUserID <= 0 {
		return 0, nil
	}
	var parent model.DistributionAgent
	err := tx.Select("id, level").Where("user_id = ? AND status = ?", inviterUserID, DistributionStatusEnabled).First(&parent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if parent.Level != DistributionAgentLevelPrimary {
		return 0, nil
	}
	return parent.Id, nil
}

func validateDistributionPackageInput(input DistributionPackageSaveInput) (DistributionPackageSaveInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Sku = strings.TrimSpace(input.Sku)
	input.Description = strings.TrimSpace(input.Description)
	input.Status = strings.TrimSpace(input.Status)
	if input.Status == "" {
		input.Status = DistributionStatusEnabled
	}
	if input.SubscriptionPlanId <= 0 {
		return input, fmt.Errorf("subscription_plan_id must be greater than 0")
	}
	if err := ValidateDistributionStatus(input.Status); err != nil {
		return input, err
	}
	if input.AgentPrice < 0 || input.SecondaryAgentPrice < 0 {
		return input, fmt.Errorf("agent prices cannot be negative")
	}
	return input, nil
}

func distributionSubscriptionPlanPriceCents(priceAmount float64) int {
	if priceAmount <= 0 {
		return 0
	}
	return int(decimal.NewFromFloat(priceAmount).Mul(decimal.NewFromInt(100)).Round(0).IntPart())
}

func calcDistributionPaymentAmountFromUSDCents(amountCents int) (int, error) {
	if amountCents <= 0 {
		return 0, nil
	}
	if common.QuotaPerUnit <= 0 {
		return 0, fmt.Errorf("quota unit config is invalid")
	}
	return int(decimal.NewFromInt(int64(amountCents)).
		Div(decimal.NewFromInt(100)).
		Mul(decimal.NewFromFloat(common.QuotaPerUnit)).
		Ceil().
		IntPart()), nil
}

func hydrateDistributionPackageFromSubscriptionPlan(tx *gorm.DB, input *DistributionPackageSaveInput) (*model.SubscriptionPlan, error) {
	if input == nil || input.SubscriptionPlanId <= 0 {
		return nil, fmt.Errorf("subscription_plan_id must be greater than 0")
	}
	var plan model.SubscriptionPlan
	if err := tx.Where("id = ? AND enabled = ?", input.SubscriptionPlanId, true).First(&plan).Error; err != nil {
		return nil, fmt.Errorf("subscription plan not found or disabled")
	}
	plan.NormalizeDefaults()
	input.Name = strings.TrimSpace(plan.Title)
	input.Description = strings.TrimSpace(plan.Subtitle)
	input.Sku = fmt.Sprintf("subscription_plan_%d", plan.Id)
	input.RetailPrice = distributionSubscriptionPlanPriceCents(plan.PriceAmount)
	input.CreditAmount = int(plan.TotalAmount)
	return &plan, nil
}

func validateDistributionPackagePrices(input DistributionPackageSaveInput) error {
	if input.AgentPrice > input.SecondaryAgentPrice {
		return errors.New(errDistributionPackageTierPriceOrder)
	}
	if input.AgentPrice > input.RetailPrice || input.SecondaryAgentPrice > input.RetailPrice {
		return errors.New(errDistributionPackagePriceTooHigh)
	}
	return nil
}

func hydrateDistributionPackagesFromSubscriptionPlans(packages []model.DistributionPackage) error {
	planIDs := make([]int, 0, len(packages))
	seen := map[int]struct{}{}
	for _, distributionPackage := range packages {
		if distributionPackage.SubscriptionPlanId <= 0 {
			continue
		}
		if _, ok := seen[distributionPackage.SubscriptionPlanId]; ok {
			continue
		}
		seen[distributionPackage.SubscriptionPlanId] = struct{}{}
		planIDs = append(planIDs, distributionPackage.SubscriptionPlanId)
	}
	if len(planIDs) == 0 {
		return nil
	}
	var plans []model.SubscriptionPlan
	if err := model.DB.Where("id IN ?", planIDs).Find(&plans).Error; err != nil {
		return err
	}
	planMap := make(map[int]model.SubscriptionPlan, len(plans))
	for _, plan := range plans {
		plan.NormalizeDefaults()
		planMap[plan.Id] = plan
	}
	for i := range packages {
		plan, ok := planMap[packages[i].SubscriptionPlanId]
		if !ok {
			continue
		}
		packages[i].SubscriptionTitle = plan.Title
		packages[i].SubscriptionSubtitle = plan.Subtitle
		packages[i].Name = strings.TrimSpace(plan.Title)
		packages[i].Description = strings.TrimSpace(plan.Subtitle)
		packages[i].RetailPrice = distributionSubscriptionPlanPriceCents(plan.PriceAmount)
		packages[i].CreditAmount = int(plan.TotalAmount)
		// 展示层同步：套餐改价后，一/二级代理价若高于当前零售价则钳到零售
		// 价，与下单时的口径一致，避免代理中心看到倒挂的旧价。
		if packages[i].RetailPrice > 0 {
			if packages[i].AgentPrice > packages[i].RetailPrice {
				packages[i].AgentPrice = packages[i].RetailPrice
			}
			if packages[i].SecondaryAgentPrice > packages[i].RetailPrice {
				packages[i].SecondaryAgentPrice = packages[i].RetailPrice
			}
		}
	}
	return nil
}

func AdminSaveDistributionAgent(input DistributionAgentSaveInput) (*model.DistributionAgent, error) {
	now := time.Now().Unix()
	var saved model.DistributionAgent
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		if input.Id > 0 {
			if err := tx.Where("id = ?", input.Id).First(&saved).Error; err != nil {
				return err
			}
			if input.UserId > 0 && input.UserId != saved.UserId {
				return fmt.Errorf("user_id cannot be changed")
			}
			if input.Level == 0 {
				input.Level = saved.Level
			}
			level, parentAgentID, err := normalizeDistributionAgentHierarchy(tx, input.Id, input.Level, input.ParentAgentId)
			if err != nil {
				return err
			}
			saved.ParentAgentId = parentAgentID
			saved.Level = level
			saved.UpdatedAt = now
			if err := tx.Save(&saved).Error; err != nil {
				return err
			}
			return nil
		}
		input.Level = DistributionAgentLevelSecondary
		input.ParentAgentId = 0
		input, err := validateDistributionAgentInput(input)
		if err != nil {
			return err
		}
		var duplicate model.DistributionAgent
		err = tx.Where("user_id = ? AND id <> ?", input.UserId, input.Id).First(&duplicate).Error
		if err == nil {
			return fmt.Errorf("user_id already has a distribution agent")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		level, parentAgentID, err := normalizeDistributionAgentHierarchy(tx, 0, input.Level, input.ParentAgentId)
		if err != nil {
			return err
		}
		saved = model.DistributionAgent{
			UserId:        input.UserId,
			Name:          input.Name,
			Status:        DistributionStatusEnabled,
			Balance:       input.Balance,
			CommissionBps: input.CommissionBps,
			ParentAgentId: parentAgentID,
			Level:         level,
			Contact:       input.Contact,
			Remark:        input.Remark,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := tx.Create(&saved).Error; err != nil {
			return err
		}
		if err := ensureDistributionAgentUserRole(tx, input.UserId); err != nil {
			return err
		}
		return syncDistributionLegacyInvitedCustomers(tx, &saved, input.UserId, now)
	})
	if err != nil {
		return nil, err
	}
	return &saved, nil
}

func ensureDistributionAgentUserRole(tx *gorm.DB, userID int) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user id")
	}
	return tx.Model(&model.User{}).
		Where("id = ? AND role < ?", userID, common.RoleAdminUser).
		Update("role", common.RoleAgentUser).Error
}

func syncDistributionLegacyInvitedCustomers(tx *gorm.DB, agent *model.DistributionAgent, inviterUserID int, now int64) error {
	if tx == nil || agent == nil || agent.Id <= 0 || inviterUserID <= 0 {
		return fmt.Errorf("invalid distribution agent sync input")
	}
	var inviter model.User
	if err := tx.Select("id, aff_code").Where("id = ?", inviterUserID).First(&inviter).Error; err != nil {
		return err
	}
	sourceNo := strings.TrimSpace(inviter.AffCode)
	if sourceNo == "" {
		sourceNo = fmt.Sprintf("aff_user_%d", inviterUserID)
	}
	var invitees []model.User
	if err := tx.Select("id").Where("inviter_id = ? AND id <> ?", inviterUserID, inviterUserID).Find(&invitees).Error; err != nil {
		return err
	}
	for _, invitee := range invitees {
		var ownership model.DistributionCustomerOwnership
		err := tx.Select("id").Where("customer_user_id = ?", invitee.Id).First(&ownership).Error
		if err == nil {
			continue
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := bindDistributionCustomer(tx, invitee.Id, agent.Id, DistributionCustomerEventBind, DistributionSourceInvitation, inviterUserID, sourceNo, 0, "legacy invitation sync", now); err != nil {
			return err
		}
	}
	return nil
}

func ensureDistributionAgentForUser(tx *gorm.DB, user *model.User, level int, parentAgentID int) (*model.DistributionAgent, error) {
	if user == nil || user.Id <= 0 {
		return nil, fmt.Errorf("invalid user")
	}
	now := time.Now().Unix()
	level, parentAgentID, err := normalizeDistributionAgentHierarchy(tx, 0, level, parentAgentID)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(user.DisplayName)
	if name == "" {
		name = strings.TrimSpace(user.Username)
	}
	if name == "" {
		name = fmt.Sprintf("user-%d", user.Id)
	}
	var agent model.DistributionAgent
	err = tx.Where("user_id = ?", user.Id).First(&agent).Error
	if err == nil {
		level, parentAgentID, err = normalizeDistributionAgentHierarchy(tx, agent.Id, level, parentAgentID)
		if err != nil {
			return nil, err
		}
		agent.Name = name
		agent.Status = DistributionStatusEnabled
		agent.Level = level
		agent.ParentAgentId = parentAgentID
		agent.UpdatedAt = now
		if err := tx.Save(&agent).Error; err != nil {
			return nil, err
		}
		if err := ensureDistributionAgentUserRole(tx, user.Id); err != nil {
			return nil, err
		}
		if err := syncDistributionLegacyInvitedCustomers(tx, &agent, user.Id, now); err != nil {
			return nil, err
		}
		return &agent, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	agent = model.DistributionAgent{
		UserId:        user.Id,
		Name:          name,
		Status:        DistributionStatusEnabled,
		Balance:       0,
		CommissionBps: 0,
		ParentAgentId: parentAgentID,
		Level:         level,
		Contact:       "",
		Remark:        "",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := tx.Create(&agent).Error; err != nil {
		return nil, err
	}
	if err := ensureDistributionAgentUserRole(tx, user.Id); err != nil {
		return nil, err
	}
	if err := syncDistributionLegacyInvitedCustomers(tx, &agent, user.Id, now); err != nil {
		return nil, err
	}
	return &agent, nil
}

func AdminEnsureDistributionAgentForUser(user *model.User) (*model.DistributionAgent, error) {
	var agent *model.DistributionAgent
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		parentAgentID, err := distributionPromotionParentAgentID(tx, user.InviterId)
		if err != nil {
			return err
		}
		saved, err := ensureDistributionAgentForUser(tx, user, DistributionAgentLevelSecondary, parentAgentID)
		if err != nil {
			return err
		}
		agent = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func AdminCreateDistributionAgentForUser(user *model.User) (*model.DistributionAgent, error) {
	var agent *model.DistributionAgent
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		saved, err := ensureDistributionAgentForUser(tx, user, DistributionAgentLevelSecondary, 0)
		if err != nil {
			return err
		}
		agent = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func AdminUpdateDistributionAgentStatus(agentID int, status string) (*model.DistributionAgent, error) {
	if agentID <= 0 {
		return nil, fmt.Errorf("invalid agent id")
	}
	status = strings.TrimSpace(status)
	if err := ValidateDistributionStatus(status); err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var agent model.DistributionAgent
	if err := model.DB.Where("id = ?", agentID).First(&agent).Error; err != nil {
		return nil, err
	}
	agent.Status = status
	agent.UpdatedAt = now
	if err := model.DB.Save(&agent).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func hydrateDistributionAgentUsers(agents []model.DistributionAgent) error {
	if len(agents) == 0 {
		return nil
	}
	userIDs := make([]int, 0, len(agents))
	for _, agent := range agents {
		if agent.UserId > 0 {
			userIDs = append(userIDs, agent.UserId)
		}
	}
	var users []model.User
	if err := model.DB.Select("id, username, display_name, email").Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return err
	}
	userMap := make(map[int]model.User, len(users))
	for _, user := range users {
		userMap[user.Id] = user
	}
	for i := range agents {
		if user, ok := userMap[agents[i].UserId]; ok {
			agents[i].Username = user.Username
			agents[i].DisplayName = user.DisplayName
			agents[i].Email = user.Email
		}
	}
	return nil
}

func AdminListDistributionAgents(keyword string, startIdx int, pageSize int) ([]model.DistributionAgent, int64, error) {
	var agents []model.DistributionAgent
	query := model.DB.Model(&model.DistributionAgent{})
	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Joins("LEFT JOIN users ON users.id = p3_agents.user_id").
			Where("p3_agents.name LIKE ? OR users.username LIKE ? OR users.display_name LIKE ? OR users.email LIKE ?", like, like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("p3_agents.id desc").Limit(pageSize).Offset(startIdx).Find(&agents).Error; err != nil {
		return nil, 0, err
	}
	if err := hydrateDistributionAgentUsers(agents); err != nil {
		return nil, 0, err
	}
	return agents, total, nil
}

func GetEnabledDistributionAgentByUserID(userID int) (*model.DistributionAgent, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	var agent model.DistributionAgent
	if err := model.DB.Where("user_id = ? AND status = ?", userID, DistributionStatusEnabled).First(&agent).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func GetDistributionAgentProfile(userID int) (*DistributionAgentProfile, error) {
	agent, err := GetEnabledDistributionAgentByUserID(userID)
	if err != nil {
		return nil, err
	}
	var user model.User
	if err := model.DB.Select("id, aff_code").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	var availableInventory int64
	if err := model.DB.Model(&model.DistributionInventory{}).
		Where("agent_id = ? AND status = ?", agent.Id, DistributionInventoryStatusAvailable).
		Count(&availableInventory).Error; err != nil {
		return nil, err
	}
	return &DistributionAgentProfile{
		Agent:              agent,
		AvailableInventory: availableInventory,
		AffCode:            user.AffCode,
	}, nil
}

func AdminListDistributionPackages(startIdx int, pageSize int) ([]model.DistributionPackage, int64, error) {
	var packages []model.DistributionPackage
	query := model.DB.Model(&model.DistributionPackage{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("sort_order desc, id desc").Limit(pageSize).Offset(startIdx).Find(&packages).Error; err != nil {
		return nil, 0, err
	}
	if err := hydrateDistributionPackagesFromSubscriptionPlans(packages); err != nil {
		return nil, 0, err
	}
	return packages, total, nil
}

func AdminCreateDistributionPackage(input DistributionPackageSaveInput) (*model.DistributionPackage, error) {
	input, err := validateDistributionPackageInput(input)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var saved model.DistributionPackage
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		plan, err := hydrateDistributionPackageFromSubscriptionPlan(tx, &input)
		if err != nil {
			return err
		}
		if err := validateDistributionPackagePrices(input); err != nil {
			return err
		}
		var duplicate model.DistributionPackage
		err = tx.Where("subscription_plan_id = ?", input.SubscriptionPlanId).First(&duplicate).Error
		if err == nil {
			return errors.New(errDistributionPackageSubscriptionExists)
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		err = tx.Where("sku = ?", input.Sku).First(&duplicate).Error
		if err == nil {
			return fmt.Errorf("sku already exists")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		saved = model.DistributionPackage{
			SubscriptionPlanId:   plan.Id,
			SubscriptionTitle:    plan.Title,
			SubscriptionSubtitle: plan.Subtitle,
			Name:                 input.Name,
			Sku:                  input.Sku,
			Description:          input.Description,
			Status:               input.Status,
			AgentPrice:           input.AgentPrice,
			RetailPrice:          input.RetailPrice,
			SecondaryAgentPrice:  input.SecondaryAgentPrice,
			CreditAmount:         input.CreditAmount,
			SortOrder:            input.SortOrder,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		return tx.Create(&saved).Error
	})
	if err != nil {
		return nil, err
	}
	return &saved, nil
}

func AdminUpdateDistributionPackage(packageID int, input DistributionPackageSaveInput) (*model.DistributionPackage, error) {
	if packageID <= 0 {
		return nil, fmt.Errorf("invalid package id")
	}
	input, err := validateDistributionPackageInput(input)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var saved model.DistributionPackage
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		plan, err := hydrateDistributionPackageFromSubscriptionPlan(tx, &input)
		if err != nil {
			return err
		}
		if err := validateDistributionPackagePrices(input); err != nil {
			return err
		}
		var duplicate model.DistributionPackage
		err = tx.Where("subscription_plan_id = ? AND id <> ?", input.SubscriptionPlanId, packageID).First(&duplicate).Error
		if err == nil {
			return errors.New(errDistributionPackageSubscriptionExists)
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		err = tx.Where("sku = ? AND id <> ?", input.Sku, packageID).First(&duplicate).Error
		if err == nil {
			return fmt.Errorf("sku already exists")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := tx.Where("id = ?", packageID).First(&saved).Error; err != nil {
			return err
		}
		saved.SubscriptionPlanId = plan.Id
		saved.SubscriptionTitle = plan.Title
		saved.SubscriptionSubtitle = plan.Subtitle
		saved.Name = input.Name
		saved.Sku = input.Sku
		saved.Description = input.Description
		saved.Status = input.Status
		saved.AgentPrice = input.AgentPrice
		saved.RetailPrice = input.RetailPrice
		saved.SecondaryAgentPrice = input.SecondaryAgentPrice
		saved.CreditAmount = input.CreditAmount
		saved.SortOrder = input.SortOrder
		saved.UpdatedAt = now
		return tx.Save(&saved).Error
	})
	if err != nil {
		return nil, err
	}
	return &saved, nil
}

func AdminUpdateDistributionPackageStatus(packageID int, status string) (*model.DistributionPackage, error) {
	if packageID <= 0 {
		return nil, fmt.Errorf("invalid package id")
	}
	status = strings.TrimSpace(status)
	if err := ValidateDistributionStatus(status); err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var distributionPackage model.DistributionPackage
	if err := model.DB.Where("id = ?", packageID).First(&distributionPackage).Error; err != nil {
		return nil, err
	}
	distributionPackage.Status = status
	distributionPackage.UpdatedAt = now
	if err := model.DB.Save(&distributionPackage).Error; err != nil {
		return nil, err
	}
	return &distributionPackage, nil
}

func ListDistributionAgentPackages(startIdx int, pageSize int) ([]model.DistributionPackage, int64, error) {
	var packages []model.DistributionPackage
	query := model.DB.Model(&model.DistributionPackage{}).Where("status = ?", DistributionStatusEnabled)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("sort_order desc, id desc").Limit(pageSize).Offset(startIdx).Find(&packages).Error; err != nil {
		return nil, 0, err
	}
	if err := hydrateDistributionPackagesFromSubscriptionPlans(packages); err != nil {
		return nil, 0, err
	}
	return packages, total, nil
}

func validateDistributionPriceConfigInput(input DistributionPriceConfigSaveInput) (DistributionPriceConfigSaveInput, error) {
	input.TargetType = strings.TrimSpace(input.TargetType)
	input.PriceType = strings.TrimSpace(input.PriceType)
	input.Status = strings.TrimSpace(input.Status)
	input.Remark = strings.TrimSpace(input.Remark)
	if input.Status == "" {
		input.Status = DistributionStatusEnabled
	}
	switch input.TargetType {
	case DistributionPriceTargetLevel:
		if input.AgentLevel != 1 && input.AgentLevel != 2 {
			return input, fmt.Errorf("agent_level must be 1 or 2")
		}
		input.CustomerUserId = 0
	case DistributionPriceTargetCustomer:
		if input.CustomerUserId <= 0 {
			return input, fmt.Errorf("customer_user_id must be greater than 0")
		}
		input.AgentLevel = 0
	default:
		return input, fmt.Errorf("invalid target_type")
	}
	if input.PackageId <= 0 {
		return input, fmt.Errorf("package_id must be greater than 0")
	}
	switch input.PriceType {
	case DistributionPriceTypeFixed:
		if input.PriceValue < 0 {
			return input, fmt.Errorf("price_value cannot be negative")
		}
	case DistributionPriceTypeDiscount:
		if input.PriceValue < 1 || input.PriceValue > 10 {
			return input, fmt.Errorf("discount price_value must be between 1 and 10")
		}
	default:
		return input, fmt.Errorf("invalid price_type")
	}
	if err := ValidateDistributionStatus(input.Status); err != nil {
		return input, err
	}
	return input, nil
}

func AdminListDistributionPriceConfigs(startIdx int, pageSize int) ([]model.DistributionPriceConfig, int64, error) {
	var configs []model.DistributionPriceConfig
	query := model.DB.Model(&model.DistributionPriceConfig{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id desc").Limit(pageSize).Offset(startIdx).Find(&configs).Error
	return configs, total, err
}

func AdminSaveDistributionPriceConfig(input DistributionPriceConfigSaveInput, operatorUserID int) (*model.DistributionPriceConfig, error) {
	input, err := validateDistributionPriceConfigInput(input)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var saved model.DistributionPriceConfig
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		if input.Id > 0 {
			if err := tx.Where("id = ?", input.Id).First(&saved).Error; err != nil {
				return err
			}
		}
		saved.PackageId = input.PackageId
		saved.TargetType = input.TargetType
		saved.CustomerUserId = input.CustomerUserId
		saved.AgentLevel = input.AgentLevel
		saved.PriceType = input.PriceType
		saved.PriceValue = input.PriceValue
		saved.Status = input.Status
		saved.Remark = input.Remark
		saved.UpdatedByUserId = operatorUserID
		saved.UpdatedAt = now
		if input.Id > 0 {
			return tx.Save(&saved).Error
		}
		saved.CreatedByUserId = operatorUserID
		saved.CreatedAt = now
		return tx.Create(&saved).Error
	})
	if err != nil {
		return nil, err
	}
	return &saved, nil
}

func AdminUpdateDistributionPriceConfigStatus(configID int, status string, operatorUserID int) (*model.DistributionPriceConfig, error) {
	if configID <= 0 {
		return nil, fmt.Errorf("invalid config id")
	}
	if err := ValidateDistributionStatus(status); err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var config model.DistributionPriceConfig
	if err := model.DB.Where("id = ?", configID).First(&config).Error; err != nil {
		return nil, err
	}
	config.Status = status
	config.UpdatedByUserId = operatorUserID
	config.UpdatedAt = now
	if err := model.DB.Save(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}
