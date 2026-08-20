---
name: audit-upstream-sync
description: Use when a Git branch has received one or more two-parent merges from a configured upstream branch and the user asks for a recent update summary, frontend/backend change intent, saved sync baselines, next-sync comparison, or exclusion of pending remote commits.
---

# Audit Upstream Sync

## Core Contract

Use Git ancestry, not author dates. Save the local result and merged upstream commit.

| Question | Range |
|---|---|
| Effect of sync merge `M` | `M^1..M` |
| Effect of outer carrier `C`, when nested | `C^1..C` |
| Newly accepted upstream content | `<old-upstream>..M^2` |
| Still pending remotely | `<saved-upstream>..<source-ref>` |

Use upstream `--first-parent` for the main table; put side-branch commits in an appendix.

## Workflow

1. Capture branch, tracking ref, HEAD, status, remotes, and timestamp; fetch the source ref.
2. Read the newest `saved_head` and `saved_synced_upstream` under `docs/sync-history/`.
3. Run `scripts/find-sync-merges.ps1`. It searches every merge newly reachable from HEAD, including syncs nested behind an outer reconciliation merge.
4. Exclude source-history merges. Accept exactly-two-parent external merges whose second parent is on the source first-parent line and advances the upstream anchor.
5. The script compares sync, upstream, and outer-carrier patch IDs. Explain every mismatch as conflict resolution, local adaptation, or additional carrier content.
6. Analyze backend, `web/default`, `web/classic`, and shared/operations changes. Separate behavior changes, reversions, and mechanical migrations.
7. Keep commits beyond the saved upstream anchor in “待下次同步”; never mix them into completed results.
8. Write a new dated report and save the new local/upstream anchors.

For fast-forward, rebase, squash, or cherry-pick synchronization, stop this merge workflow and reconstruct the range from reflog. Never invent `M^2`.

## Required Record

```yaml
generated_at:
timezone:
repository:
branch:
tracking_branch:
baseline_head:
baseline_synced_upstream:
saved_head:
saved_synced_upstream:
upstream_tracking_snapshot:
upstream_tracking_snapshot_fetched_at:
scope:
```

For a pending-main review that also assesses a separate frontend or middleware,
use `templates/pending-main-report.md`. It adds stable sections for behavior,
current-branch differences, Switcher compatibility, merge risk, and deployment
contracts. Keep claims about the separate frontend tied to its actual source
paths and classify bundled-Web work as not applicable when the deployment uses
another frontend.

## Next Sync

```powershell
$SourceRemote = 'upstream'
$SourceBranch = 'main'
$SourceRef = "$SourceRemote/$SourceBranch"
git fetch --prune $SourceRemote $SourceBranch
if ($LASTEXITCODE -ne 0) { throw "git fetch failed: $SourceRef" }

$Batches = @(
  & .\.agents\skills\audit-upstream-sync\scripts\find-sync-merges.ps1 `
    -OldHead '<saved_head>' `
    -OldUpstream '<saved_synced_upstream>' `
    -SourceRef $SourceRef `
    -Verbose
)

foreach ($Batch in $Batches) {
  git diff --stat $Batch.PreMergeHead $Batch.Merge
  git diff --name-status $Batch.PreMergeHead $Batch.Merge
  git log --first-parent --reverse --format='%H|%cI|%s' `
    "$($Batch.UpstreamFrom)..$($Batch.UpstreamTo)"
}

$Batches |
  Format-Table Merge, IntegrationMerge, SyncPatchMatches, IntegrationPatchMatches
```

Save current `HEAD` and the last batch's `UpstreamTo`; if no batch exists, keep the previous upstream anchor.

## Common Mistakes

- Using `--since ... --all`, which mixes refs and misses older newly introduced commits.
- Searching only local first-parent or `--ancestry-path`; a sync can sit on side history forked before `saved_head`.
- Accepting any ancestor of the source ref instead of requiring source first-parent membership.
- Accepting a merge already inside source history, such as “merge main into feature”.
- Treating octopus, upstream PR, feature, or outer reconciliation merges as sync batches.
- Treating Git exit codes greater than 1 as a normal “false” result.
- Saving only local HEAD, using the remote tip as completed, or including uncommitted files.
- Claiming a remote-tracking ref is live-current without a successful fetch timestamp.
