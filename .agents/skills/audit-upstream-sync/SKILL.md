---
name: audit-upstream-sync
description: Use when a Git branch has been synchronized or merged from an upstream branch and the user asks for a recent update summary, frontend/backend change intent, a saved HEAD baseline, next-sync comparison, or exclusion of not-yet-synced remote commits.
---

# Audit Upstream Sync

## Overview

Anchor the audit to Git ancestry, not commit dates. Preserve both the local sync result and the upstream commit actually merged so the next audit can separate local branch effects from pure upstream changes.

## Range Contract

| Question | Range |
|---|---|
| What changed in the local branch during merge `M`? | `M^1..M` |
| What upstream content was newly accepted? | `<old-upstream>..M^2` |
| What is still pending remotely? | `<saved-upstream>..<remote-ref>`; exclude from the completed report |

Use `git log --first-parent` for the main change table. Put PR-internal and older side-branch commits from the full ancestry range in an appendix.

## Workflow

1. Capture the branch, tracking ref, current HEAD, worktree status, remotes, and audit timestamp.
2. Search `docs/sync-history/` for the newest `saved_head` and `saved_synced_upstream` anchors.
3. Validate saved anchors with `git merge-base --is-ancestor`. If history was rewritten, reconstruct the range from reflog before continuing.
4. Identify new first-parent synchronization merges after `saved_head`. For each two-parent merge `M`, record `M^1` as the pre-merge local head and `M^2` as the new upstream anchor.
5. Compare `M^1..M` with `<old-upstream>..M^2`. If their stable patch IDs differ, inspect conflict resolutions or local adaptations explicitly.
6. Group the net diff into `web/default`, `web/classic`, backend, and shared/config/operations. Read the relevant code and commit messages; distinguish functional changes from mechanical migrations.
7. Keep remote-tracking commits not reachable from the saved local HEAD in a clearly labeled “待下次同步” section.
8. Write `docs/sync-history/YYYY-MM-DD-<branch>.md` and save the new anchors.

If the sync was fast-forward, rebase, squash, or cherry-pick, use reflog pre/post OIDs for the local tree range and record the observed source tip. Do not invent a merge second parent.

## Required Record Fields

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
scope: only changes already present in saved_head
```

The report must include totals, sync batches, frontend analysis, backend analysis, cross-layer intent, risks/validation, excluded pending changes, and next-sync commands.

## Next Sync

```powershell
$OldHead = '<saved_head>'
$OldUpstream = '<saved_synced_upstream>'

git merge-base --is-ancestor $OldHead HEAD
git log --first-parent --merges --reverse --format='%H|%P|%cI|%s' "$OldHead..HEAD"

$Merge = '<new sync merge>'
$PreMergeHead = git rev-parse "$Merge^1"
$NewUpstream = git rev-parse "$Merge^2"
git diff --stat $PreMergeHead $Merge
git log --first-parent --reverse --format='%H|%cI|%s' "$OldUpstream..$NewUpstream"
git diff --stat "$OldUpstream..$NewUpstream"
```

## Common Mistakes

- Do not use `git log --since ... --all` as the report range; it mixes unrelated refs and misses older commits first introduced by a recent merge.
- Do not use the current remote tip as the completed report upper bound.
- Do not save only the local HEAD; also save the merged upstream anchor.
- Do not describe hundreds of import or style substitutions as independent features.
- Do not include uncommitted worktree files in the committed-tree audit.
- Do not claim a remote-tracking ref is live-current unless it was refreshed and the fetch time was recorded.
