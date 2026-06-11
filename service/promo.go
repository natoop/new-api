package service

import (
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

// ValidatePromoCode 校验优惠码是否可用于指定套餐：
// 存在、type=promo、enabled、未过期、未达 MaxUses、（可选）套餐绑定匹配。
// 仅做读校验，不消耗次数；消耗在支付成功路径内原子完成。
func ValidatePromoCode(code string, planId int) (*model.Redemption, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, errors.New("未提供优惠码")
	}
	redemption, err := model.GetRedemptionByKey(code)
	if err != nil {
		return nil, errors.New("优惠码不存在")
	}
	if err := redemption.ValidatePromoUsable(planId); err != nil {
		return nil, err
	}
	return redemption, nil
}

// ApplyPromoDiscount 计算折后金额。金额折算逻辑收口在 model.ApplyPromoDiscount
// （余额购买路径在 model 层执行，model 不能 import service，故唯一实现放在 model）。
func ApplyPromoDiscount(amount float64, discountBps int) float64 {
	return model.ApplyPromoDiscount(amount, discountBps)
}

// ConsumePromoCode 原子消耗一次优惠码（WHERE used_count < max_uses OR max_uses = 0）。
// tx 可为 nil（直接使用全局 DB）。
func ConsumePromoCode(tx *gorm.DB, redemptionId int) error {
	return model.ConsumePromoCodeById(tx, redemptionId)
}
