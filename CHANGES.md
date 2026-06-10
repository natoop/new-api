# feature/njc/agent-phase2 — 商业化闭环 + UI 升级

基于 `feature/njc/agent` 的二期改造。覆盖 processInfo.txt 全部 6 项需求与 2 个 Bug，另含虎皮椒支付通道、进站签约、代理计划营销页、GosWith 品牌化与全站视觉升级。

---

## 1. 管理员订单查询（processInfo #1）

管理后台「代理管理」新增**订单查询 Tab**：按购买客户（用户名/邮箱/昵称模糊）、订单时间范围、订阅套餐（下拉）、订单状态筛选 p3_orders，分页展示订单号/买家/套餐/实付/佣金/状态/时间。

- 后端：`service/distribution_orders.go`（新）、`controller/distribution.go`、路由 `GET /api/agent-admin/orders`（AdminAuth）
- 前端：`features/distribution/agent-admin.tsx` orders Tab、`api.ts`/`types.ts`/`labels.ts`

## 2. 自然晋升代理（processInfo #2）

邀请 ≥3 名用户 **且** 其中 ≥3 人实际消费（订阅订单或充值 success），在**下单成功**时自动晋升为代理（level 2）。

- 邀请口径：`User.InviterId`（与既有客户归属同步逻辑一致）
- 触发点：`CompleteSubscriptionOrder`（四个支付 webhook 统一收口，含虎皮椒）与 `PurchaseSubscriptionWithBalance`，仅 pending→success 真实翻转后异步触发（goroutine + recover，不阻塞支付回调）
- 晋升复用既有 `ensureDistributionAgentForUser`（幂等：唯一键冲突视为已晋升；含提权与客户归属同步）
- 文件：`service/distribution_promotion.go`（新）、`model/subscription.go`（SubscriptionPaidHook 钩子）、`main.go`（注册）

## 3. 钱包页改版（processInfo #3、#4）

新结构（自上而下）：钱包统计 → **兑换码置顶卡** → 代理横幅（代理→代理中心入口；非代理→代理计划营销入口）→ **Tabs[购买套餐｜兑换码]** → 推荐奖励。移除"添加资金"独立卡与独立套餐卡（文件保留未删，便于回退）。

- 购买套餐 Tab：营销化套餐栅格（大字价格、权益清单、推荐档高亮）、促销码输入实时折价（划线原价 + 折后价）、支付方式：余额 / 微信 / 支付宝（虎皮椒，扫码 + 收银台 + 5s 轮询自动完成）/ Epay/Stripe/Creem/Waffo（按启用状态条件渲染）
- 文件：`features/wallet/index.tsx`、`components/{redemption-card,agent-banner,wallet-tabs,purchase-plans-tab}.tsx`、`dialogs/plan-purchase-dialog.tsx`

## 4. 兑换码类型化（processInfo #5）

`Redemption` 新增 `type`（balance/plan/promo）、`plan_id`、`discount_bps`、`max_uses`、`used_count`：

| 类型 | 语义 | 行为 |
|---|---|---|
| balance（默认） | 原始兑换码 | 兑余额，旧码零迁移完全兼容（空 type 视同 balance） |
| plan | 库存码 | 直接开通订阅套餐（事务内走 CreateUserSubscriptionFromPlanTx） |
| promo | 促销码 | 不可直接兑换；购买套餐时输码按 discount_bps 万分比折价，支持次数上限（原子消耗） |

- 折价唯一收口 `ApplyPromoDiscount`（decimal 分位精度）；余额购买事务内严格消耗，Epay/Stripe/虎皮椒回调时消耗；Stripe 经动态一次性 Coupon 实现
- 兑换接口（POST /api/user/topup）响应 data 由 int 改为 `{quota, type, plan_id, plan_title}`（**breaking**，前端已适配）
- 校验接口：`POST /api/subscription/promo/validate`
- 管理页：类型 Select 联动表单（编辑时类型锁定）、三色类型列、facet 筛选
- 文件：`model/redemption.go`、`service/promo.go`（新）、`controller/{user,redemption,subscription*}.go`、`features/redemption-codes/*`

## 5. 支付对接（processInfo #6）— 新增虎皮椒（XunhuPay）

微信 + 支付宝双通道（两套独立 appid/appsecret），订阅购买与余额充值双场景：

- 下单：`POST /api/subscription/xunhu/pay`、`POST /api/user/xunhu/pay`（body 含 `pay_type: wechat|alipay`，订阅可带 `promo_code`）
- 回调：`POST /api/subscription/xunhu/notify`、`POST /api/user/xunhu/notify`（MD5 验签 constant-time 比较、金额 ±1 分容差比对、幂等、成功返回纯文本 "success"）
- 管理面板「支付设置」新增虎皮椒配置卡；option keys：`XunhuEnabled / XunhuWechatAppId / XunhuWechatAppSecret / XunhuAlipayAppId / XunhuAlipayAppSecret / XunhuGatewayUrl`
- 文件：`setting/payment_xunhu.go`、`service/xunhu.go`（+测试）、`controller/{subscription_payment_xunhu,topup_xunhu}.go`

