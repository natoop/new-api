# Fixed Metered Billing Refactor

## Status

Planned on 2026-08-07. This document is the implementation boundary and
change record for replacing the earlier ZZDH-specific asynchronous task
pricing rule. Repository-wide task navigation is
`doc/DEVELOPMENT_TASK_INDEX.md`.

Implementation is in progress. The active worktree has replaced the old
calculator and settings admission with the new provider-neutral path, but it
is not releasable until the lifecycle, frontend, and release-gate checks below
are recorded. No migration, configuration rollout, or production data change
has been performed.

## Execution Plan

| Phase | Scope | Completion Evidence | Status |
| --- | --- | --- | --- |
| 0. Baseline | Record requirements, boundary, old/new compatibility, and current worktree state. | This document and focused diff review. | Complete |
| 1. Billing core | Implement and test the provider-neutral calculator, settings validation, explicit mode selection, and frozen snapshot. | Package/unit tests plus affected Go tests. | In progress |
| 2. Request integration | Connect synchronous relay pricing and ZZDH metric extraction without changing the ZZDH protocol boundary. | Adaptor/request validation tests. | In progress |
| 3. Task settlement | Preserve pre-consume/refund semantics and move terminal-success consumption to the standard log flow exactly once. | Task lifecycle/CAS/log tests. | In progress |
| 4. Frontend | Change only `D:\\code\\goswtich\\switcher\\frontend` to configure the general mode, then synchronize all locales. | Typecheck, i18n sync, lint, build. | In progress |
| 5. Compatibility | Reject new legacy-mode submissions while preserving pending historical snapshots for settlement/refund. | Compatibility regression tests. | Pending |
| 6. Release gate | Run focused tests/builds and non-production provider verification before enabling a model. | Recorded command results and test task evidence. | Pending |

## Change Recording Protocol

Every implementation change for this task is recorded in the table below
before it is treated as complete. Each row must state the date, affected
boundary, decision, reason, verification result, and rollback/revisit point.

Rules:

1. Update this plan before beginning a new phase or changing a confirmed
   decision. Do not silently widen the scope from ZZDH/task billing into
   unrelated pricing or protocol code.
2. Add an `In progress` row before editing a phase; replace it with the exact
   result after tests/builds finish. Failed checks remain recorded rather than
   being removed.
3. Record file paths and configuration/data compatibility implications for any
   persisted contract change. Do not put credentials, user data, or production
   task payloads in this file.
4. Historical snapshots, deployment configuration, and production rows are
   never rewritten merely to match the new mode. Any future migration requires
   its own rollback plan and explicit authorization.
5. A follow-up starts by reading this document, then the current diff and the
   affected lifecycle code. The change history is the navigation point for
   avoiding duplicate implementations and boundary drift.

## Confirmed Requirements

1. The new mode is system-level. It is not limited to asynchronous tasks or
   ZZDH. Synchronous and asynchronous requests select the same pricing mode.
2. The mode owns its fixed price. It must not read or require `ModelPrice`.
   Administrators must not maintain the same price in two locations.
3. A configured model selects one explicit pricing mode. The gateway must not
   scan unrelated price maps and silently pick whichever value happens to be
   present.
4. Supported fixed-metered formulas are:

   ```text
   per_request:
     price x group_ratio

   duration_seconds:
     output_seconds x unit_price x group_ratio

   duration_plus_reference_video_seconds:
     round(output_seconds + reference_video_seconds)
       x unit_price x group_ratio
   ```

5. Rounding is configurable as `none` or `ceil_total_units`. Ceiling applies
   once to the total billable unit count, before multiplication by the unit
   price. It does not round each media item independently or round quota.
6. The price, mode, rounding, measured units, group ratio, quota-per-unit and
   calculated quota are frozen before an upstream request is sent.
7. Synchronous requests write their normal consume log after a successful
   response. Asynchronous tasks reserve before submission but write one formal
   consume log only after their terminal success; failures refund the reserve
   and write a refund log.
8. The formal task consume/refund logs must use the standard `logs` table and
   standard new-api log columns. Task-specific values belong in `other`,
   including public `task_id`.

## Configuration Contract

`billing_setting.billing_mode` selects `fixed_metered` for a model.
`billing_setting.fixed_metered_billing` stores the complete price and rule in
one entry:

```json
{
  "example-video-model": {
    "version": 1,
    "unit_price": 0.12,
    "usage_mode": "duration_plus_reference_video_seconds",
    "rounding": "ceil_total_units"
  }
}
```

Validation requirements:

- `version` is exactly `1`.
- `unit_price` is finite and non-negative. Zero is an explicit free price.
- `usage_mode` and `rounding` are closed enums.
- `per_request` only accepts `none` rounding.
- Duration-derived metrics must be finite, positive when required, and bounded
  before pricing. Missing required metrics reject the request; they never
  become zero by default.
- A model with `billing_mode=fixed_metered` but no valid entry is rejected. It
  never falls back to `ModelPrice`, `ModelRatio`, or an expression.

## Mode Selection Order

The requested public model name (`OriginModelName`) is the configuration key.
Channel model mapping changes only the upstream model name, never the user
price key.

```text
fixed_metered -> fixed_metered_billing only
tiered_expr  -> billing expression only
default       -> existing ModelPrice / ModelRatio compatibility path
```

The system must not select a price based on channel order, channel count, or
which legacy map contains a value. Channel routing and user pricing remain
separate concerns.

## Implementation Boundaries

### Backend

- Replace the active `async_task_expr` calculator/settings/admission path with
  a provider-neutral `fixedmeteredbilling` package.
- Provider adaptors expose bounded usage metrics only. They do not contain
  numeric prices or formula selection.
- ZZDH continues to own V1/V8 protocol selection, request validation, status
  parsing, and reference-video duration extraction under
  `relay/channel/task/zzdh/`.
- Existing generic fixed-price and token/ratio modes stay unchanged unless a
  model explicitly selects `fixed_metered`.
- Synchronous per-request usage can use the system mode directly. A duration
  mode requires the applicable adaptor to provide verified duration metrics.

### Task Lifecycle and Logs

1. Validate required fixed-metered metrics and calculate the frozen reserve.
2. Pre-consume the reserve before submitting the task.
3. Persist the task, upstream ID, frozen billing snapshot, and submission log
   context. Do not create a consume log at accepted submission.
4. On terminal success, use the existing task-status CAS to emit one standard
   consume record and update task usage counters.
5. On terminal failure or timeout, refund the persisted reserve and write a
   standard refund record. Do not emit a consume record.

The implementation must keep task logging idempotent across poll retries and
must preserve current multi-node task CAS behavior. Separate log databases are
an explicit verification case.

### Switcher Frontend

Frontend changes belong only in `D:\code\goswtich\switcher\frontend`.
The model pricing editor will replace the `Async video` mode with `Fixed
metered` and display a single unit-price input, billing-unit selector, and
rounding selector. It will not show provider profile terms or hardcoded model
lists. Saving this mode updates only:

- `billing_setting.billing_mode`
- `billing_setting.fixed_metered_billing`

It must remove the retired `async_task_expr` configuration for the selected
model and prevent conflicting legacy mode selections in the visual editor.
All visible copy is translated through the seven existing locale files.

## File Change Inventory

This inventory is the 2026-08-07 working-tree baseline, not an acceptance
statement. Every entry below is `unverified` until its owning phase and release
gate pass. New changes must append a dated inventory row and matching
change-history row before they are merged or deployed.

