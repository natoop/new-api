---
report_type: upstream-sync-audit
generated_at: 2026-08-13T00:00:00+08:00
timezone: Asia/Shanghai
repository: D:\\code\\goswtich\\new-api
branch: feature/free/local
tracking_branch: origin/feature/free/local
baseline_head: 823e26304a396854ace30b52b98ec497c2dd9c36
baseline_synced_upstream: 823e26304a396854ace30b52b98ec497c2dd9c36
upstream_tracking_snapshot: ccd535ef8
scope: pre-merge audit of origin/main 823e26304..ccd535ef8; local worktree changes excluded
decision: merge after record; preserve local custom billing, ZZDH, Docker/Compose, Switcher-facing deployment, and uncommitted files
---

# 2026-08-13 main 与 feature/free/local 合并前审查

## 范围与基线

远端 `origin/main`、`upstream/main` 当前均指向 `ccd535ef8`，本地 `main` 不是基准，实际待同步范围为：

```text
823e26304..ccd535ef8
16 个 first-parent 提交，103 个文件，+2937/-746
```

当前分支为 `feature/free/local`，工作区存在以下未提交内容，合并前后均须保留，不应自动纳入上游合并提交：

```text
M  doc/tasks/FIXED_METERED_BILLING_TASK.md
M  setting/ratio_setting/model_ratio.go
?? ZZDH_VIDEO_API_APIFOX.openapi.yaml
?? ZZDH_VIDEO_API_APIFOX.postman_collection.json
```

## 上游功能摘要

- 渠道自动测试新增 `auto_ban_only`，只测试启用自动禁用的渠道。
- 前端搜索增加防抖；登录 Turnstile token 刷新；移动端侧栏修复；Access Token 重新生成增加确认。
- reasoning effort 使用日志修复；参数覆盖增加 `user_id`、`user_group`、`token_group`、`using_group` 上下文。
- Ollama reasoning/tool-call context 修复；Gemini 风格 `/v1/models` 列表修复。
- 后台配置增加 Unicode 长度校验；动态计费日志记录实际命中的请求条件和倍率。
- Responses 保留 presence/frequency penalty，并删除 `-openai-compact` 后缀及其价格 fallback/自动生成机制，改由 `/v1/responses/compact` 路径和渠道能力判断。
- 充值回调改为事务、行锁和幂等结算；用户钱包额度、Token quota、Redis 缓存更新增加原子预扣和并发保护，避免旧快照覆盖新余额。
- OAuth 绑定避免覆盖用户状态。

## 对本地定制的影响

1. `setting/ratio_setting/model_ratio.go` 是未提交本地改动，同时被上游 Compact 删除逻辑触及，不能使用普通自动解决结果覆盖；需保留本地计费规则并明确 Compact 兼容策略。
2. 本地仍有 Compact 后缀相关逻辑，Switcher 前端使用 `/v1/responses/compact` 路径但未发现明确依赖 `-openai-compact` 模型名。合并后仍需核对生产数据库是否存在该后缀模型、价格或映射配置。
3. 本地 fixed metered billing、ZZDH 任务计费和全局计费器需确认完整覆盖新 `PreConsume`/`Settle`/`Refund` 原子额度链路；不能仅以编译通过判断兼容。
4. `common/api_type.go` 与 `relay/common/relay_info.go` 同时包含本地扩展和上游改动，预期需要手工保留两边语义。
5. Dockerfile、Compose、固定端口、Switcher 前端和中间件不在本次上游范围内；不修改端口、Compose 或前端部署方式。

## 合并策略与验证

- 使用 `origin/main` 合并到 `feature/free/local`，不切换到本地 `main`。
- 先保留四个未提交文件，遇到重叠文件逐项人工处理；禁止覆盖本地 Docker、ZZDH、channel-affinity-admin、fixed metered billing 定制。
- 合并后执行根模块构建/聚焦测试，并单独执行 `cd relaykit; GOWORK=off go build ./...`。
- 合并后复查 Compact 模型配置、Redis 额度缓存结构和本地计费调用链，再决定是否部署。

## 结论

这批上游修复值得合入，但不是无条件覆盖式同步。记录完成后允许执行合并；合并完成不等于已完成生产部署，Compact 兼容及本地计费链路复查仍是上线前门槛。

## 合并执行结果

- 已执行 `git merge --no-edit origin/main`，生成合并提交 `659b1f2ab`，无 Git 冲突。
- 已恢复合并前保护的本地工作区；四个本地文件均未被覆盖。
- `cd relaykit; GOWORK=off go build ./...` 通过。
- 根模块 `go build ./...` 未通过，原因是当前工作区缺少嵌入文件 `web/dist/index.html`，不是 Go 合并冲突。
- 聚焦测试未全通过：`model`/`service` 中部分 session 缓存测试因本机 Redis `127.0.0.1:61406` 未启动而失败；另有 `TestObserveChannelAffinityUsageCacheByRelayFormat_UnsupportedModeKeepsEmpty` 断言失败，需要单独复核本地 channel-affinity 改动。`relay/common`、`router`、`setting/ratio_setting` 测试通过。
- 未重启容器、未修改端口、未执行生产部署。
