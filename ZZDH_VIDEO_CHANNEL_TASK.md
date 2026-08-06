# ZZDH Dedicated Channel and Async Task Integration

## Status and Authority

- Design accepted: 2026-08-04.
- Phase 1 code added: 2026-08-06, uncommitted.
- Documentation baseline refreshed: 2026-08-06.
- User-confirmed implementation scope: 2026-08-06, 47 video-endpoint model names and 4 image-output model names.
- Task navigation and change history: `ZZDH_TASK_INDEX.md`.

This document is the authoritative ZZDH task baseline. It defines what the provider documentation currently supports, what new-api implements now, and the scope that future work must preserve.

The live upstream model list is not itself runtime configuration. A model may be listed by ZZDH and still be disabled in new-api until its request contract, billing contract, and verification status are complete.

## Evidence Baseline: 2026-08-06

| Evidence | Count | File |
| --- | --- | --- |
| Upstream catalogue | 149 models | `ZZDH_MODEL_CATALOG_20260806.json` |
| Human adaptation matrix | 149 models | `ZZDH_MODEL_CATALOG_20260806.md` |
| Catalogue tag `openai-video` | 86 models | Same catalogue snapshot |
| Detail contracts captured | 52 models | `ZZDH_MODEL_DETAILS_20260806.json` |

Sources:

- `https://zizidonghua.com/api-docs/models`
- `https://zizidonghua.com/api/api-docs/models`
- `https://zizidonghua.com/api/api-docs/model/{model_name}`
- `https://zizidonghua.com/api-docs/video-generation`

When sources conflict, resolve them in this order:

1. Explicit model-specific detail override.
2. Model-specific request/query examples.
3. Catalogue endpoint map.
4. Tags and model name.

The 2026-08-06 evidence has known conflicts: the global `openai-video` map says V1, generic model examples often say V8, Seedance/upscaling detail overrides explicitly say V1, and Minimax H3 details say V8 despite a catalogue tag of ordinary `openai`.

## Complete Model Classification

All 149 models are listed individually in `ZZDH_MODEL_CATALOG_20260806.md`. The categories below are exhaustive and sum to 149.

| Category | Count | Adapter decision |
| --- | ---: | --- |
| Seedance V1 output-seconds generation | 10 | One V1 generation profile family; price base is per output second |
| Seedance V1 reference-video generation | 10 | Same V1 generation transport, but billing rule adds validated reference seconds |
| ZZDH V1 video upscaling | 19 | Same asynchronous V1 lifecycle, separate `video`/`bitrate` body and input-seconds billing |
| Explicit Kling V8 variants | 8 | Implemented target profiles |
| `kling-v3-omni` | 1 | V8 path known; enabled in the user-confirmed target after focused profile verification |
| Minimax H3 V8 variants | 4 | Detail override confirms V8 video task; enabled with H3-specific validation |
| Generic video candidates | 35 | Happyhorse, Vidu, Wan video/edit, 字字 3D/world/subtitle models; generic documentation is insufficient; disabled |
| `openai-video` metadata anomalies | 3 | Historical catalogue conflict; the three user-confirmed Wan names use the live V8 video contract |
| Audio/music models | 15 | Separate audio/music adaptors, not video tasks |
| Image-generation models | 12 | Separate image adaptor |
| Embedding model | 1 | Separate embedding adaptor |
| Ordinary `openai` models | 31 | Ordinary OpenAI relay; H3 metadata conflict is resolved by the explicit detail override |

## User-Confirmed Target Scope

The implementation scope now has two independent adapter tracks: 47 video-endpoint models and 4 image-output models. The original five supplied names below were explicitly removed because they are not present in the current evidence snapshot:

```text
doubao-seedance-2-video
happyhorse-1.0-i2v
happyhorse-1.0-t2v
wan2.6-r2v-flash
wan2.6-t2v
```

The remaining requested names are:

### Seedance (16)

```text
doubao-seedance-2-0-fast-480p
doubao-seedance-2-0-fast-720p
doubao-seedance-2-0-fast-video-480p
doubao-seedance-2-0-fast-video-720p
doubao-seedance-2-0-mini-480p
doubao-seedance-2-0-mini-720p
doubao-seedance-2-0-mini-video-480p
doubao-seedance-2-0-mini-video-720p
doubao-seedance-2-1080p
doubao-seedance-2-480p
doubao-seedance-2-4k
doubao-seedance-2-720p
doubao-seedance-2-video-1080p
doubao-seedance-2-video-480p
doubao-seedance-2-video-4k
doubao-seedance-2-video-720p
```

