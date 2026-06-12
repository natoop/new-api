package model

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	DistributionOrderTypeOriginal  = "original"
	DistributionOrderTypeInventory = "inventory"
	DistributionOrderTypeRedeem    = "redeem"

	DistributionOrderStatusPending   = "pending"
	DistributionOrderStatusFulfilled = "fulfilled"
	DistributionOrderStatusCancelled = "cancelled"
)

var (
	distributionOrderNoMu        sync.Mutex
	distributionOrderNoLastMilli int64
	distributionOrderNoSequence  int64
)

func BuildDistributionOrderNo(orderType string) string {
	return fmt.Sprintf("%d", nextDistributionOrderSnowflakeID())
}

func nextDistributionOrderSnowflakeID() int64 {
	const (
		epochMilli   = int64(1704067200000)
		workerID     = int64(2)
		sequenceMask = int64(4095)
	)
	distributionOrderNoMu.Lock()
	defer distributionOrderNoMu.Unlock()

	nowMilli := time.Now().UnixMilli()
	if nowMilli < distributionOrderNoLastMilli {
		nowMilli = distributionOrderNoLastMilli
	}
	if nowMilli == distributionOrderNoLastMilli {
		distributionOrderNoSequence = (distributionOrderNoSequence + 1) & sequenceMask
		if distributionOrderNoSequence == 0 {
			for nowMilli <= distributionOrderNoLastMilli {
				nowMilli = time.Now().UnixMilli()
			}
		}
	} else {
		distributionOrderNoSequence = 0
	}
	distributionOrderNoLastMilli = nowMilli
	return ((nowMilli - epochMilli) << 22) | (workerID << 12) | distributionOrderNoSequence
}

