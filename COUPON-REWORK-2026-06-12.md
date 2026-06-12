# 优惠券功能重做（p3_promo_codes → p3_coupons）· 2026-06-12

分支 `feature/njc/agent-phase4` · 25 文件 +757/−957 · `go build`/`go vet`/`go test` 全绿 · 前端 `bun run build` 成功 · 全库无 promo-codes 残留

## 设计概要

- **优惠券 = 包装一张上游原生余额兑换码**（redemptions 表）。创建优惠券时同步创建兑换码并把 id/code 记到券上；用户在钱包"兑换码"输入框输码即到账。
- **验券前置**：钱包兑换入口（POST /api/user/topup）最前面查券，命中走独立分支（校验 active+未过期 → 原生 Redeem → 标记 used），完全不经库存码逻辑；未命中原流程一行未改。
- **代理自申请**：用自己余额 1:1 申请，3 天有效；到期未用自动退回余额，券和兑换码物理销毁。
- **管理员发放**：手工批量多档（数量×面额×有效期天数），单次 ≤100 张，不可退，到期直接销毁。
- 旧套餐折扣码（死代码，零调用方）彻底移除，旧表 `p3_promo_codes` 启动迁移时 DropTable。

## 后端改动（Go）

### 新增

| 文件 | 内容 |
|---|---|
| `service/distribution_coupons.go` | 优惠券核心：申请（锁代理→扣余额+debit 流水→建兑换码+券）、批量发放（≤100张/次）、列表、`GetCouponByCode`/`MarkCouponUsed`（条件更新防并发）、到期清扫（条件删除防并发，self 退余额+credit 流水，券和码物理销毁；兑换已发生但未标记的补记 used 堵双付）、1 分钟定时任务（仅主节点） |
| `service/distribution_coupons_test.go` | 6 个用例：申请扣款+流水、余额不足零残留、发放校验、到期 self 退款销毁 / admin 仅销毁、已兑换补记不退款、防重复核销 |

### 修改

| 文件 | 内容 |
|---|---|
| `model/distribution.go` | 删 `DistributionPromoCode`，新增 `DistributionCoupon`（表 `p3_coupons`），迁移注册替换 |
| `model/main.go` | 启动迁移时 Drop 旧表 `p3_promo_codes`（migrateDB / migrateDBFast 两条路径都挂，失败仅记日志） |
| `controller/user.go` | TopUp 兑换入口前置验券（见设计概要） |
| `controller/distribution.go` | 删 3 个 promo handler，增 4 个 coupon handler（分页形状沿用原模式） |
| `router/api-router.go` | 删 4 条 promo-codes 路由；增 `GET /api/agent/coupons`、`POST /api/agent/coupons/apply`、`GET /api/agent-admin/coupons`、`POST /api/agent-admin/coupons/issue`（写接口挂 CriticalRateLimit） |
| `service/distribution_promotions.go` | 删全部 7 个折扣码函数，GiftRule 完整保留 |
| `service/distribution_rules.go` | 增流水来源常量 `coupon_apply` / `coupon_refund` |
| `main.go` | 注册 `StartCouponExpiryTask()` |

**红线确认**：`model/redemption.go`、`controller/redemption.go` 零改动（与上游零 diff，保住升级合并收益）。

## 前端改动（web/default/src/features/）

| 文件 | 内容 |
|---|---|
| `distribution/types.ts` / `api.ts` | 类型和 API 替换为 coupon 四函数（getAgentCoupons / applyAgentCoupon / adminGetCoupons / adminIssueCoupons） |
| `distribution/agent-center.tsx` | promo tab 重做为优惠券：列表（码可复制/面额/来源/状态/到期/使用时间）+ "申请优惠券"Dialog（面额输入、余额展示、3 天有效自动退回提示） |
| `distribution/agent-admin.tsx` | 新增 'coupons' tab：代理过滤 + 券列表（含代理名）+ 发放 Dialog（多档明细行 数量/面额/天数 可增删、合计 x/100 校验、备注） |
| `distribution/labels.ts` | 删折扣类型映射，增 active/used 状态和来源映射（self=代理申请 / admin=平台发放） |
| `i18n/locales/{en,zh,fr,ja,ru,vi}.json` | 各 +18 个 key，6 语种真实翻译，i18n:sync 校验零缺失 |

钱包（features/wallet/）、兑换码管理（features/redemption-codes/）未动。

## 接口契约

```
GET  /api/agent/coupons?p=&page_size=                  代理查自己的券（分页）
POST /api/agent/coupons/apply        {"amount": 1.5}   代理余额申请，3 天有效
GET  /api/agent-admin/coupons?agent_id=&p=&page_size=  管理员查券（agent_id=0 全部，含代理名）
POST /api/agent-admin/coupons/issue                    管理员批量发放
     {"agent_id": 1, "items": [{"count": 2, "amount": 1, "validity_days": 7}], "remark": ""}
     → {"issued_count": 2}
```

券字段：`id, agent_id, redemption_id, code, amount(美元面额), quota(=amount×QuotaPerUnit), source(self|admin), status(active|used), issued_by, used_user_id, used_at, expires_at, created_at, updated_at, remark`。到期未用的券被物理删除，列表只会出现 active/used。

## 计划外修复（phase4 HEAD 本身编译/测试是红的，不修无法验证）

- `controller/subscription_payment_stripe.go`：`:=` 重复声明 → `=`（编译错）
- `controller/subscription_payment_epay.go`：删未使用 import（编译错）
- `service/distribution_rules_test.go`：测试签名滞后（int→float64），按现行为修正
- `model/redemption_redeem_test.go`、`model/subscription_promo_order_test.go`：引用 phase4 已删的类型化兑换码字段（编译错），删失效用例、保留有效用例
- `service/task_billing_test.go`：共享 TestMain 补 Redemption/DistributionCoupon/DistributionBalanceLedger 三张表的 AutoMigrate

## 验证记录

1. `go build ./...` ✅　`go vet ./service/... ./controller/... ./model/...` ✅　`go test ./service/... ./model/... -count=1` 全绿 ✅
2. `cd web/default && bun run build` ✅（tsc 仅存一个与本次无关的预先存在错误 `components/ai-elements/code-block.tsx` 缺 'hast' 类型）
3. 全局 grep：`DistributionPromoCode` / `promo-codes` / `p3_promo_codes` 仅剩迁移清理函数一处有意引用 ✅

## 部署注意

- 启动后 GORM 自动建 `p3_coupons`、Drop `p3_promo_codes`，零手工脚本。
- 到期清扫任务仅主节点跑（IsMasterNode），间隔 1 分钟，单批 100 张循环扫净。
- 资金轨迹：申请扣款和到期退款都写 `p3_balance_ledgers`（sourceType `coupon_apply` / `coupon_refund`），券被销毁后流水仍可审计。
