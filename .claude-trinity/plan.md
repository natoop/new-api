# 派工清单 — new-api 订阅池下线

分支：`feature/njc/drop-subscription-pool`（已建，基于 `feature/njc/agent-phase4`）

## 文件所有权（硬边界：只改自己那栏的文件，不越界）

| id | 曹 | deps | 独占文件 |
|---|---|---|---|
| N1 | 计费链路曹 | [] | `service/funding_source.go` `service/billing_session.go` `service/quota.go` `service/billing.go` `service/task_billing.go` `service/log_info_generate.go` `service/subscription_reset_task.go`(整删) `relay/common/relay_info.go` `dto/user_settings.go` `common/str.go` `main.go` |
| N2 | 模型曹 | [] | `model/subscription.go` `model/main.go` `model/task.go` `controller/relay.go` `service/distribution_transactions.go` |
| N3 | 路由控制器曹 | [] | `router/api-router.go` `controller/subscription.go` `controller/subscription_payment_*.go` |
| N4 | 前端曹 | [] | `web/default/src/features/**` `web/default/src/routes/**` `web/default/src/hooks/**` |
| N5 | i18n 曹 | [] | `i18n/keys.go` `i18n/*.yaml` `web/default/src/i18n/**` |
| N6 | 编译测试曹 | [N1..N5] | 全仓（只做编译修残留 + 测试修订） |

**N1–N5 并发执行，文件互不重叠。** N6 第二波收口。

## 编译边界（重要）

N1–N5 处于同一次大重构的中途，**跨文件编译必然是红的**。因此：

- N1–N5 **不要跑 `go build ./...` 并试图修别人文件的报错**
- 只需保证自己改的文件 `gofmt -l` 干净、语法自洽
- 若发现自己文件引用了别人负责删除的符号 → **照契约删掉引用**，不要保留兼容垫片
- 全仓编译由 N6 统一收口

## 🔴 首波挖出的移交项（N6 必须逐条处理，交接文档漏了）

### T1 · 档位限购 `MaxPurchasePerUser` 失效（严重，资金侧）
`plan.MaxPurchasePerUser` 的**事务内强制**原本在 `model/subscription.go` 的
`CreateUserSubscriptionFromPlanTx`（约 :527），该函数随订阅池被删除；
N3 已删掉 controller 层的预检（它查的是 `UserSubscription`）。
**C6 的「必须保留」清单里没列这条上限 → 若无人补回，限购规则整体失效，用户可无限购买设了上限的档位。**

修复口径：在新的加余额事务内按 D1 保留的 `SubscriptionOrder` 计数
（`user_id + plan_id + status=success`）重新强制上限。必须**在事务内**，不能只做 controller 预检。

### T2 · `TaskPrivateData.BillingSource` 去留
交接文档只点名删 `TaskPrivateData.SubscriptionId`。N1 建议连 `BillingSource` 一起删。
判据：**日志里的 `billing_source` 必须保留并恒为 `wallet`**（C5，历史日志解析依赖）；
任务私有数据里的这个字段只有单一取值，属内部状态 → 可删。按此原则处理并写进审计。

### T3 · N1 的顺带删除需复核
N1 删了 `FundingSource` 接口本身（C1 允许内联）与死代码 `refundWithRetry`。
方向正确，但超出严格删除范围 → 门下省复核时确认无调用方遗留。

## 🟢 主控裁决（N2 抛出的 3 点，已定，N6 照此执行）

### R1 · `TopUp.Amount` 单位 —— 交接文档字面写错了，按事实改
交接文档说「`Amount: 0` 改成实际到账 quota」。**但 `TopUp.Amount` 存的是美元数，不是 quota 原始单位。**

证据：
- `model/topup.go:195`（虎皮椒）、`:626` —— `quotaToAdd = topUp.Amount × common.QuotaPerUnit`
- `controller/topup_creem.go:110-111` —— `Amount: 充值额度` / `Money: 支付金额`
- 前端账单按 USD 渲染 `Amount`

N2 现在写的是 `Amount: creditedQuota`（quota 原始单位）→ **差 500000 倍，一笔 $5 会显示成 $2,500,000**。

**改成**：`Amount = plan.TotalAmount / common.QuotaPerUnit`（美元整数），`Money = order.Money`（实付人民币）。
用 `decimal` 换算，不要用 float 直除。

⚠️ 已知历史遗留（**本轮不修，只记录**）：`model/topup.go:490` 的 Creem 路径是
`quota = topUp.Amount`（注释「Creem 直接使用 Amount 作为充值额度（整数）」），
与其余三条支付路径的约定相反。这是改造前就存在的不一致，不在本轮范围，写进审计备查。

