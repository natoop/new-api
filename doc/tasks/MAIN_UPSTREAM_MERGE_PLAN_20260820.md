# main 上游更新合并计划（功能完整性优先）

## 记录信息

| 项目 | 结论 |
| --- | --- |
| 计划状态 | 已按本计划完成代码合并；未重启容器、未修改部署配置 |
| 本地目标分支 | `feature/free/local`，审计时 HEAD `fe176dd5a` |
| 已吸收上游锚点 | `ccd535ef8` |
| 待合入上游 | `origin/main` = `f11641428`（同时等于 `upstream/main`） |
| 合并范围 | `ccd535ef8..f11641428`，21 个 first-parent 提交，122 个文件 |
| 实际合并提交 | `9fdbf3add7992c4dd7c89845e37adfe3b5a3e2cf`，父提交为本地 `fe176dd5a` 与上游 `f11641428` |
| 基本原则 | 新旧已验证功能都保留；不能证明兼容时先保留旧功能并显式关闭对应新功能；不以“文本无冲突”替代账务和运行语义验证。 |
| 用户自有未跟踪文件 | 保留 `ZZDH_VIDEO_API_APIFOX.openapi.yaml`、`ZZDH_VIDEO_API_APIFOX.postman_collection.json`，不得纳入此次合并或清理。 |

## 1. 合并目标、边界与禁止项

### 1.1 目标

将本范围内的后端安全性、计费正确性、协议兼容和渠道运维能力吸收到本地分支，同时完整保留本地固定计量（`fixed_metered`）、ZZDH 异步视频、冻结计费快照和 Switcher 前端接管模式。

必须达到以下结果：

1. 普通旧异步任务的失败退款、成功后的负差额结算，能回退先前已记入的用户/渠道用量，但不回退请求次数。
2. 本地固定计量任务仍坚持“提交前只预扣，成功终态才写正式消费与用量”的模型；失败时不能回减一笔从未增加的 `used_quota`。
3. 上游的充值容量保护、Responses 缓存 Token 结算、Claude 工具转换、阿里图片模型映射全部进入正式回归范围。
4. Switcher 继续是管理前端和渠道健康策略的默认责任方；new-api 内置 Web 的体验性改动不强制移植到 Switcher。

### 1.2 明确不在本次范围内

- 不修改 `docker-compose-db.yml`、`docker-compose.dev.local.yml`、`Dockerfile.host`、端口映射、卷、数据库连接地址或 Redis 地址。
- 不修改 Session、Cookie、JWT、API Key、OAuth 的服务器认证合同。上游涉及的仅是内置 Web Custom OAuth 绑定字段与测试，并非服务端会话迁移。
- 不改 Switcher 业务代码；本计划只列出需要先适配才可启用的项目。
- 不回填历史 `users.used_quota` 或渠道 `used_quota`。上游修复只保证合并后的新任务账务正确；历史修复需独立审计、备份和可回滚的数据方案。
- 不将用户自有 Apifox 文件、临时构建产物或无关工作区变化带入提交。

## 2. 必须保留的既有功能清单

| 既有能力 | 必须保留的合同 | 合并后的实现约束 | 验收证据 |
| --- | --- | --- | --- |
| 固定计量任务 | 提交前冻结 `fixed_metered` 快照和预扣额度；价格配置后续变化不得改变在途任务账单。 | `BillingContext.FixedMeteredBilling` 仍是任务计费生命周期的判据，日志继续写入冻结快照。 | 成功、失败、重试后均可从任务 `private_data` 和正式日志还原同一价格事实。 |
| 固定计量成功结算 | 只有成功终态 CAS 胜出后，才产生一次正式消费日志、用户 `used_quota`、渠道 `used_quota` 与请求数。 | 不把上游“提交期已记用量”的假设套入该分支。 | 成功轮询重复执行时，消费日志与计数均只增加一次。 |
| 固定计量失败退款 | 退还预扣的钱包/订阅/Token 额度；失败前没有正式消费时，不产生负的用户或渠道用量。 | 退款路径按计费模式分支，不能无条件调用用量回减。 | 钱包/订阅/Token 恢复，`used_quota` 和请求数不变，退款日志恰一条。 |
| ZZDH 视频协议与定价 | 保留已验证的 V8 请求、轮询、时长边界、参考视频计量、冻结快照和终态审计。 | 不用上游旧任务的计费路径覆盖 ZZDH 的固定计量路径。 | 四个指定模型和参考视频计量回归通过。 |
| 旧异步任务 | 保留原有任务创建、预扣、适配器调价、CAS 终态、失败退款和订阅/Token 支持。 | 补足上游的用量回减，但保留请求次数语义。 | 失败与负差额结算后余额、Token、订阅、用户/渠道用量均正确。 |
| Switcher 前端接管 | 使用 Switcher 编辑渠道、展示运营数据、控制 Channel Health；不依赖 new-api 内置 Web。 | 上游 `web/` 可随代码合入，但不作为当前部署用户界面。 | backend-only 镜像和既有 Switcher 路由可正常工作。 |
| 现有部署合同 | 既有 Compose、固定端口、后端专用 Dockerfile、PostgreSQL/Redis 卷不变。 | 合并前后应无 Compose/Dockerfile 差异。 | `git diff` 与容器配置核对均为空。 |

