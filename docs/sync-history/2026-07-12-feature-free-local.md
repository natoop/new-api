---
report_type: upstream-sync-audit
generated_at: 2026-07-12T03:29:38+08:00
timezone: Asia/Shanghai
repository: D:\code\goswtich\new-api
branch: feature/free/local
tracking_branch: origin/feature/free/local
requested_window_start: 2026-07-10T03:29:38+08:00
requested_window_end: 2026-07-12T03:29:38+08:00
baseline_head: adf681cb49c73f467555ee3cda5001185c4e7abd
baseline_synced_upstream: 8739c05c0e2aa96d69faec3b9f76b4d2c7f66108
saved_head: 86eddd4ac38bb4e7f4f631fe5fa432e70447ff66
saved_synced_upstream: 7c28993f6bd9e92616f3f578212577f8b7c40b45
last_sync_merge: 2d674ef347f32d8ec5f7823ba8ce9830257a1de5
upstream_tracking_snapshot: 7c28993f6bd9e92616f3f578212577f8b7c40b45
upstream_tracking_snapshot_fetched_at: 2026-07-12T02:22:36+08:00
scope: synchronized product changes through last_sync_merge; saved_head is the next-run local anchor
---

# 最近 2 天同步更新汇总

## 1. 核心结论

本报告按“最近 48 小时内进入本地分支的上游同步批次”统计，不按提交作者时间简单筛选。

- 下次整理使用的本地 HEAD：`86eddd4ac38bb4e7f4f631fe5fa432e70447ff66`。
- 当前已同步上游锚点：`7c28993f6bd9e92616f3f578212577f8b7c40b45`。
- 最后一笔真实上游同步 merge：`2d674ef347f32d8ec5f7823ba8ce9830257a1de5`，父提交为 `9f8152f0 + 7c28993f`。
- 本轮产品代码净变化范围：`adf681cb49c73f467555ee3cda5001185c4e7abd..2d674ef347f32d8ec5f7823ba8ce9830257a1de5`。
- 本轮纯上游内容范围：`8739c05c0e2aa96d69faec3b9f76b4d2c7f66108..7c28993f6bd9e92616f3f578212577f8b7c40b45`。
- 三批同步的本地 merge patch 与对应纯上游区间一致；第三批的同步 merge、纯上游范围和外层 carrier merge stable patch-id 均为 `cd082cb68d29d688a7b9a1a1bb279f5133322f49`，未发现冲突解决改写。
- 第三批不是当前 HEAD 的 first-parent 直接 merge：它先形成于远端 `2d674ef3`，随后通过外层本地整合 merge `86eddd4a` 进入当前分支。只查 local first-parent 会漏掉它；若保存 HEAD 位于另一条已前进的本地线，`--ancestry-path` 也会漏掉从其祖先处分叉的同步 merge。
- `86eddd4a` 不是第四批上游同步；它只是合并本地报告提交 `3e047f02` 与已包含第三批同步的远端分支。
- 工作区原有未跟踪目录 `.codegraph/` 未纳入统计，也未修改。

同步发生在 7 月 11—12 日，但上游范围包含 7 月 4 日起的旧提交。这是因为 merge 会一次性引入“上次已同步锚点到新锚点”的全部祖先；使用 `--since="2 days ago"` 会漏算本次实际进入分支的旧侧枝提交。

## 2. 总体统计

以下是基线到第三批同步完成后的最终净状态，不是三个批次 stat 的简单相加。第三批回退了第二批的大规模设计系统重构，因此当前净文件数显著缩小。

| 区域 | 文件数 | 新增 | 删除 | 当前净结果 |
|---|---:|---:|---:|---|
| 后端、CI、根目录及共享文件 | 127 | 13,058 | 4,814 | 协议转换、计费安全、Advanced Custom、模型元数据与测试为主 |
| `web/default` | 111 | 4,843 | 1,365 | 保留局部主题、日志、模型管理与业务功能改动 |
| `web/classic` | 1 | 30 | 4 | 仅 Codex 亲和请求头模板形成最终净变化 |
| `web/bun.lock` | 0 | 0 | 0 | 中间批次有变化，第三批回退后与基线一致 |
| 合计 | 239 | 17,931 | 6,183 | 当前同步产品树的真实净变化 |

纯上游完整祖先共 41 个提交，first-parent 主线接收点 30 个；加上本地 3 个同步 merge，产品同步结果相对基线新增 44 个可达提交。当前保存 HEAD 还包含报告提交和外层整合 merge，因此相对基线共新增 46 个可达提交。

## 3. 同步批次

| 批次 | 上游同步 merge | merge 时间 | 本地观察时间 | 上游范围 | 主线/全部祖先 | merge 净变化 |
|---|---|---|---|---|---:|---:|
| 1 | `9017be21` | 07-11 00:54:24 | 07-11 02:05:55 | `8739c05c..0cb741d8` | 8 / 15 | 52 文件，`+1402/-897` |
| 2 | `9f8152f0` | 07-11 11:19:57 | 07-11 11:23:48 | `0cb741d8..dad57a6b` | 6 / 6 | 511 文件，`+9398/-9811` |
| 3 | `2d674ef3` | 07-12 01:02:53 | 07-12 01:59:29 | `dad57a6b..7c28993f` | 16 / 20 | 639 文件，`+25836/-14180` |

