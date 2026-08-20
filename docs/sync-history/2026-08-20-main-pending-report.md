---
report_type: upstream-pending-main-audit
generated_at: 2026-08-20T12:11:14+08:00
timezone: Asia/Shanghai
repository: D:\\code\\goswtich\\new-api
branch: feature/free/local
tracking_branch: origin/feature/free/local
baseline_head: fe176dd5a
baseline_synced_upstream: ccd535ef8e50cf6e5846a59278c40b7ff59d1b7d
saved_head: 9fdbf3add7992c4dd7c89845e37adfe3b5a3e2cf
saved_synced_upstream: f116414284162ad15d8925f7bca494c109b83e93
upstream_tracking_snapshot: f116414284162ad15d8925f7bca494c109b83e93
upstream_tracking_snapshot_fetched_at: 2026-08-20T12:11:14+08:00
scope: 待合并上游范围 ccd535ef8..f11641428；不包含本地未提交的 Apifox 文件
decision: 已按固定计量/旧异步任务账务分流规则合并；代码可继续做部署前灰度检查
---

# 2026-08-20 main 待合并更新报告

## 1. 总体结论

- 范围：`ccd535ef8..f11641428`；21 个 first-parent 提交、122 个文件、`+6960/-4634`。
- 这批后端更新值得吸收，主要收益是充值额度上限保护、异步任务退款账务、渠道协议兼容和缓存 Token 计费。
- 对 `feature/free/local` 而言，**不能直接自动合并**：本地固定计量/ZZDH 任务账务与上游异步任务账务都修改了 `service/task_billing.go`。
- 已创建合并提交 `9fdbf3add`；固定计量失败不回减未产生的用量，旧异步退款/差额保留上游用量回减。
- 未重启容器、未调整端口或 Compose、未操作数据库、未修改 Switcher 代码。

## 2. 上游功能与改动

| 范围 | 新行为 | 与当前分支的区别 | 主要源码路径 | 影响级别 |
|---|---|---|---|---|
| 异步任务退款 | 失败退款和最终差额结算中的退款都会同步减少 `users.used_quota` 与渠道 `used_quota`；Midjourney 保存实际计费 Token/渠道并复用统一退款服务。 | 本地已具备固定计量任务的冻结日志和终态扣费，但退款路径尚未包含这项上游用量计数修正。 | `service/task_billing.go`、`service/quota.go`、`service/midjourney.go`、`relay/mjproxy_handler.go` | 高 |
| 充值安全 | 拒绝会使 int32 钱包额度溢出的支付请求；结算阶段用单条带条件 SQL 再次执行上限保护，避免并发回调通过过期余额判断。 | 本地官方充值路径没有这层新增的额度上限保护。 | `controller/topup*.go`、`model/topup.go` | 高 |
| Responses 缓存 Token | 将 Responses 的 `input_tokens_details` 与 `prompt_cache_hit_tokens` 写入计费使用的标准缓存 Token 字段。 | 否则缓存 Token 定价可能基于不完整的用量数据。 | `service/billing_usage.go`、`service/text_quota_test.go` | 高 |
| Claude 转换 | 不再发送空 `tools: []`；无参数函数工具会规范成 `type: object, properties: {}`。 | 修复 OpenAI/Responses 转换后被部分 Claude 上游拒绝空工具或无参数工具的问题。 | `relaykit/relayconvert/internal/.../claude/` | 中 |
| OpenAI Chat 转 Responses | 转换时保留 `prompt_cache_key`。 | 使用该字段的 Chat 客户端能在 Responses 上游保持缓存亲和性。 | `relaykit/relayconvert/internal/oai_chat/to_oai_responses_req.go` | 中 |
| 阿里图片模型映射 | 协议、异步模式和请求头按映射后的上游模型判断，而非公开别名。 | 映射到 Qwen/Wan 图片模型的本地别名当前可能选错阿里请求端点。 | `relay/channel/ali/adaptor.go` | 中 |
| Advanced Custom 余额 | 渠道可显式映射 `/v1/dashboard/billing/credit_grants`；数值 `credit_summary.total_available` 更新余额，其他合法 JSON 以格式化 `raw_response` 返回。 | 现有自定义路由只覆盖模型发现与转发路径。 | `relaykit/dto/channel_settings.go`、`relay/channel/advancedcustom/adaptor.go`、`controller/channel-billing.go` | 中 |
| 自动渠道测试 | 增加 1-32 的受限并发，以及 `scheduled_all`、`auto_ban_only`、`passive_recovery` 三种选择模式；手动全量测试改经系统任务租约队列执行。 | 本地仅有旧版自动测试行为，尚无并发设置和被动恢复模式。 | `controller/channel-test.go`、`setting/operation_setting/monitor_setting.go` | 中 |
| 内置 Web | 重构 Advanced Custom 路由编辑，展示网关字段透传控制，新增流式逐词淡入/Playground 编辑优化，并迁移到 Vitest/jsdom 测试。 | 部署使用 Switcher 前端，这些页面和测试迁移不会直接面向用户。 | `web/`、`electron/` | 运行时低 |
| 依赖与 CI | 更新 `dompurify`、Electron 相关包和 CI Node 配置。 | 不改变网关请求合同。 | `web/package.json`、`electron/`、`.github/workflows/ci.yml` | 低 |

