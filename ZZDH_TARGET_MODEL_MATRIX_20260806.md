# ZZDH Target Model Matrix

## Authority

- Verification date: 2026-08-05T20:36:01Z.
- Source template: `https://zizidonghua.com/api/api-docs/model/{model_name}`.
- This matrix records the user-confirmed target scope and the live model-detail contracts checked after the original catalogue snapshot.
- The dated catalogue and detail snapshots remain evidence; this file records the later targeted verification and does not replace them.

## Classification Rules

1. Use an explicit model-detail override before generic generated examples.
2. Determine the adapter from the advertised endpoint contract, not from `image`, `t2i`, `i2v`, or `r2v` in the model name.
3. An image reference accepted by a video model is an input capability, not an image-output model.
4. A model is dual-track only when its detail contract explicitly advertises both endpoint types. None of the 51 target names did so in this verification.

## Scope Summary

| Track | Count | Contract | Lifecycle |
| --- | ---: | --- | --- |
| Seedance video | 16 | V1 override; implemented with V1 submit/query paths | Asynchronous submit and poll |
| Kling video | 9 | `openai-video`, V8 | Asynchronous submit and poll |
| Happyhorse video | 8 | `openai-video`, V8 | Asynchronous submit and poll |
| Wan video endpoint | 10 | `openai-video`, V8 | Asynchronous submit and poll |
| Minimax H3 video | 4 | Detail override resolves catalogue `openai` tag to V8 async video | Asynchronous submit and poll |
| Qwen image output | 4 | ordinary `openai`, `POST /v1/chat/completions` | Ordinary relay response; existing per-call model pricing; no video task poll |
| **Total** | **51** | 47 video-endpoint names plus 4 image-output names | Two adapter tracks |

## Video Models

### Seedance: V1 override / output seconds (16)

The explicit detail override documents `POST /v1/video/generations` and `GET /v1/videos/{task_id}`. The generated default examples still show `/v8/videos/generations`; this conflict must be resolved by focused upstream verification before production enablement.

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

Billing evidence:

- Non-`-video`: requested output seconds.
- `-video`: requested output seconds plus validated reference-video seconds.
- Numeric prices remain new-api model-price configuration, not adapter constants.

### Kling: V8 (9)

All nine live details advertise `POST /v8/videos/generations` and `GET /v8/videos/generations/{task_id}`.

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

### Happyhorse: V8 (8)

All eight live details advertise `POST /v8/videos/generations` and `GET /v8/videos/generations/{task_id}`.

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

### Wan: V8 video endpoint (10)

All ten live details advertise `openai-video`, `POST /v8/videos/generations`, and `GET /v8/videos/generations/{task_id}`. The image-oriented names below therefore remain on the video endpoint track; their names or tags do not create an image adapter route.

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

### Minimax H3: V8 async video (4)

The catalogue metadata labels these models as `openai`, but the explicit detail override defines the authoritative V8 task contract. Runtime uses the override and applies H3-specific validation: 5-15 seconds, fixed 24 FPS, model-locked resolution, reference-role/count rules, and output-seconds billing.

```text
zzdh-Minimax-h3-480p
zzdh-Minimax-h3-720p
zzdh-Minimax-h3-1080p
zzdh-Minimax-h3-2k
```

## Image-Output Models

These four live details advertise ordinary `openai`, with `POST /v1/chat/completions`. They are not video tasks and do not use video-duration billing or polling. They use the ordinary OpenAI adaptor and the existing per-call model pricing configuration.

```text
qwen-image-2.0
qwen-image-2.0-pro
qwen-image-edit-max
qwen-image-max
```

## Explicit Exclusions

The following five names were removed from the user-confirmed scope:

```text
doubao-seedance-2-video
happyhorse-1.0-i2v
happyhorse-1.0-t2v
wan2.6-r2v-flash
wan2.6-t2v
```

## Implementation Boundary

- Video names belong under the ZZDH asynchronous task adaptor and use the existing task lifecycle, billing reservation, polling, terminal CAS, and refund path.
- Image-output names use the ordinary OpenAI endpoint/adaptor decision. They must not be admitted into the video task adaptor merely because the channel is `ChannelTypeZZDH`.
- Same-family model additions should be profile/configuration data when protocol, request shape, result shape, and billing metrics match.
- The current runtime implements all 47 confirmed video profiles; unlisted catalogue models remain disabled until their contract is verified and a matching profile exists.
