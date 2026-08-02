# 接口契约 — new-api 订阅池下线

## C1 · 资金来源（service/funding_source.go）

改造后**只剩钱包一条路径**。`FundingSource` 接口只有 `WalletFunding` 一个实现，
可顺势内联掉抽象层（若内联，须同步更新 `service/billing_session.go` 的调用方）。

```go
// 保留
type WalletFunding struct { ... }
func (w *WalletFunding) Source() string { return BillingSourceWallet }
func (w *WalletFunding) PreConsume(amount int) error
func (w *WalletFunding) Settle(delta int) error
func (w *WalletFunding) Refund() error

// 删除
type SubscriptionFunding struct { ... }   // 及其 4 个方法
```

## C2 · 计费会话（service/billing_session.go）

`NewBillingSession` 只保留 `tryWallet()`；删 `trySubscription()` 闭包和四路 `switch pref`。

- 失败语义：钱包不足 → 返回原有的额度不足错误（不再有「订阅不足则回落钱包」的二选一）
- `shouldTrust()`（:303-314）原先对订阅禁用信任额度旁路 → 删掉订阅判断后旁路自然恢复，**这是预期行为**

## C3 · 额度扣减（service/quota.go）

`DecreaseUserQuota` 成为唯一扣费路径。删 `checkAndSendSubscriptionQuotaNotify()`；
余额预警只保留钱包一条。

## C4 · 计费来源常量（service/billing.go）

```go
const BillingSourceWallet = "wallet"   // 保留
// const BillingSourceSubscription = "subscription"   // 删
```

## C5 · 日志字段（service/log_info_generate.go）

删 9 个订阅日志 key；**`billing_source` 保留并恒为 `"wallet"`** —— 历史日志里两种值都存在，
解析端必须继续认识这个 key。

## C6 · 购买链路（本轮核心行为变更）

四条创建路径全部把 `CreateUserSubscriptionFromPlanTx(...)` 换成
**在同一事务内给 `users.quota` 加额度**（复用既有 `IncreaseUserQuota` / 其事务版本，
先用 `rg` 确认真实签名，不要臆造）。

**必须保留**：`UpgradeGroup` 分组升级、`SyncSubscriptionDistributionOrderTx` 分销投影、
`fireSubscriptionPaidHook`。

| 路径 | 入口 | 契约 |
|---|---|---|
| 余额购买 | `model/subscription.go` `PurchaseSubscriptionWithBalance` | 变成「扣余额 → 加余额」的折扣兑换。**净额必须 = `plan.TotalAmount - requiredQuota`，且同一事务** |
| 在线支付 ×5 | `model/subscription.go` `CompleteSubscriptionOrder` | 加余额；且 `upsertSubscriptionTopUpTx` 的 `Amount: 0` 改成**实际到账 quota**，否则财务报表少记 |
| 管理员绑定 | `model/subscription.go` `AdminBindSubscription` | 加余额 |
| 代理库存码 | `service/distribution_transactions.go` | 加余额 |

⚠️ 「两条建订阅入口漏改一条」是本任务最高频的翻车点 —— 虎皮椒支付 + 代理库存码兑换，
**后者最容易漏**。改完用 `rg 'CreateUserSubscriptionFromPlanTx'` 确认零残留。

## C7 · 路由语义（D2：接口一个都不删）

`router/api-router.go` 全部订阅路由**保留不删**，语义改变：

| 路由 | 新行为 |
|---|---|
| `GET /subscription/self` | 返回**空订阅列表 + 用户钱包余额**（前端不炸） |
| `PUT /subscription/self/preference` | 保留接口，**恒返回成功，内部 no-op** |
| 各支付 `/pay` | 行为改为「买档位加余额」 |
| admin 套餐 CRUD | 保留（现在管的是充值档位） |
| admin 四条 `user_subscriptions` 管理路由 | 返回空 / 410 |

## C8 · 数据模型（model/）

```go
// 删除 struct + 其全部函数
type UserSubscription struct { ... }
type SubscriptionPreConsumeRecord struct { ... }

// 保留（D1）
type SubscriptionPlan struct { ... }
type SubscriptionOrder struct { ... }
```

`model/main.go` 的 AutoMigrate 去掉前两个，**保留后两个**。

## C9 · 分销口径（不改代码，但要确认）

`service/distribution_promotion.go` `CountInviterPromotionStats` 对 `SubscriptionOrder` 和 `TopUp`
取并集统计「付费邀请人数」。D1 保留了 `SubscriptionOrder` → **逻辑不动**，
但必须确认「档位购买」仍然写这张表，否则口径变松。改完在审计文件里写明确认结果。
