# 渠道亲和性管理员扩展与 new-api 的契约

## 代码边界

扩展位于 `extensions/channel-affinity-admin/`，属于根 `github.com/QuantumNous/new-api` Go 模块。它可以复用公开的根模块包，但不修改 router 以外的 new-api 核心源码，不访问 new-api 数据库，也不创建独立运行进程。

唯一核心接缝是 `router/api-router.go` 的 `channelaffinityadmin.Register(apiRouter)`。它将受 `middleware.AdminAuth()` 保护的管理路由注册到已有 `/api` 路由组。

扩展直接使用 `common.RDB`，因此要求 new-api 已启用 Redis。没有 Redis、Channel Affinity 被关闭，或下述规则不存在/不匹配时，管理接口会失败，不会写出无法被选路读取的记录。

## 管理员身份契约

权限完全由已有 `middleware.AdminAuth()` 判定：只有 new-api 的管理员 Dashboard 会话或管理员 PAT 可以访问。扩展不解析、不转发、不存储认证凭据，也不新增任何 Key、环境变量或角色表。

## 亲和 Redis 协议

针对规则 `external-admin-user-channel-v1`，键必须和 `service/channel_affinity.go` 的 `buildChannelAffinityCacheKeySuffix` 一致：

```text
new-api:channel_affinity:v1:external-admin-user-channel-v1:<model>:<actual-group>:<user-id>
```

值为十进制渠道 ID，TTL 取该规则的有效 TTL。例：

```text
key   = new-api:channel_affinity:v1:external-admin-user-channel-v1:gpt-4.1:vip:1001
value = 42
ttl   = 3600 seconds
```

模型和分组不允许包含 `:`、换行或 NUL，避免生成歧义 Redis 键；`group=auto` 被拒绝，因为 AutoGroup 是多个实际分组的回退列表，无法给出确定的单个组绑定。

## 必需的运行规则

以下规则必须置于已有 Codex/Claude CLI 规则之后。规则按顺序匹配：前面的 CLI 规则已取到 key source 时，即使 Redis 未命中也不会继续进入外部规则。这个顺序确保外部绑定不会破坏现有 CLI 亲和行为。

```json
{
  "name": "external-admin-user-channel-v1",
  "model_regex": ["^.+$"],
  "path_regex": ["^/v1/(chat/completions|completions|embeddings|images/generations|responses|messages)$"],
  "key_sources": [{ "type": "context_int", "key": "id" }],
  "value_regex": "",
  "ttl_seconds": 3600,
  "skip_retry_on_failure": false,
  "include_using_group": true,
  "include_model_name": true,
  "include_rule_name": true
}
```

全局要求：`enabled=true`，并将 `RetryTimes` 设为至少 `1`。建议保持 `switch_on_success=true`、`keep_on_channel_disabled=false`。

扩展会拒绝不满足以下条件的同名规则，防止写入错误格式的键：唯一 key source 为 `context_int:id`，`value_regex` 为空，并且三个 `include_*` 选项均为 `true`。

## 成功、失败与回写

1. 管理员将某绑定写为 B 后，下一个匹配请求优先尝试 B。
2. B 不失败时，原始 `RecordChannelAffinity` 继续将成功渠道 B 写回同一键。
3. B 失败且 `skip_retry_on_failure=false`、`RetryTimes >= 1` 时，原始重试逻辑可在同一实际分组中选择其他可用渠道。
4. 若 C 重试成功，`switch_on_success=true` 会将下一次首选回写为 C。

因此这是“指定首选、失败可回退”的亲和，而非 B 的永久硬锁定。要让回退成功后仍永远保持 B，需要 new-api 增加规则级回写开关；不能在本扩展中隔离实现。

同样，因规则按用户 ID 匹配，范围内未被管理员显式写入的请求成功后也会被 new-api 的既有逻辑自动学习为亲和记录。纯 Redis 扩展不能区分“管理员创建”与“自动学习”记录。
