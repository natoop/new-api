# ZZDH Maintenance Index

## Purpose

This is the entry point for every follow-up ZZDH task. Read this file before changing ZZDH routing, model support, protocol conversion, billing, Switcher configuration, or the provider catalogue.

It records the accepted boundary, evidence snapshots, implementation state, and the exact locations that own each concern. It exists to prevent a future task from rediscovering the design, duplicating an adaptor, or silently widening the scope.

## Canonical Records

| Record | Purpose | Refresh rule |
| --- | --- | --- |
| `ZZDH_VIDEO_CHANNEL_TASK.md` | Authoritative task baseline, protocol matrix, billing contract, model-admission rules, and change boundary | Update when behavior or the documented upstream contract changes |
| `ZZDH_MODEL_CATALOG_20260806.md` | Human-readable classification of all 149 models in the 2026-08-06 upstream catalogue | Regenerate on catalogue refresh; do not edit classifications by hand |
| `ZZDH_MODEL_CATALOG_20260806.json` | Raw response from `GET https://zizidonghua.com/api/api-docs/models` | Evidence snapshot; immutable after capture |
| `ZZDH_MODEL_DETAILS_20260806.json` | Merged model-detail evidence for the 52 models whose details determine V1/V8, request shape, or billing | Evidence snapshot; immutable after capture |
| `ZZDH_TARGET_MODEL_MATRIX_20260806.md` | Live recheck of the 47 user-confirmed target names, endpoint tracks, paths, and explicit exclusions | Refresh with a new dated matrix after any upstream model-contract change |
| `ZZDH_TARGET_MODEL_DETAILS_20260806-vidu.json` | Follow-up live detail snapshot after Vidu Q3 admission; 59 model detail pages | Use as the source for Vidu Q3 contract refreshes; preserve the prior 47-model snapshot |
| `ZZDH_ASYNC_VIDEO_PRICING_TASK.md` | Implemented `async_task_expr` contract: profile-approved terms, frozen task snapshot, lifecycle, and Switcher Model Pricing UI | Read before changing ZZDH pricing, task-billing settings, or the Switcher pricing editor |
| `scripts/zzdh-catalog-snapshot.ps1` | Repeatable snapshot tool; stable model-name ordering prevents overlapping detail batches | Use for a new dated snapshot, never overwrite the historical one |

The dated files are evidence. Runtime configuration remains in new-api model pricing and the data-driven ZZDH profile registry; never treat a catalogue snapshot as live configuration.

## Current Implementation State

The implementation is present but uncommitted.

| Concern | Current owner | Current state |
| --- | --- | --- |
| Stable channel type | `constant/channel.go` | `ChannelTypeZZDH = 61` |
| Endpoint classification | `common/endpoint_type.go` | Confirmed video names use `openai-video`; the 4 Qwen image-output names use ordinary OpenAI |
| Adaptor factory | `relay/relay_adaptor.go` | ZZDH task adaptor registered |
| Provider request/query conversion | `relay/channel/task/zzdh/adaptor.go` | V1 and V8 submit/query paths selected by model profile |
| Model admission | `relay/channel/task/zzdh/profile.go` | 59 confirmed video profiles: 16 V1 Seedance and 43 V8 Kling/Happyhorse/Wan/Vidu/Minimax H3 |
| User-confirmed target scope | `ZZDH_VIDEO_CHANNEL_TASK.md` | 59 video-endpoint names plus 4 image-output names; five supplied names explicitly removed; Vidu Q3 follows its live V8 detail contract |
| Task lifecycle | Existing task table, polling, terminal CAS, failure refund | Reused; no new route or schema |
| Task pricing | `relay/async_task_billing.go`, `relay/relay_task.go` | `async_task_expr` calculates profile-approved named metrics before submission; unselected models retain legacy `ModelPrice x seconds` |
| Async video pricing mode | `pkg/asynctaskbilling`, `setting/billing_setting`, and `ZZDH_ASYNC_VIDEO_PRICING_TASK.md` | Operator terms live in `billing_setting.async_task_billing`; profile rule, prices, metrics and reserved quota are frozen per task |
| Initial settlement policy | Existing billing session plus task billing | Pre-consume and settle the frozen accepted quota after submit; terminal success skips a new debit; terminal failure refunds the persisted quota through existing CAS |
| Switcher frontend | `D:\code\goswtich\switcher\frontend` | Model Pricing exposes profile-bound `Async video`; it remains configuration/display only and never a second billing ledger |
| new-api `web/` | No ZZDH change | This repository's frontend is explicitly outside the ZZDH frontend scope |

## Hard Boundary

Allowed without an explicit scope decision:

- `constant/channel.go`
- `common/endpoint_type.go`
- `relay/relay_adaptor.go`
- `relay/channel/task/zzdh/`
- a provider-neutral task-pricing module when the accepted work is the independent calculator
- focused new-api model/channel pricing configuration and tests

