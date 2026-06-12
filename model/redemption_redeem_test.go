package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertUserForRedeemTest(t *testing.T, id int, quota int) {
	t.Helper()
	user := &User{
		Id:       id,
		Username: "redeem_user",
		Status:   common.UserStatusEnabled,
		Quota:    quota,
	}
	require.NoError(t, DB.Create(user).Error)
}

func insertRedemptionForRedeemTest(t *testing.T, redemption *Redemption) *Redemption {
	t.Helper()
	if redemption.Status == 0 {
		redemption.Status = common.RedemptionCodeStatusEnabled
	}
	if redemption.CreatedTime == 0 {
		redemption.CreatedTime = common.GetTimestamp()
	}
	require.NoError(t, DB.Create(redemption).Error)
	return redemption
}

func TestRedeem_BalanceCode(t *testing.T) {
	truncateTables(t)

	insertUserForRedeemTest(t, 501, 100)
	insertRedemptionForRedeemTest(t, &Redemption{
		Key:   "11111111111111111111111111111111",
		Name:  "balance code",
		Quota: 5000,
	})

	quota, err := Redeem("11111111111111111111111111111111", 501)
	require.NoError(t, err)
	assert.Equal(t, 5000, quota)

	var user User
	require.NoError(t, DB.Where("id = ?", 501).First(&user).Error)
	assert.Equal(t, 5100, user.Quota)

	var redemption Redemption
	require.NoError(t, DB.Where("used_user_id = ?", 501).First(&redemption).Error)
	assert.Equal(t, common.RedemptionCodeStatusUsed, redemption.Status)
	assert.Equal(t, 501, redemption.UsedUserId)

	// 二次兑换必须失败（防重放）
	_, err = Redeem("11111111111111111111111111111111", 501)
	require.ErrorIs(t, err, ErrRedeemFailed)
}

func TestRedeem_ExpiredCodeRejected(t *testing.T) {
	truncateTables(t)

	insertUserForRedeemTest(t, 502, 0)
	insertRedemptionForRedeemTest(t, &Redemption{
		Key:         "22222222222222222222222222222222",
		Name:        "expired code",
		Quota:       5000,
		ExpiredTime: common.GetTimestamp() - 60,
	})

	_, err := Redeem("22222222222222222222222222222222", 502)
	require.ErrorIs(t, err, ErrRedeemFailed)

	// 余额不变、码未被标记使用
	var user User
	require.NoError(t, DB.Where("id = ?", 502).First(&user).Error)
	assert.Equal(t, 0, user.Quota)

	var redemption Redemption
	require.NoError(t, DB.Where("name = ?", "expired code").First(&redemption).Error)
	assert.Equal(t, common.RedemptionCodeStatusEnabled, redemption.Status)
}
