---
report_type: upstream-sync-audit
generated_at: 2026-08-10T09:13:16+08:00
timezone: Asia/Shanghai
repository: D:\code\goswtich\new-api
branch: feature/free/local
tracking_branch: origin/feature/free/local
baseline_head: ac2e1399e6faeeed573ae89ab13fdb703672bb35
baseline_synced_upstream: 0ab02020603d22e5613bc4cf46bfab06f8567769
saved_head: 5694a0ed66c339d03c98d8e68434357dd460cf49
saved_synced_upstream: 5c3abffe8572aa8a49f15c3916707d2019d66af4
upstream_tracking_snapshot: 823e26304a396854ace30b52b98ec497c2dd9c36
upstream_tracking_snapshot_fetched_at: 2026-08-08T14:52:13+08:00
upstream_refresh_attempt_at: 2026-08-10T09:08:00+08:00
upstream_refresh_result: failed-github-443-timeout
scope: effects of merge 5694a0ed and upstream range 0ab020206..5c3abffe; later upstream commits are pending
---

# 2026-08-07 上游合并功能审计

## 1. 核心结论

- 本次同步 merge：`5694a0ed66c339d03c98d8e68434357dd460cf49`。
- 第一父提交（合并前本地）：`96c139d268652487aa3d01c49dd545afe40d82f4`。
- 第二父提交（本次带入上游）：`5c3abffe8572aa8a49f15c3916707d2019d66af4`。
- 上次已同步上游锚点：`0ab02020603d22e5613bc4cf46bfab06f8567769`。
- 本次纯上游范围：`0ab020206..5c3abffe`，共 8 个 first-parent 提交。
- 净变化：35 个文件，`+2407/-552`；后端 26 个文件，前端 8 个文件，CI 1 个文件。
- merge patch-id 与纯上游范围 patch-id 都是 `c4ab427cbd1b31d5b8f256222bc07b4cff0ec431`，没有冲突解决改写或额外本地适配。
- 没有数据库字段迁移，没有 Docker/Compose/端口变化，没有 Session/Cookie 合同变化。
- 对 Switcher 现有 `/v1/*` API 调用最直接的收益，是上游 HTTP/2 短暂重置时可以安全重放完整请求体。

## 2. 提交清单

| 提交 | 时间 | 功能 |
|---|---|---|
| `d6b5ce99` | 08-06 15:56 | 为上游请求补齐可重放请求体，支持 HTTP/2 `REFUSED_STREAM`/`GOAWAY` 自动重试，并停止自动跟随上游重定向 |
| `ea4f0210` | 08-06 17:33 | 将重放元数据收敛到请求体对象，避免在 RelayInfo 跨渠道复用时残留失效状态 |
| `0cd9dc85` | 08-06 17:53 | 用户 Access Token、邀请奖励和用户资料更新改为列级/原子更新，避免并发覆盖额度数据 |
| `c9bc0386` | 08-07 13:34 | 管理前端抓取模型时扩展厂商分类并优化新模型/已有模型/已移除模型展示 |
| `b941253a` | 08-07 13:35 | 渠道测试对 Claude、Gemini 使用各自原生请求格式 |
| `1da23d6b` | 08-07 16:01 | Access Token 生成和推广额度转移增加按用户关键操作限流 |
| `e926e5ca` | 08-07 17:06 | 修复管理前端编辑兑换码时的额度精度损失和旧数据覆盖 |
| `5c3abffe` | 08-07 17:40 | GitHub Release 同步到 GitCode 的 CI 支持可选文件同步 |

## 3. 后端功能与行为变化

### 3.1 上游 HTTP/2 请求可安全重放

此前 JSON 请求体经过 `BodyStorage` 类型擦除后，Go `net/http` 无法自动生成 `Request.GetBody`。当上游 HTTP/2 在请求体已发送后返回可重试的 `REFUSED_STREAM`，或连接收到 `GOAWAY`，Transport 无法重新发送完整请求体，请求会直接失败。

本次改动：