Frontend rule: ZZDH-related frontend work, including channel-type configuration, display, and i18n, belongs exclusively in `D:\code\goswtich\switcher\frontend`. Do not implement that work in this repository's `web/` directory. Before changing the Switcher frontend, re-check its current API contract and its own repository conventions.

Do not modify these merely to add a ZZDH model:

- `relaykit/`
- public video routers or public API paths
- generic task interfaces, task schema, or existing provider adaptors
- database schema
- Switcher backend relay/settlement paths

Any exception must be recorded in the change ledger below with the concrete shared contract that makes it necessary.

## Ownership Map

| Question | First place to inspect |
| --- | --- |
| What does the upstream currently claim? | Dated catalogue/detail snapshots, then the live detail page linked from the catalogue |
| Is this model a video model at all? | `ZZDH_MODEL_CATALOG_<date>.md`; do not trust the upstream endpoint tag alone |
| Which V1/V8 request/query paths apply? | `ZZDH_VIDEO_CHANNEL_TASK.md` protocol matrix and the model's detail snapshot |
| Is the model actually enabled today? | `relay/channel/task/zzdh/profile.go` |
| Why was a request rejected? | `relay/channel/task/zzdh/adaptor.go` validation and the selected profile |
| Where is the task created, polled, and settled? | `relay/relay_task.go`, `service/task_polling.go`, `service/task_billing.go` |
| Where are configured task prices read? | `setting/billing_setting/tiered_billing.go` and `relay/async_task_billing.go`; legacy values remain in `relay/helper/price.go` |
| How will non-token task pricing work? | `pkg/asynctaskbilling` with provider-local profile metrics; do not add model-specific arithmetic to the adaptor |

## Follow-up Model Procedure

Use this sequence for every upstream model addition or change.

1. Refresh the upstream evidence with a new date. Run `scripts/zzdh-catalog-snapshot.ps1` once for the catalogue, then fetch contract details in sorted batches and merge them. Record the new snapshot date in `ZZDH_VIDEO_CHANNEL_TASK.md`.
2. Classify the model from its detail document, using this precedence: explicit detail override, model-specific detail examples, catalogue endpoint map, tags/name. A generated generic V8 example is not a production contract.
3. Select an existing profile family only when request shape, query path, status/result shape, and billing basis all match. Otherwise add a new profile family first.
4. For `async_task_expr`, configure every profile-required term in `billing_setting.async_task_billing`, then select the billing mode. For legacy models, configure `ModelPrice`. Prices are configuration, not Go constants.
5. Add only data for a model that fits an existing protocol, request schema, validation set, and billing rule. Code changes are justified only for a new protocol, media metric, request schema, result format, or billing unit.
6. Run a success task, a validation-negative task, a failure/refund task, and a provider-balance reconciliation before production enablement.
7. Append a dated entry to the change ledger and update the implementation-state table if behavior changed.

## Catalogue Refresh Commands

Use a new date stamp, for example `20260807`; preserve the previous snapshot.

```powershell
.\scripts\zzdh-catalog-snapshot.ps1 -Stamp 20260807

0,10,20,30,40,50 | ForEach-Object {
  .\scripts\zzdh-catalog-snapshot.ps1 `
    -Stamp 20260807 `
    -IncludeContractDetails `
    -DetailOffset $_ `
    -DetailLimit 10
}

.\scripts\zzdh-catalog-snapshot.ps1 -Stamp 20260807 -MergeContractDetailBatches
```

The script sorts contract models by name before batching. A failed detail request is kept with its error; retry its one-item offset, then merge again. The merged file prefers a successful response over a failed retry.

## Change Ledger

