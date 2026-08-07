# Development Task Index

## Purpose

This is the repository-level entry point for recent development work and
task-record navigation. It is based on the current `HEAD` history and the
current worktree as inspected on 2026-08-07; it is not a release note and does
not claim that every listed commit has been deployed.

Start a related follow-up by reading this file, the linked task record, the
current diff, and then the affected runtime path. Do not infer current
behavior from a historical task record alone.

## Task Documents

| Status | Task | Record | Scope and entry point |
| --- | --- | --- | --- |
| Active, implementation verification in progress | General fixed-metered billing refactor | `tasks/FIXED_METERED_BILLING_TASK.md` | Replaces active `async_task_expr` admission with a system-level mode. Backend/task/frontend implementation is present; the recorded release gates remain required before deployment. |
| Superseded for new configuration; retained for history | ZZDH async-video pricing | `tasks/ZZDH_ASYNC_VIDEO_PRICING_TASK.md` | Documents the former profile-term pricing and pending-task compatibility requirements. |
| Active protocol/operations baseline | ZZDH video channel | `tasks/ZZDH_VIDEO_CHANNEL_TASK.md` | Provider protocol, model admission, task lifecycle, and upstream-document evidence. |
| Active detailed index | ZZDH maintenance index | `tasks/ZZDH_TASK_INDEX.md` | Model snapshots, Apifox generation, paths, operational checks, and historical decisions. |

## Recent Development Summary

| Date | Area | Current branch evidence | Related record / follow-up boundary |
| --- | --- | --- | --- |
| 2026-08-07 | Fixed-metered billing | Uncommitted worktree refactor in progress. | See the active task record; do not release before all lifecycle, log, frontend, and test gates pass. |
| 2026-08-06 | ZZDH video channel and task pricing | `7427c6d11`, `48d4161d8`, `188c6a8b2`, `c06633cc2`. | See the three ZZDH task records. Current billing work supersedes only the active legacy pricing admission, not provider protocol behavior. |
| 2026-08-04 | Deployment hardening | `b9d40dab7`. | Deployment credentials and listener binding are outside the fixed-metered task boundary. |
| 2026-08-03 | Channel affinity administration | `341d6bde8`, `15a67d2c0`. | Separate extension/API behavior; do not mix with channel pricing or ZZDH routing. |
| 2026-08-01 / 2026-07-31 | Billing reliability and auto-group behavior | `cfaba1dd6`, `df43f8015`, `0ab020206`. | Preserve existing group-ratio and retry settlement behavior when changing billing selection. |
| 2026-07-27 onward | Upstream integration baseline | Current branch includes relaykit, HTTP transport, relay format, CI, and UI commits. | Inspect the current code and `git log` before changing any of these unrelated areas. |

## Mandatory File Change Inventory

Every task record must contain a `File Change Inventory` section before code
is treated as complete. The section records each added, modified, deleted,
moved, or intentionally untouched file that is material to the task.

For each entry record:

1. Repository-relative path and operation: `add`, `modify`, `delete`, `move`,
   or `retain`.
2. Responsibility and behavioral effect.
3. Compatibility/rollback impact, including persisted data, configuration,
   public API, task lifecycle, and frontend ownership where applicable.
4. Verification status. A changed file without passing verification remains
   `unverified`; it must not be described as released or complete.

When a later change alters the inventory, append a dated row to that task's
change history before merging or deploying. Do not remove past inventory rows:
mark them superseded and link the replacement decision. Generated or user-owned
untracked files must be listed as `retain` when they are deliberately outside
the task, so a cleanup or rollback cannot delete them by accident.

## Documentation Layout

```text
doc/
  DEVELOPMENT_TASK_INDEX.md    repository-level recent-work and task index
  tasks/
    FIXED_METERED_BILLING_TASK.md
    ZZDH_ASYNC_VIDEO_PRICING_TASK.md
    ZZDH_TASK_INDEX.md
    ZZDH_VIDEO_CHANNEL_TASK.md
```

Task documents are kept under `doc/tasks/`; root-level duplicates are not
maintained. Move or rename a record only with corresponding index/link updates.