### 批次 1 主线提交

| 时间 | 提交 | 主题 |
|---|---|---|
| 07-09 16:26 | `df01273b` | 调整列宽后让表格继续填满可用宽度 |
| 07-09 16:49 | `a79f9691` | 修正推广返利文案及多语言 |
| 07-09 22:03 | `246d62aa` | 删除 v1.0 提交意外复活的旧视频与预扣费死文件 |
| 07-10 10:21 | `4e570389` | 订阅重置改用有效的 GORM v2 行锁 |
| 07-10 23:28 | `4823417c` | Playground 新增参数设置面板 |
| 07-10 23:30 | `d3b01b48` | 自定义模型名允许仅大小写不同 |
| 07-10 23:33 | `f2c7cd33` | 清除定价页泄漏的示例特殊可用分组 |
| 07-11 00:49 | `0cb741d8` | 优化上游价格同步表格、徽章和批量交互 |

### 批次 2 主线提交

| 时间 | 提交 | 主题 |
|---|---|---|
| 07-11 02:02 | `262ab931` | 统一 default 前端设计系统 |
| 07-11 05:13 | `0918bdb4` | 集中设计系统组件并重构响应式数据视图 |
| 07-11 06:02 | `9d1ca545` | 完善数据卡片、定价页、渠道页和过滤工具栏 |
| 07-11 10:52 | `ca971413` | 允许自定义首页 iframe 中用户主动触发顶层导航 |
| 07-11 10:53 | `00f1cbb6` | `golang.org/x/crypto` 升级到 `0.52.0` |
| 07-11 11:04 | `dad57a6b` | 对齐 Codex Responses 字段、模型和请求/响应头 |

### 批次 3 主线提交

| 时间 | 提交 | 主题 |
|---|---|---|
| 07-11 13:02 | `b2a890e7` | 修复不同 workspace 布局下的 Fontsource 路径 |
| 07-11 14:30 | `308e3e34` | 主题化数据视图并增加任务日志详情 |
| 07-11 14:30 | `ad900bbb` | 合入预扣费饱和、图片流断连与计费修复 |
| 07-11 14:51 | `337169e0` | 回退本轮大规模 UI/design-system 重构 |
| 07-11 15:04 | `1b1b23d1` | 恢复 StatusBadge 水平内边距 |
| 07-11 15:32 | `6bbddb10` | 日志增加首 Token、总耗时与 TPS 展示 |
| 07-11 19:25 | `162f8792` | 更新主题颜色和多业务页面视觉层级 |
| 07-11 19:37 | `e4006196` | 增强陈旧实例处理并继续调整主题 |
| 07-11 20:23 | `1250fb2e` | 修正日志列中 StatusBadge 间距 |
| 07-11 20:44 | `c36418c8` | 重构文本协议转换并增强 Advanced Custom 路由 |
| 07-11 21:18 | `48068ce9` | 按缓存创建价计费 OpenAI `cache_write_tokens` |
| 07-11 22:18 | `92d3c9d1` | 修正未缓存余量并转发 compact 缓存键 |
| 07-11 22:36 | `7a2b9d86` | 模型搜索增加状态与同步过滤 |
| 07-11 22:48 | `8283df16` | 增加“未设置价格模型”页签 |
| 07-11 22:57 | `bde9b2f4` | 加固批量复制、反馈与 memo 相等判断 |
| 07-12 00:24 | `7c28993f` | 未定价页签只列出渠道实际启用模型 |

## 4. 后端改动与意图

### 4.1 删除被错误恢复的旧实现，不是删除视频能力

批次 1 删除：

- `controller/swag_video.go`：空处理器与过时 Swagger 注释。
- `controller/task_video.go`：旧异步视频轮询与独立结算路径。
- `service/pre_consume_quota.go`：旧预扣费/退款路径。

当前实际视频链路仍为：

- `router/video-router.go` → `RelayTask` / `RelayTaskFetch`。
- `controller/system_task_handlers.go:151` → `service/task_polling.go:105` → `service/task_polling.go:431`。
- `controller/relay.go:164` → `service/billing.go:19` → `service/billing_session.go:342`。

意图是消除两套轮询和结算实现并存的误用风险，不是关闭视频能力。若以后依赖 Go 注释重新生成 Swagger，需要确认视频文档仍有稳定来源。

### 4.2 订阅重置获得真正的数据库行锁

`model/subscription.go:1030`、`model/subscription.go:1052` 将被 GORM v2 静默忽略的：

```go
tx.Set("gorm:query_option", "FOR UPDATE")
```

替换为 `model/locking.go:20` 的 `lockForUpdate(tx)`。MySQL/PostgreSQL 生成 `FOR UPDATE`，SQLite 跳过不支持的语法。其意图是避免管理员重置、并发扣费、定时重置和差额结算互相覆盖。

### 4.3 第二批分页变更已被第三批回退

第二批曾将 `common.ItemsPerPage` 从 10 改为 20，并把多密钥状态默认页大小从 50 改为 20；第三批 `337169e0` 已回退：

