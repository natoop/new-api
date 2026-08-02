package model

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/pkg/cachex"
	"github.com/samber/hot"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Subscription duration units
const (
	SubscriptionDurationYear   = "year"
	SubscriptionDurationMonth  = "month"
	SubscriptionDurationDay    = "day"
	SubscriptionDurationHour   = "hour"
	SubscriptionDurationCustom = "custom"
)

var (
	ErrSubscriptionOrderNotFound      = errors.New("subscription order not found")
	ErrSubscriptionOrderStatusInvalid = errors.New("subscription order status invalid")
)

// Plan credit rejections that can never succeed on a retry. A paid order that
// hits one of them must be settled into a terminal state instead of being
// rolled back to pending, otherwise the money is taken, the quota is never
// credited and the gateway retries forever with nobody watching.
var (
	ErrPlanPurchaseCapReached = errors.New("已达到该套餐购买上限")
	ErrPlanQuotaInvalid       = errors.New("无效的充值额度")
	ErrPlanQuotaOverflow      = errors.New("充值后额度将超出钱包上限，请联系管理员")
)

// isPlanCreditTerminal reports whether a CreditPlanQuotaTx failure is a
// permanent business rejection rather than a transient infrastructure error.
func isPlanCreditTerminal(err error) bool {
	return errors.Is(err, ErrPlanPurchaseCapReached) ||
		errors.Is(err, ErrPlanQuotaInvalid) ||
		errors.Is(err, ErrPlanQuotaOverflow)
}

const subscriptionPlanCacheNamespace = "new-api:subscription_plan:v1"

var (
	subscriptionPlanCacheOnce sync.Once
	subscriptionPlanCache     *cachex.HybridCache[SubscriptionPlan]
)

func subscriptionPlanCacheTTL() time.Duration {
	ttlSeconds := common.GetEnvOrDefault("SUBSCRIPTION_PLAN_CACHE_TTL", 300)
	if ttlSeconds <= 0 {
		ttlSeconds = 300
	}
	return time.Duration(ttlSeconds) * time.Second
}

func subscriptionPlanCacheCapacity() int {
	capacity := common.GetEnvOrDefault("SUBSCRIPTION_PLAN_CACHE_CAP", 5000)
	if capacity <= 0 {
		capacity = 5000
	}
	return capacity
}

func getSubscriptionPlanCache() *cachex.HybridCache[SubscriptionPlan] {
	subscriptionPlanCacheOnce.Do(func() {
		ttl := subscriptionPlanCacheTTL()
		subscriptionPlanCache = cachex.NewHybridCache[SubscriptionPlan](cachex.HybridCacheConfig[SubscriptionPlan]{
			Namespace: cachex.Namespace(subscriptionPlanCacheNamespace),
			Redis:     common.RDB,
			RedisEnabled: func() bool {
				return common.RedisEnabled && common.RDB != nil
			},
			RedisCodec: cachex.JSONCodec[SubscriptionPlan]{},
			Memory: func() *hot.HotCache[string, SubscriptionPlan] {
				return hot.NewHotCache[string, SubscriptionPlan](hot.LRU, subscriptionPlanCacheCapacity()).
					WithTTL(ttl).
					WithJanitor().
					Build()
			},
		})
	})
	return subscriptionPlanCache
}

func subscriptionPlanCacheKey(id int) string {
	if id <= 0 {
		return ""
	}
	return strconv.Itoa(id)
}

func InvalidateSubscriptionPlanCache(planId int) {
	if planId <= 0 {
		return
	}
	cache := getSubscriptionPlanCache()
	_, _ = cache.DeleteMany([]string{subscriptionPlanCacheKey(planId)})
}

