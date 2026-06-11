package service

import (
	"strings"

	"github.com/QuantumNous/new-api/model"
)

// escapeLikeKeyword 转义用户输入中的 LIKE 通配符（% 和 _），
// 使用 ! 作为 ESCAPE 字符（兼容 MySQL/PostgreSQL/SQLite，且显式 ESCAPE 后
// 反斜杠在三库中均为普通字符，不再具有转义语义）。
func escapeLikeKeyword(s string) string {
	s = strings.ReplaceAll(s, "!", "!!")
	s = strings.ReplaceAll(s, "%", "!%")
	s = strings.ReplaceAll(s, "_", "!_")
	return s
}

type DistributionOrderListInput struct {
	Keyword   string
	PlanId    int
	Status    string
	StartTime int64
	EndTime   int64
	StartIdx  int
	PageSize  int
}

func AdminListDistributionOrders(input DistributionOrderListInput) ([]model.DistributionOrder, int64, error) {
	input.Keyword = strings.TrimSpace(input.Keyword)
	input.Status = strings.TrimSpace(input.Status)
	var orders []model.DistributionOrder
	query := model.DB.Model(&model.DistributionOrder{})
	if input.Keyword != "" {
		like := "%" + escapeLikeKeyword(input.Keyword) + "%"
		query = query.Where("buyer_username LIKE ? ESCAPE '!' OR buyer_email LIKE ? ESCAPE '!' OR buyer_display_name LIKE ? ESCAPE '!'", like, like, like)
	}
	if input.PlanId > 0 {
		query = query.Where("subscription_plan_id = ?", input.PlanId)
	}
	if input.Status != "" {
		query = query.Where("status = ?", input.Status)
	}
	if input.StartTime > 0 {
		query = query.Where("created_at >= ?", input.StartTime)
	}
	if input.EndTime > 0 {
		query = query.Where("created_at <= ?", input.EndTime)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id desc").Limit(input.PageSize).Offset(input.StartIdx).Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}