### Kling (9)

```text
kling-3.0-omni-1080p-noref-audio
kling-3.0-omni-1080p-noref-mute
kling-3.0-omni-1080p-ref-audio
kling-3.0-omni-1080p-ref-mute
kling-3.0-omni-720p-noref-audio
kling-3.0-omni-720p-noref-mute
kling-3.0-omni-720p-ref-audio
kling-3.0-omni-720p-ref-mute
kling-v3-omni
```

### Happyhorse (8)

```text
happyhorse-1.0-i2v-1080p
happyhorse-1.0-i2v-720p
happyhorse-1.0-r2v-1080p
happyhorse-1.0-r2v-720p
happyhorse-1.0-t2v-1080p
happyhorse-1.0-t2v-720p
happyhorse-1.0-video-edit-1080p
happyhorse-1.0-video-edit-720p
```

### Wan video endpoint (10)

```text
wan2.6-image
wan2.6-i2v
wan2.6-i2v-flash
wan2.6-r2v
wan2.6-t2i
wan2.7-image
wan2.7-i2v
wan2.7-r2v
wan2.7-t2v
wan2.7-videoedit
```

### Minimax H3 V8 (4)

The catalogue metadata labels these models as ordinary `openai`, but their explicit detail overrides define the authoritative asynchronous V8 contract. They use output-seconds billing and the H3-specific 5-15 second, fixed-24fps validation set.

```text
zzdh-Minimax-h3-480p
zzdh-Minimax-h3-720p
zzdh-Minimax-h3-1080p
zzdh-Minimax-h3-2k
```

### Image-output models (4)

These models are explicitly marked as image generation or image editing models. Their current detail pages advertise the ordinary `openai` endpoint (`POST /v1/chat/completions`), not the asynchronous video endpoint. They belong to the image-output track and must not enter the video task adaptor or video-duration billing rules.

```text
qwen-image-2.0
qwen-image-2.0-pro
qwen-image-edit-max
qwen-image-max
```

The names `wan2.6-image`, `wan2.6-t2i`, and `wan2.7-image` contain image-oriented labels, but their live model-detail contracts currently advertise `openai-video` with `POST /v8/videos/generations` and `GET /v8/videos/generations/{task_id}`. They therefore stay on the video endpoint track unless a later upstream contract adds an image endpoint. Image references or image-oriented model names do not by themselves create a second adapter track.

The live recheck found no target model whose detail contract advertises both `openai-video` and an image endpoint at the same time. The two tracks are currently separated by endpoint contract, not by whether a video model accepts image references.

## Protocol and Adaptation Matrix

| Profile family | Models | Submit | Query | Request shape | Billing metric | New-api status |
| --- | --- | --- | --- | --- | --- | --- |
| `v1_generation_output_seconds` | 8 confirmed non-`-video` Seedance models | `POST /v1/video/generations` | `GET /v1/videos/{task_id}` | `model`, `prompt`, optional 1-2 `images`, `duration`, V1 `metadata` | requested output seconds | Implemented |
| `v1_generation_reference_seconds` | 8 confirmed Seedance `-video` models | Same V1 paths | Same V1 paths | Requires public `metadata.reference_video`; multimodal content may be supplied in metadata | output seconds + validated reference-video seconds | Implemented |
| `v1_upscale_input_seconds` | 19 视频超分 models | `POST /v1/video/generations` | `GET /v1/videos/{task_id}` | `model`, public `video`, optional `bitrate`; no generic prompt/duration body | `ceil(input duration)`, minimum one second | Planned |
| `v8_kling_omni` | 9 confirmed Kling variants | `POST /v8/videos/generations` | `GET /v8/videos/generations/{task_id}` | top-level `resolution`, `aspect_ratio`, references, audio settings | output seconds; price configured per model | Implemented |
| `v8_happyhorse` | 8 confirmed Happyhorse variants | Same V8 paths | Same V8 paths | profile-selected image/video references, resolution and prompt fields | requested output seconds | Implemented |
| `v8_wan` | 10 confirmed Wan video-endpoint variants | Same V8 paths | Same V8 paths | profile-selected image/video references and prompt fields | requested output seconds | Implemented |
| `v8_minimax_h3` | 4 `zzdh-Minimax-h3-*` variants | `POST /v8/videos/generations` | `GET /v8/videos/generations/{task_id}` | 5-15 seconds, fixed 24fps, 480p/720p/1080p/2K, image/video/audio reference rules | requested output seconds | Implemented |
| `v8_generic_candidate` | 35 generic video candidates | Do not infer from generated examples | Do not infer | Model-specific verification required | Do not infer | Disabled |
| `image_output_openai` | 4 confirmed Qwen image-output models | `POST /v1/chat/completions` | Ordinary OpenAI response | Existing OpenAI request conversion and response handling | Existing model-price per-call semantics; configure price in the model pricing page | Implemented through ordinary OpenAI adaptor |

