package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	DistributionCouponSourceSelf   = "self"
	DistributionCouponSourceAdmin  = "admin"
	DistributionCouponStatusActive = "active"
	DistributionCouponStatusUsed   = "used"
	// DistributionCouponSelfValidityDays 代理自申请优惠券的有效期（天），到期未用自动退回余额
	DistributionCouponSelfValidityDays = 3
	// DistributionCouponIssueMaxCount 管理员单次发放的总张数上限
	DistributionCouponIssueMaxCount = 100

	distributionCouponSweepInterval  = 1 * time.Minute
	distributionCouponSweepBatchSize = 100
)

var (
	couponExpiryOnce   sync.Once
	couponSweepRunning atomic.Bool
)

type DistributionCouponIssueItem struct {
	Count        int     `json:"count"`
	Amount       float64 `json:"amount"`
	ValidityDays int     `json:"validity_days"`
}

// calcDistributionCouponQuota 面额（美元）按 QuotaPerUnit 折算为额度，向上取整
func calcDistributionCouponQuota(amount float64) (int, error) {
	if amount <= 0 {
		return 0, fmt.Errorf("amount must be greater than 0")
	}
	if common.QuotaPerUnit <= 0 {
		return 0, fmt.Errorf("额度单位配置错误")
	}
	quota := decimal.NewFromFloat(amount).
		Mul(decimal.NewFromFloat(common.QuotaPerUnit)).
		Ceil().
		IntPart()
	return int(quota), nil
}

// createDistributionCouponTx 在事务内创建原生余额兑换码并包装为优惠券
func createDistributionCouponTx(tx *gorm.DB, agentID int, creatorUserID int, amount float64, quota int, source string, issuedBy int, expiresAt int64, remark string, now int64) (*model.DistributionCoupon, error) {
	key := common.GetUUID()
	redemption := model.Redemption{
		UserId:      creatorUserID,
		Name:        "分销优惠券",
		Key:         key,
		Status:      common.RedemptionCodeStatusEnabled,
		CreatedTime: now,
		Quota:       quota,
		ExpiredTime: expiresAt,
	}
	if err := tx.Create(&redemption).Error; err != nil {
		return nil, err
	}
	coupon := model.DistributionCoupon{
		AgentId:      agentID,
		RedemptionId: redemption.Id,
		Code:         key,
		Amount:       amount,
		Quota:        quota,
		Source:       source,
		Status:       DistributionCouponStatusActive,
		IssuedBy:     issuedBy,
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
		UpdatedAt:    now,
		Remark:       remark,
	}
	if err := tx.Create(&coupon).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

// ApplyAgentCoupon 代理用自己余额 1:1 申请优惠券，3 天有效，到期未用自动退回余额
func ApplyAgentCoupon(userID int, amount float64) (*model.DistributionCoupon, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	amount = normalizeDistributionMoneyAmount(amount)
	quota, err := calcDistributionCouponQuota(amount)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	expiresAt := now + DistributionCouponSelfValidityDays*24*3600
	var coupon *model.DistributionCoupon
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		var agent model.DistributionAgent
		if err := distributionLock(tx).Where("user_id = ? AND status = ?", userID, DistributionStatusEnabled).First(&agent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("agent profile not found")
			}
			return err
		}
		if agent.Balance < amount {
			return fmt.Errorf("balance is not enough")
		}
		before := agent.Balance
		after := before - amount
		res := tx.Model(&model.DistributionAgent{}).
			Where("id = ? AND balance >= ?", agent.Id, amount).
			Updates(map[string]any{"balance": gorm.Expr("balance - ?", amount), "updated_at": now})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return fmt.Errorf("balance update failed")
		}
		created, err := createDistributionCouponTx(tx, agent.Id, userID, amount, quota, DistributionCouponSourceSelf, 0, expiresAt, "", now)
		if err != nil {
			return err
		}
		coupon = created
		return createDistributionLedger(tx, model.DistributionBalanceLedger{
			LedgerNo:      BuildLedgerNo(agent.Id, DistributionSourceCouponApply, created.Code),
			AgentId:       agent.Id,
			UserId:        userID,
			EntryType:     DistributionLedgerEntryDebit,
			SourceType:    DistributionSourceCouponApply,
			SourceId:      created.Id,
			SourceNo:      created.Code,
			Delta:         -amount,
			BalanceBefore: before,
			BalanceAfter:  after,
			Description:   "代理申请优惠券扣款",
			CreatedAt:     now,
		})
	})
	if err != nil {
		return nil, err
	}
	return coupon, nil
}