- `common/constants.go:83` 当前仍为 `ItemsPerPage = 10`。
- `controller/channel.go:1491` 当前多密钥状态默认仍为 50。
- classic 通用列表默认仍为 10。

因此“全局 API 默认分页从 10 变 20”不是当前最终行为。default 前端若显式发送 20、100 等页大小，仍按页面自身配置执行。

### 4.4 清除会进入真实业务数据的示例特殊分组

`setting/ratio_setting/group_ratio.go:28` 清空内置 `vip_special_group_1` 等演示映射，避免样例数据进入定价和分组选择。只删除内置样例，不影响管理员已保存的 `group_special_usable_group`。

### 4.5 预扣费从“饱和后继续”改为严格拒绝

第三批在原有配额饱和审计上增加严格转换：

- `common/quota_math.go:20` 为 clamp kind 增加类型，`common/quota_math.go:42` 让 `QuotaClamp` 同时可作为错误。
- `common/quota_math.go:111`、`common/quota_math.go:132` 增加 `QuotaFromFloatStrict`、`QuotaRoundStrict`。
- `relay/helper/price.go:120`、`:171`、`:217`、`:230`、`:295` 在固定价、倍率价、按次价和表达式预扣中统一使用严格转换。
- `service/billing.go:21` 在预扣前拒绝已有 clamp，`:31` 拒绝负数预扣。

意图是让异常大输入在扣费前以 400/不可重试错误失败，避免饱和值进入扣费或出现负信用。

图片计费同步增强：

- `relay/channel/openai/relay_image.go:25` 用实际返回图片数更新固定价倍率。
- 流正常结束时使用 completed event 数；`relay/channel/openai/relay_image.go:149` 在客户端提前断连时保留请求 `n`，防止只收第一张就断开而少付费。
- JSON 转 SSE 路径改用 `gjson/sjson`，避免对大体积 `b64_json` 反复复制。

### 4.6 文本协议转换变为注册表，并进入 Advanced Custom 路由

`c36418c8` 不是单纯移动文件：

- `service/convert.go` 的大型集中实现被拆到 `service/relayconvert/`。
- `service/relayconvert/request_registry.go:45`、`:151`、`:214` 建立请求转换注册表。
- `service/relayconvert/response_registry.go:52`、`:204`、`:245` 建立普通响应和流式响应转换注册表。
- 支持 OpenAI Chat、OpenAI Responses、Claude Messages、Gemini GenerateContent 之间的直接或多跳转换，并记录转换步骤和质量等级。
- `dto/channel_settings.go:86` 的 Advanced Custom route 新增按路径、模型精确值/正则、converter、上游路径和 header/query 鉴权配置；`:132` 按路径+模型匹配，`:301` 做冲突与合法性校验。
- `relay/channel/advancedcustom/adaptor.go:47`、`:223`、`:321` 将路由匹配、请求转换、鉴权和响应转换接入真实 relay。
- `model/ability.go`、`model/pricing.go` 根据 Advanced Custom route 推断模型可用 endpoint；`main.go:105` 将 pricing warmup 移到 channel cache 初始化之后。
- `relay/common/relay_utils.go:40` 新增 URL 日志脱敏，避免 query 鉴权 token 泄漏。

真实意图是让一个 Advanced Custom 渠道按模型和入口协议路由到不同上游协议，而不是只代理一个固定 OpenAI 风格地址。它会影响渠道可用性、价格页 endpoint 展示、请求/响应转换和计费 usage 来源，是本轮后端影响面最大的功能。

边界风险：

- route 使用 first-match；校验能阻止明显重复，但不能证明两个不同正则在语义上不重叠，顺序仍影响结果。
- route resolve 会执行配置校验和正则处理，规则很多时需要关注请求热路径成本。
- Query 鉴权允许自定义参数名，而 URL 脱敏只覆盖常见 key 和包含 token/secret/signature 的名称；自定义 `credential`、`accessKey` 等仍可能出现在错误 URL，Header 鉴权更稳妥。

### 4.7 转换后的响应保留原始计费语义

协议转换后，展示给客户端的协议不一定等于上游计费协议。第三批新增：

- `dto/billing_usage.go:14` 的 `BillingUsage`，保存原始 source、semantic、是否估算和协议命名 usage。
- `service/billing_usage.go:20` 的 `effectiveBillingUsage`，优先还原真实计费 usage。
- `service/text_quota.go:349` 在渠道亲和观察、tiered settlement、日志和扣费之间复用同一份 billing usage。
- `service/text_quota.go:427` 将实际走过的计费路径写入 admin log。

这避免 Claude/Gemini 响应转换为 OpenAI 后，缓存读写 token 被按错误协议语义计费。

`BillingUsage` 可序列化并随 usage DTO 传播；需要确认严格协议客户端是否接受响应中出现额外 `billing_usage` 扩展字段。

### 4.8 OpenAI cache-write 与 compact 缓存字段补齐

