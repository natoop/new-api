package service

import (
	"errors"

	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

func ValidatePromoCode(code string, planId int) (*model.Redemption, error) {
	return nil, errors.New("优惠码功能已禁用")
}

func ApplyPromoDiscount(amount float64, discountBps int) float64 {
	return amount
}

func ConsumePromoCode(tx *gorm.DB, redemptionId int) error {
	return errors.New("优惠码功能已禁用")
}
