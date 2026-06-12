package service

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/require"
)

func TestRedeemDistributionInventoryCode_AlreadyRedeemed(t *testing.T) {
	truncate(t)
	t.Cleanup(func() {
		model.DB.Exec("DELETE FROM p3_inventories")
	})

	now := time.Now().Unix()
	inventory := &model.DistributionInventory{
		AgentId:     1,
		OrderId:     1,
		PackageId:   1,
		Status:      DistributionInventoryStatusRedeemed,
		InventoryNo: "INVUSED01",
		AssignedTo:  3001,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, model.DB.Create(inventory).Error)

	// 已兑换的库存码再次兑换必须返回 sentinel，供 controller 提示"已被使用"
	_, err := RedeemDistributionInventoryCode(3001, "INVUSED01")
	require.ErrorIs(t, err, ErrInventoryAlreadyRedeemed)
}