All video task families are asynchronous. Successful video results are obtained by polling to a terminal state; they are not synchronous streamed video responses. The image track is a separate image relay path and must not be routed through the video task adaptor.

## Current Code State

Implemented files:

- `constant/channel.go`: persisted `ChannelTypeZZDH = 61`, display name, default base URL.
- `common/endpoint_type.go`: ZZDH video names use `openai-video`; the four Qwen image names use ordinary `openai`.
- `relay/relay_adaptor.go`: ZZDH task adaptor registration.
- `relay/channel/task/zzdh/profile.go`: 47 confirmed video profiles (16 V1 Seedance, 31 V8 Kling/Happyhorse/Wan/Minimax H3).
- `relay/channel/task/zzdh/adaptor.go`: V1/V8 request/query conversion, H3-specific validation, status/result extraction, authorization, and OpenAI-video conversion.
- `relay/channel/task/zzdh/adaptor_test.go`: profile, protocol, reference-video billing, URL, polling, and response regression tests.
- `pkg/asynctaskbilling/`: provider-neutral, profile-fixed async-task calculator with decimal quota conversion and a serializable frozen snapshot.
- `setting/billing_setting/tiered_billing.go`: `async_task_expr` and `async_task_billing` option storage/validation; `controller/option.go` publishes sanitized profile metadata for the configuration UI.
- `relay/async_task_billing.go`, `relay/relay_task.go`, `model/task.go`, and `service/task_billing.go`: submit-time reservation, retry reuse, persisted billing context, terminal-success no-debit guard, and refund/consume audit context.
- `common/endpoint_type.go` and `common/api_type.go`: ZZDH endpoint/API classification; Qwen image names use ordinary OpenAI, all confirmed video names use OpenAI-video.
- `relay/relay_task.go` and `service/task_polling.go`: preserve the original model in task polling so V1/V8 query paths remain stable.
- `D:\code\goswtich\switcher\frontend\src\features\channels\constants.ts`: exposes persisted channel type 61 as `ZZDH Video` in the Switcher channel-creation selector.
- `D:\code\goswtich\switcher\frontend\src\features\channels\lib\channel-utils.ts`: assigns type 61 the existing generic provider icon so the selector and channel table render it consistently.

The current code intentionally does not implement generic V8 candidates, video upscaling, or other catalogue models without a verified contract. Minimax H3 is enabled only from its explicit V8 detail override, not from the conflicting catalogue `openai` tag. The four Qwen image-output models are not admitted to the task adaptor; their ordinary OpenAI endpoint is selected by channel/model endpoint classification and they reuse existing per-call model pricing. Do not enable an unverified catalogue model from its name alone.

## Accepted Pricing Architecture

Prices belong in new-api model pricing configuration. Provider adaptors must not hardcode numeric price values. `async_task_expr` is active for models explicitly selected in `billing_setting.billing_mode`; all other models retain the legacy path.

```text
final quota = sum(profile metric x configured term price)
            x quota_per_unit
            x group ratio
```

Examples:

```text
output seconds:                 output_seconds x configured output price
reference video generation:     output_seconds x output price + reference_seconds x reference price
video upscaling (future):       input_video_seconds x configured input price
```

