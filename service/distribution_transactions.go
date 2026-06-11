package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DistributionBalanceAdjustmentInput struct {
	Delta          int    `json:"delta"`
	Description    string `json:"description"`
	IdempotencyKey string `json:"idempotency_key"`
}

type DistributionPurchaseResult struct {
	Order     *model.DistributionOrder     `json:"order"`
	Inventory *model.DistributionInventory `json:"inventory"`
}

type DistributionInventoryAssignResult struct {
	Inventory *model.DistributionInventory `json:"inventory"`
}

type DistributionInventoryRefundResult struct {
	Inventory *model.DistributionInventory     `json:"inventory"`
	Ledger    *model.DistributionBalanceLedger `json:"ledger"`
}

type DistributionInventoryListInput struct {
	Keyword  string
	Status   string
	StartIdx int
	PageSize int
}

type DistributionInventoryPackageOption struct {
	PackageId   int    `json:"package_id"`
	PackageName string `json:"package_name"`
	// RetailPrice 为套餐"当前"零售价（分），随订阅套餐改价实时刷新；
	// 前端用它作为优惠码抵扣额上限，保持与套餐定价同步。
	RetailPrice int `json:"retail_price"`
}

func distributionLock(tx *gorm.DB) *gorm.DB {
	if common.UsingSQLite {
		return tx
	}
	return tx.Set("gorm:query_option", "FOR UPDATE")
}

func createDistributionLedger(tx *gorm.DB, ledger model.DistributionBalanceLedger) error {
	if ledger.LedgerNo == "" {
		ledger.LedgerNo = BuildLedgerNo(ledger.AgentId, ledger.SourceType, ledger.SourceNo)
	}
	return tx.Create(&ledger).Error
}

func buildDistributionInventoryNo(tx *gorm.DB) (string, error) {
	for range 10 {
		code := uuid.NewString()[:8]
		var count int64
		if err := tx.Model(&model.DistributionInventory{}).Where("inventory_no = ?", code).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return code, nil
		}
	}
	return "", fmt.Errorf("failed to generate inventory code")
}

func getDistributionOrderWithInventoryByOrderNo(tx *gorm.DB, orderNo string) (*model.DistributionOrder, *model.DistributionInventory, error) {
	var order model.DistributionOrder
	if err := tx.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return nil, nil, err
	}
	var inventory model.DistributionInventory
	err := tx.Where("order_id = ?", order.Id).First(&inventory).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &order, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	return &order, &inventory, nil
}

func getDistributionOrderWithInventoryByIdempotency(tx *gorm.DB, agentID int, userID int, packageID int, key string) (*model.DistributionOrder, *model.DistributionInventory, error) {
	var order model.DistributionOrder
	if err := tx.Where("agent_id = ? AND user_id = ? AND package_id = ? AND idempotency_key = ?", agentID, userID, packageID, key).First(&order).Error; err != nil {
		return nil, nil, err
	}
	var inventory model.DistributionInventory
	err := tx.Where("order_id = ?", order.Id).First(&inventory).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &order, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	return &order, &inventory, nil
}

func verifyDistributionPurchaseIdempotency(order model.DistributionOrder, agentID int, userID int, packageID int, key string) error {
	if order.AgentId != agentID || order.UserId != userID || order.PackageId != packageID || order.IdempotencyKey != key {
		return fmt.Errorf("idempotency key conflicts")
	}
	return nil
}

func resolveDistributionAgentPrice(tx *gorm.DB, primaryAgentPrice int, secondaryAgentPrice int, agentID int) (int, error) {
	var agent model.DistributionAgent
	if err := distributionLock(tx).Select("id, level").Where("id = ?", agentID).First(&agent).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("agent profile not found")
		}
		return 0, err
	}
	if agent.Level == DistributionAgentLevelSecondary && secondaryAgentPrice > 0 {
		return secondaryAgentPrice, nil
	}
	return primaryAgentPrice, nil
}