// AdminIssueCoupons 管理员手工批量给代理发放优惠券（多档 数量×面额×有效期天数），不动代理余额，不可退
func AdminIssueCoupons(agentID int, issuedBy int, items []DistributionCouponIssueItem, remark string) (int, error) {
	if agentID <= 0 {
		return 0, fmt.Errorf("invalid agent id")
	}
	if len(items) == 0 {
		return 0, fmt.Errorf("items cannot be empty")
	}
	totalCount := 0
	for i := range items {
		if items[i].Count <= 0 {
			return 0, fmt.Errorf("count must be greater than 0")
		}
		if items[i].ValidityDays <= 0 {
			return 0, fmt.Errorf("validity_days must be greater than 0")
		}
		items[i].Amount = normalizeDistributionMoneyAmount(items[i].Amount)
		if items[i].Amount <= 0 {
			return 0, fmt.Errorf("amount must be greater than 0")
		}
		totalCount += items[i].Count
	}
	if totalCount > DistributionCouponIssueMaxCount {
		return 0, fmt.Errorf("cannot issue more than %d coupons at once", DistributionCouponIssueMaxCount)
	}
	remark = strings.TrimSpace(remark)
	now := time.Now().Unix()
	issued := 0
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var agent model.DistributionAgent
		if err := tx.Where("id = ? AND status = ?", agentID, DistributionStatusEnabled).First(&agent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("agent profile not found")
			}
			return err
		}
		for _, item := range items {
			quota, err := calcDistributionCouponQuota(item.Amount)
			if err != nil {
				return err
			}
			expiresAt := now + int64(item.ValidityDays)*24*3600
			for i := 0; i < item.Count; i++ {
				if _, err := createDistributionCouponTx(tx, agent.Id, issuedBy, item.Amount, quota, DistributionCouponSourceAdmin, issuedBy, expiresAt, remark, now); err != nil {
					return err
				}
				issued++
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return issued, nil
}

func ListAgentCoupons(userID int, startIdx int, pageSize int) ([]model.DistributionCoupon, int64, error) {
	agent, err := GetEnabledDistributionAgentByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	var coupons []model.DistributionCoupon
	query := model.DB.Model(&model.DistributionCoupon{}).Where("agent_id = ?", agent.Id)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id desc").Limit(pageSize).Offset(startIdx).Find(&coupons).Error; err != nil {
		return nil, 0, err
	}
	return coupons, total, nil
}

func hydrateDistributionCouponAgents(coupons []model.DistributionCoupon) error {
	if len(coupons) == 0 {
		return nil
	}
	agentIDs := make([]int, 0, len(coupons))
	seen := map[int]struct{}{}
	for _, coupon := range coupons {
		if coupon.AgentId <= 0 {
			continue
		}
		if _, ok := seen[coupon.AgentId]; ok {
			continue
		}
		seen[coupon.AgentId] = struct{}{}
		agentIDs = append(agentIDs, coupon.AgentId)
	}
	if len(agentIDs) == 0 {
		return nil
	}
	var agents []model.DistributionAgent
	if err := model.DB.Select("id, name").Where("id IN ?", agentIDs).Find(&agents).Error; err != nil {
		return err
	}
	agentMap := make(map[int]string, len(agents))
	for _, agent := range agents {
		agentMap[agent.Id] = agent.Name
	}
	for i := range coupons {
		coupons[i].AgentName = agentMap[coupons[i].AgentId]
	}
	return nil
}