| Operation | Path | Responsibility and rollback/revisit impact | Verification |
| --- | --- | --- | --- |
| modify | `controller/option.go` | Stops exposing retired pricing-profile metadata and will expose only the general setting contract. Revisit with the frontend API shape. | Unverified |
| modify | `controller/relay.go` | Persists the frozen fixed-metered task context. Revisit with terminal consume-log context and historical tasks. | Unverified |
| modify | `model/option.go` | Changes option validation/admission for the pricing configuration. Revisit against database option compatibility. | Unverified |
| modify | `model/task.go` | Adds fixed-metered snapshot data and retains a raw legacy snapshot for historical refund/audit only. Revisit with serialized task compatibility. | Model compatibility test passed |
| modify | `model/task_cas_test.go` | Verifies an existing persisted legacy task snapshot remains readable as opaque JSON without restoring its calculator or configuration admission path. | Passed |
| delete | `relay/async_task_billing.go` | Removes the old active relay helper. It remains a pending deletion until all callers use the new path. | Unverified |
| delete | `relay/async_task_billing_test.go` | Removes obsolete behavior coverage; replacement tests must cover the new public billing contract. | Unverified |
| modify | `relay/common/relay_info.go` | Replaces relay request snapshot state. Revisit all relay/task callers before accepting. | Unverified |
| add | `relay/fixed_metered_billing.go` | General fixed-metered task preparation and reservation entry point. Revisit with exact-once settlement. | Unverified |
| add | `relay/fixed_metered_billing_test.go` | Regression coverage for task preparation. It must be expanded only for observable contracts. | Unverified |
| modify | `relay/helper/price.go` | Selects the explicit system-level billing mode before legacy pricing. Revisit the synchronous path and mode precedence. | Unverified |
| add | `relay/helper/fixed_metered.go` | Builds generic synchronous price data. Revisit all supported relay formats before claiming broad mode support. | Unverified |
| modify | `relay/relay_task.go` | Integrates task pricing preparation. Revisit submission order, pre-consume, and retry reuse. | Unverified |
| delete | `setting/billing_setting/async_task_billing_test.go` | Deletes tests for the retired setting. Replacement validation tests are required. | Unverified |
| modify | `setting/billing_setting/tiered_billing.go` | Defines fixed-metered configuration and removes active legacy setting validation. Revisit API/config serialization compatibility. | Unverified |
| add | `setting/billing_setting/fixed_metered_billing_test.go` | Tests new setting validation. Must pass with the affected Go suite. | Unverified |
| add | `pkg/fixedmeteredbilling/` | Provider-neutral calculator/config/snapshot package. Revisit the exact package file list when implementation is finalized. | Unverified |
| delete | `pkg/asynctaskbilling/` | Removes the retired profile-term calculator and its tests. Historical task JSON is retained as raw data in `model/task.go`, so this deletion does not reactivate or reimplement the old calculation. | Unverified |
| retain | `ZZDH_VIDEO_API_APIFOX.openapi.yaml` | User-owned untracked provider-document export; not part of this refactor and must not be removed by rollback/cleanup. | Not in scope |
| modify | `ZZDH_VIDEO_API_APIFOX.postman_collection.json` | Re-generated from the checked-in generator and existing ZZDH detail snapshot so its Vidu billing notes no longer advertise the retired mode. It remains an untracked deliverable and is not removed by rollback/cleanup. | Generated; content search passed |

### 2026-08-07 implementation inventory additions