func AdminAdjustDistributionAgentBalance(agentID int, operatorUserID int, input DistributionBalanceAdjustmentInput) (*model.DistributionBalanceAdjustment, error) {
	if agentID <= 0 {
		return nil, fmt.Errorf("invalid agent id")
	}
	key, err := NormalizeIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	referenceNo := BuildBalanceRef(agentID, key)
	now := time.Now().Unix()
	var adjustment model.DistributionBalanceAdjustment
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("reference_no = ?", referenceNo).First(&adjustment).Error
		if err == nil {
			if adjustment.AgentId != agentID || adjustment.Delta != input.Delta || adjustment.IdempotencyKey != key {
				return fmt.Errorf("idempotency key conflicts")
			}
			return nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var agent model.DistributionAgent
		if err := distributionLock(tx).Where("id = ?", agentID).First(&agent).Error; err != nil {
			return err
		}
		before := agent.Balance
		if !CanApplyDelta(before, input.Delta) {
			return fmt.Errorf("balance cannot be negative")
		}
		after := before + input.Delta
		query := tx.Model(&model.DistributionAgent{}).Where("id = ?", agentID)
		if input.Delta < 0 {
			query = query.Where("balance >= ?", -input.Delta)
		}
		res := query.Updates(map[string]any{"balance": gorm.Expr("balance + ?", input.Delta), "updated_at": now})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return fmt.Errorf("balance update failed")
		}
		adjustment = model.DistributionBalanceAdjustment{
			ReferenceNo:    referenceNo,
			IdempotencyKey: key,
			AgentId:        agentID,
			Delta:          input.Delta,
			BalanceBefore:  before,
			BalanceAfter:   after,
			Description:    input.Description,
			CreatedAt:      now,
		}
		if err := tx.Create(&adjustment).Error; err != nil {
			return err
		}
		entryType := DistributionLedgerEntryCredit
		if input.Delta < 0 {
			entryType = DistributionLedgerEntryDebit
		}
		return createDistributionLedger(tx, model.DistributionBalanceLedger{
			LedgerNo:       BuildLedgerNo(agentID, DistributionSourceAdjust, referenceNo),
			IdempotencyKey: key,
			AgentId:        agentID,
			OperatorUserId: operatorUserID,
			EntryType:      entryType,
			SourceType:     DistributionSourceAdjust,
			SourceId:       adjustment.Id,
			SourceNo:       referenceNo,
			Delta:          input.Delta,
			BalanceBefore:  before,
			BalanceAfter:   after,
			Description:    input.Description,
			CreatedAt:      now,
		})
	})
	if err != nil {
		return nil, err
	}
	return &adjustment, nil
}

