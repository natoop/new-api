package model

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// GetRedemptionByKey 按兑换码 key 查询，未命中返回 (nil, nil)；
// 用 map 条件让 GORM 按方言给保留字 key 加引号，兼容 SQLite/MySQL/PG 三库
func GetRedemptionByKey(key string) (*Redemption, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, nil
	}
	var redemption Redemption
	err := DB.Where(map[string]any{"key": key}).First(&redemption).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &redemption, nil
}