| Operation | Path | Responsibility and rollback/revisit impact | Verification |
| --- | --- | --- | --- |
| modify | `model/log.go` | Adds standard task-log parameters so delayed terminal logs retain the ordinary log columns. Revisit with log database compatibility and task audit queries. | Affected package compile passed |
| modify | `service/task_billing.go` | Writes fixed-metered success logs only after terminal CAS and preserves opaque legacy task snapshots for historical refund/audit. | Targeted service tests passed |
| modify | `service/task_polling.go` | Calls the fixed-metered terminal success logger only for the CAS winner. | Affected package compile passed |
| modify | `relay/channel/task/zzdh/adaptor.go` | Exposes bounded output/reference-video metrics without provider-owned prices. | ZZDH tests passed |
| modify | `relay/channel/task/zzdh/profile.go` | Removes old pricing rule/profile-term registration while retaining protocol and capability constraints. | ZZDH tests passed |
| modify | `relay/channel/task/zzdh/adaptor_test.go` | Replaces pricing-rule expectations with fixed-metered metric coverage. | ZZDH tests passed |
| modify | `scripts/zzdh-apifox-postman-collection.mjs` | Documents the fixed-metered contract in generated Vidu request folders. | Generator completed |
| modify | `controller/relay.go` | Skips submit-time fixed-metered consume logs, persists terminal-log context, and refunds a reservation when task persistence fails. | Affected package compile passed |
| modify | `D:\\code\\goswtich\\switcher\\frontend\\src\\features\\system-settings\\models\\__tests__\\model-pricing-snapshots.test.ts` | Replaces old profile-term expectations with regression coverage for fixed-metered snapshots without `ModelPrice`. | Bun test passed |
| modify | `D:\\code\\goswtich\\switcher\\frontend\\src\\features\\system-settings\\models\\model-pricing-core.ts` | Replaces async profile terms with the one-price/usage-unit/rounding contract. | Typecheck and build passed |
| modify | `D:\\code\\goswtich\\switcher\\frontend\\src\\features\\system-settings\\models\\model-pricing-snapshots.ts` | Displays fixed-metered models and readable billing-unit labels. | Bun test passed |
| modify | `D:\\code\\goswtich\\switcher\\frontend\\src\\features\\system-settings\\models\\model-pricing-sheet.tsx` | Replaces the provider-gated async tab with the general fixed-metered editor. | Typecheck and build passed |
| modify | `D:\\code\\goswtich\\switcher\\frontend\\src\\features\\system-settings\\models\\model-ratio-visual-editor.tsx` | Saves fixed-metered configuration and removes conflicts with legacy price maps. | Typecheck and build passed |
| modify | `D:\\code\\goswtich\\switcher\\frontend\\src\\features\\system-settings\\models\\model-ratio-form.tsx` | Adds the JSON-mode configuration field to the model-pricing form. | Typecheck and build passed |
| modify | `D:\\code\\goswtich\\switcher\\frontend\\src\\features\\system-settings\\models\\model-ratio-table-columns.tsx` | Renders the general pricing-mode badge. | Typecheck and build passed |
| modify | `D:\\code\\goswtich\\switcher\\frontend\\src\\features\\system-settings\\models\\ratio-settings-card.tsx` | Stores the new setting in the model-pricing settings card. | Typecheck and build passed |
| modify | `D:\\code\\goswtich\\switcher\\frontend\\src\\features\\system-settings\\models\\index.tsx` | Initializes fixed-metered settings on the model settings page. | Typecheck and build passed |
| modify | `D:\\code\\goswtich\\switcher\\frontend\\src\\features\\system-settings\\billing\\index.tsx` | Includes the setting in the billing section defaults. | Typecheck and build passed |
| modify | `D:\\code\\goswtich\\switcher\\frontend\\src\\features\\system-settings\\billing\\section-registry.tsx` | Registers fixed-metered configuration in the system settings section. | Typecheck and build passed |
| modify | `D:\\code\\goswtich\\switcher\\frontend\\src\\features\\system-settings\\types.ts` | Defines the persisted option key in the Switcher frontend contract. | Typecheck and build passed |
| modify | `D:\\code\\goswtich\\switcher\\frontend\\src\\features\\models\\components\\drawers\\model-mutate-drawer.tsx` | Provides the default fixed-metered option when mutating model settings. | Typecheck and build passed |
| modify | `D:\\code\\goswtich\\switcher\\frontend\\src\\i18n\\static-keys.ts` | Removes retired async-video strings and registers fixed-metered UI text. | i18n sync passed |
| modify | `D:\\code\\goswtich\\switcher\\frontend\\src\\i18n\\locales\\{en,zh,zh-TW,fr,ja,ru,vi}.json` | Adds all fixed-metered labels through the sanctioned locale script and removes retired keys. | i18n sync: 0 missing / 0 extras |
| move | `ZZDH_ASYNC_VIDEO_PRICING_TASK.md`, `ZZDH_TASK_INDEX.md`, `ZZDH_VIDEO_CHANNEL_TASK.md` -> `doc/tasks/` | Consolidates historical and active task navigation under `doc/`; root duplicates are intentionally removed. | Links inspected |
| modify | `doc/tasks/ZZDH_VIDEO_CHANNEL_TASK.md`, `doc/tasks/ZZDH_TASK_INDEX.md` | Records that `/v1/video/generations` is the only public ZZDH submission route; V1/V8 describe only internal ZZDH upstream forwarding selected by the persisted original model profile. | Documentation reviewed against current router and adaptor paths |

## Retirement and Historical Compatibility

- New requests must not admit `async_task_expr` or read
  `billing_setting.async_task_billing`.
- The Switcher editor must not render or save those entries.
- No automatic conversion is safe when an old model had different output and
  reference-video prices. Administrators must explicitly choose one new unit
  price and a fixed-metered usage mode.
- Persisted tasks submitted under the old mode retain their frozen snapshot so
  that already-pending tasks can still complete or refund without a changed
  price. This is data compatibility only, not a retained active pricing rule.

## Required Tests and Release Gates