## 3. Switcher 兼容性

| 项目 | 当前 Switcher 合同 | 上游影响 | 需要动作 | 优先级 |
|---|---|---|---|---|
| Advanced Custom 路由模型 | Switcher 已在 `frontend/src/features/channels` 定义、校验和序列化 `advanced_custom`、模型发现、上游模型更新及 `pass_through_body_enabled`。 | 线协议不变；新增字段透传 UI 仅是上游内置 Web 的展示优化。 | 合并前无需修改。后续可选择性对齐编辑器体验。 | P3 |
| Advanced Custom 余额 | Switcher 的 `handleUpdateChannelBalance` 只有响应包含 `balance` 才认为成功。 | 合法的自定义余额端点现在可能返回 `success: true, raw_response`，而不是数值余额；目前会被误显示成通用失败。 | 启用此类路由前，在 Switcher 渠道余额处理里增加原始 JSON 查看/脱敏展示路径。 | P2 |
| Channel Health 探测 | Switcher 经 `OfficialChannelBackgroundApiClient` 调用 `GET /api/channel/test/{id}`，并写入自己的 Channel Health 流水线。 | 单渠道探测响应合同兼容；上游调度器可独立测试、恢复或自动禁用渠道。 | 保持上游 `AutoTestChannelEnabled=false`，或明确指定唯一责任方。不要与 Switcher Channel Health 策略并行启用。 | P1 运行策略 |
| 钱包/充值体验 | Switcher 钱包对旧官方充值调用 official amount/pay 接口，并在 `/api/user/topup/plans/*` 拥有套餐/Xunhu 流程。 | 官方路径可在支付前拒绝超上限钱包请求；Switcher 自有套餐/Xunhu 结算不在这些上游路径内。 | 正常错误处理无需改动。若产品要展示专门提示，可为上游 `message:error` 增加易读映射。未经统一余额事务审查，不要把官方保护直接复制到 Switcher 自有结算。 | P2 可选 |
| 异步任务用量面板 | Switcher 从官方用户/渠道数据读取 `used_quota`。 | 失败或过度预扣异步任务之后的未来数值会更准确；无 Schema/API 变更。 | 无需代码修改。合并后预期总数更准确；历史膨胀计数不能在没有单独审计数据方案时回填。 | P3 |
| API 转发行为 | Switcher 代理普通 `/v1/*` 请求，并有独立的视频覆盖路径。 | Claude、Responses、阿里图片和缓存 Token 修复均在服务器端，对经过 new-api 的请求生效。 | 无需代码修改。但要单测 Switcher 绕过/视频覆盖路径，因为它可能不经过正常 new-api relay 中间件。 | P2 验证 |
| Session/Cookie/OAuth | Switcher 自有登录体验；本范围没有新的服务端 Session/Cookie 改动，内置 Web 仅调整 Custom OAuth 绑定类型和测试。 | 无部署认证迁移。 | 无需修改。 | P3 |

## 4. 合并冲突与语义风险

| 文件或子系统 | 本地定制 | 上游改动 | 处理规则 | 风险 |
|---|---|---|---|---|
| `service/task_billing.go` | 固定计量/ZZDH 冻结快照、终态消费日志、请求/日志上下文和任务计费兼容。 | 退款/差额结算必须回减 `used_quota` 与渠道已用额度，且不得改变请求计数。 | 两边逻辑都保留。把上游计数回减应用到所有本地任务退款/负差额路径，再验证固定计量成功只记一次、失败任务的钱包与用量计数各退一次。 | 高 |
| `service/task_billing_test.go` | 本地覆盖固定计量快照、终态日志和异步任务生命周期。 | 上游覆盖退款、差额结算和 Midjourney 账务。 | 两套测试均保留，并新增固定计量失败/过度预扣的组合回归用例。 | 高 |
| `model/option.go` | 校验本地计费模式和固定计量配置 JSON。 | 新增上游渠道测试并发配置校验。 | 两个校验分支均保留。 | 中 |
| Channel Health 责任归属 | Switcher 当前负责策略驱动的探测、隔离、恢复和亲和性动作。 | 上游新增独立状态变更的自动测试器。 | 默认不要同时运行两个自动引擎；启用上游调度器前必须确定唯一责任方。 | 高，语义 |
| Advanced Custom 编辑器 | Switcher 有独立编辑器和表单序列化。 | 上游重构内置编辑器并新增管理余额路由。 | 不整体复制内置 UI；仅在启用功能时加入余额路由选项和原始响应体验。 | 中，语义 |
| `relaykit/` | 本地分支要求其独立构建。 | 转换语义改动位于 `relaykit`，不能依赖根模块修复。 | 原样合入上游 relaykit 代码后，在 `relaykit` 中执行 `GOWORK=off go build ./...`。 | 中 |

