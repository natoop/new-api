---
report_type: upstream-sync-audit
generated_at: 2026-07-12T01:43:22+08:00
timezone: Asia/Shanghai
repository: D:\code\goswtich\new-api
branch: feature/free/local
tracking_branch: origin/feature/free/local
requested_window_start: 2026-07-10T01:43:22+08:00
requested_window_end: 2026-07-12T01:43:22+08:00
baseline_head: adf681cb49c73f467555ee3cda5001185c4e7abd
baseline_synced_upstream: 8739c05c0e2aa96d69faec3b9f76b4d2c7f66108
saved_head: 9f8152f0f17ccccb0dd68c6f9c1fe0d88f6c60eb
saved_synced_upstream: dad57a6bb85becbb99cab26ade7a891508ed7c42
upstream_tracking_snapshot: ad900bbba74b3e9b16b1ef9c549812ada2bb14a0
upstream_tracking_snapshot_fetched_at: 2026-07-11T14:41:49+08:00
scope: only changes already present in saved_head
---

# 最近 2 天同步更新汇总

## 1. 核心结论

本报告统计的是最近 48 小时内进入当前分支的两次上游同步，而不是简单按提交日期筛选。

- 当前保存 HEAD：`9f8152f0f17ccccb0dd68c6f9c1fe0d88f6c60eb`。
- 当前已同步上游锚点：`dad57a6bb85becbb99cab26ade7a891508ed7c42`。
- 本轮本地净变化范围：`adf681cb49c73f467555ee3cda5001185c4e7abd..9f8152f0f17ccccb0dd68c6f9c1fe0d88f6c60eb`。
- 本轮纯上游内容范围：`8739c05c0e2aa96d69faec3b9f76b4d2c7f66108..dad57a6bb85becbb99cab26ade7a891508ed7c42`。
- first-parent 路径只有两次同步 merge，没有夹杂普通本地提交。
- 两次 merge 对本地分支产生的 patch 与对应上游区间一致，未发现额外冲突解决改写。
- 当前工作区原有未跟踪目录 `.codegraph/`，未纳入统计，也未修改。

为什么范围中会出现 7 月 9 日甚至 7 月 4 日的提交：两次同步发生在 7 月 11 日，merge 会把“从上次上游锚点到新锚点之间首次进入本地可达历史的全部提交”带入。若只使用 `--since="2 days ago"`，会漏掉本次同步实际带入的旧侧枝提交；若使用 `--all`，又会混入尚未同步的远端提交。

## 2. 总体统计

| 区域 | 文件数 | 新增 | 删除 | 说明 |
|---|---:|---:|---:|---|
| 后端及仓库根目录 | 16 | 84 | 559 | 删除 542 行复活死代码是主要删除量 |
| `web/default` | 494 | 10,625 | 10,079 | 设计系统、数据表、定价页和多业务页面重构 |
| `web/classic` | 16 | 46 | 20 | 分页默认值与 Codex 模板对齐 |
| `web/bun.lock` | 1 | 10 | 15 | 前端依赖图调整 |
| 合计 | 527 | 10,765 | 10,673 | 两批存在重叠文件，因此总文件数小于两批简单相加 |

上游完整祖先为 21 个提交；加上本地两个同步 merge，本地范围共可达 23 个新提交。正文使用上游 first-parent 的 14 个主线接收点，避免把 PR 内部提交重复描述为独立功能。

## 3. 同步批次

| 批次 | 本地 merge | merge 时间 | 本地观察时间 | 上游范围 | 主线/全部祖先 | 净变化 |
|---|---|---|---|---|---:|---:|
| 1 | `9017be21` | 2026-07-11 00:54:24 | 2026-07-11 02:05:55 | `8739c05c..0cb741d8` | 8 / 15 | 52 文件，`+1402/-897` |
| 2 | `9f8152f0` | 2026-07-11 11:19:57 | 2026-07-11 11:23:48 | `0cb741d8..dad57a6b` | 6 / 6 | 511 文件，`+9398/-9811` |

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
| 07-11 02:02 | `262ab931` | 统一 default 前端字体、语义色、圆角、动效和状态徽章 |
| 07-11 05:13 | `0918bdb4` | 集中设计系统组件并重构响应式数据视图 |
| 07-11 06:02 | `9d1ca545` | 完善数据卡片、定价页、渠道页和筛选工具栏 |
| 07-11 10:52 | `ca971413` | 允许自定义首页 iframe 中用户主动触发顶层导航 |
| 07-11 10:53 | `00f1cbb6` | `golang.org/x/crypto` 升级到 `0.52.0` |
| 07-11 11:04 | `dad57a6b` | 对齐 Codex Responses 字段、模型和请求/响应头 |