- `dto/openai_response.go:263` 新增 `cache_write_tokens`。
- `dto/openai_response.go:275` 用 `max(cached_creation_tokens, cache_write_tokens)` 计缓存创建量，且负数归零，防止重复计费或抵扣。
- `service/text_quota.go:301` 将 `prompt - cached - write` 的负余量钳制为 0。
- tiered 和普通计费都按缓存创建价计算 cache write。
- `dto/openai_responses_compaction_request.go:24`—`:26` 增加 `prompt_cache_key/options/retention`。
- `relay/responses_handler.go:55`—`:57` 已转发这些 compact 字段，关闭了第二批报告中“prompt_cache_key 在非 passthrough 路径丢失”的缺口。

当前 `tools`、`reasoning`、`text` 在规范化 compact 路径仍被明确解析但不转发，代码注释将其定义为客户端兼容字段。如果产品期望代理这些非文档字段，需要单独确认，不再应描述为无意遗漏。

### 4.9 模型状态、同步过滤和未定价模型来源

- `controller/model_meta.go:20`、`:49` 接收 `status`、`sync_official`。
- `model/model_meta.go:194`—`:254` 将字符串/数字状态转为跨数据库 GORM 条件。
- `controller/model.go:332` 和 `model/ability.go:50` 提供当前启用渠道模型集合。

前端“未设置价格模型”因此只检查实际渠道能力中的模型，而不是整个模型元数据表，减少无渠道模型造成的无效待办。

当前无效 `status/sync_official` 字符串会被静默解释为“不筛选”，任意数字也会直接进入过滤条件；搜索响应中的 vendor counts 仍是全量统计，不随当前过滤条件收缩。

### 4.10 依赖与 CI 供应链维护

- `golang.org/x/crypto` 保持升级到 `0.52.0`。
- Docker、release、Electron 和 PR workflow 中的 GitHub Actions 更新并固定到完整 commit SHA。
- 删除过期海报输出文件。

意图是更新构建供应链并降低 tag 漂移风险；需要在真实 CI 上确认新 major action 与现有参数完全兼容。

## 5. 前端改动与意图

### 5.1 大规模 design-system 重构已回退

第二批一度建立 `components/design-system/` 产品适配层、业务导入限制、统一响应式控件尺寸和 494 文件级迁移。第三批 `337169e0` 明确回退：

- 当前 `web/default/src/components/design-system/` 不存在。
- 当前业务代码不再导入 `components/design-system`。
- `.oxlintrc.json` 中对受管控件导入的限制被移除。
- 全局分页、Data Table 大改、定价页结构重排、Dashboard onboarding 删除和 Fontsource 路径调整一并回退。
- `1b1b23d1` 进一步恢复 StatusBadge 水平 padding。

因此本报告不能再把第二批设计系统描述为当前架构。当前仍以 `components/ui/` 为基础，只保留第三批重新加入的局部主题化和业务增强。

### 5.2 主题颜色改为集中语义 token，但没有恢复整套适配层

- `web/default/src/styles/theme.css:112`、`:185` 更新明暗主题 primary。
- `theme.css:124`—`:127`、`:197`—`:200` 明确 warning/info 语义色。
- `web/default/src/components/ui/icon-badge.tsx:24` 新增统一 `IconBadge` variant，`:73` 输出组件。
- Dashboard、Wallet、Profile、渠道抽屉、订阅和日志等页面改用 IconBadge 表达信息层级。
- 回退同时恢复卡片浮起、按钮按压和表格行 stagger 动效；这些动效尊重 reduced-motion，但密集表格翻页/过滤时仍会反复播放。
- 当前同时存在 `data-theme-font` 主题字体路径和旧 `FontProvider` Cookie/class 路径；后者没有实际消费点，形成双轨字体状态。

意图是先用较小边界统一颜色和图标容器，而不是重新引入第二批的全控件设计系统。风险是覆盖页面仍多，且缺少 lint 规则保证后续业务代码持续复用。

### 5.3 Data Table 回到旧组件架构，但保留移动视图能力

- `components/data-table/core/data-table-row.tsx:69` 使用 memo 并显式比较选中状态和列定义，避免 TanStack 稳定 row 对象漏刷新。
- `data-table-row.tsx:107` 为原始字符串/数字统一增加截断与 Tooltip。
- `data-table/layout/card-row-content.tsx:41`、`mobile-card-list.tsx:89` 继续用 column meta 生成移动卡片。
- `features/redemption-codes/components/redemptions-mobile-list.tsx:77` 新增兑换码专用移动列表。

当前存在两个需要跟进的回退风险：

- `data-table/core/data-table-view.tsx:52` 的 `colSpan` 只依赖稳定的 table instance；列显隐变化后，空状态/骨架行可能保留旧 colspan。
- 横向滚动容器不再提供此前新增的 `role="region"`、`tabIndex` 和可读标签，键盘与读屏可访问性下降。

第三批恢复了 `@tanstack/react-virtual` 依赖，但当前源码没有使用点，尚未真正实现虚拟列表。

### 5.4 Usage Logs 增加可读的时延、吞吐和审计详情

- `usage-logs/components/timing-metrics-cell.tsx:68` 展示首 Token 和总耗时，`:165` 展示流式状态和 TPS。
- `usage-logs/components/columns/common-logs-columns.tsx:641`、`:760` 将 Stream/TPS 与 Timing 接入桌面表格。
- `usage-logs/components/usage-logs-mobile-card.tsx:293`、`:301` 接入移动卡片。
- `usage-logs/components/dialogs/details-dialog.tsx:459` 统一展示 token breakdown、计费模式、管理审计、登录审计、充值来源、管理员 IP/节点等信息。
- cache write token 和计费路径也进入详情展示。