- `BodyStorage` 新增 `NewReader()`，内存模式创建独立 `bytes.Reader`，磁盘模式重新打开独立文件描述符。
- 新增 `ReplayableBody`，同时提供请求体大小和独立重放 reader。
- 请求构造时设置正确的 `ContentLength` 与 `GetBody`。
- Chat、Claude、Gemini、Responses、Embedding、Image、Rerank、Alpha Search 和部分异步任务路径统一接入。
- passthrough 请求和渠道重试也保留正确的重放能力。
- 不可重放的请求不再伪造一个会返回已消费 reader 的 `GetBody`，避免重试时静默发送空请求体。

这不会让业务层重复扣费。它属于单次上游 HTTP 传输的透明重试，不是 new-api 的渠道失败重试流程。

### 3.2 上游重定向不再自动跟随

共享 HTTP Client 会被浅复制，并设置 `CheckRedirect = http.ErrUseLastResponse`。上游返回 301/302/307/308 时，new-api 保留原始 3xx，不再自动请求重定向地址。

这是正确的安全收敛，可以避免请求体或鉴权头被带到意外目标；但如果某个渠道地址依赖 HTTP 重定向才能到达真实 API，本次升级后该渠道会暴露 3xx/上游错误，需要把 Base URL 改成最终地址。

### 3.3 用户数据并发更新加固

本次的 `Access Token` 指 `users.access_token`，即 Dashboard 用户个人管理 Token，不是 `tokens` 表里的 API Key，也不是之前讨论的 `auto_groups`。

- 生成个人 Access Token 时只更新 `access_token` 列，不再保存整份旧 User 快照。
- 普通用户资料更新明确排除 `access_token`、quota、request count 和推广统计字段。
- 邀请人数、邀请额度和历史额度使用数据库原子加法，避免并发邀请丢计数。
- 推广额度转余额的加锁查询修正为直接传入 User 指针。

主要目标是防止“用户修改资料/刷新个人 Token”和“并发计费/邀请奖励”互相覆盖数据库字段。

### 3.4 关键用户操作按用户 ID 限流

新增 `UserCriticalRateLimit(scope)`：

- `GET /api/user/token` 同时经过原有 IP 级 Critical 限流和新的用户级 `access-token` 限流。
- `POST /api/user/aff_transfer` 增加用户级 `aff-transfer` 限流。
- Redis 开启时使用用户 ID Redis key；未开启时使用内存 key。
- 默认沿用 `CRITICAL_RATE_LIMIT=20`、`CRITICAL_RATE_LIMIT_DURATION=1200` 秒。
- 超限返回 HTTP 429，并设置 `Retry-After`。

这不作用于 `/v1/*` 模型调用，也不要求 Switcher 中间件配套修改。

### 3.5 Claude/Gemini 渠道测试使用原生协议

渠道测试不再统一伪造 OpenAI Chat 请求：

- Claude 使用 `ClaudeRequest` 与原生 messages/max_tokens。
- Gemini 使用 `GeminiChatRequest` 与原生 contents/generationConfig。
- Gemini 流测试路径使用 `:streamGenerateContent`。

这减少了渠道实际可用、但因为测试请求格式错误而被判定失败的情况。生产 relay 协议没有因此改变。

## 4. 前端功能

### 4.1 抓取模型后的厂商分类增强

bundled web 的渠道“抓取模型”弹窗：

- 厂商规则从少量关键字扩展到 OpenAI、Anthropic、Gemini、xAI、Qwen、DeepSeek、Meta、Mistral、NVIDIA、Perplexity、国内模型厂商和视频/音乐平台等。
- 厂商按名称排序，`Other` 放最后。
- 新模型、已有模型、上游已移除模型分别展示。
- 默认优先打开有内容的页签。
- 分类逻辑提取成独立模块，避免每次渲染重复拆分。

你使用 Switcher 前端，因此这部分不会自动出现在 Switcher 页面；后端 API 合同没有变化。