## 4. 后端改动与意图

### 4.1 删除被错误恢复的旧实现，不是删除视频能力

删除：

- `controller/swag_video.go`：136 行空处理器与过时 Swagger 注释。
- `controller/task_video.go`：327 行旧异步视频轮询与独立结算路径。
- `service/pre_consume_quota.go`：79 行旧预扣费/退款路径。

当前实际链路仍在：

- 视频路由：`router/video-router.go` → `RelayTask` / `RelayTaskFetch`。
- 异步任务：`controller/system_task_handlers.go:151` → `service/task_polling.go:105` → `service/task_polling.go:431` → `settleTaskBillingOnComplete`。
- 计费：`controller/relay.go:164` → `service/billing.go:19` → `service/billing_session.go:342`。
- 视频 OpenAPI 路径仍在 `docs/openapi/relay.json:835`、`docs/openapi/relay.json:1166` 等位置。

改动意图是消除两套轮询/结算实现并存的误用风险。需要留意的是：以后若重新使用 Go 注释执行 `swag init`，应确认视频文档仍有稳定生成来源。

### 4.2 订阅重置获得真正的数据库行锁

`model/subscription.go:1030` 和 `model/subscription.go:1052` 将 GORM v1 的：

```go
tx.Set("gorm:query_option", "FOR UPDATE")
```

替换为 `model/locking.go:20` 的 `lockForUpdate(tx)`。

旧写法在 GORM v2 中会被静默忽略。新写法在 MySQL/PostgreSQL 生成 `SELECT ... FOR UPDATE`，SQLite 则跳过不支持的语法。其真实意图是让管理员重置、并发扣费、定时重置和退款/差额结算对同一订阅行串行化，避免清零与扣费互相覆盖。

### 4.3 全局默认分页从 10 改为 20

- `common/constants.go:83`：`ItemsPerPage = 10` → `20`。
- `controller/channel.go:1491`：多密钥状态默认页大小从固定 50 改为统一常量 20。
- classic 多个列表同步改为 20。

这不是纯样式修改。所有通过 `common.GetPageQuery()` 且客户端未显式传页大小的接口，默认返回数量会从 10 变为 20；多密钥状态接口则从 50 降到 20。

意图是统一新版数据表分页契约。影响是普通列表默认查询量和响应体增大，依赖旧默认 10 条的第三方客户端会看到数组长度与总页数变化。

### 4.4 清除会进入真实业务数据的示例特殊分组

`setting/ratio_setting/group_ratio.go:28` 将内置 `vip_special_group_1` 等演示映射清空。原数据会被可用分组链路当成真实配置，进而出现在定价和分组选择界面。

该修复只删除内置样例，不影响管理员已经保存的 `group_special_usable_group`。

### 4.5 Codex CLI / Responses 协议同步

普通 Responses 请求新增：

- `dto/openai_request.go:879`：`client_metadata`。
- `dto/openai_request.go:963`：`reasoning.mode`、`reasoning.context`。

Compaction DTO 在 `dto/openai_responses_compaction_request.go:12` 新增：

- `tools`
- `parallel_tool_calls`
- `reasoning`
- `service_tier`
- `prompt_cache_key`
- `text`

渠道亲和模板在 `setting/operation_setting/channel_affinity_setting.go:39` 增加 session/thread/client/subagent/turn-state 等 Codex 请求头；`relay/helper/stream_scanner.go:49` 将 `X-Reasoning-Included` 和 `X-Codex-Turn-State` 从上游 SSE 响应带回客户端，使 turn-state 能在下一轮请求中往返。

`relay/channel/codex/constants.go:8` 新增 `gpt-5.6-sol`、`gpt-5.6-terra`、`gpt-5.6-luna` 及自动生成的 compact 变体。

