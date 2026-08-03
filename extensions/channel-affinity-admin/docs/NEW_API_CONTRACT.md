# 与 new-api 的唯一契约

## 目的与边界

本扩展位于 `new-api/extensions/channel-affinity-admin/`，但**不改动 new-api 的 router、controller、middleware、service、model 或数据库**。它只依赖下面这个已存在的稳定行为：Channel Affinity 在 `middleware.Distribute` 开始常规随机选路前查找 Redis 中的首选渠道。

本项目不访问 new-api 数据库，也不导入其 Go 包。渠道是否启用、是否拥有对应分组/模型能力、是否支持请求路径，仍全部由 new-api 本身决定。

## 管理员身份契约

扩展不维护独立 Key。管理请求携带调用方现有的 new-api Dashboard Access Token 或个人访问令牌（PAT）到 `Authorization: Bearer ...`；扩展以 `NEW_API_BASE_URL` 调用 `GET /api/user/self`，仅当响应的 `role >= 10` 时允许读写。

这是 Switcher 现有 `CurrentUserContext -> OfficialUserApiClient -> /api/user/self` 的同一鉴权事实来源，不新增用户表、角色表或凭据。`NEW_API_BASE_URL` 应配置为 Switcher 已有 `ReverseProxy.TargetBaseAddress` 所指向的 new-api 地址。

普通 `/v1/*` 中继 Key 只用于模型请求，不能通过 `/api/user/self` 验证为管理身份；此处必须使用 Switcher 当前管理后台已经持有的 Dashboard Access Token 或 PAT。

## 对外路径

扩展是仓库内的独立进程，new-api 主路由不注册它的 HTTP 路径。若管理后台通过与 Switcher 相同的域名访问，需要在现有网关增加路径映射，并原样转发 `Authorization`：

```text
/api/extensions/channel-affinity/v1/admin/*
  -> http://127.0.0.1:8089/v1/admin/*
```

这是一项部署配置，不是 new-api 或 Switcher 的代码耦合。没有该映射时，只能从受限内网直接调用扩展端口。

## 必需的一次性运行配置

在 new-api 的「Channel Affinity」设置中，保留已有规则，并在**最后**追加下列规则。名称、TTL 和 include 三项必须与本服务的环境变量保持一致。

```json
{
  "name": "external-admin-user-channel-v1",
  "model_regex": ["^.+$"],
  "path_regex": ["^/v1/(chat/completions|completions|embeddings|images/generations|responses|messages)$"],
  "key_sources": [
    { "type": "context_int", "key": "id" }
  ],
  "ttl_seconds": 3600,
  "skip_retry_on_failure": false,
  "include_using_group": true,
  "include_model_name": true,
  "include_rule_name": true
}
```

全局项建议保持：

```json
{
  "enabled": true,
  "switch_on_success": true,
  "keep_on_channel_disabled": false
}
```

并将 new-api 的运行时 `RetryTimes` 设为至少 `1`，否则首选 B 失败后没有第二次尝试。

### 规则顺序是硬边界

new-api 对 Channel Affinity 规则按顺序查找；一个规则拿到 key source 后，即使 Redis 中没有绑定，也不会再继续检查后面的规则。因此外部规则必须位于已有 `codex cli trace`、`claude cli trace` 之后，不能置顶。

结果是：带有 `prompt_cache_key` 的 Codex Responses 请求，或带有 `metadata.user_id` 的 Claude CLI 请求，继续由其既有规则控制；外部绑定不会覆盖它们。这是维持零源码侵入、不破坏已有 CLI 亲和性的必要取舍。

若业务要求“外部管理员绑定必须覆盖所有请求，包括上述两类 CLI 请求”，当前 new-api 的规则扫描语义无法在完全隔离前提下保证，应先做独立评审；不要把外部规则置顶作为替代方案。

## Redis 协议

现有 new-api 规则构造键的方法等价于：

```text
new-api:channel_affinity:v1:<rule-name>:<model>:<using-group>:<user-id>
```

本服务以 Redis String 写入十进制 `channel_id`，TTL 为 `AFFINITY_TTL_SECONDS`。例：

```text
key   = new-api:channel_affinity:v1:external-admin-user-channel-v1:gpt-4.1:vip:1001
value = 42
ttl   = 3600 seconds
```

`rule-name`、`model`、`group` 均不能含 `:`，因为 new-api 当前协议没有转义。服务会拒绝该输入，避免生成有歧义的键。

### 自动学习范围（零源码方案的限制）

规则以 `context_int:id` 匹配后，new-api 即使第一次在 Redis 没有读到绑定，仍会在本次常规随机选路成功后回写该用户、模型、分组对应的亲和键。因此，只要请求路径和模型落入本规则，**未被管理员显式配置的用户也会形成普通亲和记录**。

管理员写入的 B 仍会覆盖这条记录，所以管理 API 的“立即切换到 B”是成立的；但这套方案不能区分“管理员创建的键”和“new-api 自动学习的键”。如果业务要求仅管理员绑定的用户启用亲和性，必须增加一个只在已绑定时才记录的规则级开关，或在请求入口提供可信的专用标识；二者都超出纯 Redis 外部实现的边界。

### TTL 必须一致

请求成功后，new-api 会以该规则的 TTL 回写亲和性。因此服务环境变量 `AFFINITY_TTL_SECONDS` 必须与规则的有效 TTL 相同；不要为单条绑定提供不同 TTL，否则首次成功后会被 new-api 回写覆盖。

## 失败与回退语义

`skip_retry_on_failure=false` 且 `RetryTimes >= 1` 时：

1. B 是本次请求的首选渠道；
2. B 上游失败后，new-api 重试并从当前实际分组的其他可用渠道中选择；
3. B 每次成功时，亲和记录仍为 B；
4. 如果 B 失败、C 重试成功，在 `switch_on_success=true`（当前默认）下，new-api 会将下次首选回写为 C。

第 4 点意味着这是“优先/学习型亲和”，不是“B 永久硬锁定”。若必须让 B 在回退成功后仍是下一次首选，现有全局 `switch_on_success=false` 会同时影响所有亲和性规则；在不影响其它规则的前提下，需要 new-api 后续增加规则级开关，属于源码改动边界。

## AutoGroup

默认禁止 `group=auto`。若显式开启 `AFFINITY_ALLOW_AUTO_GROUP=true`，服务会写入 `...:<model>:auto:<user-id>`。new-api 会再从该 Token 的 AutoGroups 列表里寻找能使用 B 的实际组；它不是“在所有渠道中选 B”，也不能指定某一个 AutoGroup 子分组。只有确实需要这个非确定行为时才开启。
