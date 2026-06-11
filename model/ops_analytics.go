package model

import (
	"github.com/QuantumNous/new-api/common"
)

// Ops analytics aggregation helpers.
//
// Cross-DB note (Rule 2): these queries deliberately avoid DB-specific date
// functions (FROM_UNIXTIME / date_trunc / strftime). Daily bucketing is done
// in Go after fetching only the lightweight columns inside a bounded time
// window, so the same code path works on SQLite / MySQL / PostgreSQL.

// TopUpStatRow is a lightweight projection of a top-up order used for ops
// revenue / order aggregation. Windowed by create_time.
type TopUpStatRow struct {
	CreateTime      int64   `json:"create_time"`
	CompleteTime    int64   `json:"complete_time"`
	Money           float64 `json:"money"`
	PaymentProvider string  `json:"payment_provider"`
	PaymentMethod   string  `json:"payment_method"`
	Status          string  `json:"status"`
}

// GetTopUpsInRange returns top-up orders created within [start, end).
func GetTopUpsInRange(start, end int64) ([]TopUpStatRow, error) {
	rows := make([]TopUpStatRow, 0)
	err := DB.Model(&TopUp{}).
		Select("create_time", "complete_time", "money", "payment_provider", "payment_method", "status").
		Where("create_time >= ? AND create_time < ?", start, end).
		Find(&rows).Error
	return rows, err
}

// GetUserCreatedTimesInRange returns created_at timestamps for users registered
// within [start, end).
func GetUserCreatedTimesInRange(start, end int64) ([]int64, error) {
	times := make([]int64, 0)
	err := DB.Model(&User{}).
		Where("created_at >= ? AND created_at < ?", start, end).
		Pluck("created_at", &times).Error
	return times, err
}

// CountUsersCreatedBefore returns how many users registered strictly before ts.
// Used as the cumulative baseline for the user-growth curve.
func CountUsersCreatedBefore(ts int64) (int64, error) {
	var n int64
	err := DB.Model(&User{}).Where("created_at < ?", ts).Count(&n).Error
	return n, err
}

// CountUsers returns total and enabled user counts.
func CountUsers() (total int64, enabled int64, err error) {
	if err = DB.Model(&User{}).Count(&total).Error; err != nil {
		return
	}
	err = DB.Model(&User{}).Where("status = ?", common.UserStatusEnabled).Count(&enabled).Error
	return
}

// CountActiveUsers returns the number of distinct users with consumption logs
// within [start, end).
func CountActiveUsers(start, end int64) (int64, error) {
	var n int64
	err := LOG_DB.Model(&Log{}).
		Where("type = ? AND created_at >= ? AND created_at < ?", LogTypeConsume, start, end).
		Distinct("user_id").Count(&n).Error
	return n, err
}

// ConsumeStat is the windowed consumption rollup for the ops console.
type ConsumeStat struct {
	Quota    int64 `json:"quota"`
	Requests int64 `json:"requests"`
	Tokens   int64 `json:"tokens"`
}

// SumConsumptionInRange returns total quota / request count / token count from
// consumption logs within [start, end). Unlike model.SumUsedQuota (whose
// rpm/tpm are a trailing-60s instantaneous rate), every figure here is the
// true window total. COALESCE keeps it cross-DB (SQLite/MySQL/PostgreSQL).
func SumConsumptionInRange(start, end int64) (ConsumeStat, error) {
	var s ConsumeStat
	err := LOG_DB.Model(&Log{}).
		Where("type = ? AND created_at >= ? AND created_at < ?", LogTypeConsume, start, end).
		Select("COALESCE(sum(quota),0) as quota, count(*) as requests, COALESCE(sum(prompt_tokens),0) + COALESCE(sum(completion_tokens),0) as tokens").
		Scan(&s).Error
	return s, err
}

// SubscriptionOrderRow is a lightweight projection of a subscription-plan
// purchase, windowed by create_time.
type SubscriptionOrderRow struct {
	CreateTime int64   `json:"create_time"`
	Money      float64 `json:"money"`
	PlanId     int     `json:"plan_id"`
	Status     string  `json:"status"`
}

// GetSubscriptionOrdersInRange returns subscription orders created within
// [start, end).
func GetSubscriptionOrdersInRange(start, end int64) ([]SubscriptionOrderRow, error) {
	rows := make([]SubscriptionOrderRow, 0)
	err := DB.Model(&SubscriptionOrder{}).
		Select("create_time", "money", "plan_id", "status").
		Where("create_time >= ? AND create_time < ?", start, end).
		Find(&rows).Error
	return rows, err
}

// PlanSales is the per-plan sales rollup for the popularity ranking.
type PlanSales struct {
	PlanId     int     `json:"plan_id"`
	Title      string  `json:"title"`
	SalesCount int64   `json:"sales_count"`
	Revenue    float64 `json:"revenue"`
}

// GetSubscriptionPlanSales aggregates successful subscription orders per plan,
// joined to plan titles, ordered by sales volume. COALESCE keeps it cross-DB.
func GetSubscriptionPlanSales(start, end int64) ([]PlanSales, error) {
	rows := make([]PlanSales, 0)
	err := DB.Table("subscription_orders as so").
		Select("sp.id as plan_id, sp.title as title, COUNT(*) as sales_count, COALESCE(SUM(so.money),0) as revenue").
		Joins("JOIN subscription_plans sp ON sp.id = so.plan_id").
		Where("so.status = ? AND so.create_time >= ? AND so.create_time < ?", common.TopUpStatusSuccess, start, end).
		Group("sp.id, sp.title").
		Order("sales_count DESC").
		Scan(&rows).Error
	return rows, err
}

// CountActiveUsersSince returns the number of distinct users with consumption
// logs since ts — used as a near-real-time "online users" approximation.
func CountActiveUsersSince(ts int64) (int64, error) {
	var n int64
	err := LOG_DB.Model(&Log{}).
		Where("type = ? AND created_at >= ?", LogTypeConsume, ts).
		Distinct("user_id").Count(&n).Error
	return n, err
}

// ChannelHealth captures channel status distribution for the ops console.
type ChannelHealth struct {
	Total            int64 `json:"total"`
	Enabled          int64 `json:"enabled"`
	ManuallyDisabled int64 `json:"manually_disabled"`
	AutoDisabled     int64 `json:"auto_disabled"`
}

// GetChannelHealth returns the channel status distribution.
func GetChannelHealth() (ChannelHealth, error) {
	var h ChannelHealth
	var err error
	if err = DB.Model(&Channel{}).Count(&h.Total).Error; err != nil {
		return h, err
	}
	if err = DB.Model(&Channel{}).Where("status = ?", common.ChannelStatusEnabled).Count(&h.Enabled).Error; err != nil {
		return h, err
	}
	if err = DB.Model(&Channel{}).Where("status = ?", common.ChannelStatusManuallyDisabled).Count(&h.ManuallyDisabled).Error; err != nil {
		return h, err
	}
	err = DB.Model(&Channel{}).Where("status = ?", common.ChannelStatusAutoDisabled).Count(&h.AutoDisabled).Error
	return h, err
}
