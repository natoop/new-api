# 渠道亲和性管理员扩展升级核验清单

每次同步或升级 new-api 后，在启用写操作前检查：

1. `router/api-router.go` 仍保留 `channelaffinityadmin.Register(apiRouter)` 这一个注册接缝，路径仍为 `/api/extensions/channel-affinity`。
2. `middleware.AdminAuth()` 仍为管理路由提供当前管理员 ID（Gin context 的 `id`）和原有审计保护。
3. `service/channel_affinity.go` 的命名空间仍是 `new-api:channel_affinity:v1`，键的组成顺序仍为 `rule-name:model:using-group:affinity-value`。
4. `middleware/distributor.go` 仍在常规随机选路前调用 `GetPreferredChannelByAffinity`，并校验渠道状态、路径、分组和模型能力。
5. `RecordChannelAffinity` 仍使用规则 TTL，且 `switch_on_success` 的回写语义未变。
6. `controller/relay.go` 的重试流程仍遵守 `skip_retry_on_failure`；生产 `RetryTimes >= 1`。
7. 配置中的外部规则仍位于 Codex/Claude CLI 规则之后，名称及 `include_*`、key source 与 [契约](NEW_API_CONTRACT.md) 一致。
8. 使用测试用户验证：写入 B 后 B 成功；B 失败后 C 重试成功；删除绑定后回到常规选路。

第 2–6 项任一变化时，先停止使用该管理接口写入；更新契约、实现和针对键协议的单元测试后再恢复。