## 3. 需要完整接入的上游能力

| 优先级 | 功能 | 上游提交与主要路径 | 合并规则 | 对 Switcher 的影响 |
| --- | --- | --- | --- | --- |
| P0 | 充值 int32 容量保护 | `2a0ce3475`、`47ba9d2c6`；`controller/topup*.go`、`model/topup.go` | 原样接入预支付校验和回调事务内的条件更新。所有官方支付通道必须走同一 `creditTopUpQuota` 容量约束。 | 无 API 结构变更；前端应按现有 `message:error` 显示失败。专门的“余额已达上限”文案为可选体验改进。 |
| P0 | Responses 缓存 Token 计费 | `f11641428`；`service/billing_usage.go` | 完整保留 `input_tokens_details` 和 `prompt_cache_hit_tokens` 向标准缓存字段归一化的优先级。 | 无需改动；经 new-api 的请求直接受益。 |
| P0 | 旧异步任务退款用量修正 | `58d4e9bd3`；`service/task_billing.go`、`service/quota.go`、Midjourney 路径 | 合入，但必须依第 4 节做计费模式分流。 | Switcher 读取的未来用户/渠道用量会更准确，无 Schema/API 变更。 |
| P1 | Midjourney 账务归因与退款 | `58d4e9bd3`；`model/midjourney.go`、`service/midjourney.go`、`relay/mjproxy_handler.go` | 完整保留任务持久化后的资金/Token 标记、计费渠道归因、失败退款和无请求数回减。 | 无需改动。部署前确认 AutoMigrate 已实际创建两个字段。 |
| P1 | Claude 工具转换兼容 | `4442bb302`、`3dda1d50c`；`relaykit/relayconvert/...` | 空工具数组必须省略；无参数 `function` 工具必须变为合法空对象 schema，不能被过滤。 | 无需改动；Switcher 代理经过标准 relay 时直接生效。 |
| P1 | Chat 转 Responses 缓存亲和 | `7d09c6954`；`relaykit/.../to_oai_responses_req.go` | 原样保留 `prompt_cache_key`。 | 无需改动。 |
| P1 | 阿里图片别名映射 | `93d2df85f`；`relay/channel/ali/adaptor.go` | 所有同步/异步协议判断和请求头均按 `UpstreamModelName`，不能继续按公开别名 `OriginModelName` 判断。 | 无需改动；需覆盖 Switcher 经过该 relay 的图片请求。 |
| P1 | Advanced Custom 管理余额 | `2b0efd848`；`controller/channel-billing.go`、`relay/channel/advancedcustom`、`relaykit/dto/channel_settings.go` | 后端路由和严格配置验证均合入；只允许显式配置 `/v1/dashboard/billing/credit_grants`，不得把普通转发路由误作管理路由。 | 见第 6 节：未适配原始 JSON 显示前，不启用非数值余额路由。 |
| P2 | 渠道测试并发与模式 | `4add708eb`；`controller/channel-test.go`、`setting/operation_setting/monitor_setting.go`、`model/option.go` | 保留 1-32 的校验、租约任务和三种模式；默认关闭自动任务、并发设为 1。 | 不能与 Switcher Channel Health 双控；见第 6 节。 |
| P3 | 内置 Web、Electron、Vitest、流式动画 | `15cfdedde`、`e90a7c48e`、`137d1171f`、`e2c7aa7b1` 及依赖提交 | 代码随上游合入以保证仓库一致性；不以此作为本次 Switcher 发布阻塞项。 | 当前部署不使用该 UI，属于不适用的内置 Web 体验改动。 |

## 4. 高风险账务合并规则

### 4.1 两种任务生命周期必须同时存在