需要重点复核：`relay/responses_handler.go:42` 把 Compaction DTO 转成普通 Responses DTO 时，目前只复制了 `parallel_tool_calls` 和 `service_tier`；`tools`、`reasoning`、`prompt_cache_key`、`text` 在非原始 body 透传路径中仍可能丢失。提交意图明确，但字段同步尚不能视为完全闭环。

### 4.6 依赖维护

`go.mod:52` 将 `golang.org/x/crypto` 从 `0.51.0` 升级到 `0.52.0`，业务源码无配套改动，属于常规安全与兼容维护。

## 5. 前端改动与意图

### 5.1 建立 shadcn 原始层与产品设计系统层

`web/default/src/components/design-system/README.md:1` 明确：

- `components/ui/` 保持为 shadcn CLI 可维护的原始源码层。
- `components/design-system/` 承担产品级响应式尺寸、variant 和复合控件策略。
- `web/default/.oxlintrc.json:47` 禁止业务代码绕过适配层直接导入 15 类受管控件。

`button.tsx:25`、`input.tsx:24`、`select.tsx:35`、`table.tsx:33` 等适配器将默认控件统一为手机 28px、桌面 32px，并提供 40→44px 的 `xl` CTA。

494 个 default 文件并不代表 494 项功能。大批文件只是：

- `components/ui/*` → `components/design-system/*`。
- 删除调用点的 `h-8`、`size-8`、`size="sm"` 等硬编码。
- 图标按钮改用 `icon/icon-sm/icon-xs`。
- 对 Base UI `render` 组合、类型导入和排版做机械统一。

改动意图是把响应式策略集中到一个可维护边界，而不是让每个业务页面自行决定尺寸。

### 5.2 视觉语言从装饰性收敛为数据工具风格

- `status-badge.tsx:29` 将徽章统一为 `neutral/info/success/warning/destructive` 五种语义，移除字符串哈希彩虹色。
- `theme.css:28` 修正 Public Sans 名称并补齐中文、日文区域字体回退。
- `fonts.css:38` 使用本地字体与 `font-display: optional`，降低首次加载字体跳动。
- `i18n/config.ts:66` 随语言切换同步 `<html lang>`，改善中日韩字形选择。
- 后台卡片悬浮、按钮按压缩放、表格 stagger 等入场动效被移除；`lib/motion.ts:39` 将页面切换限制为透明度淡入。
- 统一圆角与“描边或阴影二选一”，减少视觉噪声。

### 5.3 Data Table 统一桌面表格、移动卡片与过滤器

核心位置：

- `data-table/layout/data-table-page.tsx:227`：表格/卡片视图、持久化、移动分支。
- `tanstack-table.d.ts:22`：`cardRole/cardOrder/cardSpan/contentMode` 列元数据。
- `card-row-content.tsx:51`：根据同一列定义生成移动卡片。
- `toolbar/filter-panel.tsx:91`：桌面展开过滤器与移动 Drawer。
- `toolbar/toolbar.tsx:151`：中文输入法 composition 保护、debounce、过滤计数。
- `core/data-table-view.tsx:55`：固定高度与 sticky header。
- `core/table-sizing.ts:36`：调整列宽后表格仍填满容器。

渠道页在 `features/channels/components/channels-table.tsx:417` 正式启用桌面表格/卡片切换并持久化。用户、模型、订阅、兑换码、密钥和日志等页面主要复用同一列元数据生成移动卡片。

最终版取消移动卡片的 “More” 折叠，非隐藏字段直接可见。意图是减少移动端信息层级，但也增加单卡高度。

### 5.4 定价页与上游价格同步

公开定价页：

- `features/pricing/index.tsx:41` 从渐变 Hero、常驻侧栏和页内详情 Drawer 改为统一目录页壳。
- 默认视图由卡片改为表格；表格每页 20、卡片每页 12。
- `pricing-toolbar.tsx:170` 集中搜索、排序、价格口径、单位和视图切换。
- 模型详情改为专用路由，并保留当前筛选条件。
- 表格与卡片都加载 24 小时性能摘要。

管理员价格同步：