func PurchaseDistributionPackage(userID int, packageID int, idempotencyKey string) (*DistributionPurchaseResult, error) {
	if packageID <= 0 {
		return nil, fmt.Errorf("invalid package id")
	}
	key, err := NormalizeIdempotencyKey(idempotencyKey)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var result DistributionPurchaseResult
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		var agent model.DistributionAgent
		if err := distributionLock(tx).Where("user_id = ? AND status = ?", userID, DistributionStatusEnabled).First(&agent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("agent profile not found")
			}
			return err
		}
		var buyer model.User
		if err := tx.Select("id, username, display_name, email").Where("id = ?", userID).First(&buyer).Error; err != nil {
			return err
		}
		var distributionPackage model.DistributionPackage
		if err := distributionLock(tx).Where("id = ? AND status = ?", packageID, DistributionStatusEnabled).First(&distributionPackage).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("package not found")
			}
			return err
		}
		if distributionPackage.SubscriptionPlanId > 0 {
			var plan model.SubscriptionPlan
			if err := tx.Where("id = ? AND enabled = ?", distributionPackage.SubscriptionPlanId, true).First(&plan).Error; err != nil {
				return fmt.Errorf("subscription plan not found or disabled")
			}
			plan.NormalizeDefaults()
			distributionPackage.SubscriptionTitle = plan.Title
			distributionPackage.SubscriptionSubtitle = plan.Subtitle
			distributionPackage.Name = strings.TrimSpace(plan.Title)
			distributionPackage.Description = strings.TrimSpace(plan.Subtitle)
			distributionPackage.RetailPrice = distributionSubscriptionPlanPriceCents(plan.PriceAmount)
			distributionPackage.CreditAmount = int(plan.TotalAmount)
		}
		existingOrder, existingInventory, err := getDistributionOrderWithInventoryByIdempotency(tx, agent.Id, userID, packageID, key)
		if err == nil {
			if err := verifyDistributionPurchaseIdempotency(*existingOrder, agent.Id, userID, packageID, key); err != nil {
				return err
			}
			result.Order = existingOrder
			result.Inventory = existingInventory
			return nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		unitPrice, err := resolveDistributionAgentPrice(tx, distributionPackage.AgentPrice, distributionPackage.SecondaryAgentPrice, agent.Id)
		if err != nil {
			return err
		}
		// 与套餐定价同步：零售价已在本事务内按订阅套餐实时刷新；若套餐降价
		// 导致存量代理价(一/二级)高于当前零售价，代理拿货价自动钳到零售价，
		// 避免"套餐改价后代理价不同步"的倒挂。
		if distributionPackage.RetailPrice > 0 && unitPrice > distributionPackage.RetailPrice {
			unitPrice = distributionPackage.RetailPrice
		}
		quantity := 1
		totalPriceUSDCents := unitPrice * quantity
		paidAmount, err := calcDistributionPaymentAmountFromUSDCents(totalPriceUSDCents)
		if err != nil {
			return err
		}
		originalAmount := distributionPackage.RetailPrice * quantity
		discountAmount := originalAmount - totalPriceUSDCents
		if discountAmount < 0 {
			discountAmount = 0
		}
		commissionAmount, err := CalcCommission(unitPrice, agent.CommissionBps)
		if err != nil {
			return err
		}
		orderNo := BuildPurchaseOrderNo(userID, packageID, key)
		order := model.DistributionOrder{
			OrderNo:               orderNo,
			IdempotencyKey:        key,
			AgentId:               agent.Id,
			UserId:                userID,
			BuyerUserId:           userID,
			BuyerUsername:         buyer.Username,
			BuyerDisplayName:      buyer.DisplayName,
			BuyerEmail:            buyer.Email,
			PackageId:             packageID,
			SubscriptionPlanId:    distributionPackage.SubscriptionPlanId,
			SubscriptionTitle:     distributionPackage.SubscriptionTitle,
			SubscriptionSubtitle:  distributionPackage.SubscriptionSubtitle,
			PackageName:           distributionPackage.Name,
			PackageSku:            distributionPackage.Sku,
			PackageDescription:    distributionPackage.Description,
			PackageCreditAmount:   distributionPackage.CreditAmount,
			OriginalAmount:        originalAmount,
			DiscountAmount:        discountAmount,
			CreditDeductionAmount: 0,
			PaidAmount:            paidAmount,
			Quantity:              quantity,
			UnitAgentPrice:        unitPrice,
			TotalAgentPrice:       totalPriceUSDCents,
			RetailPrice:           distributionPackage.RetailPrice,
			CommissionBps:         agent.CommissionBps,
			CommissionAmount:      commissionAmount,
			Status:                DistributionOrderStatusPending,
			CreatedAt:             now,
			UpdatedAt:             now,
		}
		createOrder := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&order)
		if createOrder.Error != nil {
			return createOrder.Error
		}
		if createOrder.RowsAffected == 0 {
			existingOrder, existingInventory, err := getDistributionOrderWithInventoryByOrderNo(tx, orderNo)
			if err != nil {
				return err
			}
			if err := verifyDistributionPurchaseIdempotency(*existingOrder, agent.Id, userID, packageID, key); err != nil {
				return err
			}
			result.Order = existingOrder
			result.Inventory = existingInventory
			return nil
		}
		before := agent.Balance
		if before < paidAmount {
			return fmt.Errorf("insufficient balance")
		}
		after := before - paidAmount
		res := tx.Model(&model.DistributionAgent{}).
			Where("id = ? AND balance >= ?", agent.Id, paidAmount).
			Updates(map[string]any{"balance": gorm.Expr("balance - ?", paidAmount), "updated_at": now})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return fmt.Errorf("insufficient balance")
		}
		order.Status = DistributionOrderStatusFulfilled
		order.PaidAt = now
		order.FulfilledAt = now
		order.UpdatedAt = now
		inventoryNo, err := buildDistributionInventoryNo(tx)
		if err != nil {
			return err
		}
		order.RedeemCode = inventoryNo
		order.RedeemCodeOwnerUserId = userID
		order.RedeemCodeOwnerUsername = buyer.Username
		order.RedeemCodeOwnerDisplayName = buyer.DisplayName
		order.RedeemCodeOwnerEmail = buyer.Email
		inventory := model.DistributionInventory{
			AgentId:      agent.Id,
			OrderId:      order.Id,
			PackageId:    packageID,
			Status:       DistributionInventoryStatusAvailable,
			CreditAmount: distributionPackage.CreditAmount,
			RetailPrice:  distributionPackage.RetailPrice,
			InventoryNo:  inventoryNo,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := tx.Create(&inventory).Error; err != nil {
			return err
		}
		order.RedeemCodeId = inventory.Id
		if err := tx.Save(&order).Error; err != nil {
			return err
		}
		if err := createDistributionLedger(tx, model.DistributionBalanceLedger{
			LedgerNo:       BuildLedgerNo(agent.Id, DistributionSourcePurchase, orderNo),
			IdempotencyKey: key,
			AgentId:        agent.Id,
			UserId:         userID,
			EntryType:      DistributionLedgerEntryDebit,
			SourceType:     DistributionSourcePurchase,
			SourceId:       order.Id,
			SourceNo:       orderNo,
			Delta:          -paidAmount,
			BalanceBefore:  before,
			BalanceAfter:   after,
			Description:    "distribution package purchase",
			CreatedAt:      now,
		}); err != nil {
			return err
		}
		if err := tx.Create(&model.DistributionCommissionLog{
			AgentId:     agent.Id,
			OrderId:     order.Id,
			BaseAmount:  unitPrice,
			RateBps:     agent.CommissionBps,
			Amount:      commissionAmount,
			Status:      DistributionLogStatusPosted,
			Description: "distribution purchase commission",
			CreatedAt:   now,
		}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.DistributionCustomerAttributionLog{
			CustomerUserId: userID,
			AgentId:        agent.Id,
			SourceType:     DistributionSourcePurchase,
			SourceId:       order.Id,
			SourceNo:       orderNo,
			EventType:      DistributionCustomerEventPurchase,
			OrderId:        order.Id,
			Message:        "distribution package purchased",
			CreatedAt:      now,
		}).Error; err != nil {
			return err
		}
		if err := postDistributionProfit(tx, agent, order, distributionPackage.AgentPrice, distributionPackage.SecondaryAgentPrice, unitPrice, quantity, now); err != nil {
			return err
		}
		result.Order = &order
		result.Inventory = &inventory
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func postDistributionProfit(tx *gorm.DB, childAgent model.DistributionAgent, order model.DistributionOrder, primaryAgentPrice int, secondaryAgentPrice int, childPrice int, quantity int, now int64) error {
	parentID := childAgent.ParentAgentId
	visited := map[int]struct{}{childAgent.Id: {}}
	for parentID > 0 {
		if _, ok := visited[parentID]; ok {
			return fmt.Errorf("agent hierarchy cycle detected")
		}
		visited[parentID] = struct{}{}
		var parent model.DistributionAgent
		if err := distributionLock(tx).Where("id = ? AND status = ?", parentID, DistributionStatusEnabled).First(&parent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		parentCost, err := resolveDistributionAgentPrice(tx, primaryAgentPrice, secondaryAgentPrice, parent.Id)
		if err != nil {
			return err
		}
		unitProfit := childPrice - parentCost
		if unitProfit > 0 {
			amount := unitProfit * quantity
			profitNo := BuildProfitNo(order.Id, parent.Id)
			var existing model.DistributionProfitLog
			err := tx.Where("profit_no = ?", profitNo).First(&existing).Error
			if err == nil {
				childPrice = parentCost
				parentID = parent.ParentAgentId
				continue
			}
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			before := parent.Balance
			after := before + amount
			res := tx.Model(&model.DistributionAgent{}).
				Where("id = ?", parent.Id).
				Updates(map[string]any{"balance": gorm.Expr("balance + ?", amount), "updated_at": now})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected != 1 {
				return fmt.Errorf("profit balance update failed")
			}
			profitLog := model.DistributionProfitLog{
				ProfitNo:       profitNo,
				IdempotencyKey: order.OrderNo,
				AgentId:        parent.Id,
				ChildAgentId:   childAgent.Id,
				OrderId:        order.Id,
				SourceType:     DistributionSourceProfit,
				UnitProfit:     unitProfit,
				Quantity:       quantity,
				Amount:         amount,
				ParentCost:     parentCost,
				SecondaryPrice: childPrice,
				Status:         DistributionLogStatusPosted,
				Description:    "distribution parent profit",
				CreatedAt:      now,
			}
			if err := tx.Create(&profitLog).Error; err != nil {
				return err
			}
			if err := createDistributionLedger(tx, model.DistributionBalanceLedger{
				LedgerNo:       BuildLedgerNo(parent.Id, DistributionSourceProfit, profitNo),
				IdempotencyKey: order.OrderNo,
				AgentId:        parent.Id,
				UserId:         parent.UserId,
				EntryType:      DistributionLedgerEntryCredit,
				SourceType:     DistributionSourceProfit,
				SourceId:       profitLog.Id,
				SourceNo:       profitNo,
				Delta:          amount,
				BalanceBefore:  before,
				BalanceAfter:   after,
				Description:    "distribution parent profit",
				CreatedAt:      now,
			}); err != nil {
				return err
			}
		}
		childPrice = parentCost
		parentID = parent.ParentAgentId
	}
	return nil
}

func hydrateDistributionInventoryUsers(inventories []model.DistributionInventory) error {
	if len(inventories) == 0 {
		return nil
	}
	userIDs := make([]int, 0, len(inventories))
	seen := map[int]struct{}{}
	for _, inventory := range inventories {
		if inventory.AssignedTo <= 0 {
			continue
		}
		if _, ok := seen[inventory.AssignedTo]; ok {
			continue
		}
		seen[inventory.AssignedTo] = struct{}{}
		userIDs = append(userIDs, inventory.AssignedTo)
	}
	if len(userIDs) == 0 {
		return nil
	}
	var users []model.User
	if err := model.DB.Select("id, username, display_name, email").Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return err
	}
	userMap := make(map[int]model.User, len(users))
	for _, user := range users {
		userMap[user.Id] = user
	}
	for i := range inventories {
		if user, ok := userMap[inventories[i].AssignedTo]; ok {
			inventories[i].Username = user.Username
			inventories[i].DisplayName = user.DisplayName
			inventories[i].Email = user.Email
		}
	}
	return nil
}

func ListDistributionAgentInventory(userID int, input DistributionInventoryListInput) ([]model.DistributionInventory, int64, error) {
	agent, err := GetEnabledDistributionAgentByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	input.Keyword = strings.TrimSpace(input.Keyword)
	input.Status = strings.TrimSpace(input.Status)
	var inventories []model.DistributionInventory
	query := model.DB.Model(&model.DistributionInventory{}).Where("p3_inventories.agent_id = ?", agent.Id)
	if input.Status != "" {
		query = query.Where("p3_inventories.status = ?", input.Status)
	}
	if input.Keyword != "" {
		like := "%" + input.Keyword + "%"
		query = query.Joins("LEFT JOIN users ON users.id = p3_inventories.assigned_to").
			Where("users.username LIKE ? OR users.display_name LIKE ? OR users.email LIKE ?", like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("p3_inventories.id desc").Limit(input.PageSize).Offset(input.StartIdx).Find(&inventories).Error; err != nil {
		return nil, 0, err
	}
	if err := hydrateDistributionInventoryUsers(inventories); err != nil {
		return nil, 0, err
	}
	return inventories, total, nil
}

func AssignDistributionAgentInventory(userID int, inventoryID int, customerUserID int) (*DistributionInventoryAssignResult, error) {
	if inventoryID <= 0 {
		return nil, fmt.Errorf("invalid inventory id")
	}
	if customerUserID <= 0 {
		return nil, fmt.Errorf("invalid customer user id")
	}
	now := time.Now().Unix()
	var inventory model.DistributionInventory
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var agent model.DistributionAgent
		if err := distributionLock(tx).Where("user_id = ? AND status = ?", userID, DistributionStatusEnabled).First(&agent).Error; err != nil {
			return err
		}
		if err := distributionLock(tx).Where("id = ? AND agent_id = ?", inventoryID, agent.Id).First(&inventory).Error; err != nil {
			return err
		}
		if !CanAssignInventory(inventory.Status, inventory.AssignedTo) {
			return fmt.Errorf("inventory cannot be assigned")
		}
		inventory.Status = DistributionInventoryStatusAssigned
		inventory.AssignedTo = customerUserID
		inventory.UpdatedAt = now
		if err := tx.Save(&inventory).Error; err != nil {
			return err
		}
		if err := bindDistributionCustomer(tx, customerUserID, agent.Id, DistributionCustomerEventAssign, DistributionSourceInventory, inventory.Id, inventory.InventoryNo, 0, "inventory assigned", now); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &DistributionInventoryAssignResult{Inventory: &inventory}, nil
}

func CanRefundInventory(status string, assignedTo int) bool {
	status = strings.TrimSpace(status)
	return assignedTo == 0 && status == DistributionInventoryStatusAvailable
}

func RefundDistributionAgentInventory(userID int, inventoryID int) (*DistributionInventoryRefundResult, error) {
	if inventoryID <= 0 {
		return nil, fmt.Errorf("invalid inventory id")
	}
	now := time.Now().Unix()
	var inventory model.DistributionInventory
	var ledger model.DistributionBalanceLedger
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var agent model.DistributionAgent
		if err := distributionLock(tx).Where("user_id = ? AND status = ?", userID, DistributionStatusEnabled).First(&agent).Error; err != nil {
			return err
		}
		if err := distributionLock(tx).Where("id = ? AND agent_id = ?", inventoryID, agent.Id).First(&inventory).Error; err != nil {
			return err
		}
		if !CanRefundInventory(inventory.Status, inventory.AssignedTo) {
			return fmt.Errorf("inventory cannot be refunded")
		}
		var order model.DistributionOrder
		if err := distributionLock(tx).Where("id = ? AND agent_id = ?", inventory.OrderId, agent.Id).First(&order).Error; err != nil {
			return err
		}
		if order.Status != DistributionOrderStatusFulfilled {
			return fmt.Errorf("order cannot be refunded")
		}
		refundAmount := order.PaidAmount
		if refundAmount <= 0 {
			refundAmount = order.TotalAgentPrice
		}
		if refundAmount <= 0 {
			return fmt.Errorf("refund amount must be greater than 0")
		}
		before := agent.Balance
		after := before + refundAmount
		res := tx.Model(&model.DistributionAgent{}).
			Where("id = ?", agent.Id).
			Updates(map[string]any{"balance": gorm.Expr("balance + ?", refundAmount), "updated_at": now})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return fmt.Errorf("refund balance update failed")
		}
		inventory.Status = DistributionInventoryStatusRefunded
		inventory.UpdatedAt = now
		if err := tx.Save(&inventory).Error; err != nil {
			return err
		}
		order.Status = DistributionOrderStatusRefunded
		order.UpdatedAt = now
		if err := tx.Save(&order).Error; err != nil {
			return err
		}
		ledger = model.DistributionBalanceLedger{
			LedgerNo:       BuildLedgerNo(agent.Id, DistributionSourceRefund, inventory.InventoryNo),
			IdempotencyKey: order.IdempotencyKey,
			AgentId:        agent.Id,
			UserId:         userID,
			EntryType:      DistributionLedgerEntryCredit,
			SourceType:     DistributionSourceRefund,
			SourceId:       inventory.Id,
			SourceNo:       inventory.InventoryNo,
			Delta:          refundAmount,
			BalanceBefore:  before,
			BalanceAfter:   after,
			Description:    "distribution inventory refund",
			CreatedAt:      now,
		}
		if err := createDistributionLedger(tx, ledger); err != nil {
			return err
		}
		return tx.Model(&model.DistributionCommissionLog{}).
			Where("order_id = ? AND status = ?", order.Id, DistributionLogStatusPosted).
			Updates(map[string]any{"status": DistributionLogStatusRefunded}).Error
	})
	if err != nil {
		return nil, err
	}
	return &DistributionInventoryRefundResult{Inventory: &inventory, Ledger: &ledger}, nil
}

