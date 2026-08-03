# new-api 升级核验清单

此项目将 new-api 依赖压缩为一份 Redis 协议和一条运行时规则。每次升级 new-api 后，在切流前完成以下核验：

1. `service/channel_affinity.go` 仍使用命名空间 `new-api:channel_affinity:v1`。
2. `GET /api/user/self` 仍接受 Switcher 使用的 Dashboard Access Token/PAT，并返回 `data.id`、`data.role`；`role >= 10` 仍表示管理员。
3. 键顺序仍是：可选 rule name、可选 model、可选 using group、最后是 affinity value；没有转义或哈希变化。
4. Redis 值仍是十进制渠道 ID，且 Channel Affinity 读取 Redis 而非仅进程内缓存。
5. `middleware/distributor.go` 仍在随机选路前调用 `GetPreferredChannelByAffinity`，并对渠道状态、路径、分组、模型能力做二次校验。
6. `controller/relay.go` 的重试流程仍尊重 `skip_retry_on_failure`；生产 `RetryTimes >= 1`。
7. `RecordChannelAffinity` 的 TTL 回写和 `switch_on_success` 语义未变化。
8. 既有 Codex/Claude 规则仍位于外部规则之前，且抽样验证两类请求没有被外部规则吞掉。
9. 使用专用测试用户写入 B，分别验证：B 成功、B 失败后 C 重试成功、删除绑定后常规选路。

如果第 1–7 条中任何一条改变，先暂停本服务写入，不要猜测新键格式；更新本扩展的 `docs/NEW_API_CONTRACT.md` 与 `internal/affinity.Store.Key` 后，增加对应单元测试并完成灰度验证。