| 状态/动作 | 固定计量任务：`FixedMeteredBilling != nil` | 旧异步任务：无固定计量快照 | 不变量 |
| --- | --- | --- | --- |
| 提交成功 | 预扣钱包/订阅/Token，保存冻结快照；**不**写正式消费、用户/渠道用量或请求数。 | 保留现有旧逻辑：提交阶段已消费/累计时，持久化可退款标记。 | 任务记录必须先包含可追溯的资金来源、Token 和渠道归因。 |
| 成功终态 CAS 胜出 | 写一次正式消费日志，并增加用户 `used_quota`、渠道 `used_quota`、请求数；不重复扣钱包。 | 依照原有结算模型计算实际额度；若差额非零，仅调整差额，且请求数不得因结算重复增加。 | CAS 输家不得做任何账务、日志或计数写入。 |
| 成功负差额 | 若固定计量价格为冻结值且无可调价差额，保持零调整；若该模式明确允许结算差额，按固定计量专用规则退款资金/Token，但只回退已经真实增加的计数。 | 合入上游修复：退款差额时回减用户/渠道 `used_quota`，不回减请求数。 | 不能出现负计数或重复退款。 |
| 失败终态 CAS 胜出 | 退还预扣的钱包/订阅/Token；写退款审计；**不回减**用户/渠道 `used_quota`，因为尚未记入。 | 退还资金/Token，回减提交阶段已增加的用户/渠道 `used_quota`，请求数不变。 | 资金、Token、用量、日志应各精确变化一次。 |
| 终态 CAS 输家/重试 | 不结算、不退款、不写日志。 | 不结算、不退款、不写日志。 | 不依赖进程内状态；以持久化状态和条件更新为准。 |

### 4.2 `service/task_billing.go` 精确合并方案

1. 保留本地 `taskBillingOther`、`attachFixedMeteredBillingOther`、`LogFixedMeteredTaskConsumption`、固定计量上下文和现有冻结审计字段。
2. 保留上游 `model.UpdateUserUsedQuota`，使退款和差额调整不改变请求次数；不得使用 `UpdateUserUsedQuotaAndRequestCount` 做退款或已提交旧任务的差额调整。
3. 在 `RefundTaskQuota` 中，以 `task.PrivateData.BillingContext.FixedMeteredBilling != nil` 作为唯一明确的固定计量分流条件：
   - 固定计量：完成资金、订阅和 Token 退款、退款日志、持久化退款标记清除；跳过用户/渠道 `used_quota` 回减。
   - 非固定计量：保留上游的 `UpdateUserUsedQuota(user, -quota)` 与 `UpdateChannelUsedQuota(channel, -quota)`，请求数不变。
4. 在 `RecalculateTaskQuota` 中保留旧任务的上游差额调整：正差额和负差额都只变更用量，不增加请求数。固定计量任务不得误进入该通用路径；如现有调用点可能进入，先明确其结算合同并加单独分支与测试，不能靠注释假定。
5. 所有资金变动、Token 变动、用户/渠道计数变动和日志写入必须仅发生于状态 CAS 成功之后。资金退款失败时保留任务 quota 标记，不提前做计数回减或写成功退款日志。
6. 合入 Midjourney 的 `TokenId`、`BillingChannelId` 和 `UpdateBillingState`，以持久化归因完成退款；历史无该字段的记录继续回退到 `ChannelId`，Token 为 0 时只退钱包，不猜测 Token。

### 4.3 禁止的错误合并方式

- 禁止把上游 `RefundTaskQuota` 两行用量回减无条件复制到本地函数。
- 禁止为了规避冲突删除固定计量快照、成功终态正式消费日志或本地 ZZDH 计量代码。
- 禁止把失败退款写在任务状态条件更新之前，或在 CAS 输家路径补偿。
- 禁止把退款/差额处理改回会增加请求次数的 `UpdateUserUsedQuotaAndRequestCount`。
- 禁止以任务当前模型配置重算固定计量历史价格；必须使用冻结快照。

## 5. Git 冲突处理清单

