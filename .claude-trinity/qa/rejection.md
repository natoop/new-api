# 门下省驳回报告 — new-api 订阅池下线

审查对象：`feature/njc/drop-subscription-pool`（N1–N7 产出，工作区未提交）
裁决：**五维全部 rejected**（第 1 次驳回 / 上限 3）

---

## 🔴 FATAL — 必须修完才能上线

### F1 · 限购预检删除 → 用户付款后订单永久卡死，无退款路径
命中维度：1（major）+ 3（fatal）—— 两维独立发现同一问题

五个在线支付入口（`epay:54` / `stripe:57` / `creem:62` / `xunhu:79` / `waffo_pancake:55`）
的 `MaxPurchasePerUser` 预检被删（删得对：它们查的是已删除的 `UserSubscription` 表），
但**没有用 `SubscriptionOrder` 补回等价预检**。限购唯一执行点挪进了 `CreditPlanQuotaTx`。

失败链条：
1. 档位 `MaxPurchasePerUser=1`，用户已有 1 笔 success 订单
2. `GetSubscriptionPlans` 不按上限过滤，前端全库 0 处引用 `max_purchase_per_user` → 档位照常展示
3. 用户第二次下单 → 订单落 pending → 跳转网关 → **真钱被收**
4. 回调进 `CompleteSubscriptionOrder` → `ensurePlanPurchaseCapTx` 拒绝 → **整个事务回滚**
5. 订单永远停在 pending、额度一分未加、`top_ups` 无记录、**无任何退款或告警路径**
6. epay/xunhu 回调写 `"fail"`、waffo 返回 500 `"retry"` → **网关永久重试且每次都失败**，只能人工对账发现

**修法**（两处都要，复用同一 seam，不要新写第二套计数）：
- 导出 `CountPlanPurchases(userId, planId)` 包装 `countPlanPurchasesTx`，在五个 `SubscriptionRequest*Pay` 的
  plan 加载后、创建订单前做预检
- `CreditPlanQuotaTx` 里的事务内守卫**保留不动**作为兜底
- ⚠️ 兜底命中时**不能整体回滚成 pending**，必须落明确终态（订单置 failed + 告警），否则钱和额度对不上

### F2 · `gorm:query_option` 是假锁 → 限购并发穿透 + 余额可打负
`chargeWalletForPlanTx:683` 的 `tx.Set("gorm:query_option", "FOR UPDATE")` 在 **GORM v2 下完全无效**
（已核实 `gorm.io/gorm v1.25.2` 即 GORM v2，只有 `clause.Locking` 会生成 FOR UPDATE）。
函数注释 `:675-676` 写着「locks the buyer row」**是错的**——实际没有任何锁。
「文档承诺了不存在的安全属性」比没锁更危险。

叠加 `ensurePlanPurchaseCapTx` 的 COUNT 既无锁又无唯一索引，且三条落单链路都是「先 credit 后写 success 订单」：
并发 N 个 `POST /api/subscription/balance/pay` → 全部读到 count=0 → 全部通过上限 → 全部入账。

`CriticalRateLimit` 默认 20 次 / 20 分钟（`common/init.go:123-124`）→ 单窗口可穿 20 次。
具体亏钱算式：促销档位 `price=$1`（requiredQuota=500000）、`TotalAmount=$10`、cap=1，
并发 20 次 → 净白拿 `20 × (5000000 − 500000)` = 9000 万 quota ≈ **$180 / 20 分钟，可无限循环**。

同一竞态也让余额校验失效：`user.Quota` 读-判-写无锁，两个并发可把余额打成负数。

**修法**：
1. 改成 `tx.Clauses(clause.Locking{Strength:"UPDATE"})`（本文件 `:440` `:593` 已经是对的写法，照抄）
2. 限购不能只靠 COUNT —— 在 `subscription_orders` 加 `(user_id, plan_id, status)` 相关唯一约束或购买序号列做 DB 兜底
3. `CreditPlanQuotaTx` 前先对 `users` 行取排他锁，让同一用户的并发购买串行化

⚠️ **仓库级存量缺陷**（非本轮引入，建议单独立项）：同样的假锁还有 9 处 ——
`model/topup.go` ×6、`model/user.go`、`model/redemption.go`、`service/distribution_transactions.go`。
即**所有在线充值完成路径本来就没有行锁**。本轮新代码不应继续复制这个写法。