func SyncSubscriptionDistributionOrderTx(tx *gorm.DB, order *SubscriptionOrder) error {
	if tx == nil {
		tx = DB
	}
	if tx == nil || !tx.Migrator().HasTable(&DistributionOrder{}) {
		return nil
	}
	if order == nil || order.Id <= 0 {
		return errors.New("invalid subscription order")
	}
	var plan SubscriptionPlan
	planErr := tx.Where("id = ?", order.PlanId).First(&plan).Error
	if planErr != nil && !errors.Is(planErr, gorm.ErrRecordNotFound) {
		return planErr
	}
	var buyer User
	buyerErr := tx.Select("id, username, display_name, email").Where("id = ?", order.UserId).First(&buyer).Error
	if buyerErr != nil && !errors.Is(buyerErr, gorm.ErrRecordNotFound) {
		return buyerErr
	}

	now := common.GetTimestamp()
	snapshot := DistributionOrder{
		OrderType:           DistributionOrderTypeOriginal,
		SubscriptionOrderId: order.Id,
		IdempotencyKey:      strings.TrimSpace(order.TradeNo),
		UserId:              order.UserId,
		BuyUserId:           order.UserId,
		BuyUserName:         distributionUserName(buyer),
		BuyerUserId:         order.UserId,
		BuyerUsername:       buyer.Username,
		BuyerDisplayName:    buyer.DisplayName,
		BuyerEmail:          buyer.Email,
		SubscriptionPlanId:  order.PlanId,
		OriginalAmount:      order.Money,
		PaidAmount:          order.Money,
		Quantity:            1,
		Status:              distributionStatusFromSubscription(order.Status),
		PaidAt:              order.CompleteTime,
		FulfilledAt:         order.CompleteTime,
		CompletedAt:         order.CompleteTime,
		CreatedAt:           order.CreateTime,
		UpdatedAt:           now,
	}
	if snapshot.CreatedAt == 0 {
		snapshot.CreatedAt = now
	}
	if snapshot.IdempotencyKey == "" {
		snapshot.IdempotencyKey = fmt.Sprintf("subscription_order_%d", order.Id)
	}
	if planErr == nil {
		plan.NormalizeDefaults()
		snapshot.SubscriptionTitle = plan.Title
		snapshot.SubscriptionSubtitle = plan.Subtitle
		snapshot.PackageName = plan.Title
		snapshot.PackageDescription = plan.Subtitle
		snapshot.PackageCreditAmount = int(plan.TotalAmount)
		if snapshot.OriginalAmount <= 0 {
			snapshot.OriginalAmount = plan.PriceAmount
		}
		if snapshot.PaidAmount <= 0 && order.Money > 0 {
			snapshot.PaidAmount = order.Money
		}
	}
	if err := enrichDistributionOrderFromInventoryCodeTx(tx, &snapshot, order.PromoCode); err != nil {
		return err
	}

	var existing DistributionOrder
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("order_type = ? AND subscription_order_id = ?", DistributionOrderTypeOriginal, order.Id).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		snapshot.OrderNo = BuildDistributionOrderNo(DistributionOrderTypeOriginal)
		if snapshot.AgentAgentId == 0 {
			snapshot.AgentAgentId = snapshot.AgentId
		}
		return tx.Create(&snapshot).Error
	}
	if err != nil {
		return err
	}
	snapshot.Id = existing.Id
	snapshot.OrderNo = existing.OrderNo
	if snapshot.OrderNo == "" {
		snapshot.OrderNo = BuildDistributionOrderNo(DistributionOrderTypeOriginal)
	}
	if snapshot.AgentAgentId == 0 {
		snapshot.AgentAgentId = snapshot.AgentId
	}
	return tx.Model(&existing).Updates(map[string]any{
		"order_no":                       snapshot.OrderNo,
		"order_type":                     snapshot.OrderType,
		"subscription_order_id":          snapshot.SubscriptionOrderId,
		"idempotency_key":                snapshot.IdempotencyKey,
		"agent_id":                       snapshot.AgentId,
		"agent_user_id":                  snapshot.AgentUserId,
		"agent_user_name":                snapshot.AgentUserName,
		"agent_agent_id":                 snapshot.AgentAgentId,
		"user_id":                        snapshot.UserId,
		"buy_user_id":                    snapshot.BuyUserId,
		"buy_user_name":                  snapshot.BuyUserName,
		"buyer_user_id":                  snapshot.BuyerUserId,
		"buyer_username":                 snapshot.BuyerUsername,
		"buyer_display_name":             snapshot.BuyerDisplayName,
		"buyer_email":                    snapshot.BuyerEmail,
		"package_id":                     snapshot.PackageId,
		"subscription_plan_id":           snapshot.SubscriptionPlanId,
		"subscription_title":             snapshot.SubscriptionTitle,
		"subscription_subtitle":          snapshot.SubscriptionSubtitle,
		"package_name":                   snapshot.PackageName,
		"package_sku":                    snapshot.PackageSku,
		"package_description":            snapshot.PackageDescription,
		"package_credit_amount":          snapshot.PackageCreditAmount,
		"redeem_code_id":                 snapshot.RedeemCodeId,
		"redeem_code":                    snapshot.RedeemCode,
		"redeem_code_owner_user_id":      snapshot.RedeemCodeOwnerUserId,
		"redeem_code_owner_username":     snapshot.RedeemCodeOwnerUsername,
		"redeem_code_owner_display_name": snapshot.RedeemCodeOwnerDisplayName,
		"redeem_code_owner_email":        snapshot.RedeemCodeOwnerEmail,
		"agent_active_code":              snapshot.AgentActiveCode,
		"original_amount":                snapshot.OriginalAmount,
		"paid_amount":                    snapshot.PaidAmount,
		"quantity":                       snapshot.Quantity,
		"retail_price":                   snapshot.RetailPrice,
		"status":                         snapshot.Status,
		"paid_at":                        snapshot.PaidAt,
		"fulfilled_at":                   snapshot.FulfilledAt,
		"completed_at":                   snapshot.CompletedAt,
		"created_at":                     snapshot.CreatedAt,
		"updated_at":                     snapshot.UpdatedAt,
	}).Error
}