| 文件 | 冲突性质 | 保留规则 | 完成条件 |
| --- | --- | --- | --- |
| `service/task_billing.go` | 文本与账务语义双重冲突 | 按第 4 节保留本地固定计量和上游旧任务退款修正；不得只选一侧。 | 第 7 节所有账务组合测试通过。 |
| `service/task_billing_test.go` | 两侧都有关键回归覆盖 | 保留本地固定计量/终态消费测试与上游退款、差额、CAS、Token、订阅测试，并增加组合案例。 | 每个生命周期行均有可观察断言。 |
| `model/option.go` | 两套独立配置校验修改 | 同时保留本地 `billing_setting`/fixed-metered JSON 校验和上游 `monitor_setting.channel_test_concurrency` 1-32 校验。 | 两类非法配置都被拒绝。 |
| `service/quota.go`、Midjourney 相关路径 | 不是 Git 同文件冲突，但与任务计费职责重叠 | 合入上游持久化归因与部分成功处理；逐一审查调用者是否影响固定计量或 ZZDH。 | 无余额已扣但任务未标记、或退款丢失归因的路径。 |
| `relaykit/...` | 独立 Go module 的协议改动 | 原样合入转换行为，不引入根模块依赖。 | `cd relaykit; GOWORK=off go build ./...` 通过。 |
| `web/...` | 上游内置 UI 与 Switcher UI 职责不同 | 合入代码但不覆盖 Switcher 前端；不把内置 Web 表单当成运营入口。 | backend-only 部署仍不要求嵌入内置前端。 |

## 6. Switcher 兼容与责任划分

| 项目 | 分类 | 当前策略与必要动作 | 允许启用的条件 |
| --- | --- | --- | --- |
| Advanced Custom 路由编辑、字段透传开关、模型发现 | 不适用/可选体验 | Switcher 已有对应表单、校验和序列化；不复制 new-api 内置编辑器。后端协议验证照常合入。 | 无额外动作。 |
| Advanced Custom 余额查询 | 启用功能前必需 | Switcher 现有余额更新只把 `balance` 视为成功。上游可返回 `success: true, raw_response`；启用该管理路由前，Switcher 必须提供脱敏、大小受限的格式化 JSON 只读展示，不能误报失败或泄露 key。 | 数值余额与非数值合法 JSON 的成功/失败响应均有 UI 回归。 |
| 单渠道测试 API | 合并前必需验证 | Switcher 通过官方 API `/api/channel/test/{id}` 的既有调用应维持兼容。 | 对一个成功渠道和一个可控失败渠道做回归。 |
| 自动渠道测试/禁用/恢复 | 运行策略必需 | 默认 `monitor_setting.auto_test_channel_enabled=false`、`channel_test_concurrency=1`。Switcher Channel Health 是唯一自动策略责任方。 | 仅在明确切换责任方、停用另一侧策略、审计自动禁用/恢复行为后才启用 new-api 自动任务。 |
| 充值上限错误 | 可选体验跟进 | 后端先保证拒绝和原子性；Switcher 可将既有错误展示优化为“余额已达上限”。 | 不要求本次改 Switcher。 |
| Claude、Responses、阿里图像、缓存 Token | 合并前必需验证 | 这些是后端 relay 行为。须确认 Switcher 的普通代理路径经过 new-api；若视频覆盖路径绕开它，单独测试实际路径。 | 对应协议回归通过。 |
| 登录/Session/Cookie | 不适用 | 本上游范围没有服务端认证合同变更；容器重启后网站登录状态失效可重新登录，API Key 调用不应受此影响。 | 保持现有部署行为，无迁移操作。 |

## 7. 测试与验收矩阵

### 7.1 账务组合回归（阻断发布）

每个案例必须同时断言：钱包额度、订阅已用额度、Token 剩余额度与已用额度、用户 `used_quota`、渠道 `used_quota`、请求数、任务 `quota`、正式消费/退款审计日志数及冻结快照。

| 案例 | 预期 |
| --- | --- |
| 固定计量成功，CAS 胜出 | 预扣只发生一次；成功后一次正式消费、用户/渠道用量和请求数各加一次；不重复扣钱包。 |
| 固定计量失败，CAS 胜出 | 钱包/订阅/Token 恢复；退款日志一条；用户/渠道用量和请求数均不变。 |
| 固定计量失败，CAS 输家 | 无任何二次退款、日志或计数变化。 |
| 旧异步任务失败 | 资金/Token 退款，用户/渠道用量回减，请求数不回减。 |
| 旧异步任务成功且实际额度较低 | 仅差额退款；用户/渠道用量按差额回减，请求数不变。 |
| 旧异步任务成功且实际额度较高 | 补扣差额；用户/渠道用量按差额增加，请求数不重复增加。 |
| Token、订阅、无 Token 三种来源 | 不跨资金来源退款；Token 缺失时不伪造 Token 退款。 |
| Midjourney 失败退款 | 使用持久化的计费渠道/Token；历史记录能回退渠道；请求数不回减。 |