### F3 · `users.quota` 列宽溢出风险（**待生产核实，见文末**）
`model/user.go:40` 是 `gorm:"type:int"`，在 PostgreSQL 下即 `integer`（4 字节，上限 2147483647，
按 `QuotaPerUnit=500000` 只有 **$4294.97**）。而 `SubscriptionPlan.TotalAmount` 是 bigint。

改造前档位额度进的是 `user_subscriptions.amount_total`（bigint），
改造后**全部灌进 `users.quota`**，`gorm.Expr("quota + ?", plan.TotalAmount)` 无任何上溢检查。

- PostgreSQL / MySQL strict → out of range 报错 → 事务回滚 → **退化成 F1 那条「钱收了、订单卡 pending」**
- MySQL 非严格模式 → 静默截断到 2147483647，用户直接损失约 $705

**但这条与交接文档的实测数据矛盾**：文档称管理员 id=1 持 $99,997,515（≈5e13 quota），远超 int32。
GORM AutoMigrate 不会收窄既有列 → 生产列很可能早已是 bigint。
**必须先在生产库核实再决定是否要做迁移**（SQL 见文末）。

---

## 🟠 MAJOR

### M1 · 两条资金链路零测试覆盖（维度 4 驳回理由）
- `AdminBindSubscription` —— **完全零测试**。其独有逻辑（不落 `SubscriptionOrder` 以免污染付费统计、
  升组提示文案、净额传递）全部未被断言
- `RedeemDistributionInventoryCode` —— 只有「已兑换→拒绝」反向守卫，
  **成功入账那条路径无任何断言**（额度加没加、加了多少，无人验证）

### M2 · `planTopUpAmount` 截断方向恒为少记
`decimal.Div(...).IntPart()` 向零截断。`TotalAmount` 不能被 `QuotaPerUnit` 整除时余数直接丢：
`$1.5` → `Amount=1`（少记 $0.5）；`< $1` 的档位 → `Amount=0`，用户充值历史显示为 0，
且这类行若处于 pending，`ManualCompleteTopUp` 会以「无效的充值额度」硬拒。
新增测试只用了整除值（`5 × QuotaPerUnit`），完全没覆盖截断。

### M3 · 部署窗口内在途异步任务错配资金来源
`taskIsSubscription` 被删，`taskAdjustFunding` 不再按 `PrivateData.BillingSource` 分支。
部署前已从订阅池预扣的在途任务（视频/MJ 可跑数分钟到数小时）：
- 失败 → `RefundTaskQuota` → **把从未从钱包扣过的额度凭空加进钱包（白送）**
- 成功 → `RecalculateTaskQuota` 拿钱包补差额，订阅池那笔预扣永久悬空

**修法**：保留一个只读历史分支（`BillingSource=="subscription"` 的旧任务只调令牌额度并记日志跳过钱包），
或**部署前先排空异步任务队列**。

### M4 · 发布门禁：存量订阅额度无迁移则纯蒸发
两表从 `migrateDB` / `migrateDBFast` 移除、表被孤立但不删，
全仓无任何把 `amount_total - amount_used` 折算进 `users.quota` 的代码。
架构决策里这是有意识延后（D3/D8「迁移工具负责，本轮不做」），
但**延后的是工具、不是风险**：迁移工具跑通前不得部署本分支。

### M5 · 45 个前端文件改动零真机验收（维度 5 驳回理由）
`.claude-trinity/` 无任何截图 / e2e / 控制台记录；`web/default` 无 playwright/cypress/puppeteer 依赖。
本轮风险点恰好都是静态检查看不见的：
- `GET /subscription/self` 响应形状被换（新增 quota、列表 null→`[]`），消费它的组件被删，两边各自编译通过，**页面加载报不报错没人看过**
- `routeTree.gen.ts` 靠起 dev server 再掐掉产出，删掉的 `/subscriptions` 路由访问会怎样没验证过
- 兑换码 `type=plan` 分支新文案需真造代理库存码才能触发，tsc 永远测不到

**需覆盖 6 个场景**：钱包充值档位 tab / 兑换码 plan 分支 / 用户管理 row actions /
日志 Cost 列与详情弹窗（含历史 `billing_source=subscription` 旧日志不炸）/ 侧栏与系统设置 /
直接访问已删的 `/subscriptions` 路由。每页同时核 console 无 error + 网络请求返回新形状后不报错。

---

## 🟡 MINOR