意图是把日志从“请求结果表”升级为可定位慢请求、流异常和计费来源的运维工具。时延颜色阈值、移动卡片密度和 admin-only 字段裁剪需要浏览器回归。

注意：`308e3e34` 一度新增独立 `task-details-dialog.tsx`，但同批 `337169e0` 已删除；当前净树仍使用通用 `details-dialog.tsx`。TPS 采用 `completion_tokens / use_time`，代表端到端有效吞吐，不应解释为扣除首 Token 延迟后的纯模型解码速度。

### 5.5 Dashboard onboarding 已恢复，新增视觉层级而非删除功能

第二批曾删除约 670 行 onboarding、真实 Key 查询和 curl 复制；第三批回退后当前状态为：

- `overview-dashboard.tsx:170` 仍生成 curl。
- `overview-dashboard.tsx:359` 仍提供复制可运行 curl。
- `overview-dashboard.tsx:477`、`:486` 仍查询 API keys 和 models。

第三批主要使用 `IconBadge`、主题色和统计卡片层级重新整理 Overview、模型图表、性能健康、公告、FAQ、Uptime。当前不存在“Dashboard onboarding 被移除”的产品风险，但仍需确认真实 Key 的展示和复制权限符合预期。

复制预览只显示脱敏 Key；用户点击复制时才调用 `/api/token/{id}/key` 获取真实值并写入剪贴板，不保存到组件状态。仍应持续验证该接口的归属和权限校验。

### 5.6 模型列表增加状态/同步过滤

- `features/models/components/models-table.tsx:66` 将 `sync_official` 映射到 URL 状态。
- `models-table.tsx:104`—`:138` 把 status/sync 过滤传给列表与搜索 API。
- `features/models/components/models-columns.tsx:407` 展示是否同步官方模型。

意图是让管理员区分启用/禁用模型、参与/不参与官方同步的模型，减少大模型目录下的人工扫描。

### 5.7 “未设置价格模型”变成独立操作页签

- `model-ratio-form.tsx:174` 使用 `unset` variant。
- `model-ratio-form.tsx:178`—`:194` 加载启用渠道模型并处理错误。
- `model-ratio-visual-editor.tsx:220`—`:244` 只构建候选启用模型且筛选基础价格为空的行。
- `model-ratio-visual-editor.tsx:615` 实现批量复制，`:806` 将动作接入选中行工具栏。
- 最终修复确保候选来自 `/api/channel/models_enabled`，不会把模型元数据表中没有渠道能力的条目列为待设置价格。

这部分改变了管理员定价工作流，批量复制、保存反馈、表达式/倍率字段配对和 memo equality 都需要专项回归。

### 5.8 Advanced Custom 编辑器支持按模型拆分同一路径

- `features/channels/lib/advanced-custom.ts:60` 增加 OpenAI Responses → Gemini converter。
- Advanced Custom route DTO 增加 `models?: string[]`，支持精确模型、`re:` 正则和空列表 catch-all。
- `advanced-custom.ts:521`、`:690` 校验重复模型、多个 catch-all、fallback 顺序、入口/上游路径、converter 与鉴权。
- `advanced-custom-editor-dialog.tsx:403`、`:523`、`:790`、`:1013` 提供 JSON 校验、Visual/JSON 切换、模型分路说明和 fallback 置底。

意图是让同一个 `/v1/chat/completions` 或 `/v1/responses` 入口按客户端原始模型选择不同上游协议。前端只验证正则非空，不本地编译语法；无效正则会在后端保存校验时才显示错误。

### 5.9 第一、二批中仍保留的功能

未被 UI 回退撤销的主要功能：

- Playground 参数面板：temperature、top_p、penalty、max_tokens、seed；只发送已启用字段，保留显式 0 和负 penalty。
- 自定义模型名大小写敏感判重。
- 上游价格同步的批量状态处理和字段配对优化。
- 自定义首页 iframe 的 `allow-top-navigation-by-user-activation`。
- Codex 请求头、响应头和模型模板同步。
- 特殊分组样例清理与推广文案修正。

### 5.10 实例管理、Profile、Wallet 与移动列表

- `features/system-info/components/system-instances-panel.tsx:513` 识别 stale 实例，`:677` 支持批量删除。
- 同文件 `:303`—`:332` 对自动 hostname 给出 `NODE_NAME` 配置指引。
- Redemption 增加移动列表。
- Wallet、Profile、安全卡片、订阅抽屉和 API Key 时间展示使用新的图标层级和紧凑布局。

意图是提高多实例运维可操作性，并在不恢复整套 design-system 的情况下统一关键页面。

### 5.11 公开定价页统一计费模式标识

- `features/pricing/components/model-billing-mode-badge.tsx:32` 统一显示 Dynamic Pricing、Token-based、Per Request。
- 卡片、列表、详情页和后端信息区复用同一徽章。
- 缓存价格提高到与输入/输出价格相同的视觉权重。

