package service

import (
	"errors"
	"fmt"
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

func getDistributionOrderWithInventory(tx *gorm.DB, orderNo string) (*model.DistributionOrder, *model.DistributionInventory, error) {
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

func verifyDistributionPurchaseIdempotency(order model.DistributionOrder, agentID int, userID int, packageID int, key string) error {
	if order.AgentId != agentID || order.UserId != userID || order.PackageId != packageID || order.IdempotencyKey != key {
		return fmt.Errorf("idempotency key conflicts")
	}
	return nil
}

func getDistributionAgentLevel(tx *gorm.DB, agentID int) (int, error) {
	level := 0
	visited := map[int]struct{}{}
	currentID := agentID
	for currentID > 0 {
		if _, ok := visited[currentID]; ok {
			return 0, fmt.Errorf("agent hierarchy cycle detected")
		}
		visited[currentID] = struct{}{}
		var agent model.DistributionAgent
		if err := distributionLock(tx).Where("id = ?", currentID).First(&agent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if currentID == agentID {
					return 0, fmt.Errorf("agent profile not found")
				}
				break
			}
			return 0, err
		}
		if agent.ParentAgentId <= 0 {
			break
		}
		level++
		currentID = agent.ParentAgentId
	}
	return level, nil
}

func resolveDistributionAgentPrice(tx *gorm.DB, packageID int, packageDefaultPrice int, agentID int) (int, error) {
	level, err := getDistributionAgentLevel(tx, agentID)
	if err != nil {
		return 0, err
	}
	var configs []model.DistributionPriceConfig
	if err := tx.Where("package_id = ? AND status = ?", packageID, DistributionStatusEnabled).Order("id desc").Find(&configs).Error; err != nil {
		return 0, err
	}
	rules := make([]DistributionPriceConfigRule, 0, len(configs))
	for _, cfg := range configs {
		rules = append(rules, DistributionPriceConfigRule{
			ScopeType: cfg.ScopeType,
			AgentId:   cfg.AgentId,
			Level:     cfg.Level,
			UnitPrice: cfg.UnitPrice,
			Status:    cfg.Status,
		})
	}
	return ResolveDistributionAgentPrice(packageDefaultPrice, agentID, level, rules), nil
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
		res := tx.Model(&model.DistributionAgent{}).
			Where("id = ? AND balance = ?", agentID, before).
			Updates(map[string]any{"balance": after, "updated_at": now})
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
	orderNo := BuildPurchaseOrderNo(userID, packageID, key)
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
		var distributionPackage model.DistributionPackage
		if err := distributionLock(tx).Where("id = ? AND status = ?", packageID, DistributionStatusEnabled).First(&distributionPackage).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("package not found")
			}
			return err
		}
		existingOrder, existingInventory, err := getDistributionOrderWithInventory(tx, orderNo)
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
		unitPrice, err := resolveDistributionAgentPrice(tx, distributionPackage.Id, distributionPackage.AgentPrice, agent.Id)
		if err != nil {
			return err
		}
		quantity := 1
		totalPrice := unitPrice * quantity
		commissionAmount, err := CalcCommission(unitPrice, agent.CommissionBps)
		if err != nil {
			return err
		}
		order := model.DistributionOrder{
			OrderNo:          orderNo,
			IdempotencyKey:   key,
			AgentId:          agent.Id,
			UserId:           userID,
			PackageId:        packageID,
			Quantity:         quantity,
			UnitAgentPrice:   unitPrice,
			TotalAgentPrice:  totalPrice,
			RetailPrice:      distributionPackage.RetailPrice,
			CommissionBps:    agent.CommissionBps,
			CommissionAmount: commissionAmount,
			Status:           DistributionOrderStatusPending,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		createOrder := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&order)
		if createOrder.Error != nil {
			return createOrder.Error
		}
		if createOrder.RowsAffected == 0 {
			existingOrder, existingInventory, err := getDistributionOrderWithInventory(tx, orderNo)
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
		if before < totalPrice {
			return fmt.Errorf("insufficient balance")
		}
		after := before - totalPrice
		res := tx.Model(&model.DistributionAgent{}).
			Where("id = ? AND balance >= ?", agent.Id, totalPrice).
			Updates(map[string]any{"balance": gorm.Expr("balance - ?", totalPrice), "updated_at": now})
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
		if err := tx.Save(&order).Error; err != nil {
			return err
		}
		inventoryNo, err := buildDistributionInventoryNo(tx)
		if err != nil {
			return err
		}
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
		if err := createDistributionLedger(tx, model.DistributionBalanceLedger{
			LedgerNo:       BuildLedgerNo(agent.Id, DistributionSourcePurchase, orderNo),
			IdempotencyKey: key,
			AgentId:        agent.Id,
			UserId:         userID,
			EntryType:      DistributionLedgerEntryDebit,
			SourceType:     DistributionSourcePurchase,
			SourceId:       order.Id,
			SourceNo:       orderNo,
			Delta:          -totalPrice,
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
		if err := postDistributionProfit(tx, agent, order, distributionPackage.AgentPrice, unitPrice, quantity, now); err != nil {
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

func postDistributionProfit(tx *gorm.DB, childAgent model.DistributionAgent, order model.DistributionOrder, packageDefaultPrice int, secondaryPrice int, quantity int, now int64) error {
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
		parentCost, err := resolveDistributionAgentPrice(tx, order.PackageId, packageDefaultPrice, parent.Id)
		if err != nil {
			return err
		}
		unitProfit := secondaryPrice - parentCost
		if unitProfit > 0 {
			amount := unitProfit * quantity
			profitNo := BuildProfitNo(order.Id, parent.Id)
			var existing model.DistributionProfitLog
			err := tx.Where("profit_no = ?", profitNo).First(&existing).Error
			if err == nil {
				secondaryPrice = parentCost
				parentID = parent.ParentAgentId
				continue
			}
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			before := parent.Balance
			after := before + amount
			res := tx.Model(&model.DistributionAgent{}).
				Where("id = ? AND balance = ?", parent.Id, before).
				Updates(map[string]any{"balance": after, "updated_at": now})
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
				SecondaryPrice: secondaryPrice,
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
		secondaryPrice = parentCost
		parentID = parent.ParentAgentId
	}
	return nil
}

func ListDistributionAgentInventory(userID int) ([]model.DistributionInventory, error) {
	agent, err := GetEnabledDistributionAgentByUserID(userID)
	if err != nil {
		return nil, err
	}
	var inventories []model.DistributionInventory
	err = model.DB.Where("agent_id = ?", agent.Id).Order("id desc").Find(&inventories).Error
	return inventories, err
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

func ListDistributionAgentLedger(userID int) ([]model.DistributionBalanceLedger, error) {
	agent, err := GetEnabledDistributionAgentByUserID(userID)
	if err != nil {
		return nil, err
	}
	var ledgers []model.DistributionBalanceLedger
	err = model.DB.Where("agent_id = ?", agent.Id).Order("id desc").Find(&ledgers).Error
	return ledgers, err
}

func ListDistributionAgentProfit(userID int) ([]model.DistributionProfitLog, error) {
	agent, err := GetEnabledDistributionAgentByUserID(userID)
	if err != nil {
		return nil, err
	}
	var profits []model.DistributionProfitLog
	err = model.DB.Where("agent_id = ?", agent.Id).Order("id desc").Find(&profits).Error
	return profits, err
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

func ListDistributionCustomers(userID int) ([]model.DistributionCustomerOwnership, error) {
	agent, err := GetEnabledDistributionAgentByUserID(userID)
	if err != nil {
		return nil, err
	}
	var customers []model.DistributionCustomerOwnership
	err = model.DB.Where("agent_id = ?", agent.Id).Order("id desc").Find(&customers).Error
	return customers, err
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
