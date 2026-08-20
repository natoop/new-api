---
report_type: upstream-pending-main-audit
generated_at: <ISO-8601 time>
timezone: Asia/Shanghai
repository: <absolute path>
branch: <local integration branch>
tracking_branch: <tracking branch>
baseline_head: <last local merge head>
baseline_synced_upstream: <last upstream commit already merged>
upstream_tracking_snapshot: <fetched origin/main or upstream/main>
upstream_tracking_snapshot_fetched_at: <ISO-8601 time>
scope: 待合并上游范围 <基线>..<上游提交>；不包含本地未提交内容
decision: <暂不合并 / 修复后合并 / 可合并>
---

# <日期> main 待合并更新报告

## 1. 总体结论

- 范围：`<基线>..<上游提交>`；`<提交数量>` 个 first-parent 提交；`<文件数量>` 个文件；`+<新增>/-<删除>`。
- 部署与 API 合同结论：<一句明确结论>。
- 明确列出未修改或未检查的内容。

## 2. 上游功能与改动

| 范围 | 新行为 | 与当前分支的区别 | 主要源码路径 | 影响级别 |
|---|---|---|---|---|
| <范围> | <行为> | <差异> | `<路径>` | <高/中/低> |

后端协议/计费、内置 Web、Electron、CI、依赖和仅测试改动应分开说明。

## 3. Switcher 兼容性

| 项目 | 当前 Switcher 合同 | 上游影响 | 需要动作 | 优先级 |
|---|---|---|---|---|
| <项目> | <已核验路径或 API> | <影响> | <无需修改 / 实现 / 运维设置> | <P0-P3> |

每项必须归类为：合并前必需、启用功能前必需、可选体验跟进，或因 Switcher 替代内置 Web 而不适用。

## 4. 合并冲突与语义风险

| 文件或子系统 | 本地定制 | 上游改动 | 处理规则 | 风险 |
|---|---|---|---|---|
| <路径> | <本地行为> | <上游行为> | <两边保留 / 选择明确责任方> | <高/中/低> |

同时列出 Git 双方修改的文件，以及路径不同但职责重叠的语义冲突。

## 5. 部署与数据合同

- 数据库/Schema 迁移：<需要/不需要，附证据>。
- Docker/Compose/端口改动：<需要/不需要，附证据>。
- Session/Cookie/API Key 合同：<需要/不需要，附证据>。
- 启用前需核对的配置：<键和安全默认值>。

## 6. 验证计划

1. <合并或本地准备步骤>
2. <聚焦测试，以及适用时独立 relaykit 构建>
3. <Switcher 兼容性检查>
4. <运行时/部署检查>

## 7. 提交附录

| 提交 | 主题 | 分类 |
|---|---|---|
| `<哈希>` | <主题> | <行为 / UI / 依赖 / 测试> |