// AdminListCoupons agentID 为 0 时列出全部代理的优惠券
func AdminListCoupons(agentID int, startIdx int, pageSize int) ([]model.DistributionCoupon, int64, error) {
	var coupons []model.DistributionCoupon
	query := model.DB.Model(&model.DistributionCoupon{})
	if agentID > 0 {
		query = query.Where("agent_id = ?", agentID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id desc").Limit(pageSize).Offset(startIdx).Find(&coupons).Error; err != nil {
		return nil, 0, err
	}
	if err := hydrateDistributionCouponAgents(coupons); err != nil {
		return nil, 0, err
	}
	return coupons, total, nil
}

// GetCouponByCode 未命中返回 (nil, nil)，供钱包兑换入口前置判断
func GetCouponByCode(code string) (*model.DistributionCoupon, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, nil
	}
	var coupon model.DistributionCoupon
	err := model.DB.Where("code = ?", code).First(&coupon).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

// MarkCouponUsed 条件更新防并发：仅 active 状态的券能被标记为已使用
func MarkCouponUsed(couponID int, usedUserID int) error {
	if couponID <= 0 {
		return fmt.Errorf("invalid coupon id")
	}
	now := time.Now().Unix()
	res := model.DB.Model(&model.DistributionCoupon{}).
		Where("id = ? AND status = ?", couponID, DistributionCouponStatusActive).
		Updates(map[string]any{
			"status":       DistributionCouponStatusUsed,
			"used_user_id": usedUserID,
			"used_at":      now,
			"updated_at":   now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("coupon is not active")
	}
	return nil
}

// SweepExpiredCoupons 清扫到期未用的优惠券：券和兑换码都物理销毁，
// self 来源退回代理余额并记 credit 流水；admin 来源仅销毁不退款。
func SweepExpiredCoupons() (int, error) {
	now := time.Now().Unix()
	var coupons []model.DistributionCoupon
	if err := model.DB.
		Where("status = ? AND expires_at > 0 AND expires_at < ?", DistributionCouponStatusActive, now).
		Order("id asc").
		Limit(distributionCouponSweepBatchSize).
		Find(&coupons).Error; err != nil {
		return 0, err
	}
	swept := 0
	for i := range coupons {
		if err := sweepExpiredCoupon(&coupons[i], now); err != nil {
			common.SysLog(fmt.Sprintf("failed to sweep expired distribution coupon %d: %s", coupons[i].Id, err.Error()))
			continue
		}
		swept++
	}
	return swept, nil
}

func sweepExpiredCoupon(coupon *model.DistributionCoupon, now int64) error {
	return model.DB.Transaction(func(tx *gorm.DB) error {
		var redemption model.Redemption
		err := distributionLock(tx).Where("id = ?", coupon.RedemptionId).First(&redemption).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil && redemption.Status != common.RedemptionCodeStatusEnabled {
			// 兑换已发生但 MarkCouponUsed 尚未落库：补记 used，不退款不销毁
			return tx.Model(&model.DistributionCoupon{}).
				Where("id = ? AND status = ?", coupon.Id, DistributionCouponStatusActive).
				Updates(map[string]any{
					"status":       DistributionCouponStatusUsed,
					"used_user_id": redemption.UsedUserId,
					"used_at":      redemption.RedeemedTime,
					"updated_at":   now,
				}).Error
		}
		// 条件 DELETE 防与兑换并发：影响行数不为 1 说明已被并发处理，直接跳过
		res := tx.Where("id = ? AND status = ?", coupon.Id, DistributionCouponStatusActive).Delete(&model.DistributionCoupon{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return nil
		}
		if err == nil {
			if err := tx.Delete(&redemption).Error; err != nil {
				return err
			}
		}
		if coupon.Source != DistributionCouponSourceSelf {
			return nil
		}
		var agent model.DistributionAgent
		if err := distributionLock(tx).Where("id = ?", coupon.AgentId).First(&agent).Error; err != nil {
			return err
		}
		before := agent.Balance
		after := before + coupon.Amount
		balanceRes := tx.Model(&model.DistributionAgent{}).
			Where("id = ?", agent.Id).
			Updates(map[string]any{"balance": gorm.Expr("balance + ?", coupon.Amount), "updated_at": now})
		if balanceRes.Error != nil {
			return balanceRes.Error
		}
		if balanceRes.RowsAffected != 1 {
			return fmt.Errorf("refund balance update failed")
		}
		return createDistributionLedger(tx, model.DistributionBalanceLedger{
			LedgerNo:      BuildLedgerNo(agent.Id, DistributionSourceCouponRefund, coupon.Code),
			AgentId:       agent.Id,
			UserId:        agent.UserId,
			EntryType:     DistributionLedgerEntryCredit,
			SourceType:    DistributionSourceCouponRefund,
			SourceId:      coupon.Id,
			SourceNo:      coupon.Code,
			Delta:         coupon.Amount,
			BalanceBefore: before,
			BalanceAfter:  after,
			Description:   "优惠券到期退回",
			CreatedAt:     now,
		})
	})
}

// StartCouponExpiryTask 优惠券到期清扫定时任务（仅主节点运行）
func StartCouponExpiryTask() {
	couponExpiryOnce.Do(func() {
		if !common.IsMasterNode {
			return
		}
		gopool.Go(func() {
			logger.LogInfo(context.Background(), fmt.Sprintf("distribution coupon expiry task started: tick=%s", distributionCouponSweepInterval))
			ticker := time.NewTicker(distributionCouponSweepInterval)
			defer ticker.Stop()

			runCouponExpirySweepOnce()
			for range ticker.C {
				runCouponExpirySweepOnce()
			}
		})
	})
}

func runCouponExpirySweepOnce() {
	if !couponSweepRunning.CompareAndSwap(false, true) {
		return
	}
	defer couponSweepRunning.Store(false)

	for {
		n, err := SweepExpiredCoupons()
		if err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("distribution coupon expiry sweep failed: %v", err))
			return
		}
		if n < distributionCouponSweepBatchSize {
			break
		}
	}
}