// Subscription plan
type SubscriptionPlan struct {
	Id int `json:"id"`

	Title    string `json:"title" gorm:"type:varchar(128);not null"`
	Subtitle string `json:"subtitle" gorm:"type:varchar(255);default:''"`

	// Display money amount (follow existing code style: float64 for money)
	PriceAmount float64 `json:"price_amount" gorm:"type:decimal(10,6);not null;default:0"`
	Currency    string  `json:"currency" gorm:"type:varchar(8);not null;default:'USD'"`

	DurationUnit  string `json:"duration_unit" gorm:"type:varchar(16);not null;default:'month'"`
	DurationValue int    `json:"duration_value" gorm:"type:int;not null;default:1"`
	CustomSeconds int64  `json:"custom_seconds" gorm:"type:bigint;not null;default:0"`

	Enabled   bool `json:"enabled" gorm:"default:true"`
	SortOrder int  `json:"sort_order" gorm:"type:int;default:0"`

	AllowBalancePay *bool `json:"allow_balance_pay" gorm:"default:true"`

	StripePriceId         string `json:"stripe_price_id" gorm:"type:varchar(128);default:''"`
	CreemProductId        string `json:"creem_product_id" gorm:"type:varchar(128);default:''"`
	WaffoPancakeProductId string `json:"waffo_pancake_product_id" gorm:"type:varchar(128);default:''"`

	// Max purchases per user (0 = unlimited)
	MaxPurchasePerUser int `json:"max_purchase_per_user" gorm:"type:int;default:0"`

	// Upgrade user group after purchase (empty = no change)
	UpgradeGroup string `json:"upgrade_group" gorm:"type:varchar(64);default:''"`

	// Quota credited to the buyer's wallet on purchase (amount in quota units)
	TotalAmount int64 `json:"total_amount" gorm:"type:bigint;not null;default:0"`

	// Quota reset period for plan
	QuotaResetPeriod        string `json:"quota_reset_period" gorm:"type:varchar(16);default:'never'"`
	QuotaResetCustomSeconds int64  `json:"quota_reset_custom_seconds" gorm:"type:bigint;default:0"`

	CreatedAt int64 `json:"created_at" gorm:"bigint"`
	UpdatedAt int64 `json:"updated_at" gorm:"bigint"`
}

func (p *SubscriptionPlan) BeforeCreate(tx *gorm.DB) error {
	now := common.GetTimestamp()
	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

func (p *SubscriptionPlan) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = common.GetTimestamp()
	return nil
}

func (p *SubscriptionPlan) NormalizeDefaults() {
	if p.AllowBalancePay == nil {
		p.AllowBalancePay = common.GetPointer(true)
	}
}

// Subscription order (payment -> webhook -> wallet credit)
type SubscriptionOrder struct {
	Id     int     `json:"id"`
	UserId int     `json:"user_id" gorm:"index"`
	PlanId int     `json:"plan_id" gorm:"index"`
	Money  float64 `json:"money"`

	TradeNo         string `json:"trade_no" gorm:"unique;type:varchar(255);index"`
	PaymentMethod   string `json:"payment_method" gorm:"type:varchar(50)"`
	PaymentProvider string `json:"payment_provider" gorm:"type:varchar(50);default:''"`
	Status          string `json:"status"`
	CreateTime      int64  `json:"create_time"`
	CompleteTime    int64  `json:"complete_time"`

	ProviderPayload string `json:"provider_payload" gorm:"type:text"`

	// PromoCode records the inventory code used by agent-inventory redemption orders.
	// Ordinary redemptions are balance codes and are not stored here.
	PromoCode string `json:"promo_code" gorm:"type:varchar(64);default:''"`
}

// subscriptionPaidHook 订阅付费成功钩子（pending→success 真正翻转后异步触发）。
var subscriptionPaidHook func(userId int)

// RegisterSubscriptionPaidHook 注册订阅付费成功钩子（在 main 启动时注册，避免 model→service 依赖）。
func RegisterSubscriptionPaidHook(f func(userId int)) {
	subscriptionPaidHook = f
}

// FireSubscriptionPaidHook 对外导出付费成功钩子触发，供充值成功路径复用
// （口径对齐：晋升统计含 top_ups，触发点也应覆盖充值成功）。内部异步 + recover。
func FireSubscriptionPaidHook(userId int) {
	fireSubscriptionPaidHook(userId)
}

func fireSubscriptionPaidHook(userId int) {
	hook := subscriptionPaidHook
	if hook == nil || userId <= 0 {
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				common.SysError(fmt.Sprintf("subscription paid hook panic: %v", r))
			}
		}()
		hook(userId)
	}()
}

func (o *SubscriptionOrder) Insert() error {
	if o.CreateTime == 0 {
		o.CreateTime = common.GetTimestamp()
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(o).Error; err != nil {
			return err
		}
		return SyncSubscriptionDistributionOrderTx(tx, o)
	})
}

func (o *SubscriptionOrder) Update() error {
	return DB.Save(o).Error
}