func bindDistributionCustomer(tx *gorm.DB, customerUserID int, agentID int, eventType string, sourceType string, sourceID int, sourceNo string, orderID int, message string, now int64) error {
	var ownership model.DistributionCustomerOwnership
	err := tx.Where("customer_user_id = ?", customerUserID).First(&ownership).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ownership = model.DistributionCustomerOwnership{
			CustomerUserId: customerUserID,
			AgentId:        agentID,
			SourceType:     sourceType,
			SourceId:       sourceID,
			SourceNo:       sourceNo,
			BoundAt:        now,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := tx.Create(&ownership).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return tx.Create(&model.DistributionCustomerAttributionLog{
		CustomerUserId: customerUserID,
		AgentId:        agentID,
		SourceType:     sourceType,
		SourceId:       sourceID,
		SourceNo:       sourceNo,
		EventType:      eventType,
		OrderId:        orderID,
		Message:        message,
		CreatedAt:      now,
	}).Error
}

func ListDistributionAgentLedger(userID int, startIdx int, pageSize int) ([]model.DistributionBalanceLedger, int64, error) {
	agent, err := GetEnabledDistributionAgentByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	var ledgers []model.DistributionBalanceLedger
	query := model.DB.Model(&model.DistributionBalanceLedger{}).Where("agent_id = ?", agent.Id)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err = query.Order("id desc").Limit(pageSize).Offset(startIdx).Find(&ledgers).Error
	return ledgers, total, err
}

func ListDistributionAgentProfit(userID int, startIdx int, pageSize int) ([]model.DistributionProfitLog, int64, error) {
	agent, err := GetEnabledDistributionAgentByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	var profits []model.DistributionProfitLog
	query := model.DB.Model(&model.DistributionProfitLog{}).Where("agent_id = ?", agent.Id)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err = query.Order("id desc").Limit(pageSize).Offset(startIdx).Find(&profits).Error
	return profits, total, err
}

func AdminListDistributionProfit(startIdx int, pageSize int) ([]model.DistributionProfitLog, int64, error) {
	var profits []model.DistributionProfitLog
	query := model.DB.Model(&model.DistributionProfitLog{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id desc").Limit(pageSize).Offset(startIdx).Find(&profits).Error
	return profits, total, err
}

func hydrateDistributionCustomerUsers(customers []model.DistributionCustomerOwnership) error {
	if len(customers) == 0 {
		return nil
	}
	userIDs := make([]int, 0, len(customers))
	for _, customer := range customers {
		if customer.CustomerUserId > 0 {
			userIDs = append(userIDs, customer.CustomerUserId)
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
	for i := range customers {
		if user, ok := userMap[customers[i].CustomerUserId]; ok {
			customers[i].Username = user.Username
			customers[i].DisplayName = user.DisplayName
			customers[i].Email = user.Email
		}
	}
	return nil
}

func ListDistributionCustomers(userID int, keyword string, startIdx int, pageSize int) ([]model.DistributionCustomerOwnership, int64, error) {
	agent, err := GetEnabledDistributionAgentByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	keyword = strings.TrimSpace(keyword)
	var customers []model.DistributionCustomerOwnership
	query := model.DB.Model(&model.DistributionCustomerOwnership{}).Where("p3_customer_ownerships.agent_id = ?", agent.Id)
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Joins("LEFT JOIN users ON users.id = p3_customer_ownerships.customer_user_id").
			Where("users.username LIKE ? OR users.display_name LIKE ? OR users.email LIKE ?", like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("p3_customer_ownerships.id desc").Limit(pageSize).Offset(startIdx).Find(&customers).Error; err != nil {
		return nil, 0, err
	}
	if err := hydrateDistributionCustomerUsers(customers); err != nil {
		return nil, 0, err
	}
	return customers, total, nil
}

func ListDistributionAgentInventoryPackageOptions(userID int) ([]DistributionInventoryPackageOption, error) {
	agent, err := GetEnabledDistributionAgentByUserID(userID)
	if err != nil {
		return nil, err
	}
	var options []DistributionInventoryPackageOption
	if err := model.DB.Model(&model.DistributionInventory{}).
		Select("package_id").
		Where("agent_id = ?", agent.Id).
		Where("status <> ? AND status <> ?", DistributionInventoryStatusRefunded, DistributionInventoryStatusVoided).
		Group("package_id").
		Scan(&options).Error; err != nil {
		return nil, err
	}
	if len(options) == 0 {
		return options, nil
	}
	packageIDs := make([]int, 0, len(options))
	for _, option := range options {
		packageIDs = append(packageIDs, option.PackageId)
	}
	var packages []model.DistributionPackage
	if err := model.DB.Select("id, name, subscription_plan_id, retail_price, agent_price, secondary_agent_price").Where("id IN ?", packageIDs).Find(&packages).Error; err != nil {
		return nil, err
	}
	// 实时刷新零售价（套餐改价后立即生效），供前端做优惠码额度上限。
	if err := hydrateDistributionPackagesFromSubscriptionPlans(packages); err != nil {
		return nil, err
	}
	packageMap := make(map[int]model.DistributionPackage, len(packages))
	for _, distributionPackage := range packages {
		packageMap[distributionPackage.Id] = distributionPackage
	}
	for i := range options {
		if pkg, ok := packageMap[options[i].PackageId]; ok {
			options[i].PackageName = pkg.Name
			options[i].RetailPrice = pkg.RetailPrice
		}
	}
	return options, nil
}

func AdminListDistributionAttributionLogs(startIdx int, pageSize int) ([]model.DistributionCustomerAttributionLog, int64, error) {
	var logs []model.DistributionCustomerAttributionLog
	query := model.DB.Model(&model.DistributionCustomerAttributionLog{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id desc").Limit(pageSize).Offset(startIdx).Find(&logs).Error
	return logs, total, err
}