The task profile declares the fixed term list, metric bounds and rounding choices. `pkg/asynctaskbilling` receives only those named metrics and operator coefficients, validates the full rule, converts the result safely, and freezes the rule version, price terms, metrics, group ratio, quota-per-unit and reservation. The ZZDH adaptor supplies no numeric price. Existing `ModelPrice x seconds` behavior stays intact only when a model has not selected `async_task_expr`.

## Phase 1 Compatibility Settlement Policy

The initial ZZDH rollout will use the existing new-api task-compatible timing for settlement, as explicitly accepted during requirement alignment:

1. Validate the request and extract every enabled task metric that is knowable before submission.
2. Calculate the configured task expression synchronously from those bounded metrics.
3. Pre-consume that calculated amount and submit the upstream task.
4. On accepted submission, settle the same amount immediately and persist it in the task billing context.
5. If the asynchronous task reaches a terminal failure, refund the persisted task quota exactly once.

This is a compatibility phase, not a claim that every provider's final upstream deduction has been reconciled. A model is eligible only when its billable metric is authoritative at submission. `PerCallBilling` is stored with the frozen snapshot so terminal success cannot add an unreserved positive delta. A later terminal-settlement mode may reserve a bounded maximum and refund a proven difference.

The page/configuration contract remains data-driven: the frontend sets values for Profile-approved terms and rounding modes, while the provider-local Profile owns the formula structure and metric extraction. No numeric upstream price is hardcoded in the adaptor.

Safety requirements:

- Bound all request- and media-derived quantities before they reach pricing.
- Pre-consume the maximum validated price before upstream submission.
- A terminal success may refund down to proven actual price, but must never create an unreserved additional debit.
- Preserve a frozen price/rule snapshot for each task; later configuration changes do not rewrite an existing task.
- Failed tasks refund exactly once through the existing terminal CAS and task refund flow.

## Model Addition Rule

Adding a model must not automatically mean adding Go code.

| Change | Required work |
| --- | --- |
| Price-only change | Update the active async term prices, or legacy `ModelPrice` only when the model remains in legacy mode; no adaptor change |
| New model matching an existing protocol, schema, validation set, and billing rule | Add model/profile configuration, all required async term prices, then run focused verification |
| New model with different request/query/status/result contract | Add or extend a provider-local profile family and converter |
| New billing metric or settlement semantics | Extend the provider-neutral task calculator once, add tests, then configure affected models |

Never enable a model from a name match alone. The profile must state protocol, endpoint kind, validation limits, status/result parser, and billing rule. A configured model price is necessary but not sufficient.

## Scope Boundary

Keep ZZDH provider behavior under `relay/channel/task/zzdh/` and use the existing public task route, task table, polling, CAS, billing session, and refund/audit path.

Do not modify `relaykit`, public video routers, generic task interfaces, database schema, unrelated provider adaptors, this repository's `web/` frontend, or Switcher backend settlement solely to add ZZDH support. Switcher frontend work, when later needed, belongs in `D:\code\goswtich\switcher\frontend`, is configuration/display only, and must not relay tasks or create a separate ledger.

## Required Verification Before Production Enablement

1. Configure every required async term price and select `async_task_expr`, or deliberately retain a legacy `ModelPrice` model.
2. Submit a valid task and poll to a terminal state.
3. Confirm output URL, status normalization, and public task response.
4. Submit an invalid request and verify rejection occurs before pre-consume.
5. Reconcile upstream balance, reserved quota, task billing snapshot, consume/refund logs, and user/token quota.
6. Force or observe a failed task and confirm an exactly-once full refund.
7. Run focused Go tests, `go vet` on changed packages, and `git diff --check`.

## Historical Verification

Current focused verification completed:

```text
go test ./pkg/asynctaskbilling ./relay/channel/task/zzdh ./relay ./setting/billing_setting
go test ./service -run "TestSettle_PerCallBilling|Test.*Refund" -count=1
go vet ./pkg/asynctaskbilling ./relay/channel/task/zzdh ./relay ./setting/billing_setting ./service
git diff --check
```

Broader Windows test execution may still have unrelated Redis/session, channel-affinity, missing `web/dist/index.html`, or test executable access-denied failures. Treat focused verification separately from those environmental failures. No real ZZDH submission/poll/failure-refund integration test has been run against the upstream service.
