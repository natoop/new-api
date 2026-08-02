package model

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// ---------------------------------------------------------------------------
// CreditPlanQuotaTx — 档位额度入账（本轮新增的资金 seam）
// ---------------------------------------------------------------------------

func seedPlanCreditUser(t *testing.T, id int, quota int) {
	t.Helper()
	user := &User{
		Id:       id,
		Username: "plan_credit_user",
		Status:   common.UserStatusEnabled,
		Quota:    quota,
	}
	require.NoError(t, DB.Create(user).Error)
}

func seedPlanCreditPlan(t *testing.T, plan *SubscriptionPlan) *SubscriptionPlan {
	t.Helper()
	plan.Title = "Credit Plan"
	plan.Currency = "USD"
	plan.DurationUnit = SubscriptionDurationMonth
	plan.DurationValue = 1
	plan.Enabled = true
	require.NoError(t, DB.Create(plan).Error)
	InvalidateSubscriptionPlanCache(plan.Id)
	return plan
}

func seedPlanCreditOrder(t *testing.T, userID int, planID int, tradeNo string, status string) {
	t.Helper()
	order := &SubscriptionOrder{
		UserId:          userID,
		PlanId:          planID,
		Money:           1,
		TradeNo:         tradeNo,
		PaymentMethod:   "alipay",
		PaymentProvider: PaymentProviderEpay,
		Status:          status,
		CreateTime:      common.GetTimestamp(),
	}
	require.NoError(t, DB.Create(order).Error)
}

func seedPlanCreditSuccessOrder(t *testing.T, userID int, planID int, tradeNo string) {
	t.Helper()
	seedPlanCreditOrder(t, userID, planID, tradeNo, common.TopUpStatusSuccess)
}

func planCreditUserQuota(t *testing.T, userID int) int {
	t.Helper()
	var user User
	require.NoError(t, DB.Select("quota").Where("id = ?", userID).First(&user).Error)
	return user.Quota
}

func countSubscriptionOrders(t *testing.T, userID int) int64 {
	t.Helper()
	var count int64
	require.NoError(t, DB.Model(&SubscriptionOrder{}).Where("user_id = ?", userID).Count(&count).Error)
	return count
}

func TestCreditPlanQuotaTx_CreditsPlanTotalAmount(t *testing.T) {
	truncateTables(t)

	seedPlanCreditUser(t, 801, 100)
	plan := seedPlanCreditPlan(t, &SubscriptionPlan{Id: 801, PriceAmount: 5, TotalAmount: 2_500_000})

	var credited int64
	require.NoError(t, DB.Transaction(func(tx *gorm.DB) error {
		var err error
		credited, err = CreditPlanQuotaTx(tx, 801, plan)
		return err
	}))

	assert.Equal(t, int64(2_500_000), credited)
	assert.Equal(t, 100+2_500_000, planCreditUserQuota(t, 801))
}