- `upstream-ratio-sync-table.tsx:135` 预计算可选项、已选数量与移除计划。
- `upstream-ratio-sync-helpers.ts:303` 统一处理 price、ratio、tiered 与 `billing_mode/billing_expr` 配对。
- `upstream-ratio-sync.tsx:274` 把逐项多次 `setState` 改为单次批处理，减少大模型列表卡顿。
- 固定表头、内部滚动、固定分页和跨上游字段对齐改善了大量数据下的可操作性。

这部分同时改变业务选择与保存逻辑，是前端风险最高的区域，不能只按样式重构验收。

### 5.5 Playground 参数面板与模型选择器

`features/playground/lib/parameters/playground-parameters.ts:35` 集中定义：

- temperature
- top_p
- frequency_penalty
- presence_penalty
- max_tokens
- seed

`playground-parameter-panel.tsx:104` 提供开关、滑杆和数字输入；桌面使用 Popover，移动端使用 Sheet。`payload-builder.ts:47` 只把启用的参数加入请求，保留“未出现”和“显式 0/负 penalty”的语义差异。

模型选择器在打开后自动滚动到当前分组和模型，双栏固定高度并独立滚动。`components/multi-select.tsx:145` 将自定义模型判重改为大小写敏感，允许 `Model-A` 与 `model-a` 同时存在。

### 5.6 Profile、Wallet、Dashboard、Usage Logs

Profile：

- 原账户绑定/通知设置页签展平成直接可见卡片。
- 密码、访问令牌、2FA、Passkey 收敛到统一安全卡片。
- `features/profile/index.tsx:47` 改为主内容加侧栏布局。

Wallet：

- `features/wallet/index.tsx:254` 改为统计 → 充值 → 订阅 → 推广的线性顺序。
- 支付方式和订阅计划改成更紧凑、可键盘操作的按钮/卡片。
- `affiliate-rewards-card.tsx:70` 将推广文案从“下级充值才奖励”修正为“用户通过推荐链接加入后奖励”。

Dashboard：

- `overview-dashboard.tsx:35` 删除约 670 行首次使用引导、快捷操作、真实密钥 curl 复制及相关 API key/model 查询。
- 现在只保留统计、性能、API 信息、公告、FAQ 和 Uptime。
- 这降低了请求量和敏感密钥读取路径，但属于真实用户功能移除，需要产品确认，不应归入机械样式迁移。

Usage Logs：

- 默认分页统一为 20，并按日志类别和角色持久化。
- 移动日志卡片改用通用列元数据。
- 日期保留在移动过滤栏，其余过滤进入 Drawer。
- `hooks/use-group-ratios.ts:23` 新增当前分组倍率查询，在历史日志未携带倍率时做补充展示。
- 管理员可看到 quota saturation 标记与危险行底色。

### 5.7 自定义首页、i18n、Classic 与依赖

- `features/home/index.tsx:76` 增加 `allow-top-navigation-by-user-activation`，允许可信管理员配置的 iframe 页面在用户点击后跳转顶层窗口，不授予 same-origin 权限。
- 英文基准新增 26 个 i18n key、删除 30 个；删除项主要属于被移除的 Dashboard onboarding。七个 locale 当前均为 5,100 个 key。
- `i18n/static-keys.ts:233` 登记 Playground 参数静态 key。
- classic 主要把列表默认分页从 10 改为 20，并同步 Codex 请求头模板。
- `package.json:30` 移除 `@tanstack/react-virtual`、`vaul`，新增 `date-fns`，`recharts` 从 3.9.1 调整为 3.8.0。

## 6. 跨前后端改动意图

1. **统一数据密度与响应式行为**：后端、default、classic 的默认分页都收敛为 20；default 的控件和数据表由统一适配层负责手机/桌面差异。
2. **把后台从展示型 UI 改为操作型 UI**：降低颜色、阴影、圆角和动画干扰，强调数据对齐、过滤和批量操作。
3. **减少重复实现**：后端删除旧任务/计费文件，前端将表格、卡片、过滤器和受管控件集中到通用层。
4. **提高大数据量操作效率**：价格同步批量状态更新、固定表头和内部滚动；列宽调整后继续利用容器空间。
5. **跟进 Codex CLI 协议演进**：请求 DTO、亲和模板、default/classic 管理配置和 SSE 响应头同时更新，避免只改某一层。
6. **收紧默认数据真实性**：移除特殊分组演示数据和不再使用的 Dashboard onboarding，减少展示与真实配置不一致。

