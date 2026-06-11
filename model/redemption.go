package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Redemption code types
const (
	RedemptionTypeBalance = "balance" // 余额码：兑换后增加用户额度
	RedemptionTypePlan    = "plan"    // 套餐码：兑换后直接开通订阅套餐
	RedemptionTypePromo   = "promo"   // 优惠码：购买套餐时抵扣折扣，不可直接兑换
)

type Redemption struct {
	Id           int            `json:"id"`
	UserId       int            `json:"user_id"`
	Key          string         `json:"key" gorm:"type:char(32);uniqueIndex"`
	Status       int            `json:"status" gorm:"default:1"`
	Name         string         `json:"name" gorm:"index"`
	Quota        int            `json:"quota" gorm:"default:100"`
	CreatedTime  int64          `json:"created_time" gorm:"bigint"`
	RedeemedTime int64          `json:"redeemed_time" gorm:"bigint"`
	Count        int            `json:"count" gorm:"-:all"` // only for api request
	UsedUserId   int            `json:"used_user_id"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	ExpiredTime  int64          `json:"expired_time" gorm:"bigint"`                           // 过期时间，0 表示不过期
	Type         string         `json:"type" gorm:"type:varchar(16);index;default:'balance'"` // balance/plan/promo
	PlanId       int            `json:"plan_id" gorm:"default:0"`                             // plan 码对应的套餐；promo 码可选限定套餐
	DiscountBps  int            `json:"discount_bps" gorm:"default:0"`                        // promo 折扣，万分比，如 2000=立减20%
	MaxUses      int            `json:"max_uses" gorm:"default:0"`                            // promo 码可用次数，0=不限
	UsedCount    int            `json:"used_count" gorm:"default:0"`                          // promo 码已用次数
}

// applyRedemptionTypeFilter 按兑换码类型过滤（空 typeFilter = 全部）。
// 历史数据可能存在空 type，视同 balance。
func applyRedemptionTypeFilter(query *gorm.DB, typeFilter string) *gorm.DB {
	switch typeFilter {
	case "":
		return query
	case RedemptionTypeBalance:
		return query.Where("(type = ? OR type = '' OR type IS NULL)", RedemptionTypeBalance)
	default:
		return query.Where("type = ?", typeFilter)
	}
}

func GetAllRedemptions(startIdx int, num int, typeFilter string) (redemptions []*Redemption, total int64, err error) {
	// 开始事务
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	query := applyRedemptionTypeFilter(tx.Model(&Redemption{}), typeFilter)

	// 获取总数
	err = query.Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// 获取分页数据
	err = query.Order("id desc").Limit(num).Offset(startIdx).Find(&redemptions).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return redemptions, total, nil
}

func SearchRedemptions(keyword string, typeFilter string, startIdx int, num int) (redemptions []*Redemption, total int64, err error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Build query based on keyword type
	query := applyRedemptionTypeFilter(tx.Model(&Redemption{}), typeFilter)

	// Only try to convert to ID if the string represents a valid integer
	if id, err := strconv.Atoi(keyword); err == nil {
		query = query.Where("id = ? OR name LIKE ?", id, keyword+"%")
	} else {
		query = query.Where("name LIKE ?", keyword+"%")
	}

	// Get total count
	err = query.Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// Get paginated data
	err = query.Order("id desc").Limit(num).Offset(startIdx).Find(&redemptions).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return redemptions, total, nil
}

func GetRedemptionById(id int) (*Redemption, error) {
	if id == 0 {
		return nil, errors.New("id 为空！")
	}
	redemption := Redemption{Id: id}
	var err error = nil
	err = DB.First(&redemption, "id = ?", id).Error
	return &redemption, err
}

// ErrPromoCodeNotRedeemable 优惠码不允许直接兑换，需在购买套餐时使用。
var ErrPromoCodeNotRedeemable = errors.New("优惠码请在购买套餐时使用")

// RedeemResult 兑换结果，按兑换码类型区分。
type RedeemResult struct {
	Type      string `json:"type"`
	Quota     int    `json:"quota"`
	PlanId    int    `json:"plan_id"`
	PlanTitle string `json:"plan_title"`
}

func redemptionKeyCol() string {
	if common.UsingPostgreSQL {
		return `"key"`
	}
	return "`key`"
}

func Redeem(key string, userId int) (*RedeemResult, error) {
	if key == "" {
		return nil, errors.New("未提供兑换码")
	}
	if userId == 0 {
		return nil, errors.New("无效的 user id")
	}
	redemption := &Redemption{}
	result := &RedeemResult{}
	var upgradeGroup string

	common.RandomSleep()
	err := DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(redemptionKeyCol()+" = ?", key).First(redemption).Error
		if err != nil {
			return errors.New("无效的兑换码")
		}
		if redemption.Status != common.RedemptionCodeStatusEnabled {
			return errors.New("该兑换码已被使用")
		}
		if redemption.ExpiredTime != 0 && redemption.ExpiredTime < common.GetTimestamp() {
			return errors.New("该兑换码已过期")
		}
		// 条件 UPDATE 抢占“标记已用”（WHERE status=enabled），并发双兑换时只有一个事务能成功，
		// 失败方整体回滚，不发放任何权益（行锁失效时的最后一道保险）。
		preempt := tx.Model(&Redemption{}).
			Where("id = ? AND status = ?", redemption.Id, common.RedemptionCodeStatusEnabled).
			Updates(map[string]interface{}{
				"redeemed_time": common.GetTimestamp(),
				"status":        common.RedemptionCodeStatusUsed,
				"used_user_id":  userId,
			})
		if preempt.Error != nil {
			return preempt.Error
		}
		if preempt.RowsAffected != 1 {
			return errors.New("该兑换码已被使用")
		}
		switch redemption.Type {
		case "", RedemptionTypeBalance:
			result.Type = RedemptionTypeBalance
			result.Quota = redemption.Quota
			err = tx.Model(&User{}).Where("id = ?", userId).Update("quota", gorm.Expr("quota + ?", redemption.Quota)).Error
			if err != nil {
				return err
			}
		case RedemptionTypePlan:
			plan, err := getSubscriptionPlanByIdTx(tx, redemption.PlanId)
			if err != nil {
				return err
			}
			if !plan.Enabled {
				// 套餐已停用：与其他失败一致由外层掩码为 ErrRedeemFailed
				return errors.New("套餐未启用")
			}
			if _, err := CreateUserSubscriptionFromPlanTx(tx, userId, plan, "redemption"); err != nil {
				return err
			}
			upgradeGroup = strings.TrimSpace(plan.UpgradeGroup)
			result.Type = RedemptionTypePlan
			result.PlanId = plan.Id
			result.PlanTitle = plan.Title
		case RedemptionTypePromo:
			return ErrPromoCodeNotRedeemable
		default:
			return fmt.Errorf("未知的兑换码类型: %s", redemption.Type)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrPromoCodeNotRedeemable) {
			return nil, err
		}
		common.SysError("redemption failed: " + err.Error())
		return nil, ErrRedeemFailed
	}
	switch result.Type {
	case RedemptionTypePlan:
		if upgradeGroup != "" {
			_ = UpdateUserGroupCache(userId, upgradeGroup)
		}
		RecordLog(userId, LogTypeTopup, fmt.Sprintf("通过兑换码开通套餐 %s，兑换码ID %d", result.PlanTitle, redemption.Id))
	default:
		RecordLog(userId, LogTypeTopup, fmt.Sprintf("通过兑换码充值 %s，兑换码ID %d", logger.LogQuota(redemption.Quota), redemption.Id))
	}
	return result, nil
}

// GetRedemptionByKey 按 key 查询兑换码。
func GetRedemptionByKey(key string) (*Redemption, error) {
	if key == "" {
		return nil, errors.New("key 为空！")
	}
	var redemption Redemption
	if err := DB.Where(redemptionKeyCol()+" = ?", key).First(&redemption).Error; err != nil {
		return nil, err
	}
	return &redemption, nil
}

// ValidatePromoUsable 校验优惠码当前是否可用于指定套餐（planId<=0 时跳过套餐绑定校验）。
func (redemption *Redemption) ValidatePromoUsable(planId int) error {
	if redemption.Type != RedemptionTypePromo {
		return errors.New("该兑换码不是优惠码")
	}
	if redemption.Status != common.RedemptionCodeStatusEnabled {
		return errors.New("该优惠码不可用")
	}
	if redemption.ExpiredTime != 0 && redemption.ExpiredTime < common.GetTimestamp() {
		return errors.New("该优惠码已过期")
	}
	if redemption.DiscountBps <= 0 || redemption.DiscountBps >= 10000 {
		return errors.New("优惠码折扣配置无效")
	}
	if redemption.PlanId != 0 && planId > 0 && redemption.PlanId != planId {
		return errors.New("该优惠码不适用于此套餐")
	}
	if redemption.MaxUses > 0 && redemption.UsedCount >= redemption.MaxUses {
		return errors.New("优惠码已被用尽")
	}
	return nil
}

// ApplyPromoDiscount 金额折算唯一收口：amount * (10000 - discountBps) / 10000，
// 用 decimal 计算并保留两位小数（分位），避免浮点误差。
func ApplyPromoDiscount(amount float64, discountBps int) float64 {
	if amount <= 0 || discountBps <= 0 {
		return amount
	}
	if discountBps >= 10000 {
		return 0
	}
	result, _ := decimal.NewFromFloat(amount).
		Mul(decimal.NewFromInt(int64(10000 - discountBps))).
		Div(decimal.NewFromInt(10000)).
		Round(2).
		Float64()
	return result
}

// consumePromoUseRowsTx 原子地为优惠码 +1 次使用（带 max_uses 上限保护），返回受影响行数。
func consumePromoUseRowsTx(tx *gorm.DB, code string) (int64, error) {
	if tx == nil {
		tx = DB
	}
	res := tx.Model(&Redemption{}).
		Where(redemptionKeyCol()+" = ? AND type = ? AND (max_uses = 0 OR used_count < max_uses)", code, RedemptionTypePromo).
		Update("used_count", gorm.Expr("used_count + 1"))
	return res.RowsAffected, res.Error
}

// refundPromoUseRowsTx 原子回补一次优惠码使用次数（订单过期/取消时回滚预占），
// WHERE used_count > 0 防止回补成负数；返回受影响行数。
func refundPromoUseRowsTx(tx *gorm.DB, code string) (int64, error) {
	if tx == nil {
		tx = DB
	}
	res := tx.Model(&Redemption{}).
		Where(redemptionKeyCol()+" = ? AND type = ? AND used_count > 0", code, RedemptionTypePromo).
		Update("used_count", gorm.Expr("used_count - 1"))
	return res.RowsAffected, res.Error
}

// ConsumePromoCodeById 原子消耗一次优惠码（WHERE used_count < max_uses OR max_uses = 0）。
func ConsumePromoCodeById(tx *gorm.DB, redemptionId int) error {
	if redemptionId <= 0 {
		return errors.New("无效的优惠码 id")
	}
	if tx == nil {
		tx = DB
	}
	res := tx.Model(&Redemption{}).
		Where("id = ? AND type = ? AND (max_uses = 0 OR used_count < max_uses)", redemptionId, RedemptionTypePromo).
		Update("used_count", gorm.Expr("used_count + 1"))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("优惠码已被用尽")
	}
	return nil
}

// ConsumePromoRedemptionTx 在事务内严格校验并消耗一次优惠码，返回该优惠码（用于读取折扣）。
// 余额购买套餐路径使用：校验失败/用尽时返回错误使整笔交易回滚。
func ConsumePromoRedemptionTx(tx *gorm.DB, code string, planId int) (*Redemption, error) {
	if tx == nil {
		return nil, errors.New("tx is nil")
	}
	var redemption Redemption
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(redemptionKeyCol()+" = ?", code).First(&redemption).Error; err != nil {
		return nil, errors.New("优惠码不存在")
	}
	if err := redemption.ValidatePromoUsable(planId); err != nil {
		return nil, err
	}
	rows, err := consumePromoUseRowsTx(tx, code)
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, errors.New("优惠码已被用尽")
	}
	redemption.UsedCount++
	return &redemption, nil
}

func (redemption *Redemption) Insert() error {
	var err error
	err = DB.Create(redemption).Error
	return err
}

func (redemption *Redemption) SelectUpdate() error {
	// This can update zero values
	return DB.Model(redemption).Select("redeemed_time", "status").Updates(redemption).Error
}

// Update Make sure your token's fields is completed, because this will update non-zero values
func (redemption *Redemption) Update() error {
	var err error
	err = DB.Model(redemption).Select("name", "status", "quota", "redeemed_time", "expired_time", "type", "plan_id", "discount_bps", "max_uses").Updates(redemption).Error
	return err
}

func (redemption *Redemption) Delete() error {
	var err error
	err = DB.Delete(redemption).Error
	return err
}

func DeleteRedemptionById(id int) (err error) {
	if id == 0 {
		return errors.New("id 为空！")
	}
	redemption := Redemption{Id: id}
	err = DB.Where(redemption).First(&redemption).Error
	if err != nil {
		return err
	}
	return redemption.Delete()
}

func DeleteInvalidRedemptions() (int64, error) {
	now := common.GetTimestamp()
	result := DB.Where("status IN ? OR (status = ? AND expired_time != 0 AND expired_time < ?)", []int{common.RedemptionCodeStatusUsed, common.RedemptionCodeStatusDisabled}, common.RedemptionCodeStatusEnabled, now).Delete(&Redemption{})
	return result.RowsAffected, result.Error
}
