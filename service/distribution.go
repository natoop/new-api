package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

type DistributionAgentSaveInput struct {
	Id            int    `json:"id"`
	UserId        int    `json:"user_id"`
	Name          string `json:"name"`
	Balance       int    `json:"balance"`
	CommissionBps int    `json:"commission_bps"`
	ParentAgentId int    `json:"parent_agent_id"`
	Contact       string `json:"contact"`
	Remark        string `json:"remark"`
}

type DistributionAgentProfile struct {
	Agent              *model.DistributionAgent `json:"agent"`
	AvailableInventory int64                    `json:"available_inventory"`
	AffCode            string                   `json:"aff_code"`
}

type DistributionPackageSaveInput struct {
	Name         string `json:"name"`
	Sku          string `json:"sku"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	AgentPrice   int    `json:"agent_price"`
	RetailPrice  int    `json:"retail_price"`
	CreditAmount int    `json:"credit_amount"`
	SortOrder    int    `json:"sort_order"`
}

type DistributionPriceConfigSaveInput struct {
	Id             int    `json:"id"`
	ScopeType      string `json:"scope_type"`
	PackageId      int    `json:"package_id"`
	Level          int    `json:"level"`
	ParentAgentId  int    `json:"parent_agent_id"`
	AgentId        int    `json:"agent_id"`
	UnitPrice      int    `json:"unit_price"`
	Tier1CostPrice int    `json:"tier1_cost_price"`
	Tier2CostPrice int    `json:"tier2_cost_price"`
	Status         string `json:"status"`
	Remark         string `json:"remark"`
}

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
	if input.Id > 0 && input.ParentAgentId == input.Id {
		return input, fmt.Errorf("parent_agent_id cannot point to the same agent")
	}
	return input, nil
}

func validateDistributionPackageInput(input DistributionPackageSaveInput) (DistributionPackageSaveInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Sku = strings.TrimSpace(input.Sku)
	input.Description = strings.TrimSpace(input.Description)
	input.Status = strings.TrimSpace(input.Status)
	if input.Status == "" {
		input.Status = DistributionStatusEnabled
	}
	if input.Name == "" {
		return input, fmt.Errorf("name cannot be empty")
	}
	if input.Sku == "" {
		return input, fmt.Errorf("sku cannot be empty")
	}
	if err := ValidateDistributionStatus(input.Status); err != nil {
		return input, err
	}
	if input.AgentPrice < 0 || input.RetailPrice < 0 || input.CreditAmount < 0 {
		return input, fmt.Errorf("amount fields cannot be negative")
	}
	return input, nil
}

func AdminSaveDistributionAgent(input DistributionAgentSaveInput) (*model.DistributionAgent, error) {
	input, err := validateDistributionAgentInput(input)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var saved model.DistributionAgent
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		var duplicate model.DistributionAgent
		err := tx.Where("user_id = ? AND id <> ?", input.UserId, input.Id).First(&duplicate).Error
		if err == nil {
			return fmt.Errorf("user_id already has a distribution agent")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if input.ParentAgentId > 0 {
			var parent model.DistributionAgent
			if err := tx.Where("id = ?", input.ParentAgentId).First(&parent).Error; err != nil {
				return fmt.Errorf("parent agent not found")
			}
		}
		if input.Id > 0 {
			if err := tx.Where("id = ?", input.Id).First(&saved).Error; err != nil {
				return err
			}
			saved.UserId = input.UserId
			saved.Name = input.Name
			saved.Balance = input.Balance
			saved.CommissionBps = input.CommissionBps
			saved.ParentAgentId = input.ParentAgentId
			saved.Contact = input.Contact
			saved.Remark = input.Remark
			saved.UpdatedAt = now
			if err := tx.Save(&saved).Error; err != nil {
				return err
			}
			if err := ensureDistributionAgentUserRole(tx, input.UserId); err != nil {
				return err
			}
			return syncDistributionLegacyInvitedCustomers(tx, &saved, input.UserId, now)
		}
		saved = model.DistributionAgent{
			UserId:        input.UserId,
			Name:          input.Name,
			Status:        DistributionStatusEnabled,
			Balance:       input.Balance,
			CommissionBps: input.CommissionBps,
			ParentAgentId: input.ParentAgentId,
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

func ensureDistributionAgentForUser(tx *gorm.DB, user *model.User) (*model.DistributionAgent, error) {
	if user == nil || user.Id <= 0 {
		return nil, fmt.Errorf("invalid user")
	}
	now := time.Now().Unix()
	name := strings.TrimSpace(user.DisplayName)
	if name == "" {
		name = strings.TrimSpace(user.Username)
	}
	if name == "" {
		name = fmt.Sprintf("user-%d", user.Id)
	}
	var agent model.DistributionAgent
	err := tx.Where("user_id = ?", user.Id).First(&agent).Error
	if err == nil {
		agent.Name = name
		agent.Status = DistributionStatusEnabled
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
		ParentAgentId: 0,
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
		saved, err := ensureDistributionAgentForUser(tx, user)
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
	err := query.Order("sort_order desc, id desc").Limit(pageSize).Offset(startIdx).Find(&packages).Error
	return packages, total, err
}

func AdminCreateDistributionPackage(input DistributionPackageSaveInput) (*model.DistributionPackage, error) {
	input, err := validateDistributionPackageInput(input)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var saved model.DistributionPackage
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		var duplicate model.DistributionPackage
		err := tx.Where("sku = ?", input.Sku).First(&duplicate).Error
		if err == nil {
			return fmt.Errorf("sku already exists")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		saved = model.DistributionPackage{
			Name:         input.Name,
			Sku:          input.Sku,
			Description:  input.Description,
			Status:       input.Status,
			AgentPrice:   input.AgentPrice,
			RetailPrice:  input.RetailPrice,
			CreditAmount: input.CreditAmount,
			SortOrder:    input.SortOrder,
			CreatedAt:    now,
			UpdatedAt:    now,
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
		var duplicate model.DistributionPackage
		err := tx.Where("sku = ? AND id <> ?", input.Sku, packageID).First(&duplicate).Error
		if err == nil {
			return fmt.Errorf("sku already exists")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := tx.Where("id = ?", packageID).First(&saved).Error; err != nil {
			return err
		}
		saved.Name = input.Name
		saved.Sku = input.Sku
		saved.Description = input.Description
		saved.Status = input.Status
		saved.AgentPrice = input.AgentPrice
		saved.RetailPrice = input.RetailPrice
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

func ListDistributionAgentPackages() ([]model.DistributionPackage, error) {
	var packages []model.DistributionPackage
	err := model.DB.Where("status = ?", DistributionStatusEnabled).
		Order("sort_order desc, id desc").
		Find(&packages).Error
	return packages, err
}

func validateDistributionPriceConfigInput(input DistributionPriceConfigSaveInput) (DistributionPriceConfigSaveInput, error) {
	input.ScopeType = strings.TrimSpace(input.ScopeType)
	input.Status = strings.TrimSpace(input.Status)
	input.Remark = strings.TrimSpace(input.Remark)
	if input.Status == "" {
		input.Status = DistributionStatusEnabled
	}
	switch input.ScopeType {
	case DistributionPriceScopeGlobal:
		input.AgentId = 0
		input.Level = 0
		input.ParentAgentId = 0
	case DistributionPriceScopeLevel:
		if input.Level < 0 {
			return input, fmt.Errorf("level cannot be negative")
		}
		input.AgentId = 0
	case DistributionPriceScopeAgent:
		if input.AgentId <= 0 {
			return input, fmt.Errorf("agent_id must be greater than 0")
		}
	default:
		return input, fmt.Errorf("invalid scope_type")
	}
	if input.PackageId <= 0 {
		return input, fmt.Errorf("package_id must be greater than 0")
	}
	if input.UnitPrice < 0 || input.Tier1CostPrice < 0 || input.Tier2CostPrice < 0 {
		return input, fmt.Errorf("price fields cannot be negative")
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
		saved.ScopeType = input.ScopeType
		saved.PackageId = input.PackageId
		saved.Level = input.Level
		saved.ParentAgentId = input.ParentAgentId
		saved.AgentId = input.AgentId
		saved.UnitPrice = input.UnitPrice
		saved.Tier1CostPrice = input.Tier1CostPrice
		saved.Tier2CostPrice = input.Tier2CostPrice
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