## 7. 风险与验证状态

### 高优先级风险

1. Codex compact 新增字段在非 passthrough 路径可能没有全部转发。
2. 上游价格同步重写了选择、互斥和保存逻辑，但没有发现对应新 helper 的专项测试。
3. Dashboard onboarding 是明确功能删除，需要产品确认。
4. 全局默认分页是 API 默认行为变化，可能影响未显式传页大小的第三方客户端。
5. Data Table 影响面极广，需要桌面与 375px 移动端浏览器回归。

### 建议浏览器回归

- 渠道、用户、模型、密钥、兑换码、订阅、使用日志。
- 表格/卡片切换与持久化。
- Drawer 过滤、重置、中文输入法、sticky header、横向滚动、列显隐和行操作。
- Playground 参数显式 0、负 penalty、空 seed、启用/禁用后的实际请求体。
- 定价页筛选到详情路由的状态保留。
- 自定义首页 iframe 顶层导航。
- Profile、Wallet 与被移除 Dashboard onboarding 的产品验收。

### 本地验证

- `git diff --check adf681cb..9f8152f0`：通过。
- `go test ./model`：通过。
- `go test ./relay/channel/codex`：该包无测试文件。
- `go test ./service -run '^TestObserveChannelAffinityUsageCacheByRelayFormat_MixedMode$' -count=1`：单独重跑通过。
- 相关包组合测试中，`relay/helper` 的 uint 溢出错误文案断言失败；对应测试与实现文件不在本次范围内，记录为当前分支既有基线失败，不归因于本轮同步。
- 前端 `bun run typecheck` 未执行成功：当前 `web/default` 没有 `node_modules`，脚本报 `bun: command not found: tsgo`。这是依赖未安装，不是类型检查结果。
- 未执行完整前端 build、lint 或浏览器回归。

## 8. 当前尚未同步的内容

本地最近 fetch 的 `upstream/main` 快照：

- commit：`ad900bbba74b3e9b16b1ef9c549812ada2bb14a0`。
- fetch 时间：2026-07-11 14:41:49 +08:00。
- `dad57a6b..ad900bbb`：3 个 first-parent、6 个全部祖先提交。
- 94 个文件，`+2098/-816`。

已知主题：

- Fontsource 在不同 workspace 布局下的资源路径修复。
- themed data views 与任务日志详情增强。
- 预扣费饱和拒绝与错误处理。
- 图片流客户端断连处理与计费调整。

这些内容尚未进入保存 HEAD `9f8152f0`，必须留给下次同步报告，不能混入本次“已同步”结论。

本次尝试实时读取 GitHub 远端时发生 443 连接超时，因此 `ad900bbb` 仅代表本地 remote-tracking 快照，不宣称是当前 GitHub 最新 HEAD。`origin/main` 是更靠前的 fork 主线，也不作为本次同步源。

## 9. 下次同步整理流程

保留本文件，不覆盖当前锚点。下次同步后新建下一份日期报告，并从以下值开始：

```powershell
$OldHead = '9f8152f0f17ccccb0dd68c6f9c1fe0d88f6c60eb'
$OldUpstream = 'dad57a6bb85becbb99cab26ade7a891508ed7c42'

git merge-base --is-ancestor $OldHead HEAD
git merge-base --is-ancestor $OldUpstream upstream/main

# 找出保存 HEAD 之后的同步 merge
git log --first-parent --merges --reverse `
  --format='%H|%P|%cI|%s' "$OldHead..HEAD"

# 对每个新的同步 merge
$Merge = '<new-sync-merge>'
$PreMergeHead = git rev-parse "$Merge^1"
$NewUpstream = git rev-parse "$Merge^2"

# merge 对本地分支产生的实际效果
git diff --stat $PreMergeHead $Merge
git diff --name-status $PreMergeHead $Merge