| # | 位置 | 问题 |
|---|---|---|
| m1 | `model/subscription.go:726` | `PurchaseSubscriptionWithBalance` 42 行 > 40（抽 `newBalancePurchaseOutcome` 可降到 ~32） |
| m2 | `model/subscription.go:730` | `_ = strings.TrimSpace(promoCode)` 纯 no-op 占位符 |
| m3 | `service/distribution_transactions.go:902` | 本轮新增的 `_ = model.InvalidateUserCache(userID)` 静默丢错，与本轮自己写的 `applyPlanPurchaseSideEffects`（会 SysLog）不一致 |
| m4 | `purchase-plans-tab.tsx:67` | 本轮删掉 `catch` → 从「静默吞」变「未捕获 rejection」；接口挂了与没配档位界面上无法区分 |
| m5 | `model/subscription.go:292` | `getSubscriptionPlanByIdTx` 缓存命中即返回、不走 tx；「决定给多少钱」的读取不该走缓存（TTL 300s 窗口内按旧价入账） |
| m6 | `model/subscription.go:531/643` | `planTopUpAmount` / `calcSubscriptionBalanceQuota` 直读全局 `common.QuotaPerUnit`，应由调用方注入才能做边界测试 |
| m7 | `CHANGES.md:37` | 仍写「事务内走 `CreateUserSubscriptionFromPlanTx`」，函数已删、行为已反 |
| m8 | `plan_credit_test.go:125` | 测试名 `...FailedOrders` 实际种的是 pending 订单 |
| m9 | `controller/subscription.go:187/254` | reset 体系已下线，仍用 `NormalizeResetPeriod` 校验废弃的 `quota_reset_period` |
| m10 | fr/ja/ru/vi | 缺 8 个新充值档位 key（需真机确认回落英文而非渲染裸 key） |
| m11 | `controller/user.go:1232/1234` | 库存码兑换仍返回「套餐已开通」且 `quota` 硬编码 0，前端只能报档位名不报金额 |

---

## ✅ 审查通过的部分（不需返工）

- **契约面**：C1–C9 全部逐条实地验证通过；D1 保留项（`SubscriptionPlan`/`SubscriptionOrder` + AutoMigrate）完好；
  红线四文件 `git diff --stat` 为空逐字未改；R1/R2/R3 实现与裁决文本一致
- **事务纯度**：`CreditPlanQuotaTx` 三步全落在传入的 tx 上，无一步偷用全局 DB
- **同事务回滚**：实跑验证扣款（:738）先于入账（:742），构造入账失败后余额回到初值、无残留订单
- **回调重放幂等**：五条支付路径逐条核实，`loadPendingSubscriptionOrderTx` 用的是**真行锁** `clause.Locking`，
  六个调用点全包在 LockOrder/UnlockOrder 内，重放不会重复加额度
- **跨表双花已排除**：订阅镜像的 `top_ups` 行恒 `PaymentProvider` 为空，所有钱包入账路径要么校验 provider 要么幂等返回
- **限购计数口径无 off-by-one**：三条落单链路均 credit 在前、写 success 订单在后，COUNT 正确排除本次在途购买
- **`shouldTrust` 旁路恢复不产生超扣/少扣**：信任命中时 preConsumed=0，Settle 全额扣，`needsRefundLocked` 正确返回 false
- **没有靠删测试掩盖问题**：`payment_method_guard_test.go` 把 `assert.Zero(countUserSubscriptions)`
  换成等强度新不变量 `assert.Equal(0, getUserQuota)`，语义强度不降反升
- **平行路径干净**：新组件是纯展示、新 API 加进既有 `wallet/api.ts`、`planTopUpAmount` 是补齐反向换算非重复实现
- **粒度**：新建文件最大 248 行、新增函数最高圈复杂度 10、最深嵌套 3、参数全 ≤4；
  TODO/硬编码/注释代码块全部 0 命中

---

## 上线前必须由人执行的核实

**F3 的前提**——在生产站1（23.80.83.29）的库上跑：

```sql
SELECT column_name, data_type, numeric_precision
FROM information_schema.columns
WHERE table_name = 'users' AND column_name IN ('quota','used_quota');
```

- 返回 `bigint` → F3 不成立，降级为「补一个溢出守卫」的 minor
- 返回 `integer` → F3 成立且是**上线阻塞**，必须先做列加宽迁移
  （同时说明交接文档里管理员 $99,997,515 那个数字的来源需要重新核对）