func enrichDistributionOrderFromInventoryCodeTx(tx *gorm.DB, order *DistributionOrder, code string) error {
	code = strings.TrimSpace(code)
	if tx == nil || order == nil || code == "" {
		return nil
	}
	var inventory DistributionInventory
	err := tx.Where("inventory_no = ?", code).First(&inventory).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	order.AgentId = inventory.AgentId
	order.AgentAgentId = inventory.AgentId
	order.PackageId = inventory.PackageId
	order.RedeemCodeId = inventory.Id
	order.RedeemCode = inventory.InventoryNo
	order.AgentActiveCode = inventory.InventoryNo
	order.RetailPrice = inventory.RetailPrice
	if order.PackageCreditAmount == 0 {
		order.PackageCreditAmount = inventory.CreditAmount
	}

	var agent DistributionAgent
	if err := tx.Where("id = ?", inventory.AgentId).First(&agent).Error; err == nil {
		order.AgentUserId = agent.UserId
		order.RedeemCodeOwnerUserId = agent.UserId
		var agentUser User
		if err := tx.Select("id, username, display_name, email").Where("id = ?", agent.UserId).First(&agentUser).Error; err == nil {
			order.AgentUserName = distributionUserName(agentUser)
			order.RedeemCodeOwnerUsername = agentUser.Username
			order.RedeemCodeOwnerDisplayName = agentUser.DisplayName
			order.RedeemCodeOwnerEmail = agentUser.Email
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	var pkg DistributionPackage
	if err := tx.Where("id = ?", inventory.PackageId).First(&pkg).Error; err == nil {
		order.PackageName = firstNonEmpty(order.PackageName, pkg.Name)
		order.PackageSku = firstNonEmpty(order.PackageSku, pkg.Sku)
		order.PackageDescription = firstNonEmpty(order.PackageDescription, pkg.Description)
		if order.PackageCreditAmount == 0 {
			order.PackageCreditAmount = pkg.CreditAmount
		}
		if order.SubscriptionPlanId == 0 {
			order.SubscriptionPlanId = pkg.SubscriptionPlanId
		}
		order.SubscriptionTitle = firstNonEmpty(order.SubscriptionTitle, pkg.SubscriptionTitle)
		order.SubscriptionSubtitle = firstNonEmpty(order.SubscriptionSubtitle, pkg.SubscriptionSubtitle)
		if order.RetailPrice <= 0 {
			order.RetailPrice = pkg.RetailPrice
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	var purchaseOrder DistributionOrder
	if err := tx.Where("id = ?", inventory.OrderId).First(&purchaseOrder).Error; err == nil {
		order.SubscriptionTitle = firstNonEmpty(order.SubscriptionTitle, purchaseOrder.SubscriptionTitle)
		order.SubscriptionSubtitle = firstNonEmpty(order.SubscriptionSubtitle, purchaseOrder.SubscriptionSubtitle)
		order.PackageName = firstNonEmpty(order.PackageName, purchaseOrder.PackageName)
		order.PackageSku = firstNonEmpty(order.PackageSku, purchaseOrder.PackageSku)
		order.PackageDescription = firstNonEmpty(order.PackageDescription, purchaseOrder.PackageDescription)
		if order.RetailPrice <= 0 {
			order.RetailPrice = purchaseOrder.RetailPrice
		}
		if order.OriginalAmount <= 0 {
			order.OriginalAmount = purchaseOrder.OriginalAmount
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return nil
}

func distributionStatusFromSubscription(status string) string {
	switch strings.TrimSpace(status) {
	case common.TopUpStatusSuccess:
		return DistributionOrderStatusFulfilled
	case common.TopUpStatusExpired, common.TopUpStatusFailed:
		return DistributionOrderStatusCancelled
	default:
		return DistributionOrderStatusPending
	}
}

func distributionUserName(user User) string {
	if strings.TrimSpace(user.DisplayName) != "" {
		return user.DisplayName
	}
	if strings.TrimSpace(user.Username) != "" {
		return user.Username
	}
	return user.Email
}

func firstNonEmpty(current string, fallback string) string {
	if strings.TrimSpace(current) != "" {
		return current
	}
	return fallback
}

func BackfillDistributionOrders() error {
	if DB == nil {
		return nil
	}
	if err := DB.Model(&DistributionOrder{}).Where("order_type = '' OR order_type IS NULL").
		Update("order_type", DistributionOrderTypeInventory).Error; err != nil {
		return err
	}
	var orders []SubscriptionOrder
	if err := DB.Order("id asc").Find(&orders).Error; err != nil {
		return err
	}
	for i := range orders {
		if err := DB.Transaction(func(tx *gorm.DB) error {
			return SyncSubscriptionDistributionOrderTx(tx, &orders[i])
		}); err != nil {
			return err
		}
	}
	return nil
}