风险较低，但 `tiered_expr + billing_expr` 当前统一显示 Dynamic Pricing，详情头部不再细分特殊表达式；原始表达式仍可在详细说明中查看。

### 5.12 i18n、Classic 与依赖最终状态

- 新增 timing、stale instance、unset price 等多语言键，七个 locale 同步维护。
- 当前七份 locale 均为 5,179 个 key，现有 `_sync-report.json` 的 missing/extras/untranslated 均为 0。
- classic 当前最终净变化只剩 Codex 亲和请求头模板；第二批分页与样式变化已回退。
- `web/bun.lock` 中间发生依赖调整，但最终与基线无净差异。

## 6. 跨前后端改动意图

1. **协议兼容从散落函数变成可组合能力**：后端注册表负责 Chat/Responses/Claude/Gemini 的请求、响应和流转换，Advanced Custom 决定每个路径/模型走哪条转换链。
2. **计费语义与客户端返回协议解耦**：BillingUsage 保存原始协议 usage，前端日志显示缓存写入和计费路径，避免“转换成功但账算错”。
3. **在预扣阶段阻断极端值**：strict quota、负数拒绝、图片实际数量和断连保护共同覆盖估算→预扣→结算。
4. **前端从大范围重构回到小边界增量**：第三批明确撤销高风险 design-system 迁移，只保留主题 token、IconBadge、日志和模型管理等可验证增强。
5. **把模型配置缺口变成可操作队列**：后端过滤与启用模型集合支撑前端“未设置价格模型”页签。
6. **Codex 缓存和 turn-state 继续闭环**：请求头、SSE 响应头、compact 缓存键和 cache write 计费在 DTO、relay、计费和日志之间同步。

## 7. 风险与验证状态

### 高优先级风险

1. `service/relayconvert` 和 Advanced Custom 同时覆盖多协议、多跳、流式工具调用、usage 还原和 header/query 鉴权，功能面广，需真实上游矩阵验证。
2. `relay/helper` 新增的 `TestModelPriceHelperTieredRejectsPreConsumeOverflow` 当前失败：测试表达式中的超大整数字面量先被表达式编译器拒绝，未进入预期 `QuotaClamp` 路径。安全实现存在，但回归测试没有验证到目标分支。
3. 图片流在正常结束、上游错误和客户端断连下使用不同计费数量，需确认各上游事件类型和“已生成但未传完”语义。
4. cache read、cache write、Claude cache creation 与转换后的 BillingUsage 交叉复杂，必须用真实 usage 样本核账。
5. 未定价模型批量复制和定价保存会同时修改 price、ratio、tiered、billing_mode、billing_expr，属于高风险管理操作。
6. 大规模 design-system 已回退，当前 UI 不再有集中导入规则；后续视觉一致性依赖人工约束。
7. compact 的 `tools/reasoning/text` 当前为“解析但有意不转发”，需要产品确认是否与所接上游期望一致。
8. Advanced Custom Query 鉴权的自定义参数名不一定会被 URL 日志脱敏；正则路由也可能存在无法静态识别的重叠。
9. Data Table 当前有列显隐后的 colspan 缓存风险，并回退了横向滚动区域的键盘/读屏属性。

### 建议浏览器回归

- 模型列表 status/sync 过滤、URL 持久化、搜索和分页组合。
- 未设置价格页签：空状态、只列渠道模型、单模型编辑、批量复制、保存/失败反馈。
- Usage Logs：首 Token、总耗时、TPS、流错误、移动卡片、详情 Drawer、admin-only 信息。
- Dashboard onboarding：Key/模型加载、curl 生成与复制权限。
- Data Table：列显隐后的空状态 colspan、横向滚动键盘操作、读屏标签、翻页/过滤动效。
- Advanced Custom：精确模型、正则、catch-all 顺序、Visual/JSON 切换、Header/Query 鉴权和不同协议转换。
- stale instance 单条/批量删除、实例重新上报后的竞态保护、`NODE_NAME` 提示。
- Playground 显式 0、负 penalty、空 seed 和启用/禁用参数的实际请求体。
- 自定义首页 iframe 顶层导航。

### 本地验证

- `git fetch --prune upstream main`：成功；02:22:36 时 `upstream/main = 7c28993f`。
- 第三批 `9f8152f0..2d674ef3`、纯上游 `dad57a6b..7c28993f`、当前整合 `3e047f02..86eddd4a` 的 stable patch-id 一致。
- `find-sync-merges.ps1` 已覆盖直接同步、`9f815..86eddd` 嵌套同步、`3e047..86eddd` 祖先处分叉后再合入、同 carrier 多批、无新增批次、无效 revision 抛错和 source-history false-positive；第三批 carrier 被识别为 `86eddd4a`。
- `git diff --shortstat adf681cb..2d674ef3`：239 文件，`+17931/-6183`。
- `go test ./model`：通过。
- `go test ./relay/channel/openai`：通过。
- `go test ./service/relayconvert`：通过。
- Advanced Custom 直接 `go test` 在 Windows 临时目录执行测试二进制时报 `Access is denied`；改为 `go test -c` 后显式运行同一测试二进制，结果 `PASS`。
- `go test ./service -run 'Test(PreConsumeBillingRejects|CacheWriteTokensTotal|CalculateTextQuotaSummaryBillsOpenAICacheWriteTokens|CalculateTextQuotaSummaryUses.*BillingUsage|UsageBillingPathForLog|AppendUsageBillingPathForLog)' -count=1`：通过。
- `go test ./relay/common -run '^TestSanitizeURLForLog' -count=1`：通过。
- `go test ./relay/helper -count=1`：存在两类失败。基线已有的超大 uint max-token/image `n` 用例实际已在 JSON 解析阶段拒绝，但错误文本与测试期待不一致；第三批新增的 tiered overflow 用例则被超大整数字面量编译错误提前截断。
- `go test ./relay/common` 的两条 `TestTaskDurationBounds` 断言仍失败；测试与实现均来自基线前的 `d0bd8aac`，不归因于本轮同步。
- 前端 `bun run typecheck` 未执行成功：`web/default` 无可用 `node_modules/tsgo`，报 `bun: command not found: tsgo`。
- locale 只读核对：七个文件均为 5,179 个 key；现有 i18n sync report 的 missing/extras/untranslated 均为 0。
- 未执行完整前端 build、lint 和浏览器回归。