## 6. 进站签约（新增）

登录进入控制台后，若未签署当前版本协议 → 阻断式弹窗（Esc/点外无效、无关闭钮），Markdown 渲染协议内容，勾选确认后【同意并继续】，【拒绝并退出登录】走完整登出。

- 签署记录入库 `user_agreement_consents`（user_id + version 复合唯一，含 IP，保留各版本审计）；改版本号即要求全员重签
- 逃生门：未启用或版本号为空绝不弹窗
- 设置项：`legal.user_agreement_version`、`legal.console_agreement_enabled`（管理面板法务区可配）
- API：`GET /api/user/agreement/status`、`POST /api/user/agreement/consent`
- 文件：`model/user_agreement.go`（新）、`controller/user_agreement.go`（新）、`components/agreement-gate.tsx`（新）
- **默认协议文案见文末附录 A**（管理员粘贴进「用户协议」设置即可）

## 7. 代理计划营销页（新增）

`/agent/guide`（所有登录用户可访问）：Hero + 晋升进度卡（已邀请 x/3 · 已付费 x/3 双进度条 + 邀请链接复制）+ 权益四卡 + 三步时间线 + FAQ + 「想了解更多合作权益？联系我们」CTA。已是代理则显示欢迎态 + 进入代理中心。

- **保密约束**：页面不出现层级结构、一级/二级、上级代理、具体佣金比例等任何敏感信息
- 配套接口：`GET /api/user/agent-progress`
- 文件：`features/distribution/agent-guide.tsx`、`routes/_authenticated/agent/guide.tsx`、`controller/distribution_progress.go`

## 8. Bug 修复：后台导航显隐失效（processInfo bug #1）

管理员隐藏的顶栏/侧栏模块此前用户端仍显示、且可在 Profile 个性化重新打开。修复：

- 管理员配置为**硬上限**（最终可见性 = 管理员 AND 用户），用户只能做减法
- Profile 个性化不再渲染被管理员关闭的模块开关，保存时剔除越权 key
- 堵住三处 fail-open 旁路：header 静态链接回退、布尔简写 section 绕过、公开页顶栏同款
- 防"导航全消失"：配置整体缺失/坏值时回退默认全显
- 文件：`lib/nav-modules.ts`、`hooks/use-sidebar-config.ts`、`features/profile/components/sidebar-modules-card.tsx`、`components/layout/components/{app-header,public-header,public-navigation}.tsx`

## 9. Bug 修复：i18n（processInfo bug #2）

- 本期全部 114 个新增文案同步写入 6 个 locale（zh/en/fr/ru/ja/vi 各 4811 key，零缺失、占位符校验通过）
- 语言切换器只显示 **简体中文 / English**（locale 文件全保留；fr/ru/ja/vi 既有用户偏好自动回退 English）
- 顺带修复基线已存在的 4 处占位符损坏（ja×3、vi×1）

## 10. UI 升级 + GosWith 品牌化

- **设计语言**：靛蓝→紫品牌色体系（light `#4f46e5` 级）、`--radius` 0.75rem、弱阴影 + hairline 边框、系统字体栈（SF Pro/PingFang）、深色模式近黑灰底
- **逐页**：侧栏 active 左侧品牌色细条 + 浅靛蓝高亮；Header backdrop-blur；Dashboard 统计大字化；登录/注册居中卡片化；图表色轴品牌化
- **品牌**：默认站名 GosWith（`SystemName` 默认值，面板可改）、`public/goswith-logo.svg` 渐变 G 字标、favicon、前端 fallback 收口到 `lib/constants.ts`
- **未动**：上游项目署名（页脚归属/关于页/README/NOTICE/LICENSE/模块路径）一律保留，符合 Apache-2.0 署名要求

## 11. 终审加固（三轮独立审查后的修复）

**安全/资损面：**
- 促销码改为**下单时原子预占**次数（Epay/Stripe/虎皮椒统一走 `CreateSubscriptionOrderWithPromoReserve`），用尽即拒单；订单过期/拉起失败自动回补，支付完成不再二次消耗——堵死"挂多笔 pending 单绕过次数上限"
- 修复 GORM v2 下 `FOR UPDATE` 行锁静默失效（改 `clause.Locking`），兑换码并发双兑换/订阅双开通不再可能；Redeem 另加条件 UPDATE 抢占双保险
- 虎皮椒充值回调"翻单 + 加额度"原子化（同一事务，失败整体回滚由网关重试），杜绝丢额度
- 促销码校验接口补防枚举 RandomSleep；协议版本号长度校验；停用套餐的库存码拒绝兑换

