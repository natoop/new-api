package model

import (
	"errors"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

// UserAgreementConsent 用户协议签署记录
type UserAgreementConsent struct {
	Id       int    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId   int    `json:"user_id" gorm:"not null;uniqueIndex:idx_user_agreement_user_version"`
	Version  string `json:"version" gorm:"type:varchar(32);not null;uniqueIndex:idx_user_agreement_user_version"`
	AgreedAt int64  `json:"agreed_at" gorm:"bigint"`
	Ip       string `json:"ip" gorm:"type:varchar(64)"`
}

func (UserAgreementConsent) TableName() string {
	return "user_agreement_consents"
}

// HasUserConsented 检查用户是否已签署指定版本的协议
func HasUserConsented(userId int, version string) (bool, error) {
	var count int64
	err := DB.Model(&UserAgreementConsent{}).
		Where("user_id = ? AND version = ?", userId, version).
		Count(&count).Error
	return count > 0, err
}

// RecordUserConsent 记录用户签署协议，重复签署幂等（唯一键冲突视为成功）
func RecordUserConsent(userId int, version, ip string) error {
	consent := UserAgreementConsent{
		UserId:   userId,
		Version:  version,
		AgreedAt: common.GetTimestamp(),
		Ip:       ip,
	}
	err := DB.Create(&consent).Error
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil
	}
	// 部分驱动未翻译唯一键冲突错误，回查确认是否已存在记录
	consented, checkErr := HasUserConsented(userId, version)
	if checkErr == nil && consented {
		return nil
	}
	return err
}
