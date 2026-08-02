# 架构决策 — 订阅下线 · 计费口径切换到 ¥1=$1（站1）

## 目标

站1 存在两套并行额度体系（钱包 ¥1=$1 / 订阅池杠杆 10×~35.7×），价格差 30 倍无法解释。
**只留钱包一套额度**：套餐退化为「折扣充值档位」，买完直接进 `users.quota`；
分组倍率整体 ÷30 把批发杠杆挪出来；模型价格保持官方价不动。

额度**永久有效、不限时** → duration / reset period 全是死概念。

## 锁定决策（不要再问，直接执行）

| # | 决策 | 取值 |
|---|---|---|
| D1 | 删到哪一层 | **只删额度池**：`user_subscriptions` + `subscription_pre_consume_records` 两表及全部订阅计费分支。`subscription_plans` / `subscription_orders` **保留**，语义降级为「充值档位」+「档位购买订单」 |
| D2 | 前端接口 | **一个都不删**，只删展示入口和管理端的订阅列 |
| D3 | 存量订阅剩余 | 按实付人民币比例折进钱包（迁移工具负责，本轮不做） |
| D4 | 存量钱包 | 数字原封不动 |
| D5 | 分组倍率除数 | ÷30（迁移工具负责，本轮不做） |
| D6 | 模型价格 | ModelRatio / CompletionRatio / CacheRatio 一个都不动 |
| D7 | 充值档位 | 9 档阶梯，10 折 → 9.2 折封顶，永久有效（迁移工具负责，本轮不做） |
| D8 | 迁移工具形态 | 后台常驻功能，放 switcher，**本轮不做** |
| D9 | 前端范围 | switcher/frontend（生产）和 new-api/web/default（备用）都改，保持一致 |
| D10 | 分销体系 | **零改动**（D1 保留了 plans/orders，外键和订单投影都还在） |

## 本轮范围

**只做第一节（new-api Go + web/default）**。迁移工具（第三节）不做。

## 红线

1. **new-api 保持趋近于零的改动面**，方便将来跟上游 — 只删订阅池相关，不做无关重构
2. **不要误删**（这三处与本地订阅体系无关）：
   - `router/dashboard.go:18-19` → `controller/billing.go:11-67` 的 `GET /dashboard/billing/subscription` 是 **OpenAI 兼容 API**，只是把用户 quota 包装成 OpenAI billing 响应
   - `controller/channel-billing.go` 同理（探测上游渠道余额）
   - `controller/subscription_promo.go` 已是禁用状态（`:28-41` 无条件返回「优惠码功能已禁用」），**不用动**
3. **保留 `subscription_plans` / `subscription_orders`** 两个 model 与 AutoMigrate — D1 明确
4. **`billing_source` 日志字段保留并恒为 `wallet`** — 避免历史日志解析炸掉
5. **不造平行路径**：改造既有 seam，不新增重复实现

## 已知行为变化（可接受，需写进公告）

- `downgradeUserGroupForSubscriptionTx` 随订阅池消失 → **分组升级从此单向**，需要降级由管理员手工改
- `shouldTrust()` 旁路重新生效：原先对订阅禁用信任额度旁路，纯钱包后恢复（性能变好，但要跑高并发回归确认无超扣）