## 8. 当前尚未同步的内容

2026-07-12 02:22:36 +08:00 已成功刷新 `upstream/main`：

- `upstream/main = 7c28993f6bd9e92616f3f578212577f8b7c40b45`。
- `saved_synced_upstream = 7c28993f6bd9e92616f3f578212577f8b7c40b45`。
- 已知待同步范围为空。

这只代表该 fetch 时刻的远端状态。下次整理仍应先 fetch，再比较新 remote-tracking tip；不得把未来远端提交预先写入本次已同步结论。

## 9. 下次同步整理流程

保留本文件作为历史记录，下次同步后新建下一份日期报告。起点：

```powershell
$OldHead = '86eddd4ac38bb4e7f4f631fe5fa432e70447ff66'
$OldUpstream = '7c28993f6bd9e92616f3f578212577f8b7c40b45'
$SourceRemote = 'upstream'
$SourceBranch = 'main'
$SourceRef = "$SourceRemote/$SourceBranch"

git fetch --prune $SourceRemote $SourceBranch
if ($LASTEXITCODE -ne 0) { throw "git fetch failed: $SourceRef" }

$Batches = @(
  & .\.agents\skills\audit-upstream-sync\scripts\find-sync-merges.ps1 `
    -OldHead $OldHead `
    -OldUpstream $OldUpstream `
    -SourceRef $SourceRef `
    -Verbose
)

foreach ($Batch in $Batches) {
  git diff --stat $Batch.PreMergeHead $Batch.Merge
  git diff --name-status $Batch.PreMergeHead $Batch.Merge

  git log --first-parent --reverse `
    --format='%H|%cI|%an|%s' `
    "$($Batch.UpstreamFrom)..$($Batch.UpstreamTo)"
  git diff --stat "$($Batch.UpstreamFrom)..$($Batch.UpstreamTo)"
}

$Batches |
  Format-Table Merge, IntegrationMerge, SyncPatchMatches, IntegrationPatchMatches

$Batches |
  Group-Object IntegrationMerge |
  ForEach-Object {
    $First = $_.Group[0]
    git diff --stat $First.IntegrationBase $First.IntegrationMerge
  }

$NewSavedHead = git rev-parse HEAD
$NewSyncedUpstream = if ($Batches.Count -gt 0) {
  $Batches[-1].UpstreamTo
} else {
  $OldUpstream
}
```

关键约束：

- 候选发现使用完整的 `saved_head..HEAD` 可达差集；不能限制为 local first-parent 或 `--ancestry-path`，否则会漏掉从保存 HEAD 的祖先处分叉、再由外层 merge 带回的真实同步。
- 只接受恰好两个父提交的 merge，排除 octopus。
- 候选 merge 本身若已属于 `$SourceRef` 历史，必须排除；这可避免把上游特性分支中的“merge main into feature”误判为本地同步。
- `M^2` 必须位于 `$SourceRef` 的 first-parent 主线，不能只验证它是任意祖先。
- `merge-base --is-ancestor` 的 exit `0/1/>1` 分别表示 true、false、Git 错误；Git 错误必须中止。
- 每批本地效果统计 `M^1..M`；纯上游范围从上次锚点推进到 `M^2`。
- 若同步经外层 carrier merge 进入当前 first-parent，脚本还会比较 `carrier^1..carrier` 与同一 carrier 内合并后的纯上游范围。
- 完成后保存新的本地 HEAD 和最后一个确认的 `M^2`。

## 10. 重复流程打包结论