### 7.2 上游功能回归（阻断发布）

| 领域 | 最小验证 |
| --- | --- |
| 充值 | 接近 `common.MaxQuota` 的支付前请求被拒绝；两个并发回调只能有一个成功入账；Epay、Stripe、Creem、Waffo、Waffo Pancake 与手工完成路径均受保护。 |
| Responses 计费 | `input_tokens_details.cached_tokens`、cache write/creation、`prompt_cache_hit_tokens` 均归一到缓存计费字段；已有字段不被覆盖。 |
| Claude | 无 tools 时上游请求不含 `tools`；无参数 function 工具有 `type: object` 和空 `properties`。 |
| Chat 转 Responses | 非空 `prompt_cache_key` 在上游请求保留；空值仍省略。 |
| 阿里图片 | 公开别名映射到同步模型/旧 Wan/Wan 时，请求 URL、异步头和转换器均按映射后的上游模型选择。 |
| Advanced Custom | `/v1/models` 与余额路由只能各配置一次，管理路由不允许模型或 converter；数值余额和原始 JSON 回应均有后端测试。 |
| 渠道测试 | 并发配置边界 0、1、32、33；三种模式的选取、禁用许可、恢复行为及取消/租约退出。 |

### 7.3 推荐执行顺序

1. 合并工作树先运行 `git diff --check`，确认没有冲突标记、空白错误或遗漏用户自有未跟踪文件。
2. 运行涉及 `service/task_billing.go`、`service/quota.go`、`service/midjourney.go`、`model/topup.go`、`controller/topup*.go` 的聚焦 Go 测试。
3. 运行 `service/billing_usage.go`、`relay/channel/ali`、`relaykit/dto`、`relaykit/relayconvert`、`controller/channel-test.go` 的聚焦测试。
4. 在 `relaykit` 目录运行 `GOWORK=off go build ./...`；根模块构建成功不能替代这一步。
5. 运行根模块受影响包构建/测试，并用后端专用 Dockerfile 构建镜像；不得改 Compose 或端口来绕过失败。
6. 使用隔离测试数据进行第 7.1 的真实持久化账务验证；之后进行单渠道 Switcher 回归，不启用任何自动渠道任务。

## 8. 分阶段实施、部署与回滚

### 阶段 A：合并前冻结与可恢复点

1. 记录当前 `HEAD`、`origin/main`、工作区状态和运行容器版本；检查用户未跟踪 Apifox 文件仍在。
2. 在保留工作区的前提下创建恢复锚点/备份分支或 stash 方案；不得覆盖本地固定计量提交和未跟踪用户文件。
3. 复查 `origin/main` 是否仍为已审计的 `f11641428`。若变更，重新生成待合并范围，不沿用本计划直接合入。

### 阶段 B：受控代码合并

1. 将 `origin/main` 合入隔离的合并工作树或临时集成分支。
2. 按第 5 节逐个处理三处共同修改；优先完成账务状态机，再处理配置校验。
3. 对非冲突上游功能完整吸收，不因当前不用内置 Web 而跳过后端 DTO、转换、测试或安全修复。
4. 在提交前更新本计划的“文件变更清单”和已验证状态；未通过的项必须保留为 `unverified`，不得宣称可发布。

### 阶段 C：运行前检查

1. 数据库：本范围无显式 migration。Midjourney 新增 `token_id`、`billing_channel_id` GORM 字段，确认生产实例的 AutoMigrate 已开启且实际字段存在；不得为此创建 Switcher 迁移。
2. Redis：确认现有 Redis 可用，尤其是渠道测试系统任务租约和既有任务/缓存依赖；不可用时不启用自动渠道测试。
3. 配置：保持 `auto_test_channel_enabled=false`、并发 1；不要因为配置字段已合入就自动打开新调度器。
4. Compose：核对 `docker-compose-db.yml`、`docker-compose.dev.local.yml` 和 `Dockerfile.host` 与合并前无变化，固定端口不变。

### 阶段 D：灰度与观测

1. 首先部署后端，保持 Switcher 前端不变，使用现有 API Key 做文本、Responses、Claude、阿里图像和一个异步视频样本验证。
2. 人工完成一笔小额官方充值与一个失败任务样本；核对账务日志和四类计数，不做历史修数。
3. 初期不启用 Advanced Custom 非数值余额路由，也不启用 new-api 自动渠道测试；观察日志、退款、Redis 任务和渠道用量。
4. 出现账务不变量失败、额度负数、同一任务多条终态账务日志、渠道双重禁用/恢复时，立即停止灰度并回到阶段 A 锚点。