// CreateSubscriptionOrderWithPromoReserve keeps the historical API name used by
// payment controllers. Redemption codes are no longer typed promo codes here;
// orders are created at the plan price. PromoCode is retained for inventory-code
// provenance when that path creates subscription orders directly.
func CreateSubscriptionOrderWithPromoReserve(order *SubscriptionOrder, planPrice float64, minMoney float64) (*Redemption, error) {
	if order == nil {
		return nil, errors.New("order is nil")
	}
	err := DB.Transaction(func(tx *gorm.DB) error {
		order.Money = planPrice
		if minMoney > 0 && order.Money < minMoney {
			return errors.New("订单金额过低")
		}
		if order.CreateTime == 0 {
			order.CreateTime = common.GetTimestamp()
		}
		if err := tx.Create(order).Error; err != nil {
			common.SysError("failed to create subscription order: " + err.Error())
			return errors.New("创建订单失败")
		}
		return SyncSubscriptionDistributionOrderTx(tx, order)
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func GetSubscriptionOrderByTradeNo(tradeNo string) *SubscriptionOrder {
	if tradeNo == "" {
		return nil
	}
	var order SubscriptionOrder
	if err := DB.Where("trade_no = ?", tradeNo).First(&order).Error; err != nil {
		return nil
	}
	return &order
}

// subscriptionTradeNoColumn returns the trade_no column quoted for the active DB.
func subscriptionTradeNoColumn() string {
	if common.UsingPostgreSQL {
		return `"trade_no"`
	}
	return "`trade_no`"
}

// GetSubscriptionPlanById reads a plan through the display cache. Money
// decisions must not use it: within a purchase transaction the price and the
// credited amount have to be read back from the DB, see
// getSubscriptionPlanForUpdateTx.
func GetSubscriptionPlanById(id int) (*SubscriptionPlan, error) {
	if id <= 0 {
		return nil, errors.New("invalid plan id")
	}
	key := subscriptionPlanCacheKey(id)
	if key != "" {
		if cached, found, err := getSubscriptionPlanCache().Get(key); err == nil && found {
			cached.NormalizeDefaults()
			return &cached, nil
		}
	}
	var plan SubscriptionPlan
	if err := DB.Where("id = ?", id).First(&plan).Error; err != nil {
		return nil, err
	}
	plan.NormalizeDefaults()
	_ = getSubscriptionPlanCache().SetWithTTL(key, plan, subscriptionPlanCacheTTL())
	return &plan, nil
}

// getSubscriptionPlanForUpdateTx reads a plan straight from tx, bypassing the
// display cache. Anything that decides how much money moves must go through
// here so a plan edited during the cache TTL cannot be billed at the old price.
func getSubscriptionPlanForUpdateTx(tx *gorm.DB, id int) (*SubscriptionPlan, error) {
	if tx == nil {
		return nil, errors.New("tx is nil")
	}
	if id <= 0 {
		return nil, errors.New("invalid plan id")
	}
	var plan SubscriptionPlan
	if err := tx.Where("id = ?", id).First(&plan).Error; err != nil {
		return nil, err
	}
	plan.NormalizeDefaults()
	return &plan, nil
}

// CountPlanPurchases returns how many successful purchases of planId the user
// already has. Payment controllers call it before creating a pending order so a
// buyer who already hit the cap is rejected before any money is taken.
func CountPlanPurchases(userId int, planId int) (int64, error) {
	return countPlanPurchasesTx(nil, userId, planId)
}

// countPlanPurchasesTx counts a user's successful purchases of a plan. Purchases
// are tracked by subscription orders now that the subscription quota pool is gone.
//
// Inside a transaction the count is a locking read. The users row lock the caller
// already holds only serialises the buyers; it does not refresh what they can see.
// Under MySQL REPEATABLE READ the transaction snapshot is pinned by its first
// plain SELECT, which happens before that lock is taken, so a plain count would
// still read the pre-lock snapshot and wave a second purchase past the cap.
// A locking read always sees the latest committed rows. PostgreSQL takes a fresh
// snapshot per statement and the SQLite dialector drops the clause, so neither is
// affected. tx == nil is the controller pre-check, which stays a plain read on
// purpose: it is best-effort and must not lock rows outside a transaction.
func countPlanPurchasesTx(tx *gorm.DB, userId int, planId int) (int64, error) {
	if userId <= 0 || planId <= 0 {
		return 0, errors.New("invalid userId or planId")
	}
	query := DB
	if tx != nil {
		query = tx.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var count int64
	if err := query.Model(&SubscriptionOrder{}).
		Where("user_id = ? AND plan_id = ? AND status = ?", userId, planId, common.TopUpStatusSuccess).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func getUserGroupByIdTx(tx *gorm.DB, userId int) (string, error) {
	if userId <= 0 {
		return "", errors.New("invalid userId")
	}
	if tx == nil {
		tx = DB
	}
	var group string
	if err := tx.Model(&User{}).Where("id = ?", userId).Select(commonGroupCol).Find(&group).Error; err != nil {
		return "", err
	}
	return group, nil
}

// ensurePlanPurchaseCapTx rejects the purchase when the user already reached the
// plan's per-user purchase cap. It runs after lockUserQuotaTx, so concurrent
// purchases by the same buyer are serialised and the count cannot be stale.
func ensurePlanPurchaseCapTx(tx *gorm.DB, userId int, plan *SubscriptionPlan) error {
	if plan.MaxPurchasePerUser <= 0 {
		return nil
	}
	count, err := countPlanPurchasesTx(tx, userId, plan.Id)
	if err != nil {
		return err
	}
	if count >= int64(plan.MaxPurchasePerUser) {
		return ErrPlanPurchaseCapReached
	}
	return nil
}

// lockUserQuotaTx takes an exclusive lock on the buyer row and returns the
// current wallet balance. Holding it for the whole credit serialises the same
// user's concurrent purchases, which is what keeps the purchase-cap count and
// the balance check from being read-then-write races.
func lockUserQuotaTx(tx *gorm.DB, userId int) (int, error) {
	var user User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id, quota").Where("id = ?", userId).First(&user).Error; err != nil {
		return 0, err
	}
	return user.Quota, nil
}

// walletQuotaCeiling returns the highest balance a plan credit may produce.
// users.quota is declared `type:int`, so deployments whose column is still four
// bytes can lower the ceiling through USER_QUOTA_CEILING; the default keeps the
// full 64-bit range and only guards the addition itself against overflow.
func walletQuotaCeiling() int64 {
	ceiling := int64(common.GetEnvOrDefault("USER_QUOTA_CEILING", math.MaxInt64))
	if ceiling <= 0 {
		return math.MaxInt64
	}
	return ceiling
}

// ensurePlanQuotaFits rejects a credit that would push the wallet past the
// ceiling. Failing here turns an over-range balance into a business error the
// caller can settle, instead of a driver range error that would roll a paid
// order back to pending with the money already taken.
func ensurePlanQuotaFits(currentQuota int, amount int64) error {
	headroom := currentQuota
	if headroom < 0 {
		headroom = 0
	}
	if amount > walletQuotaCeiling()-int64(headroom) {
		return ErrPlanQuotaOverflow
	}
	return nil
}

// applyPlanUpgradeGroupTx moves the user into the plan's group. Wallet quota
// never expires, so the group upgrade is one-way: there is no automatic
// downgrade any more.
func applyPlanUpgradeGroupTx(tx *gorm.DB, userId int, plan *SubscriptionPlan) error {
	upgradeGroup := strings.TrimSpace(plan.UpgradeGroup)
	if upgradeGroup == "" {
		return nil
	}
	currentGroup, err := getUserGroupByIdTx(tx, userId)
	if err != nil {
		return err
	}
	if currentGroup == upgradeGroup {
		return nil
	}
	return tx.Model(&User{}).Where("id = ?", userId).Update("group", upgradeGroup).Error
}

// CreditPlanQuotaTx credits the plan quota into the user's wallet inside tx and
// applies the plan's group upgrade, returning the credited quota. Callers must
// sync the quota cache and fire the paid hook after the transaction commits.
//
// The plan is re-read from tx, so a stale cached price can never decide how much
// money moves. The buyer row is locked first, which serialises the same user's
// concurrent purchases. Every rejection (misconfigured amount, purchase cap,
// wallet ceiling) happens before the first write and is one of the terminal
// sentinels, so a caller that already took money can settle the order instead of
// rolling it back to pending.
func CreditPlanQuotaTx(tx *gorm.DB, userId int, plan *SubscriptionPlan) (int64, error) {
	if tx == nil {
		return 0, errors.New("tx is nil")
	}
	if plan == nil || plan.Id == 0 {
		return 0, errors.New("invalid plan")
	}
	if userId <= 0 {
		return 0, errors.New("invalid user id")
	}
	current, err := lockUserQuotaTx(tx, userId)
	if err != nil {
		return 0, err
	}
	fresh, err := getSubscriptionPlanForUpdateTx(tx, plan.Id)
	if err != nil {
		return 0, err
	}
	if fresh.TotalAmount <= 0 {
		return 0, ErrPlanQuotaInvalid
	}
	if err := ensurePlanPurchaseCapTx(tx, userId, fresh); err != nil {
		return 0, err
	}
	if err := ensurePlanQuotaFits(current, fresh.TotalAmount); err != nil {
		return 0, err
	}
	if err := applyPlanUpgradeGroupTx(tx, userId, fresh); err != nil {
		return 0, err
	}
	if err := tx.Model(&User{}).Where("id = ?", userId).
		Update("quota", gorm.Expr("quota + ?", fresh.TotalAmount)).Error; err != nil {
		return 0, err
	}
	return fresh.TotalAmount, nil
}

// planPurchaseOutcome carries data out of a purchase transaction: cache sync and
// logging must only run once the transaction has committed.
type planPurchaseOutcome struct {
	userId        int
	planTitle     string
	money         float64
	paymentMethod string
	chargedQuota  int64
	creditedQuota int64
	upgradeGroup  string
}

// applyPlanPurchaseSideEffects syncs the quota/group caches and records the
// wallet log after a plan purchase transaction has committed. quotaDelta is the
// net wallet change (credited minus charged).
func applyPlanPurchaseSideEffects(userId int, quotaDelta int64, upgradeGroup string, logMsg string) {
	if quotaDelta != 0 {
		if err := cacheIncrUserQuota(userId, quotaDelta); err != nil {
			common.SysLog("failed to sync user quota cache after plan purchase: " + err.Error())
		}
	}
	if upgradeGroup != "" {
		if err := UpdateUserGroupCache(userId, upgradeGroup); err != nil {
			common.SysLog("failed to sync user group cache after plan purchase: " + err.Error())
		}
	}
	RecordLog(userId, LogTypeTopup, logMsg)
}

// loadPendingSubscriptionOrderTx locks a plan order by trade no and checks it is
// payable. It returns (nil, nil) when the order is already successful, which
// keeps order completion idempotent.
func loadPendingSubscriptionOrderTx(tx *gorm.DB, tradeNo string, expectedPaymentProvider string) (*SubscriptionOrder, error) {
	var order SubscriptionOrder
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(subscriptionTradeNoColumn()+" = ?", tradeNo).First(&order).Error; err != nil {
		return nil, ErrSubscriptionOrderNotFound
	}
	if expectedPaymentProvider != "" && order.PaymentProvider != expectedPaymentProvider {
		return nil, ErrPaymentMethodMismatch
	}
	if order.Status == common.TopUpStatusSuccess {
		return nil, nil
	}
	if order.Status != common.TopUpStatusPending {
		return nil, ErrSubscriptionOrderStatusInvalid
	}
	return &order, nil
}

// applySubscriptionOrderGatewayEvidence stamps the gateway receipt onto the
// order. Both settlement paths need it: the failed one is precisely the order a
// human has to refund by hand, so its provider payload and real payment method
// must be on record instead of being dropped with the callback.
func applySubscriptionOrderGatewayEvidence(order *SubscriptionOrder, providerPayload string, actualPaymentMethod string) {
	if providerPayload != "" {
		order.ProviderPayload = providerPayload
	}
	if actualPaymentMethod != "" && order.PaymentMethod != actualPaymentMethod {
		order.PaymentMethod = actualPaymentMethod
	}
}

// markSubscriptionOrderPaidTx flips the order to success and re-projects it into
// the distribution order table.
func markSubscriptionOrderPaidTx(tx *gorm.DB, order *SubscriptionOrder, providerPayload string, actualPaymentMethod string) error {
	applySubscriptionOrderGatewayEvidence(order, providerPayload, actualPaymentMethod)
	order.Status = common.TopUpStatusSuccess
	order.CompleteTime = common.GetTimestamp()
	if err := tx.Save(order).Error; err != nil {
		return err
	}
	return SyncSubscriptionDistributionOrderTx(tx, order)
}

// failSubscriptionOrderTx settles a paid order that can never be credited into a
// terminal failed state. Committing the failure is deliberate: leaving the order
// pending would hide a taken payment behind an endless gateway retry loop. The
// gateway evidence is written before the status flip, because this row is what a
// human refund has to be reconciled against.
func failSubscriptionOrderTx(tx *gorm.DB, order *SubscriptionOrder, providerPayload string, actualPaymentMethod string) error {
	applySubscriptionOrderGatewayEvidence(order, providerPayload, actualPaymentMethod)
	order.Status = common.TopUpStatusFailed
	order.CompleteTime = common.GetTimestamp()
	if err := tx.Save(order).Error; err != nil {
		return err
	}
	return SyncSubscriptionDistributionOrderTx(tx, order)
}

// CompleteSubscriptionOrder settles a paid plan order (idempotent): inside one
// transaction it credits the plan quota to the buyer's wallet, mirrors the order
// into top_ups and flips the order to success.
// expectedPaymentProvider guards against cross-gateway callback attacks (empty skips the check).
// actualPaymentMethod updates the order's PaymentMethod to reflect the real payment type used (empty skips update).
// When the credit is rejected for good (cap, misconfigured amount, wallet
// ceiling) the order is committed as failed and an alert is logged, because the
// buyer has already been charged and the money now needs a human refund.
func CompleteSubscriptionOrder(tradeNo string, providerPayload string, expectedPaymentProvider string, actualPaymentMethod string) error {
	if tradeNo == "" {
		return errors.New("tradeNo is empty")
	}
	var settlement subscriptionSettlement
	err := DB.Transaction(func(tx *gorm.DB) error {
		order, err := loadPendingSubscriptionOrderTx(tx, tradeNo, expectedPaymentProvider)
		if err != nil || order == nil {
			return err
		}
		return settlePaidOrderTx(tx, order, providerPayload, actualPaymentMethod, &settlement)
	})
	if err != nil {
		return err
	}
	if settlement.rejected != nil {
		common.SysLog(fmt.Sprintf(
			"订阅订单已收款但额度入账被拒，订单已置失败需人工退款 trade_no=%s reason=%s",
			tradeNo, settlement.rejected.Error()))
		return settlement.rejected
	}
	outcome := settlement.outcome
	if outcome == nil {
		return nil
	}
	msg := fmt.Sprintf("档位购买成功，档位: %s，支付金额: %.2f，支付方式: %s，到账额度: %d",
		outcome.planTitle, outcome.money, outcome.paymentMethod, outcome.creditedQuota)
	applyPlanPurchaseSideEffects(outcome.userId, outcome.creditedQuota, outcome.upgradeGroup, msg)
	fireSubscriptionPaidHook(outcome.userId)
	return nil
}

// subscriptionSettlement is what one CompleteSubscriptionOrder transaction
// produced: either a committed purchase outcome, or a terminal rejection whose
// failed-order state was committed alongside it.
type subscriptionSettlement struct {
	outcome  *planPurchaseOutcome
	rejected error
}

// settlePaidOrderTx credits a paid order inside tx and mirrors it into top_ups.
// A terminal rejection is recorded on the settlement and the order is flipped to
// failed, and the transaction still commits: rolling back would leave a buyer
// who has already been charged sitting on a pending order forever.
func settlePaidOrderTx(tx *gorm.DB, order *SubscriptionOrder, providerPayload string, actualPaymentMethod string, settlement *subscriptionSettlement) error {
	// 事务内必须复用 tx 查询，避免占用第二个连接（单连接场景会死锁）
	plan, err := getSubscriptionPlanForUpdateTx(tx, order.PlanId)
	if err != nil {
		return err
	}
	credited, err := CreditPlanQuotaTx(tx, order.UserId, plan)
	if err != nil {
		if !isPlanCreditTerminal(err) {
			return err
		}
		settlement.rejected = err
		return failSubscriptionOrderTx(tx, order, providerPayload, actualPaymentMethod)
	}
	if err := upsertSubscriptionTopUpTx(tx, order, credited, common.QuotaPerUnit); err != nil {
		return err
	}
	if err := markSubscriptionOrderPaidTx(tx, order, providerPayload, actualPaymentMethod); err != nil {
		return err
	}
	settlement.outcome = newOrderPurchaseOutcome(order, plan, credited)
	return nil
}

func newOrderPurchaseOutcome(order *SubscriptionOrder, plan *SubscriptionPlan, creditedQuota int64) *planPurchaseOutcome {
	return &planPurchaseOutcome{
		userId:        order.UserId,
		planTitle:     plan.Title,
		money:         order.Money,
		paymentMethod: order.PaymentMethod,
		creditedQuota: creditedQuota,
		upgradeGroup:  strings.TrimSpace(plan.UpgradeGroup),
	}
}

// planTopUpAmount converts credited quota into the unit stored in TopUp.Amount.
// That column holds a whole-dollar figure everywhere else in the codebase
// (model/topup.go derives quota as Amount * quotaPerUnit), so the raw quota must
// be divided back down before it is mirrored into top_ups.
//
// The column cannot hold fractions, so the rounding rule is: round to nearest
// (half away from zero) and never report zero for a credit that did happen. Plain
// truncation would under-report every tier whose amount is not a whole multiple
// of quotaPerUnit and would write Amount=0 for sub-dollar tiers, which
// ManualCompleteTopUp then rejects as an invalid top-up.
func planTopUpAmount(creditedQuota int64, quotaPerUnit float64) (int64, error) {
	if creditedQuota <= 0 {
		return 0, ErrPlanQuotaInvalid
	}
	if quotaPerUnit <= 0 {
		return 0, errors.New("额度单位配置错误")
	}
	amount := decimal.NewFromInt(creditedQuota).
		Div(decimal.NewFromFloat(quotaPerUnit)).
		Round(0).
		IntPart()
	if amount <= 0 {
		return 1, nil
	}
	return amount, nil
}

// upsertSubscriptionTopUpTx mirrors a paid plan order into top_ups so wallet
// history and finance reports see it. creditedQuota is the quota that actually
// landed in the buyer's wallet; it is stored as the equivalent dollar amount.
func upsertSubscriptionTopUpTx(tx *gorm.DB, order *SubscriptionOrder, creditedQuota int64, quotaPerUnit float64) error {
	if tx == nil || order == nil {
		return errors.New("invalid subscription order")
	}
	amount, err := planTopUpAmount(creditedQuota, quotaPerUnit)
	if err != nil {
		return err
	}
	now := common.GetTimestamp()
	var topup TopUp
	if err := tx.Where("trade_no = ?", order.TradeNo).First(&topup).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			topup = TopUp{
				UserId:        order.UserId,
				Amount:        amount,
				Money:         order.Money,
				TradeNo:       order.TradeNo,
				PaymentMethod: order.PaymentMethod,
				CreateTime:    order.CreateTime,
				CompleteTime:  now,
				Status:        common.TopUpStatusSuccess,
			}
			return tx.Create(&topup).Error
		}
		return err
	}
	topup.Amount = amount
	topup.Money = order.Money
	if topup.PaymentMethod == "" {
		topup.PaymentMethod = order.PaymentMethod
	} else if topup.PaymentMethod != order.PaymentMethod {
		return ErrPaymentMethodMismatch
	}
	if topup.CreateTime == 0 {
		topup.CreateTime = order.CreateTime
	}
	topup.CompleteTime = now
	topup.Status = common.TopUpStatusSuccess
	return tx.Save(&topup).Error
}

func ExpireSubscriptionOrder(tradeNo string, expectedPaymentProvider string) error {
	if tradeNo == "" {
		return errors.New("tradeNo is empty")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		var order SubscriptionOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(subscriptionTradeNoColumn()+" = ?", tradeNo).First(&order).Error; err != nil {
			return ErrSubscriptionOrderNotFound
		}
		if expectedPaymentProvider != "" && order.PaymentProvider != expectedPaymentProvider {
			return ErrPaymentMethodMismatch
		}
		if order.Status != common.TopUpStatusPending {
			return nil
		}
		order.Status = common.TopUpStatusExpired
		order.CompleteTime = common.GetTimestamp()
		if err := tx.Save(&order).Error; err != nil {
			return err
		}
		return SyncSubscriptionDistributionOrderTx(tx, &order)
	})
}

// AdminBindSubscription credits a plan's quota to a user without payment.
// No order is recorded, so admin grants stay out of the paid-purchase statistics.
func AdminBindSubscription(userId int, planId int, sourceNote string) (string, error) {
	if userId <= 0 || planId <= 0 {
		return "", errors.New("invalid userId or planId")
	}
	plan, err := GetSubscriptionPlanById(planId)
	if err != nil {
		return "", err
	}
	var creditedQuota int64
	err = DB.Transaction(func(tx *gorm.DB) error {
		credited, txErr := CreditPlanQuotaTx(tx, userId, plan)
		if txErr != nil {
			return txErr
		}
		creditedQuota = credited
		return nil
	})
	if err != nil {
		return "", err
	}
	upgradeGroup := strings.TrimSpace(plan.UpgradeGroup)
	msg := fmt.Sprintf("管理员发放档位额度，档位: %s，到账额度: %d，来源: %s", plan.Title, creditedQuota, sourceNote)
	applyPlanPurchaseSideEffects(userId, creditedQuota, upgradeGroup, msg)
	if upgradeGroup != "" {
		return fmt.Sprintf("用户分组将升级到 %s", upgradeGroup), nil
	}
	return "", nil
}

// calcSubscriptionBalanceQuota converts a plan price into the wallet quota it
// costs. quotaPerUnit is injected by the caller so the conversion stays testable
// at its boundaries instead of reading the global config.
func calcSubscriptionBalanceQuota(priceAmount float64, quotaPerUnit float64) (int, error) {
	if priceAmount <= 0 {
		return 0, nil
	}
	if quotaPerUnit <= 0 {
		return 0, errors.New("额度单位配置错误")
	}
	quota := decimal.NewFromFloat(priceAmount).
		Mul(decimal.NewFromFloat(quotaPerUnit)).
		Ceil().
		IntPart()
	return int(quota), nil
}

// loadBalancePayablePlanTx loads a plan and asserts it can be bought with wallet balance.
func loadBalancePayablePlanTx(tx *gorm.DB, planId int) (*SubscriptionPlan, error) {
	plan, err := getSubscriptionPlanForUpdateTx(tx, planId)
	if err != nil {
		return nil, err
	}
	if !plan.Enabled {
		return nil, errors.New("套餐未启用")
	}
	if plan.PriceAmount < 0 {
		return nil, errors.New("套餐价格不能为负数")
	}
	if plan.AllowBalancePay != nil && !*plan.AllowBalancePay {
		return nil, errors.New("该套餐不允许使用余额兑换")
	}
	return plan, nil
}

// chargeWalletForPlanTx takes an exclusive lock on the buyer row (a real
// SELECT ... FOR UPDATE via clause.Locking — gorm:query_option is a no-op under
// GORM v2) and deducts the plan price from the wallet, returning the deducted
// quota. The lock is what makes the balance check safe against concurrent
// purchases by the same user.
func chargeWalletForPlanTx(tx *gorm.DB, userId int, plan *SubscriptionPlan, quotaPerUnit float64) (int, error) {
	requiredQuota, err := calcSubscriptionBalanceQuota(plan.PriceAmount, quotaPerUnit)
	if err != nil {
		return 0, err
	}
	currentQuota, err := lockUserQuotaTx(tx, userId)
	if err != nil {
		return 0, err
	}
	if requiredQuota <= 0 {
		return 0, nil
	}
	if currentQuota < requiredQuota {
		return 0, errors.New("余额不足")
	}
	if err := tx.Model(&User{}).Where("id = ?", userId).
		Update("quota", gorm.Expr("quota - ?", requiredQuota)).Error; err != nil {
		return 0, err
	}
	return requiredQuota, nil
}

// createBalancePurchaseOrderTx records the wallet purchase as a successful order
// so distribution projection and paid-user statistics keep seeing it.
func createBalancePurchaseOrderTx(tx *gorm.DB, userId int, plan *SubscriptionPlan, chargedQuota int) error {
	now := common.GetTimestamp()
	order := &SubscriptionOrder{
		UserId:          userId,
		PlanId:          plan.Id,
		Money:           plan.PriceAmount,
		TradeNo:         fmt.Sprintf("SUBBALUSR%dNO%s%d", userId, common.GetRandomString(6), time.Now().UnixNano()),
		PaymentMethod:   PaymentMethodBalance,
		PaymentProvider: PaymentProviderBalance,
		Status:          common.TopUpStatusSuccess,
		CreateTime:      now,
		CompleteTime:    now,
		ProviderPayload: fmt.Sprintf("charged_quota=%d", chargedQuota),
	}
	if err := tx.Create(order).Error; err != nil {
		return err
	}
	return SyncSubscriptionDistributionOrderTx(tx, order)
}

// newBalancePurchaseOutcome mirrors newOrderPurchaseOutcome for the wallet
// purchase path, where the buyer is charged as well as credited.
func newBalancePurchaseOutcome(userId int, plan *SubscriptionPlan, chargedQuota int, creditedQuota int64) *planPurchaseOutcome {
	return &planPurchaseOutcome{
		userId:        userId,
		planTitle:     plan.Title,
		money:         plan.PriceAmount,
		paymentMethod: PaymentMethodBalance,
		chargedQuota:  int64(chargedQuota),
		creditedQuota: creditedQuota,
		upgradeGroup:  strings.TrimSpace(plan.UpgradeGroup),
	}
}

// PurchaseSubscriptionWithBalance exchanges wallet quota for a plan's quota at
// the plan price (a discounted top-up). The deduction and the credit share one
// transaction, so the net wallet change is exactly
// plan.TotalAmount - calcSubscriptionBalanceQuota(plan.PriceAmount, quotaPerUnit).
// promoCode is an unused compatibility parameter: ordinary redemption codes are
// no longer promo discounts, and plan orders are always created at list price.
func PurchaseSubscriptionWithBalance(userId int, planId int, promoCode string) error {
	if userId <= 0 || planId <= 0 {
		return errors.New("invalid userId or planId")
	}
	var outcome *planPurchaseOutcome
	err := DB.Transaction(func(tx *gorm.DB) error {
		plan, err := loadBalancePayablePlanTx(tx, planId)
		if err != nil {
			return err
		}
		chargedQuota, err := chargeWalletForPlanTx(tx, userId, plan, common.QuotaPerUnit)
		if err != nil {
			return err
		}
		creditedQuota, err := CreditPlanQuotaTx(tx, userId, plan)
		if err != nil {
			return err
		}
		if err := createBalancePurchaseOrderTx(tx, userId, plan, chargedQuota); err != nil {
			return err
		}
		outcome = newBalancePurchaseOutcome(userId, plan, chargedQuota, creditedQuota)
		return nil
	})
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("使用余额兑换档位成功，档位: %s，支付金额: %.2f，扣除额度: %d，到账额度: %d",
		outcome.planTitle, outcome.money, outcome.chargedQuota, outcome.creditedQuota)
	applyPlanPurchaseSideEffects(userId, outcome.creditedQuota-outcome.chargedQuota, outcome.upgradeGroup, msg)
	fireSubscriptionPaidHook(userId)
	return nil
}