### 4.2 兑换码编辑不再损失额度精度

旧逻辑会把 quota 转成用于展示的有限小数，再在保存时转回整数 quota。管理员即使没有修改额度，也可能因为显示舍入导致原额度被改小或改大。

新逻辑：

- 编辑表单使用与当前货币配置匹配的可编辑精度和 input step。
- 如果额度字段没有被用户修改，更新时直接保留后端加载的原始 quota 整数。
- 加载未完成或目标兑换码不一致时禁止提交。
- 提交期间禁用按钮，避免重复请求。

这是 bundled web 的管理端修复，兑换码后端 API 没有变化。Switcher 若有自己的兑换码编辑页，需要单独确认是否存在相同的“展示值回写”问题。

## 5. CI 变化

`.github/workflows/sync-release-to-gitcode.yml` 增加可选文件同步能力并整理 workflow 参数。它只影响 Release/GitCode 发布流程，不影响容器运行、数据库或 API 请求。

## 6. 对现有部署和 Switcher 的影响

- `docker-compose-db.yml`、`docker-compose.dev.local.yml`、`Dockerfile.host` 和端口映射均未被本次上游范围修改。
- 没有新增数据库列，不需要手工迁移。
- Session/Cookie/JWT 核心文件没有变化。
- Switcher `/v1/*` 反向代理和 API Key 认证合同没有变化。
- 不需要修改 Switcher 中间件。
- API 调用会直接获得 HTTP/2 短暂重置重试能力。
- bundled web 的模型分类与兑换码精度修复不会显示在 Switcher 前端。
- 需要检查所有渠道 Base URL 是否依赖 301/302 重定向；应配置为最终地址。

## 7. 验证结果

- `find-sync-merges.ps1`：识别到单笔同步 `5694a0ed`，同步 patch 与纯上游 patch 完全一致。
- `git diff --check 0ab020206..5c3abffe`：通过。
- `go test ./common -run '^TestNewReplayableBodyReaderKeepsStorageLifecycleWithCaller$'`：通过。
- relay 请求体 metadata、`REFUSED_STREAM` 重试、passthrough 重试、不可重放边界和 3xx 不跟随测试：通过。
- `go test ./model` 中用户并发更新与 Access Token 列级更新测试：通过。
- Sora passthrough replayable body 测试：通过。
- controller/middleware 编译检查：通过。
- `TestUpstreamGetBody_HTTP2RetryAfterGracefulGoAway_PassThrough` 在当前 Windows 环境重复失败，表现为本机 TCP 被远端强制关闭，未观察到预期的新连接重试。`REFUSED_STREAM` 路径通过，因此需要在 Linux/实际容器环境再次确认 GOAWAY 分支。
- bundled web `bun run typecheck` 未通过：当前依赖缺少 `yace` 与 `happy-dom` 类型/模块，同时存在由缺失类型引发的隐式 any。未执行 `bun install`，不能把该结果直接归因于本次 8 个提交。
- 未重启容器，未执行正式部署。

## 8. 待下次同步

本地最后成功保存的上游跟踪快照来自 2026-08-08 14:52:13 +08:00：

- `upstream/main = 823e26304a396854ace30b52b98ec497c2dd9c36`。
- `5c3abffe..823e2630` 有 2 个尚未合入提交：
  - `2399de97`：阿里渠道在客户端未提供 `top_p` 时不再自动注入。
  - `823e2630`：修正 Qwen TTS 模型分类。

2026-08-10 再次 fetch 时 GitHub 443 超时，因此不能声称上述快照是远端当前最新状态，也不能把可能存在的后续提交计入本次已合并结果。

## 9. 下次审计锚点

```text
saved_head=5694a0ed66c339d03c98d8e68434357dd460cf49
saved_synced_upstream=5c3abffe8572aa8a49f15c3916707d2019d66af4
```

工作区原有的 `Dockerfile.dev` 修改和两个 `ZZDH_VIDEO_API_APIFOX` 未跟踪文件未纳入本次统计，也未修改。