| Date | Change | Scope and evidence | Verification |
| --- | --- | --- | --- |
| 2026-08-04 | Dedicated ZZDH async-channel design accepted | Dedicated task adaptor; reuse lifecycle and keep Switcher as configuration only | Design record in Codex memory and `ZZDH_VIDEO_CHANNEL_TASK.md` |
| 2026-08-06 | ZZDH target adaptor completed | Type 61, endpoint classification, adaptor registration, 47 video profiles across V1/V8, ordinary OpenAI path for 4 Qwen image models | Focused tests/vet and `git diff --check` |
| 2026-08-06 | Minimax H3 V8 profiles added | Four `zzdh-Minimax-h3-*` models enabled from explicit detail overrides; fixed 24fps, 5-15 seconds, 2K tier, reference-role/count validation, data URL support; output-seconds billing reused | H3 adaptor regression tests, focused tests/vet, `git diff --check` |
| 2026-08-06 | Catalogue/task baseline recorded | 149 catalogue models, 86 `openai-video` tags, 52 contract-detail snapshots, explicit metadata conflicts | Snapshot generation and merged detail count check: 52 models, 0 fetch errors |
| 2026-08-06 | Frontend ownership clarified | Any ZZDH frontend work is in `D:\code\goswtich\switcher\frontend`; `new-api/web/` is excluded | User-confirmed scope alignment |
| 2026-08-06 | Switcher channel-type selector synchronized | Added type 61 `ZZDH Video`, static i18n discovery, and a regression test in `D:\code\goswtich\switcher\frontend`; it only controls channel configuration/display and creates no billing path | Focused Bun test, typecheck, targeted lint, and production build passed |
| 2026-08-06 | Async-video pricing completion implemented | Added provider-neutral calculator, `billing_setting.async_task_billing`, profile registry, submit-time pre-consume, frozen task snapshot, retry reuse, audit fields, and legacy compatibility | Focused Go tests and vet pass; real upstream reconciliation remains a production gate |
| 2026-08-06 | Switcher async-video pricing editor implemented | Added profile-bound mode, term inputs, rounding selector, badge/filter/copy/JSON round-trip, seven-locale i18n, and snapshot regression test; no Switcher backend ledger | Focused Bun test, typecheck, changed-file lint, i18n sync, and production build pass |
| 2026-08-06 | Live model-detail recheck corrected adapter tracks | Retain 47 video-endpoint names after removing `doubao-seedance-2-video`, `happyhorse-1.0-i2v`, `happyhorse-1.0-t2v`, `wan2.6-r2v-flash`, and `wan2.6-t2v`; keep `wan2.6-image`, `wan2.6-t2i`, and `wan2.7-image` on the V8 video endpoint because their live contracts advertise `/v8/videos/generations`; keep the four Qwen image models on the ordinary OpenAI image-output track | Live detail recheck at 2026-08-05T20:36:01Z; no target model advertised both endpoint types; full matrix in `ZZDH_TARGET_MODEL_MATRIX_20260806.md` |
| 2026-08-06 | Duration input hardening | Reject malformed or overflowed `duration`/`seconds` values before task pricing; prevents a 32-bit parse failure from falling back to the default 5-second charge | `go test ./relay/common` including `TestTaskDurationBounds` |
| 2026-08-06 | Apifox collection export reorganized | `ZZDH_VIDEO_API_APIFOX.postman_collection.json` imports as direct folders for the five model families plus query/download; all 47 model requests retain the same public submit URL. `scripts/zzdh-apifox-postman-collection.mjs` regenerates it from the OpenAPI examples | Generated collection has 7 top-level folders, 47 distinct model requests, and the expected submit/query/download URLs |
| 2026-08-06 | Apifox model-contract details completed | `scripts/zzdh-apifox-postman-collection.mjs [ZZDH_TARGET_MODEL_DETAILS_<date>.json]` regenerates the Postman/Apifox import from the 47-model source-detail snapshot; it defaults to `ZZDH_TARGET_MODEL_DETAILS_20260806.json`. The hierarchy is provider -> model -> valid generation mode, with 97 mode-specific requests. Every model folder links its source detail page and distinguishes upstream-declared limits from current local ZZDH validation; Seedance and H3 include all documented optional fields and mutually exclusive examples. The public image aliases `image`, `images`, and `input_reference` are documented for image-reference-capable models; the sparse Kling/Happyhorse/Wan pages explicitly avoid inventing undocumented upstream fields, and video-edit/r2v models include a separately labelled adapter-permitted image-reference example. All submission examples use only the public `/v1/video/generations` route; query/download remain separate user-token-protected folders. | Regenerated collection: 7 top-level folders, 47 model folders, 97 submission examples; 37 model folders carry an image-input example. Every submission has matching model body, explicit Bearer token, source link, parameter section, limit section, and public route; no upstream `/v8` path or automatic V1/V8 wording remains |
| 2026-08-06 | Vidu Q3 V8 basic contract admitted | Live recheck of all 12 `vidu-q3-*` detail pages: each consistently declares V8 submit/query with `model`, `prompt`, `resolution`, and `duration`, and no reference-media fields. Added a `vidu_q3_v8` output-seconds profile that rejects reference media; prices remain operator-configured. Captured a new immutable 59-model detail snapshot and regenerated the collection. Each Vidu Apifox model folder documents the `duration` / `output_seconds` metric, default model-price formula, async-task-expression term, failure refund, and missing-price rejection. | `go test ./relay/channel/task/zzdh`; generated import has 8 top-level folders, 59 model folders, 109 submission examples, and Vidu requests use only the public route |
| 2026-08-06 | Provider-neutral independent task calculator | `pkg/asynctaskbilling` evaluates profile-fixed named terms, uses decimal quota conversion, freezes the billing context, and leaves token `tiered_expr` unchanged | Focused package, relay, ZZDH, task-refund, and vet checks pass |
| 2026-08-06 | ZZDH task-query response parity fix | Both public query paths (`/v1/video/generations/{task_id}` and `/v1/videos/{task_id}`) now use the OpenAI-video converter. ZZDH responses preserve `model` and `task_id`; the pre-polling `NOT_START` task state is exposed as `queued` instead of `unknown`. | `go test ./relay/channel/task/zzdh ./relay`; route and converter regression tests added |

## Repeated-Workflow Check

The recurring, costly workflow identified here is upstream catalogue drift and model-contract verification. The snapshot script plus this index cover it. No separate skill or automation was created because this process is provider-specific, manually triggered by upstream model changes, and now has a deterministic command, evidence output, and stopping condition.