func TestCreditPlanQuotaTx_RejectsNonPositiveTotalAmount(t *testing.T) {
	truncateTables(t)

	seedPlanCreditUser(t, 802, 100)
	plan := seedPlanCreditPlan(t, &SubscriptionPlan{Id: 802, PriceAmount: 5, TotalAmount: 0})

	err := DB.Transaction(func(tx *gorm.DB) error {
		_, txErr := CreditPlanQuotaTx(tx, 802, plan)
		return txErr
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "无效的充值额度")
	// 事务内拒绝：额度一分未加，不会出现「收钱不发额度」
	assert.Equal(t, 100, planCreditUserQuota(t, 802))
}

func TestCreditPlanQuotaTx_RejectsWhenPurchaseCapReached(t *testing.T) {
	truncateTables(t)

	seedPlanCreditUser(t, 803, 100)
	plan := seedPlanCreditPlan(t, &SubscriptionPlan{
		Id:                 803,
		PriceAmount:        5,
		TotalAmount:        2_500_000,
		MaxPurchasePerUser: 1,
	})
	// 已有一笔成功订单 → 达到上限
	seedPlanCreditSuccessOrder(t, 803, plan.Id, "PLAN_CAP_ORDER_1")

	err := DB.Transaction(func(tx *gorm.DB) error {
		_, txErr := CreditPlanQuotaTx(tx, 803, plan)
		return txErr
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "购买上限")
	assert.Equal(t, 100, planCreditUserQuota(t, 803))
}

func TestCreditPlanQuotaTx_CapIgnoresOtherPlansAndNonSuccessOrders(t *testing.T) {
	truncateTables(t)

	seedPlanCreditUser(t, 804, 0)
	plan := seedPlanCreditPlan(t, &SubscriptionPlan{
		Id:                 804,
		PriceAmount:        5,
		TotalAmount:        1_000,
		MaxPurchasePerUser: 1,
	})
	other := seedPlanCreditPlan(t, &SubscriptionPlan{Id: 805, PriceAmount: 5, TotalAmount: 1_000})
	seedPlanCreditSuccessOrder(t, 804, other.Id, "PLAN_CAP_OTHER_1")

	// 同一档位上的在途单与终态失败单都不计入上限
	seedPlanCreditOrder(t, 804, plan.Id, "PLAN_CAP_PENDING_1", common.TopUpStatusPending)
	seedPlanCreditOrder(t, 804, plan.Id, "PLAN_CAP_FAILED_1", common.TopUpStatusFailed)

	require.NoError(t, DB.Transaction(func(tx *gorm.DB) error {
		_, txErr := CreditPlanQuotaTx(tx, 804, plan)
		return txErr
	}))

	assert.Equal(t, 1_000, planCreditUserQuota(t, 804))
}

// ---------------------------------------------------------------------------
// countPlanPurchasesTx — 限购计数必须是加锁读
//
// 不变量：事务内的限购 COUNT 必须挂 FOR UPDATE。持有 users 行锁只做到「串行化」，
// 做不到「读最新」——MySQL 默认 REPEATABLE READ 下事务快照由事务内第一条普通
// SELECT 固定，而那条 SELECT 发生在抢到 users 锁之前，普通读会一直看到抢锁前的
// 旧快照，让第二笔购买绕过上限白拿额度。
//
// 测试环境限制：测试库是 SQLite（TestMain 固定 :memory:），写串行且 GORM 的
// sqlite dialector 会直接丢弃 Locking 子句，真实的 RR 快照穿透在这里无法复现。
// 因此下面从两个角度锁死不变量：查询确实申请了 FOR UPDATE 锁（子句层），以及该
// 子句在 MySQL dialector 上确实渲染成 SQL 里的 FOR UPDATE（SQL 层）。任一处被
// 改回普通读，测试即红。
// ---------------------------------------------------------------------------

// captureCountPlanPurchases 跑一次 countPlanPurchasesTx，回传该查询实际生成的
// SQL 与它申请的锁子句。挂 query 回调是唯一能同时看到这两者、又不复制一份查询
// 语句（平行路径）的办法。
func captureCountPlanPurchases(t *testing.T, hookOn *gorm.DB, tx *gorm.DB) (string, clause.Expression) {
	t.Helper()
	const hookName = "test:capture_count_plan_purchases"
	var (
		sql        string
		lockClause clause.Expression
	)
	require.NoError(t, hookOn.Callback().Query().After("gorm:query").Register(hookName, func(d *gorm.DB) {
		if d.Statement.Table != "subscription_orders" {
			return
		}
		sql = d.Statement.SQL.String()
		if forClause, ok := d.Statement.Clauses["FOR"]; ok {
			lockClause = forClause.Expression
		}
	}))
	t.Cleanup(func() { _ = hookOn.Callback().Query().Remove(hookName) })

	_, err := countPlanPurchasesTx(tx, 1, 1)
	require.NoError(t, err)
	return sql, lockClause
}

func TestCountPlanPurchasesTx_TakesLockingReadInsideTransaction(t *testing.T) {
	truncateTables(t)

	seedPlanCreditUser(t, 850, 0)
	plan := seedPlanCreditPlan(t, &SubscriptionPlan{Id: 850, PriceAmount: 5, TotalAmount: 1_000})
	seedPlanCreditSuccessOrder(t, 850, plan.Id, "PLAN_LOCKREAD_OK_1")
	seedPlanCreditSuccessOrder(t, 850, plan.Id, "PLAN_LOCKREAD_OK_2")
	seedPlanCreditOrder(t, 850, plan.Id, "PLAN_LOCKREAD_PENDING_1", common.TopUpStatusPending)

	var (
		count      int64
		sql        string
		lockClause clause.Expression
	)
	require.NoError(t, DB.Transaction(func(tx *gorm.DB) error {
		sql, lockClause = captureCountPlanPurchases(t, DB, tx)
		var err error
		count, err = countPlanPurchasesTx(tx, 850, plan.Id)
		return err
	}))

	locking, ok := lockClause.(clause.Locking)
	require.True(t, ok, "事务内的限购 COUNT 必须申请行锁，否则 MySQL RR 快照会穿透")
	assert.Equal(t, "UPDATE", locking.Strength)
	// SQLite 不支持行锁，dialector 会丢弃子句 —— 三库共用同一份代码的前提
	assert.NotContains(t, strings.ToUpper(sql), "FOR UPDATE")
	// 加锁读不改变计数口径：只算同档位的成功单
	assert.Equal(t, int64(2), count)
}

func TestCountPlanPurchases_PrecheckStaysNonLocking(t *testing.T) {
	truncateTables(t)

	// controller 下单前的预检走全局 DB（tx == nil），本来就是尽力而为，
	// 不能在事务外把订单行锁住
	_, lockClause := captureCountPlanPurchases(t, DB, nil)
	assert.Nil(t, lockClause)
}

// ---------------------------------------------------------------------------
// 限购索引 —— 加锁读的锁范围靠它收窄
//
// 加锁读会把扫到的行全部锁住（MySQL RR 下不匹配的行也不放锁），所以「扫多少行」
// 直接决定死锁面：走 user_id 单列索引时锁的是该买家的全部订单，回调链路手里
// 攥着的那张 pending 单也在里面，与 users 行锁形成反向环。只有三列逐个等值匹配
// 到同一个索引，区间才收到「该买家该档位的成功单」。
// 下面两条分别锁住：索引列序（逻辑）与索引真的被 AutoMigrate 建了出来（落库）。
// ---------------------------------------------------------------------------

// purchaseCapIndexColumns 必须与 countPlanPurchasesTx 的 WHERE 列序逐列一致。
var purchaseCapIndexColumns = []string{"user_id", "plan_id", "status"}

// lookupPurchaseCapIndex 从 SubscriptionOrder 的 GORM 声明里取出覆盖限购 COUNT
// 的索引。按列序找而不是按名字找：名字换了不影响锁行为，列序变了才影响。
func lookupPurchaseCapIndex(t *testing.T, db *gorm.DB) schema.Index {
	t.Helper()
	stmt := &gorm.Statement{DB: db}
	require.NoError(t, stmt.Parse(&SubscriptionOrder{}))
	for _, index := range stmt.Schema.ParseIndexes() {
		columns := make([]string, 0, len(index.Fields))
		for _, field := range index.Fields {
			columns = append(columns, field.DBName)
		}
		if strings.Join(columns, ",") == strings.Join(purchaseCapIndexColumns, ",") {
			return index
		}
	}
	t.Fatalf("SubscriptionOrder 没有声明 (%s) 复合索引，限购加锁读会退回 user_id 区间锁",
		strings.Join(purchaseCapIndexColumns, ", "))
	return schema.Index{}
}

// renderSubscriptionOrderDDL 用 DryRun 连接（不连库）渲染建表语句，拿到 GORM
// 实际会发给该数据库的 DDL。
func renderSubscriptionOrderDDL(t *testing.T, dialector gorm.Dialector) []string {
	t.Helper()
	db, err := gorm.Open(dialector, &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	require.NoError(t, err)

	var statements []string
	require.NoError(t, db.Callback().Raw().Before("gorm:raw").
		Register("test:capture_subscription_order_ddl", func(d *gorm.DB) {
			statements = append(statements, d.Statement.SQL.String())
		}))
	require.NoError(t, db.Migrator().CreateTable(&SubscriptionOrder{}))
	return statements
}

// squashSQL 抹掉引号与空白，让三库的引号风格（“ / ""）和排版差异不进入断言。
func squashSQL(sql string) string {
	return strings.NewReplacer("`", "", `"`, "", " ", "", "\n", "", "\t", "").Replace(sql)
}

func TestSubscriptionOrderPurchaseCapIndex_IsCreatedByAutoMigrate(t *testing.T) {
	index := lookupPurchaseCapIndex(t, DB)

	for _, field := range index.Fields {
		// 前缀长度只有 MySQL 认，PG 会在 col(n) 处语法报错，三库共用一份声明就不能带
		assert.Zero(t, field.Length, "限购索引不能带前缀长度，PostgreSQL 不支持 col(n)")
	}
	// TestMain 已经对 SubscriptionOrder 跑过 AutoMigrate，这里查的是真实建出来的库
	assert.True(t, DB.Migrator().HasIndex(&SubscriptionOrder{}, index.Name),
		"AutoMigrate 之后 %s 必须存在，否则加锁读仍然走 user_id 区间", index.Name)
}

func TestSubscriptionOrderPurchaseCapIndex_RendersOnMySQLAndPostgres(t *testing.T) {
	index := lookupPurchaseCapIndex(t, DB)

	cases := []struct {
		name         string
		dialector    gorm.Dialector
		assertStatus func(t *testing.T, ddl string)
	}{
		{
			name: "mysql",
			dialector: mysql.New(mysql.Config{
				DSN:                       "gorm:gorm@tcp(127.0.0.1:3306)/gorm?parseTime=true",
				SkipInitializeWithVersion: true,
			}),
			assertStatus: func(t *testing.T, ddl string) {
				// MySQL 拒绝给 longtext 建无前缀索引，而前缀长度又不是合法 PG 语法，
				// 所以 status 必须落成有界 varchar，索引才建得出来
				assert.NotContains(t, ddl, "statuslongtext")
				assert.Contains(t, ddl, "statusvarchar(")
			},
		},
		{
			name: "postgres",
			dialector: postgres.New(postgres.Config{
				DSN: "host=127.0.0.1 user=gorm password=gorm dbname=gorm port=5432",
			}),
			assertStatus: func(t *testing.T, ddl string) {
				// PG 的 text 本来就能进 btree 索引，列不必跟着 MySQL 一起动
				assert.Contains(t, ddl, "statustext")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			statements := renderSubscriptionOrderDDL(t, tc.dialector)

			var indexStatement string
			for _, statement := range statements {
				if strings.Contains(statement, index.Name) {
					indexStatement = squashSQL(statement)
					break
				}
			}
			require.NotEmpty(t, indexStatement, "%s 的建表 DDL 里没有 %s", tc.name, index.Name)
			assert.Contains(t, indexStatement, "("+strings.Join(purchaseCapIndexColumns, ",")+")")

			tc.assertStatus(t, squashSQL(strings.Join(statements, ";")))
		})
	}
}

func TestCountPlanPurchasesTx_RendersForUpdateOnMySQL(t *testing.T) {
	// 洞只在 MySQL 上是实的，所以这里用 MySQL dialector 干跑一次，确认锁子句
	// 真的落进了 SQL，而不是停在 GORM 的子句层被静默丢掉。
	mysqlDB, err := gorm.Open(
		mysql.New(mysql.Config{
			DSN:                       "gorm:gorm@tcp(127.0.0.1:3306)/gorm?parseTime=true",
			SkipInitializeWithVersion: true,
		}),
		&gorm.Config{DryRun: true, DisableAutomaticPing: true},
	)
	require.NoError(t, err)

	sql, lockClause := captureCountPlanPurchases(t, mysqlDB, mysqlDB)

	require.NotNil(t, lockClause)
	assert.Contains(t, strings.ToUpper(sql), "FOR UPDATE")
}

// ---------------------------------------------------------------------------
// PurchaseSubscriptionWithBalance — 折扣兑换：扣与加必须同事务
// ---------------------------------------------------------------------------

func TestPurchaseSubscriptionWithBalance_NetQuotaIsTotalMinusPrice(t *testing.T) {
	truncateTables(t)

	const price = 1.0
	requiredQuota, err := calcSubscriptionBalanceQuota(price, common.QuotaPerUnit)
	require.NoError(t, err)

	const initQuota = 5_000_000
	seedPlanCreditUser(t, 810, initQuota)
	plan := seedPlanCreditPlan(t, &SubscriptionPlan{Id: 810, PriceAmount: price, TotalAmount: 3_000_000})

	require.NoError(t, PurchaseSubscriptionWithBalance(810, plan.Id, ""))

	// 净额恒 = TotalAmount - requiredQuota
	assert.Equal(t, initQuota+int(plan.TotalAmount)-requiredQuota, planCreditUserQuota(t, 810))
	assert.Equal(t, int64(1), countSubscriptionOrders(t, 810))
}

func TestPurchaseSubscriptionWithBalance_RollsBackChargeWhenCreditFails(t *testing.T) {
	truncateTables(t)

	const initQuota = 5_000_000
	seedPlanCreditUser(t, 811, initQuota)
	plan := seedPlanCreditPlan(t, &SubscriptionPlan{
		Id:                 811,
		PriceAmount:        1.0,
		TotalAmount:        3_000_000,
		MaxPurchasePerUser: 1,
	})
	// 上限已满 → 扣款发生在入账之前，入账失败必须把扣款一并回滚
	seedPlanCreditSuccessOrder(t, 811, plan.Id, "PLAN_ROLLBACK_ORDER_1")

	err := PurchaseSubscriptionWithBalance(811, plan.Id, "")
	require.Error(t, err)

	assert.Equal(t, initQuota, planCreditUserQuota(t, 811))
	// 只剩预置的那一笔，本次购买没有落单
	assert.Equal(t, int64(1), countSubscriptionOrders(t, 811))
}

func TestPurchaseSubscriptionWithBalance_RejectsInsufficientBalance(t *testing.T) {
	truncateTables(t)

	seedPlanCreditUser(t, 812, 10)
	plan := seedPlanCreditPlan(t, &SubscriptionPlan{Id: 812, PriceAmount: 1.0, TotalAmount: 3_000_000})

	err := PurchaseSubscriptionWithBalance(812, plan.Id, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "余额不足")
	assert.Equal(t, 10, planCreditUserQuota(t, 812))
	assert.Zero(t, countSubscriptionOrders(t, 812))
}

// ---------------------------------------------------------------------------
// CompleteSubscriptionOrder — top_ups 镜像的 Amount 记的是美元，不是 quota
// ---------------------------------------------------------------------------

func TestCompleteSubscriptionOrder_MirrorsTopUpAmountInDollars(t *testing.T) {
	truncateTables(t)

	seedPlanCreditUser(t, 820, 0)
	totalAmount := int64(5 * common.QuotaPerUnit)
	plan := seedPlanCreditPlan(t, &SubscriptionPlan{Id: 820, PriceAmount: 5, TotalAmount: totalAmount})

	order := &SubscriptionOrder{
		UserId:          820,
		PlanId:          plan.Id,
		Money:           35.5,
		TradeNo:         "PLAN_TOPUP_MIRROR_1",
		PaymentMethod:   "alipay",
		PaymentProvider: PaymentProviderEpay,
		Status:          common.TopUpStatusPending,
		CreateTime:      common.GetTimestamp(),
	}
	require.NoError(t, DB.Create(order).Error)

	require.NoError(t, CompleteSubscriptionOrder(order.TradeNo, "", PaymentProviderEpay, "alipay"))

	assert.Equal(t, int(totalAmount), planCreditUserQuota(t, 820))

	topUp := GetTopUpByTradeNo(order.TradeNo)
	require.NotNil(t, topUp)
	// Amount 是美元数（quota = Amount * QuotaPerUnit），不是 quota 原始单位
	assert.Equal(t, int64(5), topUp.Amount)
	assert.InDelta(t, 35.5, topUp.Money, 1e-9)
	assert.Equal(t, common.TopUpStatusSuccess, topUp.Status)
}

// planTopUpAmount 的取整口径：四舍五入，且任何正额度至少记 1 美元。
func TestPlanTopUpAmount_RoundingBoundaries(t *testing.T) {
	const perUnit = 500_000.0

	cases := []struct {
		name     string
		credited int64
		want     int64
	}{
		{"整除", 5 * 500_000, 5},
		{"向上进位", 1_750_000, 4},    // $3.5 -> 4
		{"向下舍去", 1_600_000, 3},    // $3.2 -> 3
		{"半美元档不再少记", 750_000, 2},  // $1.5 -> 2
		{"不足一美元不落 0", 250_000, 1}, // $0.5 -> 1
		{"极小额度不落 0", 1, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := planTopUpAmount(tc.credited, perUnit)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestPlanTopUpAmount_RejectsInvalidInput(t *testing.T) {
	_, err := planTopUpAmount(0, 500_000)
	require.ErrorIs(t, err, ErrPlanQuotaInvalid)

	_, err = planTopUpAmount(1_000, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "额度单位配置错误")
}

// 收了钱却入不了账时，订单必须落终态 failed，而不是回滚成 pending 让网关无限重试。
func TestCompleteSubscriptionOrder_SettlesOrderFailedWhenCapReached(t *testing.T) {
	truncateTables(t)

	seedPlanCreditUser(t, 830, 0)
	plan := seedPlanCreditPlan(t, &SubscriptionPlan{
		Id:                 830,
		PriceAmount:        5,
		TotalAmount:        2_500_000,
		MaxPurchasePerUser: 1,
	})
	seedPlanCreditSuccessOrder(t, 830, plan.Id, "PLAN_CAP_PAID_1")

	paid := &SubscriptionOrder{
		UserId:          830,
		PlanId:          plan.Id,
		Money:           5,
		TradeNo:         "PLAN_CAP_PAID_2",
		PaymentMethod:   "alipay",
		PaymentProvider: PaymentProviderEpay,
		Status:          common.TopUpStatusPending,
		CreateTime:      common.GetTimestamp(),
	}
	require.NoError(t, DB.Create(paid).Error)

	const payload = `{"gateway_trade_no":"EPAY-REFUND-ME-1"}`
	err := CompleteSubscriptionOrder(paid.TradeNo, payload, PaymentProviderEpay, "wxpay")
	require.ErrorIs(t, err, ErrPlanPurchaseCapReached)

	var settled SubscriptionOrder
	require.NoError(t, DB.Where("trade_no = ?", paid.TradeNo).First(&settled).Error)
	assert.Equal(t, common.TopUpStatusFailed, settled.Status)
	assert.NotZero(t, settled.CompleteTime)
	// 这笔单存在的意义就是「钱已收、需人工退款」，网关流水凭证必须落库
	assert.Equal(t, payload, settled.ProviderPayload)
	assert.Equal(t, "wxpay", settled.PaymentMethod)
	// 额度一分未加，也没有伪造 top_ups 记录
	assert.Equal(t, 0, planCreditUserQuota(t, 830))
	assert.Nil(t, GetTopUpByTradeNo(paid.TradeNo))
}

// ---------------------------------------------------------------------------
// AdminBindSubscription — 管理员发放：加额度但绝不进付费统计
// ---------------------------------------------------------------------------

func TestAdminBindSubscription_CreditsQuotaWithoutOrder(t *testing.T) {
	truncateTables(t)

	const initQuota = 1_000
	seedPlanCreditUser(t, 840, initQuota)
	plan := seedPlanCreditPlan(t, &SubscriptionPlan{Id: 840, PriceAmount: 5, TotalAmount: 2_500_000})

	hint, err := AdminBindSubscription(840, plan.Id, "manual grant")
	require.NoError(t, err)
	assert.Empty(t, hint)

	assert.Equal(t, initQuota+int(plan.TotalAmount), planCreditUserQuota(t, 840))
	// 核心契约：管理员发放不落订单，付费统计与限购计数都不受影响
	assert.Zero(t, countSubscriptionOrders(t, 840))
}

func TestAdminBindSubscription_RejectsWhenCapReached(t *testing.T) {
	truncateTables(t)

	const initQuota = 1_000
	seedPlanCreditUser(t, 841, initQuota)
	plan := seedPlanCreditPlan(t, &SubscriptionPlan{
		Id:                 841,
		PriceAmount:        5,
		TotalAmount:        2_500_000,
		MaxPurchasePerUser: 1,
	})
	seedPlanCreditSuccessOrder(t, 841, plan.Id, "PLAN_ADMIN_CAP_1")

	_, err := AdminBindSubscription(841, plan.Id, "manual grant")
	require.ErrorIs(t, err, ErrPlanPurchaseCapReached)
	assert.Equal(t, initQuota, planCreditUserQuota(t, 841))
}

func TestAdminBindSubscription_RejectsNonPositiveTotalAmount(t *testing.T) {
	truncateTables(t)

	const initQuota = 1_000
	seedPlanCreditUser(t, 842, initQuota)
	plan := seedPlanCreditPlan(t, &SubscriptionPlan{Id: 842, PriceAmount: 5, TotalAmount: 0})

	_, err := AdminBindSubscription(842, plan.Id, "manual grant")
	require.ErrorIs(t, err, ErrPlanQuotaInvalid)
	assert.Equal(t, initQuota, planCreditUserQuota(t, 842))
	assert.Zero(t, countSubscriptionOrders(t, 842))
}
