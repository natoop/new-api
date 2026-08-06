# ZZDH Async Video Pricing

## Status

- Implementation status: implemented, uncommitted (2026-08-06).
- Billing runtime: `D:\code\goswtich\new-api`.
- Configuration UI only: `D:\code\goswtich\switcher\frontend`.
- Read with: `ZZDH_TASK_INDEX.md` and `ZZDH_VIDEO_CHANNEL_TASK.md`.

This record defines the current `async_task_expr` contract. It replaces the
earlier planned-only record. `ModelPrice x seconds` remains a historical
compatibility path for models that have not selected this mode; it is not the
async-video pricing model.

## Persistent Contract

No schema or public route was added. Existing `options` rows hold:

```json
{
  "billing_setting.billing_mode": {
    "doubao-seedance-2-video-720p": "async_task_expr"
  },
  "billing_setting.async_task_billing": {
    "doubao-seedance-2-video-720p": {
      "version": 1,
      "rounding": "none",
      "terms": {
        "output_seconds": 0.12,
        "reference_video_seconds": 0.04
      }
    }
  }
}
```

`AsyncTaskBillingProfiles` is a read-only option returned by new-api. It
contains only the profile version, permitted metric names and bounds, and
supported rounding choices. The Switcher editor can choose `Async video` only
for a matching profile, and can edit only those term prices. It does not store
or evaluate an arbitrary formula.

## Runtime Contract

1. `relay/channel/task/zzdh/profile.go` owns the fixed rule and bounded metric
   admission. Current rules use `output_seconds` and, for reference-video
   models, `reference_video_seconds`.
2. `pkg/asynctaskbilling` validates positive finite operator term prices,
   profile membership, metric bounds, group ratio, and quota conversion. It
   calculates `sum(metric * term_price) * quota_per_unit * group_ratio` with
   decimal arithmetic and fails closed on saturation.
3. `relay/relay_task.go` selects this path before legacy `ModelPrice` pricing,
   pre-consumes its reservation before upstream submission, then freezes rule
   version, term prices, metrics, group ratio, quota conversion and reserved
   quota on `RelayInfo`.
4. Automatic upstream retries reuse that frozen snapshot. A successfully
   submitted task persists it in
   `tasks.private_data.billing_context.async_task_billing` and is marked
   per-call billed, so terminal success cannot add an unreserved debit.
5. Existing terminal-state CAS and `RefundTaskQuota` refund the persisted task
   quota once on failure. Submit and refund audit entries include the frozen
   async billing context and quota-saturation marker.

The async path ignores a stale legacy `ModelPrice` value. Leaving that value in
configuration is intentional: switching a model back to a legacy mode remains
reversible and historical models keep their original behavior.

## Boundaries

- No migration, Switcher backend work, `relaykit` change, public route change,
  generic task-adaptor interface change, or new-api `web/` change.
- The provider adaptor supplies named metrics only. Numeric prices and formula
  evaluation remain outside provider transport code.
- A missing profile-required term, unsupported model/profile, invalid metric,
  invalid rounding, non-finite value, or saturated calculation rejects the
  request before upstream submission.

## Focused Verification

Completed locally:

```text
go test ./pkg/asynctaskbilling ./relay/channel/task/zzdh ./relay ./setting/billing_setting
go test ./service -run "TestSettle_PerCallBilling|Test.*Refund" -count=1
go vet ./pkg/asynctaskbilling ./relay/channel/task/zzdh ./relay ./setting/billing_setting ./service
bun test src/features/system-settings/models/__tests__/model-pricing-snapshots.test.ts
bun run typecheck
bunx oxlint -c .oxlintrc.json <changed model-pricing files>
bun run i18n:sync
bun run build
git diff --check
```

No real upstream ZZDH submit/poll/failure-refund run was made. Before production
enablement, use a non-production channel to verify one output-only and one
reference-video model, then reconcile request metrics, persisted snapshot,
pre-consume, task/consume/refund logs, user and token quota, and provider
balance.
