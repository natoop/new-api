# ZZDH Model Catalogue Snapshot

- Snapshot date: 20260806
- Fetched at (UTC): 2026-08-05T17:41:17.4212170Z
- Source: [https://zizidonghua.com/api/api-docs/models](https://zizidonghua.com/api/api-docs/models)
- Total models: 149
- Models tagged ``openai-video``: 86
- Raw catalogue: ``ZZDH_MODEL_CATALOG_20260806.json``
- Contract-detail batch: ``ZZDH_MODEL_DETAILS_20260806.json``

This file is an evidence snapshot, not runtime configuration. A model is not production-enabled merely because it appears here.

## non-video / audio or music adapter (15)

| Model | Catalog endpoint | Confirmed or expected task paths | Catalogue note |
| --- | --- | --- | --- |
| eleven_flash_v2 | audio |  |  |
| eleven_flash_v2_5 | audio |  |  |
| eleven_monolingual_v1 | audio |  |  |
| eleven_multilingual_v1 | audio |  |  |
| eleven_multilingual_v2 | audio |  |  |
| eleven_music_v1 | music |  |  |
| eleven_music_v2 | music |  |  |
| eleven_turbo_v2 | audio |  |  |
| eleven_turbo_v2_5 | audio |  |  |
| eleven_v3 | audio |  |  |
| indextts2-v1 | audio |  |  |
| music-2.6 | music |  |  |
| music-2.6-free | music |  |  |
| music-cover | music |  |  |
| music-cover-free | music |  |  |

## non-video / embedding adapter (1)

| Model | Catalog endpoint | Confirmed or expected task paths | Catalogue note |
| --- | --- | --- | --- |
| tongyi-embedding-vision-plus-2026-03-06 | embeddings |  |  |

## non-video / image adapter (12)

| Model | Catalog endpoint | Confirmed or expected task paths | Catalogue note |
| --- | --- | --- | --- |
| 双子生图 | image-generation |  |  |
| 双子生图flash | image-generation |  |  |
| 字字-幻图 | image-generation |  |  |
| 字字-幻图-pro | image-generation |  |  |
| Omni-Image2 | image-generation |  | 通用图像生成模型，适合日常文生图与素材生成。 |
| Omni-Image2-pro | image-generation |  | 高规格图像生成模型，适合高清素材与精细构图。 |
| Omni-Nano | image-generation |  | 1K 图像生成模型，支持文生图与参考图生成。 |
| Omni-Nano-pro | image-generation |  | 默认 1K 的高质量图像生成模型，支持参考图生成。 |
| Omni-Nano-pro-1K | image-generation |  | 固定 1K 输出档，支持文生图与参考图生成。 |
| Omni-Nano-pro-2K | image-generation |  | 固定 2K 输出档，支持文生图与参考图生成。 |
| Omni-Nano-pro-vip-1K | image-generation |  | 固定 1K 输出档，按实际输入/输出 Token 用量计费，费用不固定。 |
| Omni-Nano-pro-vip-2K | image-generation |  | 固定 2K 输出档，按实际输入/输出 Token 用量计费，费用不固定。 |

## non-video / ordinary OpenAI adapter (31)

| Model | Catalog endpoint | Confirmed or expected task paths | Catalogue note |
| --- | --- | --- | --- |
| 字字小说-C | openai |  |  |
| 字字语言模型-O | openai |  |  |
| deepseek-v3.2 | openai |  | DeepSeek-V3.2是引入DeepSeek Sparse Attention（一种稀疏注意力机制）的正式版模型，也是DeepSeek推出的首个将思考融入工具使用的模型，同时支持思考模式与非思考模式的工具调用。 |
| deepseek-v4-flash | openai |  |  |
| deepseek-v4-pro | openai |  |  |
| eleven_text_to_sound_v2 | openai |  |  |
| glm-5 | openai |  |  |
| glm-5.1 | openai |  |  |
| kimi-k2.5 | openai |  | kimi-k2.5是月之暗面迄今发布最全能的模型，原生多模态架构设计，同时支持视觉与文本输入、思考与非思考模式、对话与Agent任务。 |
| kimi-k2.7-code | openai |  |  |
| kimi-k3 | openai |  |  |
| MiniMax-M2.1 | openai |  | MiniMax-M2.1是MiniMax推出的旗舰级开源大模型，聚焦真实世界复杂任务，以多语言编程与长链 Agent 能力为核心优势。 |
| MiniMax-M2.5 | openai |  | MiniMax-M2.5是MiniMax推出的旗舰级开源大模型，经过数十万个真实复杂环境中的大规模强化学习训练，M2.5 在编程、工具调用和搜索、办公等生产力场景都达到或者刷新了行业的 SOTA。 |
| Moonshot-Kimi-K2-Instruct | openai |  | Kimi-K2是月之暗面提供的国内首个开源万亿参数MoE模型，激活参数达 320 亿，具有卓越的编码和工具调用能力。 |
| qwen-flash-character | openai |  | 千问系列多语言角色扮演模型，适合拟人化的角色扮演，同时优化了限定人设指令遵循、话题推进、倾听共情等能力，支持个性化角色的深度还原。 |
| qwen-image-2.0 | openai |  |  |
| qwen-image-2.0-pro | openai |  |  |
| qwen-image-edit-max | openai |  |  |
| qwen-image-max | openai |  | 千问图像生成模型Max系列，在各类生成任务中表现出色，相较Plus系列大幅度降低生成图片的AI感，提升图像真实性；具备更真实的人物质感、更细腻的自然纹理、更美观的文字渲染。 |
| qwen-vl-max | openai |  | qwen-vl-max 是由阿里云百炼提供的人工智能模型，千问VL-Max（qwen-vl-max），即千问超大规模视觉语言模型。相比增强版，再次提升视觉推理能力和指令遵循能力，提供更高的视觉感知和认知水平。在更多复杂任务上提供最佳的性能。 |
| qwen-vl-ocr | openai |  | 千问VL-OCR（qwen-vl-ocr），即基于Qwen-VL训练的OCR识别大模型。通过统一模型的方式聚合多种图文识别、解析、处理类任务，提供强大的图文识别能力。 |
| qwen3-vl-plus | openai |  |  |
| qwen3.5-plus | openai |  | Qwen3.5 Plus 是由阿里云百炼提供的人工智能模型。 |
| qwen3.6-plus | openai |  |  |
| qwen3.7-max | openai |  |  |
| qwen3.7-plus | openai |  |  |
| qwen3.8-max | openai |  |  |
| qwq-plus | openai |  | 千问QwQ推理模型增强版，基于Qwen2.5模型训练的QwQ推理模型，通过强化学习大幅度提升了模型推理能力。模型数学代码等核心指标（AIME 24/25、livecodebench）以及部分通用指标（IFEval、LiveBench等）达到DeepSeek-R1 满血版水平。 |
| xiaomi/mimo-v2.5-pro | openai |  |  |
| z-image-turbo | openai |  | Z Image Turbo 是由 阿里云百炼提供的人工智能模型。 |
| zimage | openai |  | 由国家超算平台提供的z-image |

## openai-video metadata anomaly / image candidate (3)

| Model | Catalog endpoint | Confirmed or expected task paths | Catalogue note |
| --- | --- | --- | --- |
| wan2.6-image | openai-video |  |  |
| wan2.6-t2i | openai-video |  | 阿里通义千问的生图模型 |
| wan2.7-image | openai-video |  |  |

## V1 generation / output-seconds billing (10)

| Model | Catalog endpoint | Confirmed or expected task paths | Catalogue note |
| --- | --- | --- | --- |
| doubao-seedance-2-0-fast-480p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | fast 档（速度更快）。无视频参考模型，计费方式为 (生成视频时长)x【每秒价格】，视频参考模型请调用-video |
| doubao-seedance-2-0-fast-720p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | fast 档（速度更快）。无视频参考模型，计费方式为 (生成视频时长)x【每秒价格】，视频参考模型请调用-video |
| doubao-seedance-2-0-mini-480p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | mini 档（价格更省）。无视频参考模型，计费方式为 (生成视频时长)x【每秒价格】，视频参考模型请调用-video |
| doubao-seedance-2-0-mini-720p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | mini 档（价格更省）。无视频参考模型，计费方式为 (生成视频时长)x【每秒价格】，视频参考模型请调用-video |
| doubao-seedance-2-1080p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 无视频参考模型，计费方式为 (生成视频时长)x【每秒价格】，视频参考模型请调用-video |
| doubao-seedance-2-480p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 无视频参考模型，计费方式为 (生成视频时长)x【每秒价格】，视频参考模型请调用-video |
| doubao-seedance-2-4k | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 4K 分辨率直出。无视频参考模型，计费方式为 (生成视频时长)x【每秒价格】，视频参考模型请调用-video |
| doubao-seedance-2-720p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 无视频参考模型，计费方式为 (生成视频时长)x【每秒价格】，视频参考模型请调用-video |
| doubao-seedance-2-优惠版-1080p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 优惠版（价格更低）。无视频参考模型，计费方式为 (生成视频时长)x【每秒价格】，视频参考模型请调用-video |
| doubao-seedance-2-优惠版-720p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 优惠版（价格更低）。无视频参考模型，计费方式为 (生成视频时长)x【每秒价格】，视频参考模型请调用-video |

## V1 generation / reference-video billing (10)

| Model | Catalog endpoint | Confirmed or expected task paths | Catalogue note |
| --- | --- | --- | --- |
| doubao-seedance-2-0-fast-video-480p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | fast 档（速度更快）。视频参考模型，计费方式为 (参考视频时长+生成视频时长)x【每秒价格】 |
| doubao-seedance-2-0-fast-video-720p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | fast 档（速度更快）。视频参考模型，计费方式为 (参考视频时长+生成视频时长)x【每秒价格】 |
| doubao-seedance-2-0-mini-video-480p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | mini 档（价格更省）。视频参考模型，计费方式为 (参考视频时长+生成视频时长)x【每秒价格】 |
| doubao-seedance-2-0-mini-video-720p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | mini 档（价格更省）。视频参考模型，计费方式为 (参考视频时长+生成视频时长)x【每秒价格】 |
| doubao-seedance-2-video-1080p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频参考模型，计费方式为 (参考视频时长+生成视频时长)x【每秒价格】 |
| doubao-seedance-2-video-480p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频参考模型，计费方式为 (参考视频时长+生成视频时长)x【每秒价格】 |
| doubao-seedance-2-video-4k | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 4K 分辨率直出。视频参考模型，计费方式为 (参考视频时长+生成视频时长)x【每秒价格】 |
| doubao-seedance-2-video-720p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频参考模型，计费方式为 (参考视频时长+生成视频时长)x【每秒价格】 |
| doubao-seedance-2-video-优惠版-1080p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 优惠版（价格更低）。视频参考模型，计费方式为 (参考视频时长+生成视频时长)x【每秒价格】 |
| doubao-seedance-2-video-优惠版-720p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 优惠版（价格更低）。视频参考模型，计费方式为 (参考视频时长+生成视频时长)x【每秒价格】 |

## V1 upscale / input-seconds billing (19)

| Model | Catalog endpoint | Confirmed or expected task paths | Catalogue note |
| --- | --- | --- | --- |
| 字字动画视频超分-1080p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 1080p（极速档）：保守增强、速度最快，适合批量处理；输入 ≤30fps、≤600 秒。 |
| 字字动画视频超分-1080p-标准 | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 1080p（标准档）：细节与对比度比极速档更强；输入 ≤30fps、≤300 秒。 |
| 字字动画视频超分-1080p-高级 | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 1080p（高级档）：大模型重绘、精修感最强；输入 ≤30fps、≤300 秒。 |
| 字字动画视频超分-1080p-高帧率 | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 1080p（极速-高帧率档）：支持 30~60fps 高帧率输入；≤600 秒。 |
| 字字动画视频超分-1080p-专业 | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 1080p（专业档）：细节增强最强、纹理锐利；输入 ≤30fps、≤300 秒。 |
| 字字动画视频超分-2k | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 2K（极速档）：保守增强、速度最快，适合批量处理；输入 ≤30fps、≤300 秒。 |
| 字字动画视频超分-2k-标准 | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 2K（标准档）：细节与对比度比极速档更强；输入 ≤30fps、≤300 秒。 |
| 字字动画视频超分-2k-高级 | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 2K（高级档）：大模型重绘、精修感最强；输入 ≤30fps、≤300 秒。 |
| 字字动画视频超分-2k-高帧率 | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 2K（极速-高帧率档）：支持 30~60fps 高帧率输入；≤300 秒。 |
| 字字动画视频超分-2k-专业 | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 2K（专业档）：细节增强最强、纹理锐利；输入 ≤30fps、≤120 秒。 |
| 字字动画视频超分-4k | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 4K（极速档）：保守增强、速度最快，适合批量处理；输入 ≤30fps、≤300 秒。 |
| 字字动画视频超分-4k-标准 | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 4K（标准档）：细节与对比度比极速档更强；输入 ≤30fps、≤300 秒。 |
| 字字动画视频超分-4k-高帧率 | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 4K（极速-高帧率档）：支持 30~60fps 高帧率输入；≤300 秒。 |
| 字字动画视频超分-4k-专业 | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 4K（专业档）：细节增强最强、纹理锐利；输入 ≤30fps、≤120 秒。 |
| 字字动画视频超分-720p | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 720p（极速档）：保守增强、速度最快，适合批量处理；输入 ≤30fps、≤600 秒。 |
| 字字动画视频超分-720p-标准 | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 720p（标准档）：细节与对比度比极速档更强；输入 ≤30fps、≤300 秒。 |
| 字字动画视频超分-720p-高级 | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 720p（高级档）：大模型重绘、精修感最强；输入 ≤30fps、≤300 秒。 |
| 字字动画视频超分-720p-高帧率 | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 720p（极速-高帧率档）：支持 30~60fps 高帧率输入；≤600 秒。 |
| 字字动画视频超分-720p-专业 | openai-video | POST /v1/video/generations; GET /v1/videos/{task_id} | 视频超分至 720p（专业档）：细节增强最强、纹理锐利；输入 ≤30fps、≤300 秒。 |

## V8 generation / ambiguous generic variant (1)

| Model | Catalog endpoint | Confirmed or expected task paths | Catalogue note |
| --- | --- | --- | --- |
| kling-v3-omni | openai-video | POST /v8/videos/generations; GET /v8/videos/generations/{task_id} | Kling 3.0 Omni 多模态视频模型，支持参考输入；分辨率、参考模式和音频能力由具体模型变体决定。 |

## V8 generation / confirmed variant (8)

| Model | Catalog endpoint | Confirmed or expected task paths | Catalogue note |
| --- | --- | --- | --- |
| kling-3.0-omni-1080p-noref-audio | openai-video | POST /v8/videos/generations; GET /v8/videos/generations/{task_id} | Kling 3.0 Omni 多模态视频模型，支持参考输入；分辨率、参考模式和音频能力由具体模型变体决定。 |
| kling-3.0-omni-1080p-noref-mute | openai-video | POST /v8/videos/generations; GET /v8/videos/generations/{task_id} | Kling 3.0 Omni 多模态视频模型，支持参考输入；分辨率、参考模式和音频能力由具体模型变体决定。 |
| kling-3.0-omni-1080p-ref-audio | openai-video | POST /v8/videos/generations; GET /v8/videos/generations/{task_id} | Kling 3.0 Omni 多模态视频模型，支持参考输入；分辨率、参考模式和音频能力由具体模型变体决定。 |
| kling-3.0-omni-1080p-ref-mute | openai-video | POST /v8/videos/generations; GET /v8/videos/generations/{task_id} | Kling 3.0 Omni 多模态视频模型，支持参考输入；分辨率、参考模式和音频能力由具体模型变体决定。 |
| kling-3.0-omni-720p-noref-audio | openai-video | POST /v8/videos/generations; GET /v8/videos/generations/{task_id} | Kling 3.0 Omni 多模态视频模型，支持参考输入；分辨率、参考模式和音频能力由具体模型变体决定。 |
| kling-3.0-omni-720p-noref-mute | openai-video | POST /v8/videos/generations; GET /v8/videos/generations/{task_id} | Kling 3.0 Omni 多模态视频模型，支持参考输入；分辨率、参考模式和音频能力由具体模型变体决定。 |
| kling-3.0-omni-720p-ref-audio | openai-video | POST /v8/videos/generations; GET /v8/videos/generations/{task_id} | Kling 3.0 Omni 多模态视频模型，支持参考输入；分辨率、参考模式和音频能力由具体模型变体决定。 |
| kling-3.0-omni-720p-ref-mute | openai-video | POST /v8/videos/generations; GET /v8/videos/generations/{task_id} | Kling 3.0 Omni 多模态视频模型，支持参考输入；分辨率、参考模式和音频能力由具体模型变体决定。 |

## V8 generation / endpoint metadata conflict (4)

| Model | Catalog endpoint | Confirmed or expected task paths | Catalogue note |
| --- | --- | --- | --- |
| zzdh-Minimax-h3-1080p | openai | POST /v8/videos/generations; GET /v8/videos/generations/{task_id} | Minimax H3 视频生成模型（1080P 档），支持文生视频、首帧/首尾帧生视频、参考生视频，生成结果原生自带音轨。时长 5~15 秒可调，按秒计费：费用 =（生成视频秒数）×（每秒价格）。 |
| zzdh-Minimax-h3-2k | openai | POST /v8/videos/generations; GET /v8/videos/generations/{task_id} | Minimax H3 视频生成模型（2K 档），支持文生视频、首帧/首尾帧生视频、参考生视频，生成结果原生自带音轨。时长 5~15 秒可调，按秒计费：费用 =（生成视频秒数）×（每秒价格）。 |
| zzdh-Minimax-h3-480p | openai | POST /v8/videos/generations; GET /v8/videos/generations/{task_id} | Minimax H3 视频生成模型（480P 档），支持文生视频、首帧/首尾帧生视频、参考生视频，生成结果原生自带音轨。时长 5~15 秒可调，按秒计费：费用 =（生成视频秒数）×（每秒价格）。 |
| zzdh-Minimax-h3-720p | openai | POST /v8/videos/generations; GET /v8/videos/generations/{task_id} | Minimax H3 视频生成模型（720P 档），支持文生视频、首帧/首尾帧生视频、参考生视频，生成结果原生自带音轨。时长 5~15 秒可调，按秒计费：费用 =（生成视频秒数）×（每秒价格）。 |

## video candidate / generic documentation only (35)

| Model | Catalog endpoint | Confirmed or expected task paths | Catalogue note |
| --- | --- | --- | --- |
| 字字3D | openai-video |  |  |
| 字字3D-flash | openai-video |  |  |
| 字字动画去字幕 | openai-video |  |  |
| 字字动画去字幕-pro | openai-video |  |  |
| 字字世界模型 | openai-video |  |  |
| 字字世界模型pro | openai-video |  |  |
| happyhorse-1.0-i2v-1080p | openai-video |  |  |
| happyhorse-1.0-i2v-720p | openai-video |  |  |
| happyhorse-1.0-r2v-1080p | openai-video |  |  |
| happyhorse-1.0-r2v-720p | openai-video |  |  |
| happyhorse-1.0-t2v-1080p | openai-video |  |  |
| happyhorse-1.0-t2v-720p | openai-video |  |  |
| happyhorse-1.0-video-edit-1080p | openai-video |  |  |
| happyhorse-1.0-video-edit-720p | openai-video |  |  |
| vidu-q3-pro-1080p | openai-video |  | 按秒计费，即时生成模型 |
| vidu-q3-pro-1080p-offpeak | openai-video |  | 错峰模式，谨慎使用，会进入超长时间排队可能导致失败 |
| vidu-q3-pro-540p | openai-video |  |  |
| vidu-q3-pro-540p-offpeak | openai-video |  |  |
| vidu-q3-pro-720p | openai-video |  |  |
| vidu-q3-pro-720p-offpeak | openai-video |  |  |
| vidu-q3-turbo-1080p | openai-video |  |  |
| vidu-q3-turbo-1080p-offpeak | openai-video |  |  |
| vidu-q3-turbo-540p | openai-video |  |  |
| vidu-q3-turbo-540p-offpeak | openai-video |  |  |
| vidu-q3-turbo-720p | openai-video |  |  |
| vidu-q3-turbo-720p-offpeak | openai-video |  |  |
| wan2.2-i2v-plus | openai-video |  | wan2.2-i2v-plus 是由 aliyun-bailian 提供的人工智能模型。计费类型为按秒计费 480p：0.084元/秒 1080p：0.42元/秒 |
| wan2.2-kf2v-flash | openai-video |  | wan2.2-kf2v-flash 是由阿里云提供的首尾帧生成模型，计费类型为按秒计费，每秒价格同每次价格 480p 0.06元/秒 720p 0.12元/秒 1080p 0.288元/秒 |
| wan2.6-i2v | openai-video |  | 通义千问的wan2.6图生视频模型，计费类型为按秒计费 720p：0.36元/秒 1080p：0.6元/秒 |
| wan2.6-i2v-flash | openai-video |  | 万相2.6-图生视频-Flash，生成更快更高性价比。智能分镜调度支持多镜头叙事，多人稳定对话，更自然真实音色，最高支持15秒时长生成。计费类型为按秒计费 720p  0.18元/秒 1080p 0.3元/秒 |
| wan2.6-r2v | openai-video |  |  |
| wan2.7-i2v | openai-video |  |  |
| wan2.7-r2v | openai-video |  |  |
| wan2.7-t2v | openai-video |  |  |
| wan2.7-videoedit | openai-video |  |  |