### 回滚界限

- 代码回滚可回到合并前锚点；不应删除数据库字段，新增 Midjourney 字段可保留且对旧代码无害。
- 已发生的充值、消费或退款不得用代码回滚抵消。必须保留审计记录并用单独、经备份和复核的数据修复流程处理。
- 不通过改端口、替换 Compose、清空数据库/Redis 或删除用户文件来完成回滚。

## 9. 只能保留旧功能的明确条件

以下不是默认取舍。只有满足前置条件仍无法证明兼容，或可能破坏既有功能时，才允许暂不启用新能力；每项都要在发布记录中写明。

| 条件 | 保留旧行为 | 暂不启用/放弃的新行为 | 恢复新能力的门槛 |
| --- | --- | --- | --- |
| 固定计量退款无法证明提交前已写用量 | 固定计量失败只退预扣资金/Token，不回减用户/渠道用量。 | 不将上游通用退款计数回减用于固定计量。 | 有持久化证据证明固定计量在提交期增加了哪些计数，并有完整组合回归。 |
| Switcher 无原始余额响应的安全展示 | 保留数值 `balance` 查询和既有错误提示。 | 不配置/不使用 Advanced Custom 的非数值余额路由。 | Switcher 完成脱敏、大小限制、格式化只读展示及错误回归。 |
| Switcher Channel Health 仍在自动运行 | 保留 Switcher 为唯一策略执行者。 | new-api 的 `auto_test_channel_enabled` 保持关闭。 | 明确迁移责任、停用另一侧、审计禁用/恢复和告警链路。 |
| Midjourney 表字段未在目标库创建 | 保留当前部署，不发布依赖持久化归因的新退款路径。 | 延后 Midjourney 新账务归因功能。 | 已验证 AutoMigrate/字段存在，且升级演练通过。 |
| relaykit 独立构建失败 | 保留当前可构建 relaykit，不发布该批转换更新。 | 延后 Claude/Responses 转换修复，不能只把根模块强行编过。 | `GOWORK=off go build ./...` 通过并有协议回归。 |

## 10. 文件变更清单（计划态）

| 路径 | 操作 | 责任与兼容影响 | 验证状态 |
| --- | --- | --- | --- |
| `doc/tasks/MAIN_UPSTREAM_MERGE_PLAN_20260820.md` | add | 本次合并的中文执行、兼容、回滚和验收基线；不改变运行行为。 | 已按 `9fdbf3add` 执行并验证。 |
| `doc/DEVELOPMENT_TASK_INDEX.md` | modify | 为后续任务定位增加本计划入口；不改变运行行为。 | 已恢复并保留，待单独提交文档变更。 |
| `service/task_billing.go` | retain then merge | 最关键账务状态机；同时保留固定计量与旧任务计数修正。 | 已合并；固定计量分流测试通过。 |
| `service/task_billing_test.go` | retain then merge | 固定计量、退款、差额、CAS 组合回归。 | 已合并；聚焦账务测试通过。 |
| `model/option.go` | retain then merge | 固定计量配置与渠道测试并发配置的双重校验。 | 已合并；配置测试通过。 |
| `ZZDH_VIDEO_API_APIFOX.openapi.yaml` | retain | 用户自有未跟踪接口定义，不纳入提交或清理。 | 已确认保留。 |
| `ZZDH_VIDEO_API_APIFOX.postman_collection.json` | retain | 用户自有未跟踪 Postman 集合，不纳入提交或清理。 | 已确认保留。 |

## 11. 合并后发布判定

代码合并已完成，但只有同时满足以下条件，才可判定为“可灰度”：

1. 第 4 节两类任务的计数语义均已在代码和测试中落实。
2. 第 7 节 P0/P1 验收通过，relaykit 独立构建通过，且没有未解释的测试跳过。
3. Switcher 的单渠道测试兼容已验证；new-api 自动渠道测试仍关闭，或已有唯一责任方切换记录。
4. 生产数据库字段、Redis、Docker 后端镜像、Compose 端口不变均已核对。
5. `git diff --check` 通过，用户自有未跟踪文件仍保留，计划态文件已更新为实际合并结果。

在任一条件不满足时，结论只能是“暂不灰度”；不得把账务冲突当作普通冲突用一侧代码覆盖另一侧。
