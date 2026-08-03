# new-api 渠道亲和性管理员服务

这是 `new-api/extensions/` 下的独立扩展模块。它不导入、不修改、不编译 `new-api` 的任何源码；运行时只在同一个 Redis 中维护 new-api 已有的 Channel Affinity 缓存键。

它实现的语义是：

`用户 ID + 实际分组 + 模型 -> 优先渠道 ID`

渠道可用时，new-api 的原始选路会先尝试该渠道；渠道上游失败时，由 new-api 现有重试机制选择同分组的其他可用渠道。

## 先读边界

部署前必须按 [docs/NEW_API_CONTRACT.md](docs/NEW_API_CONTRACT.md) 配置一条 new-api 的亲和性规则。该规则是一次性的运行配置，不是源码改动。

本服务默认拒绝 `group=auto`。AutoGroup 不是“全部渠道池”，而是多个实际分组的有序回退列表；第一版绑定实际分组才有确定语义。

外部规则按用户 ID 匹配后，new-api 会在请求成功时自动回写亲和性；这意味着符合该规则范围、但未被管理员显式绑定的用户也会开始形成普通亲和记录。它不是“仅管理员创建的键才启用亲和性”的开关；这个限制和后续最小源码接缝见 [docs/NEW_API_CONTRACT.md](docs/NEW_API_CONTRACT.md)。

## 安全模型

- 服务默认只监听 `127.0.0.1:8089`；远程访问应通过内网 TLS 反向代理。
- 所有 `/v1/admin/*` 操作都要求调用方当前的 new-api `Authorization: Bearer <token>`。扩展会调用 new-api 的 `/api/user/self`，仅 `role >= 10` 的管理员可操作。
- 不增加独立 Key；Switcher 已使用这同一身份查询链路。部署时将 `NEW_API_BASE_URL` 配成 Switcher 的 `ReverseProxy.TargetBaseAddress` 所指向的 new-api 地址即可。
- 这里的 Token 是 Dashboard Access Token 或个人访问令牌（PAT）；普通 `/v1/*` 中继 Key 不能调用 `/api/user/self`，因此不能作为管理凭据。
- 可额外用 `AFFINITY_ALLOWED_CIDRS` 限制直连来源；服务只相信 TCP 对端地址，不相信可伪造的 `X-Forwarded-For`。
- 每次写入/删除同时写入 Redis Stream：`new-api:channel_affinity_admin:v1:audit`；凭据本身不会写入日志或 Stream。

复制 `.env.example` 为部署环境变量后启动：

```powershell
go mod tidy
go run ./cmd/affinity-admin
```

## 管理 API

所有管理接口都需要管理员 Bearer Key。

写入或立即改绑（幂等）：

```http
PUT /v1/admin/channel-affinities
Authorization: Bearer <new-api-dashboard-access-token-or-PAT>
Content-Type: application/json

{
  "user_id": 1001,
  "group": "vip",
  "model": "gpt-4.1",
  "channel_id": 42
}
```

查询：

```http
GET /v1/admin/channel-affinities?user_id=1001&group=vip&model=gpt-4.1
Authorization: Bearer <new-api-dashboard-access-token-or-PAT>
```

删除绑定：

```http
DELETE /v1/admin/channel-affinities?user_id=1001&group=vip&model=gpt-4.1
Authorization: Bearer <new-api-dashboard-access-token-or-PAT>
```

`channel_id`、渠道状态、渠道的分组/模型能力由 new-api 在真实选路前继续校验；本服务不会绕过这些保护。

## 挂载到现有 Switcher 域名

Switcher 已有的 `ReverseProxy.TargetBaseAddress` 可直接作为本扩展的 `NEW_API_BASE_URL`，用于复用管理员身份校验。但 Switcher 的现有反向代理不会自动把新路径转发到扩展进程；若要通过同一域名访问，需要在部署层增加一条路径转发，并保留 `Authorization` 请求头。例如将：

```text
/api/extensions/channel-affinity/v1/admin/*
```

转发为扩展进程的：

```text
http://127.0.0.1:8089/v1/admin/*
```

这是网关/反向代理配置，不修改 new-api 或 Switcher 的源码。若只在内网运维调用，也可直接访问扩展监听端口。

## 运维命令

```powershell
go test ./extensions/channel-affinity-admin/...
go build ./extensions/channel-affinity-admin/cmd/affinity-admin
docker build -f extensions/channel-affinity-admin/Dockerfile -t new-api-affinity-admin:local .
```

升级 new-api 前后，请按 [docs/UPGRADE_CHECKLIST.md](docs/UPGRADE_CHECKLIST.md) 做协议核验。
