package controller

import (
	"sort"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

const opsDaySeconds int64 = 86400

// parseDays reads the ?days= query param, defaulting to 30 and clamping to
// [1, 365]. Returns the resolved day count plus the [start, end) window.
func parseOpsWindow(c *gin.Context) (days int, start int64, end int64) {
	days = 30
	if v := c.Query("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			days = n
		}
	}
	if days < 1 {
		days = 1
	}
	if days > 365 {
		days = 365
	}
	end = common.GetTimestamp()
	start = end - int64(days)*opsDaySeconds
	return
}

// quotaToUSD converts internal quota units to a USD figure for display.
func quotaToUSD(quota int64) float64 {
	if common.QuotaPerUnit == 0 {
		return 0
	}
	return float64(quota) / common.QuotaPerUnit
}

// GetOpsOverview returns the headline KPIs for the operations console.
func GetOpsOverview(c *gin.Context) {
	days, start, end := parseOpsWindow(c)

	topups, err := model.GetTopUpsInRange(start, end)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	var revenue float64
	var successOrders, totalOrders int
	for _, r := range topups {
		totalOrders++
		if r.Status == common.TopUpStatusSuccess {
			successOrders++
			revenue += r.Money
		}
	}
	successRate := 0.0
	if totalOrders > 0 {
		successRate = float64(successOrders) / float64(totalOrders)
	}

	subOrders, err := model.GetSubscriptionOrdersInRange(start, end)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	var subRevenue float64
	var subSuccessOrders, subTotalOrders int
	for _, r := range subOrders {
		subTotalOrders++
		if r.Status == common.TopUpStatusSuccess {
			subSuccessOrders++
			subRevenue += r.Money
		}
	}

	newUserTimes, err := model.GetUserCreatedTimesInRange(start, end)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	totalUsers, enabledUsers, err := model.CountUsers()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	activeUsers, err := model.CountActiveUsers(start, end)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	consume, err := model.SumConsumptionInRange(start, end)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	channels, err := model.GetChannelHealth()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// "Online" ≈ distinct users with a consumption log in the last 15 minutes
	// (no socket/heartbeat layer exists; this is the closest near-real-time
	// signal). Labeled accordingly on the frontend.
	onlineUsers, err := model.CountActiveUsersSince(end - 15*60)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, gin.H{
		"days":                 days,
		"start_timestamp":      start,
		"end_timestamp":        end,
		"revenue":              revenue,
		"order_total":          totalOrders,
		"order_success":        successOrders,
		"order_success_rate":   successRate,
		"subscription_revenue": subRevenue,
		"subscription_orders":  subTotalOrders,
		"subscription_success": subSuccessOrders,
		"total_revenue":        revenue + subRevenue,
		"new_users":            len(newUserTimes),
		"total_users":          totalUsers,
		"enabled_users":        enabledUsers,
		"active_users":         activeUsers,
		"online_users":         onlineUsers,
		"consumption_quota":    consume.Quota,
		"consumption_usd":      quotaToUSD(consume.Quota),
		"request_count":        consume.Requests,
		"token_count":          consume.Tokens,
		"channel":              channels,
	})
}

// GetOpsRevenueTrend returns per-day revenue + order counts within the window.
func GetOpsRevenueTrend(c *gin.Context) {
	days, start, end := parseOpsWindow(c)
	topups, err := model.GetTopUpsInRange(start, end)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	subOrders, err := model.GetSubscriptionOrdersInRange(start, end)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// Platform revenue per day = recharge (top-up) + subscription sales.
	type bucket struct {
		Date         int64   `json:"date"`
		Revenue      float64 `json:"revenue"`
		TopupRevenue float64 `json:"topup_revenue"`
		SubRevenue   float64 `json:"subscription_revenue"`
		OrderCount   int     `json:"order_count"`
		SuccessCount int     `json:"success_count"`
	}
	buckets := make(map[int64]*bucket)
	getBucket := func(ts int64) *bucket {
		day := (ts / opsDaySeconds) * opsDaySeconds
		b := buckets[day]
		if b == nil {
			b = &bucket{Date: day}
			buckets[day] = b
		}
		return b
	}
	for _, r := range topups {
		b := getBucket(r.CreateTime)
		b.OrderCount++
		if r.Status == common.TopUpStatusSuccess {
			b.SuccessCount++
			b.TopupRevenue += r.Money
			b.Revenue += r.Money
		}
	}
	for _, r := range subOrders {
		b := getBucket(r.CreateTime)
		b.OrderCount++
		if r.Status == common.TopUpStatusSuccess {
			b.SuccessCount++
			b.SubRevenue += r.Money
			b.Revenue += r.Money
		}
	}

	// Emit a continuous series so the chart has no gaps.
	startDay := (start / opsDaySeconds) * opsDaySeconds
	endDay := (end / opsDaySeconds) * opsDaySeconds
	series := make([]*bucket, 0, days+1)
	for d := startDay; d <= endDay; d += opsDaySeconds {
		if b := buckets[d]; b != nil {
			series = append(series, b)
		} else {
			series = append(series, &bucket{Date: d})
		}
	}

	common.ApiSuccess(c, gin.H{
		"days":            days,
		"start_timestamp": start,
		"end_timestamp":   end,
		"series":          series,
	})
}