# 纯上游新增内容
git log --first-parent --reverse `
  --format='%H|%cI|%an|%s' "$OldUpstream..$NewUpstream"
git diff --stat "$OldUpstream..$NewUpstream"
```

重新整理完成后，新报告更新：

```yaml
saved_head: <new local sync head>
saved_synced_upstream: <new merge second parent>
```

如果同步期间还夹杂普通本地提交，不得把 `$OldHead..HEAD` 全部描述成上游同步；必须继续按每个 merge 的 `M^1..M` 隔离本地实际效果，并以 `$OldUpstream..M^2` 隔离纯上游内容。

## 10. 重复流程打包结论

| 重复流程 | 证据/日期 | 频率与信心 | 形式 | 结论 |
|---|---|---|---|---|
| 上游同步后按前端/后端审计、保存双锚点、下次续算 | 2026-07-11 两次同步；用户明确要求下次再次整理 | 明确会复发，高信心 | Skill | 已创建 `.agents/skills/audit-upstream-sync/` |
| 为每次同步保留常驻专用子代理 | 本次可临时并行分派前端/后端审计 | 中等频率 | Skip | 领域边界随改动变化，固定角色会过窄；技能内按需分派即可 |
| 定时生成同步报告 | 同步由人工事件触发，暂无固定周期证据 | 证据不足 | Skip | 定时任务容易产生无变化噪声；等形成稳定同步节奏再考虑 |
| 扩展 `classic-to-default-sync` | 现有技能只处理 classic→default 的单提交功能对齐 | 高信心不适用 | Skip | 当前流程是上游分支同步审计，输入、范围和输出契约不同，避免重叠 |

创建的技能只规定 Git 锚点、报告边界和输出契约，不复制本报告中的一次性业务结论。

## 附录：完整上游祖先提交

### 批次 1

| 提交 | 时间 | 作者 | 主题 |
|---|---|---|---|
| `81808d24` | 2026-07-04 23:42 | feitianbubu | remove sample special usable groups leaking into pricing page |
| `df01273b` | 2026-07-09 16:26 | zuiho | let resized tables fill available width |
| `a79f9691` | 2026-07-09 16:49 | CaIon | update referral message |
| `4645ad9d` | 2026-07-09 21:47 | QuentinHsu | keep model selector lists in sync |
| `246d62aa` | 2026-07-09 22:03 | feitianbubu | remove dead files resurrected by v1.0 launch commit |
| `928b4750` | 2026-07-09 22:10 | QuentinHsu | add chat parameter settings panel |
| `4e570389` | 2026-07-10 10:21 | Seefs | use GORM v2 row locking for subscription resets |
| `e8596cab` | 2026-07-10 10:54 | feitianbubu | allow adding custom model names that differ only by case |
| `4823417c` | 2026-07-10 23:28 | 同語 | merge Playground parameter settings panel |
| `d3b01b48` | 2026-07-10 23:30 | 同語 | merge case-sensitive custom model names |
| `f2c7cd33` | 2026-07-10 23:33 | 同語 | merge special usable group sample removal |
| `489c0458` | 2026-07-10 23:37 | QuentinHsu | optimize upstream price sync table |
| `43783286` | 2026-07-10 23:56 | QuentinHsu | polish sync channel dialog layout |
| `6869cd94` | 2026-07-11 00:28 | QuentinHsu | align table badge spacing |
| `0cb741d8` | 2026-07-11 00:49 | 同語 | merge upstream price sync optimization |

### 批次 2

| 提交 | 时间 | 作者 | 主题 |
|---|---|---|---|
| `262ab931` | 2026-07-11 02:02 | t0ng7u | unify design system across default frontend |
| `0918bdb4` | 2026-07-11 05:13 | t0ng7u | consolidate design-system primitives and responsive data views |
| `9d1ca545` | 2026-07-11 06:02 | t0ng7u | refine data-table cards and pricing page layout |
| `ca971413` | 2026-07-11 10:52 | 乾L | allow user-activated top navigation for custom home iframe |
| `00f1cbb6` | 2026-07-11 10:53 | dependabot[bot] | bump golang.org/x/crypto to 0.52.0 |
| `dad57a6b` | 2026-07-11 11:04 | Seefs | sync Codex fields |