**口径/体验：**
- 充值成功也触发晋升检查（与"实际消费含充值"的统计口径对齐，全部充值通道挂钩）
- 协议内容为空时不弹签约（第二道逃生门）；管理员/Root 不再被误导向代理中心 403（角色语义统一）
- 订阅购买弹窗剔除误入的 waffo 支付项；Stripe 拉起失败补提示；金额符号按套餐币种显示；语言偏好残留错位自动纠正
- 顺手修复两处存量事务内死锁（订阅开通/完成路径）

## 数据迁移

全部 GORM AutoMigrate 自动完成，零手工脚本：`redemptions` 新列（旧数据默认 balance）、新表 `user_agreement_consents`、`subscription_orders.promo_code` 列。

## 升级注意

1. 兑换接口响应结构变化（int → 对象），仅影响自定义客户端
2. 启用进站签约：法务设置中开启「控制台强制签署」并填版本号（如 v1.0），协议内容粘贴附录 A
3. 虎皮椒：支付设置中配置对应通道 appid/appsecret；微信/支付宝按配置自动出现在购买面板
4. 促销码次数为下单时预占：挂起未付的 pending 订单会占用一次名额直到过期自动回补（属预期行为）
5. Stripe 促销折扣经一次性 Coupon 实现，仅首期生效，每次折扣下单会在 Stripe 后台产生一个 Coupon 对象

---

## 附录 A：默认《用户服务协议》文案

> 管理员可直接粘贴至「系统设置 → 法务 → 用户协议」，并设置版本号（如 v1.0）后开启控制台强制签署。请在【】处填入实际主体与联系方式，并建议由法律顾问复核后正式启用。

### 用户服务协议

**版本：v1.0｜生效日期：【填写】**

欢迎使用本平台（以下称"本服务"）。本服务由【运营主体名称】（以下称"我们"）提供。请您在使用前仔细阅读并充分理解本协议全部条款；您勾选同意或实际使用本服务，即视为已接受本协议。

**一、服务性质**
1. 本服务为人工智能模型应用程序编程接口（API）的聚合与转发平台，为您提供统一接入、用量计费与账户管理功能。
2. 模型能力由第三方模型服务商提供，其输出内容由模型自动生成，不代表我们的观点；我们不对生成内容的准确性、完整性、适用性作出保证。

**二、账户与安全**
1. 您应提供真实有效的注册信息，并妥善保管账户凭据与 API 密钥；因保管不善造成的损失由您自行承担。
2. 账户仅限您本人（或您所代表的组织）使用，不得出借、转售或共享；我们有权对异常使用行为采取限制、冻结等措施。

**三、付费、退款与发票**
1. 本服务按套餐订阅或按用量预付费计费，具体价格以购买页面公示为准。
2. 充值余额与已开通套餐一经使用即发生消耗；除法律法规另有规定或我们书面同意外，已消耗部分不予退还。未消耗部分的退款申请将在核实后于合理期限内处理。
3. 促销码、兑换码等优惠权益不可兑现、不可转让，解释权在法律允许范围内归我们所有。
4. 如需发票，请通过本协议文末联系方式与我们联系。

**四、使用规范（禁止性条款）**
您承诺不利用本服务从事下列行为，否则我们有权立即中止或终止服务且不退还费用，并保留追责权利：
1. 违反中华人民共和国及您所在司法辖区法律法规的行为；
2. 生成、传播危害国家安全、淫秽色情、暴力恐怖、虚假信息、侵害他人名誉权/隐私权/知识产权的内容；
3. 实施网络攻击、恶意爬取、绕过计费、共享转售接口等破坏服务秩序的行为；
4. 将服务用于医疗诊断、法律意见等高风险场景而不加人工审核。

**五、数据与隐私**
1. 我们遵循最小必要原则收集与处理您的信息（账户信息、用量日志、支付记录），用于提供服务、计费结算与安全审计。
2. 您提交的请求内容将被转发至相应模型服务商以完成调用；我们不会主动将其用于与服务无关的目的。
3. 具体规则详见《隐私政策》；法律法规要求披露的情形除外。

**六、服务可用性与免责**
1. 我们以"现状"提供服务，并尽商业上合理的努力保障可用性，但不承诺服务不中断、无错误。
2. 因第三方模型服务商故障、不可抗力、网络原因、您自身操作导致的损失，我们在法律允许的最大范围内免责。
3. 在任何情况下，我们对您的全部赔偿责任以您过去十二个月内实际支付的费用总额为限。

**七、协议变更与终止**
1. 我们可根据业务调整本协议，更新后将通过站内公告或弹窗提示；您继续使用即视为接受更新版本。
2. 您可随时停止使用并注销账户；注销前请自行处理余额与数据。

**八、争议解决**
本协议适用中华人民共和国法律。因本协议产生的争议，双方应友好协商；协商不成的，任何一方可向【运营主体所在地】有管辖权的人民法院提起诉讼。

**九、联系我们**
如对本协议或服务有任何疑问，请联系：【联系邮箱/客服渠道】。