### R2 · `TaskPrivateData.BillingSource` —— 保留（N2 的做法对，N1 的建议不采纳）
保留该字段。与 C5「日志 `billing_source` 恒为 wallet」一致，删掉反而要动更多调用点。

### R3 · `plan.TotalAmount = 0` 的档位「收钱不发额度」—— 加事务内守卫
照抄既有 seam `model/topup.go:196` 的写法：
```go
if quotaToAdd <= 0 { return errors.New("无效的充值额度") }
```
在 `CreditPlanQuotaTx` 里对 `plan.TotalAmount <= 0` 做同样的**事务内**拒绝，让它响亮失败而不是静默收钱。
（9 档新价目要到迁移那一轮才灌，这道守卫是防呆底线。）

### R4 · 管理端「充值档位 CRUD」前端入口本轮不补，推到下一轮（用户已裁决）
N4 整删 `features/subscriptions/` 时连带删掉了 `subscriptions-table` / `subscriptions-mutate-drawer` / `lib/plan-form`
—— 这是**唯一的档位管理界面**，与 C7「admin 套餐 CRUD 保留」正面冲突。

**用户裁决：不本轮补，下一轮跟 D7 九档价目一起做。**
理由：旧 6 档要到迁移那一轮才下架、新九档才写入，期间本来也不需要改档位。

**必须同时记录的风险（写进上线公告与 D8 迁移工具的必做项）**：
- 后端 4 条 admin plan 路由仍活着，但 POST/PUT/PATCH `/api/subscription/admin/plans` **零前端调用方**
- 叠加 R3（`plan.TotalAmount <= 0` 事务内硬失败）→ 九档价目一旦配错，
  管理员只能手搓 API 或直接改库才能修
- **D8 迁移工具必须自带档位写入与校验能力**，不能假设有 UI 可用
- N7 已在 `static-keys.ts` 保住了 4 语种现成译文，下轮重建成本可控

### R5 · 追认 N4 删除 `subscription-plans-card.tsx`（派工写的是「改造」）
门下省已在干净基线独立复扫：3 处命中**全在该文件自身**，零 import、零 barrel 再导出；
对照组 `my-subscription-card.tsx` 用同样扫法命中了真实 import，证明扫法有效非假阴性。
确系死代码，改造它反而会凭空造出第二条档位展示路径，违反「不造平行路径」。**追认删除。**

### R6 · 追认 N2 给 `AdminBindSubscription` 新增 `RecordLog(LogTypeTopup)`
契约无此项，净增 1 行。「真金白银入账必须留痕」的判断成立，改造前无日志本身是缺陷。**追认。**

### R7 · 追认 N6 删 `CountUserSubscriptionsByPlan`
N2 曾主张保留（依据「5 个支付 controller 依赖」），但 N3 已把 5 处 controller 预检全删
（预检查的是被删掉的 `UserSubscription` 表）。全仓 grep 零命中，确系零调用方的兼容垫片。
限购规则没丢：`CreditPlanQuotaTx` → `ensurePlanPurchaseCapTx` → `countPlanPurchasesTx` 数
`subscription_orders` 的 success 单，在事务内，四条链路全覆盖。**追认删除。**

### R8 · `footer.newapi.projectAttributionSuffix` 恢复原转义写法
fr/ja/ru/vi 四语种该 key 里的 `a` 被 N5 的 JSON 往返去转义成明文。
解析结果等价、无功能风险，但属与本轮无关的噪音 diff，违反红线 1「趋近于零的改动面」。
**裁决：恢复原转义写法**（4 处单行 revert），不要留给下一轮的人重新纠结。

## N6 收口任务

1. `go build ./...` → 修所有残留引用（不得新增兼容垫片，按契约删干净）
2. `rg 'CreateUserSubscriptionFromPlanTx|UserSubscription\b|SubscriptionPreConsumeRecord|BillingPreference|SubscriptionFunding'` → 确认零残留（`SubscriptionOrder`/`SubscriptionPlan` 应仍在）
3. 测试修订（1.5 节）：
   - 需删：`model/subscription_promo_order_test.go` 大部分、`model/payment_method_guard_test.go:142,161`、`service/task_billing_test.go:103,139,175,234`、`service/waffo_pancake_test.go:174-272`、`model/task_cas_test.go:45-47` 三行 AutoMigrate
   - 需改写：`service/distribution_promotion_test.go:27` 的 `seedSuccessSubscriptionOrder` 被 4 个晋升测试依赖 → 改成用 TopUp 或档位订单造数据
4. `go test ./...` 全绿
