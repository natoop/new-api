# 渠道亲和性管理员扩展

该扩展是 `new-api` 根 Go 模块中的一个独立目录，不是独立项目或进程。它为管理员提供：

```text
用户 ID + 实际分组 + 模型 -> 优先渠道 ID
```

它直接复用 `new-api` 的管理员鉴权、Redis 连接和 Channel Affinity Redis 协议；不新增环境变量、端口、Docker 服务、数据库表或独立 Key。

## 路由与权限

路由由 `router/api-router.go` 中唯一的注册接缝挂载：

```text
/api/extensions/channel-affinity
```

该路由组使用现有的 `middleware.AdminAuth()`。调用时携带当前 new-api 管理后台已经使用的 Dashboard Access Token 或 PAT；普通中继 API Key 不具备管理员权限，无法调用此接口。

| 方法 | 用途 |
| --- | --- |
| `PUT /api/extensions/channel-affinity` | 创建或立即改绑到指定渠道 |
| `GET /api/extensions/channel-affinity?user_id=...&group=...&model=...` | 查询绑定及剩余 TTL |
| `DELETE /api/extensions/channel-affinity?user_id=...&group=...&model=...` | 删除绑定 |

写入示例：

```http
PUT /api/extensions/channel-affinity
Authorization: Bearer <new-api-admin-dashboard-token-or-pat>
Content-Type: application/json

{
  "user_id": 1001,
  "group": "vip",
  "model": "gpt-4.1",
  "channel_id": 42
}
```

`group` 必须是一个实际分组，不能是 `auto`；`channel_id` 是否启用、是否具备该分组和模型能力仍由原有选路逻辑校验。

## 运行前提

在系统设置的 Channel Affinity 规则中，将以下规则添加在既有 `codex cli trace` 和 `claude cli trace` 规则之后：

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

同时保持 Channel Affinity 已启用，并将全局 `RetryTimes` 设为至少 `1`。这是运行配置，不要求为扩展配置任何环境变量。

## 行为边界

- 写入会立即覆盖同一 `用户 + 分组 + 模型` 的 Redis 亲和记录；下一个请求优先选 B。
- B 成功时，B 会继续保持为首选；B 失败且重试成功到 C 时，在 `switch_on_success=true` 下，下次首选会回写为 C。
- 亲和记录的 TTL 取当前规则配置；扩展与请求成功后的原有回写使用同一 TTL。
- 每次创建、改绑或删除都会写入 Redis Stream `new-api:channel_affinity_admin:v1:audit`。现有 `AdminAuth` 也会照常记录管理操作审计。

升级前后请按 [升级核验清单](docs/UPGRADE_CHECKLIST.md) 检查；完整兼容性边界见 [new-api 契约](docs/NEW_API_CONTRACT.md)。