Git 祖先比较确认，只有前三行属于双方同文件改动。该上游范围没有 Dockerfile、`docker-compose*.yml`、固定端口映射、数据库迁移或 Switcher 源码变更。

## 5. 部署与数据合同

- 数据库/Schema 迁移：本上游范围无迁移。Midjourney 新增两个 GORM 字段（`token_id`、`billing_channel_id`），在依赖新退款归因前须确认运行中的 new-api 开启 AutoMigrate；不要为 new-api 自有表创建 Switcher 迁移。
- Docker/Compose/端口：上游无改动。现有后端专用 `Dockerfile.host`、`docker-compose-db.yml`、`docker-compose.dev.local.yml`、卷和端口分配继续是部署合同。
- Session/Cookie/API Key：无服务端 Session 或 Cookie 合同变化；API Key 与 `/v1/*` 认证不受影响。
- 启用前配置：Switcher Channel Health 持有自动化期间，保持 `monitor_setting.auto_test_channel_enabled=false`。若以后启用上游测试，应明确设置 `channel_test_mode`，并从 `channel_test_concurrency=1` 开始；`passive_recovery` 最保守，因为它不会自动禁用渠道。
- 数据预期：新退款逻辑会修正未来的 `used_quota`/渠道用量；历史上已膨胀的计数不会自动修复。

## 6. 验证计划

1. 在受保护工作树中将 `origin/main` 合入 `feature/free/local`；按上述规则手工处理三个双方同文件改动。
2. 为固定计量/ZZDH 的失败退款和负差额结算补充聚焦测试，断言钱包额度、Token 额度、`users.used_quota`、渠道已用额度、请求计数和各一条审计日志。
3. 执行根模块 `service`、`model`、`controller`、`relay/channel/ali` 聚焦测试；再在 `relaykit` 中执行 `GOWORK=off go build ./...`。
4. 对 `/api/channel/test/{id}` 执行 Switcher Channel Health 单探测回归；在责任归属确定前不要测试上游自动调度。
5. 若配置 Advanced Custom 余额路由，启用前先更新 Switcher 原始响应展示，并测试数值与非数值 JSON 的成功响应。
6. 正式部署前验证生成的后端专用镜像、Midjourney 字段的实时 AutoMigrate 状态、Redis 可用性和未改变的 Compose 端口映射。

## 7. 提交附录

| 提交 | 主题 | 分类 |
|---|---|---|
| `58d4e9bd3` | 异步任务退款回减已用额度 | 后端计费正确性 |
| `15cfdedde` | 抓取的模型选择与表单保持同步 | 内置 Web 体验 |
| `93d2df85f` | 阿里图片映射使用上游模型协议 | 转发兼容 |
| `626058075` 至 `bbf67df04` | Electron/Web 依赖升级 | 依赖维护 |
| `2a0ce3475`、`47ba9d2c6` | 拒绝/保护不可入账充值 | 钱包安全 |
| `7d09c6954` | 向 Responses 保留 `prompt_cache_key` | 转发兼容 |
| `e90a7c48e` | 网关字段透传控制 | 内置 Web 体验 |
| `4442bb302`、`3dda1d50c` | Claude 空/无参数工具处理 | 转发兼容 |
| `116255f07` | Custom OAuth 绑定字段对齐 | 内置 Web 合同清理 |
| `e2c7aa7b1` | Web 测试统一到 Vitest | 测试基础设施 |
| `2b0efd848` | Advanced Custom 路由编辑器与余额路由支持 | 后端加内置 Web |
| `4add708eb` | 渠道测试模式与受限并发 | 运维 |
| `137d1171f` | 流式淡入与 Playground 编辑加固 | 内置 Web 体验 |
| `f11641428` | 结算 Responses 缓存 Token 用量 | 计费正确性 |