| 重复流程 | 支持证据与日期 | 频率/信心 | 推荐形式 | 结论 |
|---|---|---|---|---|
| 上游同步后按前端/后端审计、隔离纯上游与本地 merge 效果、保存双锚点 | 2026-07-11—12 三批同步；2026-07-02 文档分支精确同步；用户明确要求下次再次整理 | 已多次发生，高信心 | Skill | 已创建并修订 `.agents/skills/audit-upstream-sync/` |
| 同步后区分增量失败与仓库既有基线噪声 | 2026-06-29 同步后 lint 债务分流；本轮 Go/前端验证再次出现 | 多次发生，高信心 | Extend existing | 已写入同步报告与技能的验证/范围约束，不另建重叠技能 |
| 常驻专用同步子代理 | 本次适合临时拆分前端、后端、Git 审计 | 中等 | Skip | 改动边界每次不同，固定角色过窄；技能内按需分派即可 |
| 定时生成同步报告 | 当前同步仍由人工 merge 事件触发 | 证据不足 | Skip | 定时任务会产生无变化噪声，等形成固定节奏再考虑 |
| 扩展 `classic-to-default-sync` | 该技能处理 classic→default 的单提交功能对齐 | 高信心不适用 | Skip | 当前输入是上游分支 merge 和双锚点，契约不同，避免重叠 |

创建的技能只保存可复用的 Git 范围、候选识别、报告边界和双锚点流程，不复制本报告中的一次性业务结论。

## 附录：完整上游祖先提交

### 批次 1

| 提交 | 时间 | 作者 | 主题 |
|---|---|---|---|
| `81808d24` | 07-04 23:42 | feitianbubu | remove sample special usable groups leaking into pricing page |
| `df01273b` | 07-09 16:26 | zuiho | let resized tables fill available width |
| `a79f9691` | 07-09 16:49 | CaIon | update referral message |
| `4645ad9d` | 07-09 21:47 | QuentinHsu | keep model selector lists in sync |
| `246d62aa` | 07-09 22:03 | feitianbubu | remove dead files resurrected by v1.0 launch commit |
| `928b4750` | 07-09 22:10 | QuentinHsu | add chat parameter settings panel |
| `4e570389` | 07-10 10:21 | Seefs | use GORM v2 row locking for subscription resets |
| `e8596cab` | 07-10 10:54 | feitianbubu | allow adding custom model names that differ only by case |
| `4823417c` | 07-10 23:28 | 同語 | merge Playground parameter settings panel |
| `d3b01b48` | 07-10 23:30 | 同語 | merge case-sensitive custom model names |
| `f2c7cd33` | 07-10 23:33 | 同語 | merge special usable group sample removal |
| `489c0458` | 07-10 23:37 | QuentinHsu | optimize upstream price sync table |
| `43783286` | 07-10 23:56 | QuentinHsu | polish sync channel dialog layout |
| `6869cd94` | 07-11 00:28 | QuentinHsu | align table badge spacing |
| `0cb741d8` | 07-11 00:49 | 同語 | merge upstream price sync optimization |

### 批次 2

| 提交 | 时间 | 作者 | 主题 |
|---|---|---|---|
| `262ab931` | 07-11 02:02 | t0ng7u | unify design system across default frontend |
| `0918bdb4` | 07-11 05:13 | t0ng7u | consolidate design-system primitives and responsive data views |
| `9d1ca545` | 07-11 06:02 | t0ng7u | refine data-table cards and pricing page layout |
| `ca971413` | 07-11 10:52 | 乾L | allow user-activated top navigation for custom home iframe |
| `00f1cbb6` | 07-11 10:53 | dependabot[bot] | bump golang.org/x/crypto to 0.52.0 |
| `dad57a6b` | 07-11 11:04 | Seefs | sync Codex fields |

### 批次 3

| 提交 | 时间 | 作者 | 主题 |
|---|---|---|---|
| `b2a890e7` | 07-11 13:02 | t0ng7u | fix Fontsource asset resolution |
| `308e3e34` | 07-11 14:30 | t0ng7u | polish themed data views and add task log details |
| `621927f7` | 07-10 21:53 | CaIon | reject saturated pre-consume quota |
| `d9595831` | 07-10 22:30 | CaIon | improve pre-consume quota error handling |
| `269e4ff3` | 07-10 22:48 | CaIon | image stream disconnect and billing adjustments |
| `ad900bbb` | 07-11 14:30 | t0ng7u | merge origin/main |
| `337169e0` | 07-11 14:51 | CaIon | undo UI design-system refactor |
| `1b1b23d1` | 07-11 15:04 | CaIon | restore StatusBadge horizontal padding |
| `6bbddb10` | 07-11 15:32 | CaIon | add timing metrics display for stream logs |
| `162f8792` | 07-11 19:25 | CaIon | update theme colors |
| `e4006196` | 07-11 19:37 | CaIon | enhance stale instance handling |
| `1250fb2e` | 07-11 20:23 | CaIon | adjust StatusBadge margin in logs |
| `c36418c8` | 07-11 20:44 | Calcium-Ion | enhance text conversion and Advanced Custom routing |
| `48068ce9` | 07-11 21:18 | CaIon | bill OpenAI cache_write_tokens |
| `92d3c9d1` | 07-11 22:18 | CaIon | bound uncached remainder and forward compact cache key |
| `7a2b9d86` | 07-11 22:36 | CaIon | add model status and sync filters |
| `8283df16` | 07-11 22:48 | feitianbubu | add unset price models tab |
| `bde9b2f4` | 07-11 22:57 | CaIon | harden unset price batch copy |
| `93e936f7` | 07-11 23:41 | feitianbubu | list only channel models in unset price tab |
| `7c28993f` | 07-12 00:24 | 同語 | merge channel-model-only unset price fix |