1. Fixed per-request pricing works without `ModelPrice` or `ModelRatio`.
2. Output-only seconds pricing preserves fractional seconds with `none`.
3. Reference-video pricing sums all billable reference durations and applies
   `ceil_total_units` once to the sum.
4. Missing fixed config, missing price, invalid enum, invalid metric, and
   overflow all fail before upstream submission.
5. Changed configuration after submission does not change the frozen task
   charge.
6. Synchronous success writes the standard consume format once.
7. Async success writes one standard consume log with public `task_id`; async
   failure writes one refund and no consume log.
8. Terminal CAS races do not duplicate debit, refund, usage counters, or logs.
9. Existing `async_task_expr` task snapshots remain readable for pending
   historical tasks, while newly submitted requests reject the retired mode.
10. Run affected Go tests, root build, Switcher typecheck/lint/build, locale
    sync, and a non-production ZZDH output-only plus reference-video task.

## Change History

| Date | Boundary | Decision and result | Reason / verification / revisit point |
| --- | --- | --- | --- |
| 2026-08-07 | Task governance | Added phased execution plan and mandatory change-record protocol before continuing implementation. | Establishes the required handoff/navigation baseline. No code verification or rollout has occurred; review current scaffolding in Phase 1 before accepting it. |
| 2026-08-07 | Documentation | Moved task records to `doc/tasks/` and added the repository-level index plus a working-tree file inventory. | Future rollback and related work must use the inventory and keep it updated. Documentation-only change; runtime verification is unchanged. |
| 2026-08-07 | Billing-core implementation | Began a dependency and compatibility audit before completing removal of `async_task_expr`. | Do not delete legacy package/snapshot data until runtime callers, task settlement, profile registration, tests, and historical pending-task handling have been verified. |
| 2026-08-07 | Pricing configuration | Replaced the prior proposal to reuse `ModelPrice`. | A fixed-metered price must be configured once, inside its own mode. |
| 2026-08-07 | Pricing scope | Generalized the mode beyond async/ZZDH. | Billing mode describes pricing, not transport timing. |
| 2026-08-07 | Legacy mode | Retired active `async_task_expr` term pricing. | Separate output/reference prices and provider-gated UI no longer match the confirmed single-price rule. |
| 2026-08-07 | Historical data | Retained historical snapshots only. | Pending old tasks must not be repriced or lose refund compatibility. |
| 2026-08-07 | Historical data regression | Added a serializer regression test for the raw legacy task snapshot. | It proves old persisted task data can still be read for refund/audit while the old calculator remains deleted and inactive. |
| 2026-08-07 | Implementation resumption | Reviewed the active diff against the confirmed boundary before continuing backend/task/frontend work. | The old executable configuration and calculator are deleted; only an opaque legacy task snapshot remains for already-persisted refunds and audit. Lifecycle logging and frontend cleanup require focused verification before release. |
| 2026-08-07 | Task-log safety | Added an explicit reservation refund when a fixed-metered upstream task succeeds but its local task row fails to insert. | Without the row, terminal polling cannot produce a consume/refund record; the BillingSession refund prevents an untracked reserve. A database-failure integration test remains a release-gate follow-up. |
| 2026-08-07 | Frontend and locale cleanup | Replaced the provider-specific async-video editor with fixed price, billing unit, and rounding controls; synchronized all seven locales and removed retired UI text. | `bun test` for snapshots, `bun run typecheck`, `bun run i18n:sync`, and `bun run build` passed. |
| 2026-08-07 | API collection | Re-generated the Postman collection from `ZZDH_TARGET_MODEL_DETAILS_20260806-vidu.json` after updating the generator's billing notes. | The collection contains 59 model folders and no longer contains `async_task_expr` or `async_task_billing`. |
| 2026-08-07 | Verification limits | `go test ./... -run '^$'` could not compile the root package because `web/dist/index.html` is absent from this worktree. | All affected Go packages compiled; fixed-metered, settings, relay, ZZDH, and targeted service tests passed. Full frontend lint remains blocked by pre-existing violations in unrelated files; production build passed. |
| 2026-08-07 | API-layer terminology | Clarified the single public ZZDH V1 route versus internal model-selected ZZDH V1/V8 forwarding in the protocol baseline and maintenance index. | Prevents client/API documentation from exposing an internal V8 forwarding path; no runtime code changed. |