// GetOpsUserGrowth returns per-day new-user counts plus a cumulative curve.
func GetOpsUserGrowth(c *gin.Context) {
	days, start, end := parseOpsWindow(c)
	times, err := model.GetUserCreatedTimesInRange(start, end)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	baseline, err := model.CountUsersCreatedBefore(start)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	dayCount := make(map[int64]int)
	for _, t := range times {
		day := (t / opsDaySeconds) * opsDaySeconds
		dayCount[day]++
	}

	type point struct {
		Date       int64 `json:"date"`
		NewUsers   int   `json:"new_users"`
		Cumulative int64 `json:"cumulative"`
	}
	startDay := (start / opsDaySeconds) * opsDaySeconds
	endDay := (end / opsDaySeconds) * opsDaySeconds
	series := make([]point, 0, days+1)
	cumulative := baseline
	for d := startDay; d <= endDay; d += opsDaySeconds {
		n := dayCount[d]
		cumulative += int64(n)
		series = append(series, point{Date: d, NewUsers: n, Cumulative: cumulative})
	}

	common.ApiSuccess(c, gin.H{
		"days":            days,
		"start_timestamp": start,
		"end_timestamp":   end,
		"baseline":        baseline,
		"series":          series,
	})
}

// GetOpsPaymentProviders aggregates revenue / order counts per payment provider.
func GetOpsPaymentProviders(c *gin.Context) {
	days, start, end := parseOpsWindow(c)
	topups, err := model.GetTopUpsInRange(start, end)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	type providerStat struct {
		Provider     string  `json:"provider"`
		Revenue      float64 `json:"revenue"`
		OrderTotal   int     `json:"order_total"`
		OrderSuccess int     `json:"order_success"`
		SuccessRate  float64 `json:"success_rate"`
	}
	stats := make(map[string]*providerStat)
	for _, r := range topups {
		key := r.PaymentProvider
		if key == "" {
			key = r.PaymentMethod
		}
		if key == "" {
			key = "unknown"
		}
		s := stats[key]
		if s == nil {
			s = &providerStat{Provider: key}
			stats[key] = s
		}
		s.OrderTotal++
		if r.Status == common.TopUpStatusSuccess {
			s.OrderSuccess++
			s.Revenue += r.Money
		}
	}

	list := make([]*providerStat, 0, len(stats))
	for _, s := range stats {
		if s.OrderTotal > 0 {
			s.SuccessRate = float64(s.OrderSuccess) / float64(s.OrderTotal)
		}
		list = append(list, s)
	}
	// Stable order: highest revenue first.
	sort.Slice(list, func(i, j int) bool { return list[i].Revenue > list[j].Revenue })

	common.ApiSuccess(c, gin.H{
		"days":            days,
		"start_timestamp": start,
		"end_timestamp":   end,
		"providers":       list,
	})
}

// GetOpsPlanSales returns per-subscription-plan sales counts + revenue,
// ordered by popularity (sales volume).
func GetOpsPlanSales(c *gin.Context) {
	days, start, end := parseOpsWindow(c)
	plans, err := model.GetSubscriptionPlanSales(start, end)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	var totalSales int64
	var totalRevenue float64
	for _, p := range plans {
		totalSales += p.SalesCount
		totalRevenue += p.Revenue
	}
	common.ApiSuccess(c, gin.H{
		"days":            days,
		"start_timestamp": start,
		"end_timestamp":   end,
		"plans":           plans,
		"total_sales":     totalSales,
		"total_revenue":   totalRevenue,
	})
}
