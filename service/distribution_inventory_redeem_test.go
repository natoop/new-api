package service

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
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

// seedRedeemableInventory 铺一条可兑换的库存码及其档位/套餐/代理，返回档位。
func seedRedeemableInventory(t *testing.T, userID int, code string) *model.SubscriptionPlan {
	t.Helper()
	now := time.Now().Unix()

	plan := &model.SubscriptionPlan{
		Id:           7001,
		Title:        "Inventory Plan",
		Currency:     "USD",
		PriceAmount:  20,
		TotalAmount:  4_000_000,
		DurationUnit: "month", DurationValue: 1,
		Enabled: true,
	}
	require.NoError(t, model.DB.Create(plan).Error)
	model.InvalidateSubscriptionPlanCache(plan.Id)

	pkg := &model.DistributionPackage{
		Id:                 7001,
		SubscriptionPlanId: plan.Id,
		Name:               "Inventory Package",
		Sku:                "SKU-INV-1",
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	require.NoError(t, model.DB.Create(pkg).Error)

	agent := &model.DistributionAgent{Id: 7001, UserId: 9001, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, model.DB.Create(agent).Error)

	// aff_code 有唯一索引，代理与买家不能共用空串
	agentUser := &model.User{Id: 9001, Username: "inv_agent", Status: common.UserStatusEnabled, AffCode: "invagent"}
	require.NoError(t, model.DB.Create(agentUser).Error)

	inventory := &model.DistributionInventory{
		AgentId:     agent.Id,
		OrderId:     7001,
		PackageId:   pkg.Id,
		Status:      DistributionInventoryStatusAvailable,
		InventoryNo: code,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, model.DB.Create(inventory).Error)

	buyer := &model.User{Id: userID, Username: "inv_buyer", Status: common.UserStatusEnabled, AffCode: "invbuyer"}
	require.NoError(t, model.DB.Create(buyer).Error)
	return plan
}

func TestRedeemDistributionInventoryCode_CreditsWalletAndRecordsOrder(t *testing.T) {
	truncate(t)
	t.Cleanup(func() {
		model.DB.Exec("DELETE FROM p3_inventories")
	})

	const userID = 3002
	plan := seedRedeemableInventory(t, userID, "INVOK0001")

	result, err := RedeemDistributionInventoryCode(userID, "INVOK0001")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Matched)

	// 额度真的进了钱包，且对外透出的到账额度与档位一致
	assert.Equal(t, plan.TotalAmount, result.CreditedQuota)
	assert.Equal(t, int(plan.TotalAmount), getUserQuota(t, userID))
	assert.Equal(t, plan.Id, result.PlanId)
	assert.Equal(t, plan.Title, result.PlanTitle)

	// 落一笔成功订单，支付方式必须是库存码，否则代理台账与付费统计对不上
	var order model.SubscriptionOrder
	require.NoError(t, model.DB.Where("user_id = ?", userID).First(&order).Error)
	assert.Equal(t, model.SubscriptionPaymentMethodAgentInventory, order.PaymentMethod)
	assert.Equal(t, model.SubscriptionPaymentMethodAgentInventory, order.PaymentProvider)
	assert.Equal(t, common.TopUpStatusSuccess, order.Status)
	assert.Equal(t, "INVOK0001", order.PromoCode)

	// 库存状态转已兑换并归属到兑换人
	var inventory model.DistributionInventory
	require.NoError(t, model.DB.Where("inventory_no = ?", "INVOK0001").First(&inventory).Error)
	assert.Equal(t, DistributionInventoryStatusRedeemed, inventory.Status)
	assert.Equal(t, userID, inventory.AssignedTo)
}
