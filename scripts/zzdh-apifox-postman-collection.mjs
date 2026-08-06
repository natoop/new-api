const openAPIPath = 'ZZDH_VIDEO_API_APIFOX.openapi.yaml';
const detailSnapshotPath = process.argv[2] || 'ZZDH_TARGET_MODEL_DETAILS_20260806.json';
const outputPath = 'ZZDH_VIDEO_API_APIFOX.postman_collection.json';

const submitURL = '{{base_url}}/v1/video/generations';
const sourceImage = '{{image_url}}';
const sourceVideo = '{{video_url}}';
const sourceAudio = '{{audio_url}}';
const aspectRatios = '`16:9` / `9:16` / `1:1` / `4:3` / `3:4` / `21:9`';

const openAPI = Bun.YAML.parse(await Bun.file(openAPIPath).text());
const expectedModels = openAPI?.components?.schemas?.ZZDHVideoModel?.enum;
const detailSnapshot = await Bun.file(detailSnapshotPath).json();

if (!Array.isArray(expectedModels) || expectedModels.length === 0) {
  throw new Error(`Unable to read ZZDH model enum from ${openAPIPath}`);
}
if (!Array.isArray(detailSnapshot?.data)) {
  throw new Error(`Unable to read ZZDH model details from ${detailSnapshotPath}`);
}

const detailByModel = new Map(detailSnapshot.data.map((entry) => [entry.model_name, entry]));
const missingDetails = expectedModels.filter((model) => !detailByModel.has(model));
const unexpectedDetails = [...detailByModel.keys()].filter((model) => !expectedModels.includes(model));
if (missingDetails.length > 0 || unexpectedDetails.length > 0) {
  throw new Error([
    missingDetails.length > 0 ? `Missing source details: ${missingDetails.join(', ')}` : '',
    unexpectedDetails.length > 0 ? `Unexpected source details: ${unexpectedDetails.join(', ')}` : '',
  ].filter(Boolean).join('; '));
}

const folders = [
  { name: '豆包 Seedance', matches: (model) => model.startsWith('doubao-seedance-') },
  { name: '可灵 Kling', matches: (model) => model.startsWith('kling-') },
  { name: 'Happyhorse', matches: (model) => model.startsWith('happyhorse-') },
  { name: 'Wan', matches: (model) => model.startsWith('wan') },
  { name: 'Minimax H3', matches: (model) => model.startsWith('zzdh-Minimax-h3-') },
];

function markdown(lines) {
  return lines.filter((line) => line !== undefined && line !== null && line !== '').join('\n\n');
}

function table(rows) {
  return [
    '| 字段 | 类型 | 必填 | 说明 |',
    '| --- | --- | --- | --- |',
    ...rows.map((row) => `| ${row.join(' | ')} |`),
  ].join('\n');
}

function submitHeaders() {
  return [
    { key: 'Authorization', value: 'Bearer {{api_token}}', type: 'text' },
    { key: 'Content-Type', value: 'application/json', type: 'text' },
    { key: 'Accept', value: 'application/json', type: 'text' },
  ];
}

function createSubmitRequest(name, body, description) {
  return {
    name,
    request: {
      method: 'POST',
      header: submitHeaders(),
      body: {
        mode: 'raw',
        raw: JSON.stringify(body, null, 2),
        options: { raw: { language: 'json' } },
      },
      url: submitURL,
      description,
    },
    response: [],
  };
}

function sourceEvidence(entry) {
  const tags = entry.detail?.tags ? `；上游标签：${entry.detail.tags}` : '';
  return markdown([
    '## 原始文档依据',
    `详情页：${entry.detail_url}${tags}`,
    `本集合基于 ${detailSnapshotPath} 中 ${detailSnapshot.fetched_at} 抓取的详情页内容整理。以下“本地校验”仅描述 new-api 当前 ZZDH 适配器；它不等同于上游新增的能力承诺。`,
  ]);
}

function publicRouteNote() {
  return markdown([
    '## 公共调用地址',
    `提交：\`POST ${submitURL}\`。所有示例均使用这个公共地址，认证为 \`Authorization: Bearer {{api_token}}\`。`,
    '请求成功后取得 `task_id`，请使用集合中的“查询任务”和“下载任务”目录；两者都必须使用创建任务的同一用户 API Token。',
  ]);
}

function publicImageAliasRows() {
  return [
    ['`image`', 'string', '否', '公共兼容字段；单张图片会映射为首帧。仅适用于允许图片参考的模型。'],
    ['`images`', 'string[]', '否', '公共兼容字段；图片会按首帧/尾帧顺序转换。具体模型的原始数量限制仍以该目录说明为准。'],
    ['`input_reference`', 'string', '否', '公共兼容字段；在没有 `metadata.reference_images` 时作为首帧图片。'],
  ];
}

function standardTaskFields(model, durationDescription = '整数秒；省略时默认 `5`；本地校验为 `1~3600`。') {
  return table([
    ['`model`', 'string', '是', `固定为 \`${model}\`。`],
    ['`prompt`', 'string', '是', '提示词；不能只传参考素材。'],
    ['`duration`', 'integer', '否', durationDescription],
  ]);
}

function seedanceStandardDescription(entry) {
  const model = entry.model_name;
  return markdown([
    `# ${model}`,
    sourceEvidence(entry),
    publicRouteNote(),
    '## 参数',
    table([
      ['`model`', 'string', '是', `固定为 \`${model}\`。`],
      ['`prompt`', 'string', '是', '提示词。'],
      ['`images`', 'string[]', '否', '原始文档支持 0、1、2 张公网图片：1 张为首帧，2 张依次为首帧和尾帧。不要传参考视频。'],
      ['`image` / `input_reference`', 'string', '否', '公共单图兼容字段；没有 `images` 时映射为首帧。'],
      ['`duration`', 'integer', '否', '输出秒数；默认 `5`；本地校验 `1~3600`。原始页面未给出更窄的时长范围。'],
      ['`metadata.ratio`', 'string', '否', `画幅；原始文档给出 ${aspectRatios}，默认 \`16:9\`。`],
      ['`metadata.seed`', 'integer', '否', '随机种子；本地校验必须为非负数。'],
      ['`metadata.generate_audio`', 'boolean', '否', '是否生成音频。'],
      ['`metadata.watermark`', 'boolean', '否', '水印开关，按原始 Seedance 文档透传。'],
      ['`metadata.callback_url`', 'string', '否', '回调地址，按原始 Seedance 文档透传。'],
      ['`metadata.content`', 'array', '否', '多模态数组；原始文档列出 `text`、`image_url`、`audio_url` 与可选 `role`。示例见“多模态与控制项”。'],
    ]),
    '## 限制与互斥规则',
    '- 模型名已固定分辨率；不要依赖请求中的 `resolution` 改变交付档位。',
    '- 该类模型不接受参考视频；需要参考视频时选择同档位带 `-video-` 的模型。',
    '- 图片、音频和回调地址都应使用可由上游访问的 URL。原始文档未声明额外的大小/数量上限，除 `images` 的 0/1/2 模式外不要自行推断。',
  ]);
}

function seedanceReferenceDescription(entry) {
  const model = entry.model_name;
  return markdown([
    `# ${model}`,
    sourceEvidence(entry),
    publicRouteNote(),
    '## 参数',
    table([
      ['`model`', 'string', '是', `固定为 \`${model}\`。`],
      ['`prompt`', 'string', '是', '提示词；使用 `metadata.content` 时仍保留非空提示词。'],
      ['`duration`', 'integer', '否', '输出秒数；默认 `5`；本地校验 `1~3600`。'],
      ['`image` / `images` / `input_reference`', 'string / string[]', '否', '公共图片兼容字段；只能作为参考视频之外的附加图片，不能替代必需的参考视频。'],
      ['`metadata.reference_video`', 'string', '二选一', '公网参考视频 URL。当前适配器将其纳入参考视频输入，并读取时长用于计费。'],
      ['`metadata.content`', 'array', '二选一', '可用 `video_url` 且 `role: "reference_video"` 传入参考视频；还可组合 `text`、`image_url`、`audio_url`。'],
      ['`metadata.ratio`', 'string', '否', `画幅；原始文档给出 ${aspectRatios}。`],
      ['`metadata.seed`', 'integer', '否', '随机种子；本地校验必须为非负数。'],
      ['`metadata.generate_audio`', 'boolean', '否', '是否生成音频。'],
      ['`metadata.watermark`', 'boolean', '否', '水印开关，按原始文档透传。'],
      ['`metadata.callback_url`', 'string', '否', '回调地址，按原始文档透传。'],
    ]),
    '## 限制与互斥规则',
    '- 必须至少提供一个参考视频：`metadata.reference_video` 或 `metadata.content` 中的 `video_url`。',
    '- 参考视频必须是绝对 HTTP(S) URL。本地还会读取其时长，时长必须大于 0 且不超过 3600 秒。',
    '- 模型名固定分辨率；不要依赖请求里的 `resolution` 改变交付档位。',
  ]);
}

function klingVariant(model) {
  if (model === 'kling-v3-omni') {
    return { tier: '兼容别名', references: '由实际参考输入决定', audio: '由 `metadata.generate_audio` 决定', explicit: false };
  }
  return {
    tier: model.includes('-1080p-') ? '1080P' : '720P',
    references: model.includes('-noref-') ? '禁止任何参考输入' : '至少一个参考输入',
    audio: model.endsWith('-audio') ? '必须生成音频' : '必须静音',
    explicit: true,
  };
}

function klingDescription(entry) {
  const model = entry.model_name;
  const variant = klingVariant(model);
  const referenceField = variant.references === '禁止任何参考输入'
    ? '不允许发送。'
    : variant.references === '至少一个参考输入'
      ? '至少从这三类中选择一种；具体数量上限未在原始详情页声明。'
      : '可选；原始详情页未定义具体限制。';
  const audioField = variant.audio === '必须生成音频'
    ? '可省略，适配器默认补为 `true`；显式传入只能为 `true`。'
    : variant.audio === '必须静音'
      ? '可省略，适配器默认补为 `false`；显式传入只能为 `false`。'
      : '可选布尔值；建议优先使用带 `audio` 或 `mute` 后缀的显式模型。';
  const limits = [
    `- 原始文档指定分辨率档位：${variant.tier}。${variant.explicit ? '模型名锁定档位；示例不传 `resolution`。' : '该别名未锁定档位，原始文档建议改用显式变体。'}`,
    `- 参考规则：${variant.references}。`,
    `- 音频规则：${variant.audio}。`,
    '- 原始详情页未提供 prompt 长度、时长档位或参考素材数量的完整参数表；除以下适配器字段外，不要从上游最小 cURL 样例推断额外参数。',
  ];
  return markdown([
    `# ${model}`,
    sourceEvidence(entry),
    publicRouteNote(),
    '## 参数',
    table([
      ['`model`', 'string', '是', `固定为 \`${model}\`。`],
      ['`prompt`', 'string', '是', '提示词。'],
      ['`duration`', 'integer', '否', '默认 `5`；本地校验 `1~3600`。原始详情页没有声明更窄范围。'],
      ['`metadata.generate_audio`', 'boolean', '否', audioField],
      ['`metadata.reference_images`', 'array', '按模型', referenceField],
      ['`metadata.reference_videos`', 'array', '按模型', referenceField],
      ['`metadata.reference_audios`', 'array', '按模型', referenceField],
      ['`metadata.content`', 'array', '按模型', '可用 `image_url` / `video_url` / `audio_url` 表示参考素材；参考变体至少包含一种素材。'],
      ['`metadata.aspect_ratio` 或 `metadata.ratio`', 'string', '否', `适配器接受 ${aspectRatios}。原始 Kling 详情页未列出其完整取值表。`],
      ['`metadata.seed`', 'integer', '否', '适配器接受非负整数；原始 Kling 详情页未进一步说明。'],
      ...(variant.references === '禁止任何参考输入' ? [] : publicImageAliasRows()),
    ]),
    '## 限制与互斥规则',
    ...limits,
  ]);
}

function sparseModelDescription(entry, family) {
  const model = entry.model_name;
  const isHappyhorse = family === 'happyhorse';
  let mode;
  let rows;
  let restrictions;

  if (isHappyhorse && model.includes('-i2v-')) {
    mode = '首帧生视频';
    rows = [['`metadata.reference_images`', 'array', '是', '至少一张参考图片。原始页标签为“首帧生视频”；本地适配器据此要求图片参考。']];
    restrictions = ['必须含至少一张参考图片。'];
  } else if (isHappyhorse && model.includes('-r2v-')) {
    mode = '参考图片生视频';
    rows = [['`metadata.reference_images`', 'array', '是', '至少一张参考图片。原始页标签为“参考图片生视频”；本地适配器据此要求图片参考。']];
    restrictions = ['必须含至少一张参考图片。'];
  } else if (isHappyhorse && model.includes('-t2v-')) {
    mode = '文生视频';
    rows = [['参考素材字段', '-', '否', '不允许发送 `reference_images`、`reference_videos`、`reference_audios` 或 `content` 中的媒体素材。']];
    restrictions = ['本地适配器拒绝任何参考输入。'];
  } else if (isHappyhorse && model.includes('-video-edit-')) {
    mode = '视频编辑';
    rows = [['`metadata.reference_images` / `reference_videos` / `reference_audios`', 'array', '是', '至少一种参考素材。本地适配器只验证“存在参考输入”；原始页面未声明必须是视频或各类素材上限。']];
    restrictions = ['本地适配器要求至少一个参考输入；示例使用参考视频。'];
  } else if (!isHappyhorse && model.includes('-i2v')) {
    mode = '图生视频';
    rows = [['`metadata.reference_images`', 'array', '是', '至少一张参考图片；这是当前本地适配器校验。原始页面没有完整字段表。']];
    restrictions = ['本地适配器要求至少一张参考图片。'];
  } else if (!isHappyhorse && (model.includes('-r2v') || model.includes('videoedit'))) {
    mode = '参考/编辑视频';
    rows = [['`metadata.reference_images` / `reference_videos` / `reference_audios`', 'array', '是', '至少一种参考素材；这是当前本地适配器校验。原始页面没有完整字段表。']];
    restrictions = ['本地适配器要求至少一个参考输入。'];
  } else if (!isHappyhorse && model.includes('-t2v')) {
    mode = '文生视频';
    rows = [['参考素材字段', '-', '否', '不允许发送参考输入；这是当前本地适配器校验。']];
    restrictions = ['本地适配器拒绝任何参考输入。'];
  } else {
    mode = entry.detail?.tags || '通用模型';
    rows = [['`metadata.aspect_ratio` 或 `metadata.ratio`', 'string', '否', `适配器接受 ${aspectRatios}；原始页面未提供完整参数表。`]];
    restrictions = ['原始页面只提供最小调用样例，未明确参考素材、时长档位或扩展参数限制；不要把该样例扩展为上游保证。'];
  }

  const flashLimit = model === 'wan2.6-i2v-flash'
    ? '原始文档明确“最高支持 15 秒”。示例使用 5 秒；当前本地通用校验仍为 1~3600 秒，超过 15 秒是否接受由上游决定。'
    : '原始页面未声明更窄的时长范围。';
  return markdown([
    `# ${model}`,
    sourceEvidence(entry),
    publicRouteNote(),
    '## 原始文档覆盖范围',
    `原始详情页将其标为“${mode}”，但除最小提交样例外没有给出完整参数表。下方字段分为明确的公共字段与当前本地适配器校验，避免将未文档化参数误写为上游能力。`,
    '## 参数',
    table([
      ['`model`', 'string', '是', `固定为 \`${model}\`。`],
      ['`prompt`', 'string', '是', '提示词。'],
      ['`duration`', 'integer', '否', `默认 \`5\`；本地校验 \`1~3600\`。${flashLimit}`],
      ['`metadata.aspect_ratio` 或 `metadata.ratio`', 'string', '否', `适配器接受 ${aspectRatios}；原始页面未给出完整枚举。`],
      ['`metadata.seed`', 'integer', '否', '适配器接受非负整数；原始页面未进一步说明。'],
      ...(rows.some((row) => row[0].includes('reference_images')) ? publicImageAliasRows() : []),
      ...rows,
    ]),
    '## 限制与互斥规则',
    ...restrictions.map((restriction) => `- ${restriction}`),
    '- 参考素材使用 `metadata.reference_images`、`metadata.reference_videos`、`metadata.reference_audios`，每个元素为 `{ "url": "https://...", "role": "..." }`；URL 必须是绝对 HTTP(S) 地址。',
    '- 原始详情页未说明的字段，不在本集合中伪造为可用参数。',
  ]);
}

function h3Description(entry) {
  const model = entry.model_name;
  return markdown([
    `# ${model}`,
    sourceEvidence(entry),
    publicRouteNote(),
    '## 参数',
    table([
      ['`model`', 'string', '是', `固定为 \`${model}\`；分辨率由模型名锁定。`],
      ['`prompt`', 'string', '是', '最长 10000 字符。'],
      ['`duration`', 'integer', '否', '默认 `5`；必须为 `5~15` 秒。'],
      ['`metadata.aspect_ratio`', 'string', '否', `默认 \`16:9\`；支持 ${aspectRatios}。`],
      ['`metadata.fps`', 'integer', '否', '固定 `24`；省略时适配器补为 `24`，传其它值会被拒绝。'],
      ['`metadata.seed`', 'integer', '否', '非负整数；省略时随机。'],
      ...publicImageAliasRows(),
      ['`metadata.reference_images`', 'array', '按模式', '最多 9 张；每项为 `{ "url": "...", "role": "first_frame" | "last_frame" | "reference_image" }`。'],
      ['`metadata.reference_videos`', 'array', '否', '最多 3 条；仅参考生模式；每项 `role` 只能省略或为 `reference_video`。'],
      ['`metadata.reference_audios`', 'array', '否', '最多 3 条；仅参考生模式；每项 `role` 只能省略或为 `reference_audio`。'],
      ['`metadata.extra.reference_video_audio`', 'boolean', '否', '默认 `true`；是否把参考视频中的音轨纳入生成条件。'],
      ['`metadata.negative_prompt`', 'string', '不支持', '原始文档明确不支持；不要发送。'],
    ]),
    '## 限制与互斥规则',
    '- 素材 URL 支持公网 HTTP(S) 或 `data:` Base64。上游文档的大小限制：图片 20MB、视频 200MB、音频 50MB。',
    '- 不传参考素材为文生视频。1~2 张无 `role` 图片为首帧/首尾帧；`reference_image`、参考视频或参考音频为参考生模式。',
    '- 首尾帧不能与 `reference_image`、参考视频或参考音频混用；`last_frame` 必须与 `first_frame` 同时存在。',
    '- 如果图片数组中有一个元素填写了 `role`，其余图片也应填写 `role`。参考视频超过 15 秒的部分按原始文档会被截断。',
  ]);
}

function seedanceStandardExamples(model) {
  return [
    createSubmitRequest('文生视频', { model, prompt: '清晨的海边，电影级自然光，镜头缓慢推进。', duration: 5, metadata: { ratio: '16:9' } }, '不传参考素材的基础文生视频。'),
    createSubmitRequest('首帧生视频', { model, prompt: '保持首帧人物身份和服装，人物自然回头。', duration: 5, images: [sourceImage], metadata: { ratio: '9:16' } }, '原始文档的 1 张图片模式：首帧生视频。'),
    createSubmitRequest('首尾帧生视频', { model, prompt: '从开始画面平滑过渡到结束画面，镜头连续自然。', duration: 5, images: [sourceImage, '{{last_frame_image_url}}'], metadata: { ratio: '16:9' } }, '原始文档的 2 张图片模式：第一张首帧，第二张尾帧。'),
    createSubmitRequest('多模态与控制项', {
      model,
      prompt: '以参考图的主体和参考音频的节奏制作短片。',
      duration: 5,
      metadata: {
        ratio: '1:1',
        seed: 42,
        generate_audio: true,
        watermark: false,
        callback_url: '{{callback_url}}',
        content: [
          { type: 'text', text: '主体表情自然，镜头稳定。' },
          { type: 'image_url', image_url: { url: sourceImage }, role: 'reference_image' },
          { type: 'audio_url', audio_url: { url: sourceAudio }, role: 'reference_audio' },
        ],
      },
    }, '展示 Seedance 原始文档列出的 `ratio`、`seed`、音频、水印、回调与多模态 `content`。'),
  ];
}

function seedanceReferenceExamples(model) {
  return [
    createSubmitRequest('基础参考视频', {
      model,
      prompt: '保留参考视频中人物的动作节奏，改为电影级夕阳光线。',
      duration: 5,
      metadata: { ratio: '16:9', reference_video: sourceVideo },
    }, '通过 `metadata.reference_video` 提供必需的参考视频。'),
    createSubmitRequest('多模态参考视频', {
      model,
      prompt: '以参考视频的运镜、参考图的人物身份和参考音频的情绪生成新片段。',
      duration: 5,
      metadata: {
        ratio: '9:16',
        seed: 42,
        generate_audio: true,
        content: [
          { type: 'video_url', video_url: { url: sourceVideo }, role: 'reference_video' },
          { type: 'image_url', image_url: { url: sourceImage }, role: 'reference_image' },
          { type: 'audio_url', audio_url: { url: sourceAudio }, role: 'reference_audio' },
        ],
      },
    }, '以 `metadata.content` 内的 `video_url` 满足参考视频要求，同时组合图片与音频。'),
  ];
}

function klingExamples(model) {
  const variant = klingVariant(model);
  const metadata = { aspect_ratio: '16:9' };
  if (variant.audio === '必须生成音频') metadata.generate_audio = true;
  if (variant.audio === '必须静音') metadata.generate_audio = false;
  if (variant.references === '至少一个参考输入') {
    metadata.reference_images = [{ url: sourceImage, role: 'reference_image' }];
  }
  return [createSubmitRequest(
    variant.references === '至少一个参考输入' ? '参考输入模式' : '基础生成',
    { model, prompt: '电影级写实镜头，主体自然运动，镜头稳定。', duration: 5, metadata },
    variant.references === '至少一个参考输入'
      ? '显式 `ref` 变体必须包含至少一种实际参考输入。'
      : '显式 `noref` 变体不包含任何参考素材。',
  )];
}

function klingAliasExamples(model) {
  return [
    createSubmitRequest('无参考输入', {
      model,
      prompt: '电影级写实镜头，主体自然运动，镜头稳定。',
      duration: 5,
      metadata: { aspect_ratio: '16:9', generate_audio: true },
    }, '兼容别名的无参考示例；生产调用优先选择显式分辨率、参考和音频变体。'),
    createSubmitRequest('图片参考输入', {
      model,
      prompt: '保持参考图人物身份，镜头从中景缓慢推进。',
      duration: 5,
      metadata: {
        aspect_ratio: '9:16',
        reference_images: [{ url: sourceImage, role: 'reference_image' }],
        generate_audio: false,
      },
    }, '兼容别名的图片参考示例；生产调用优先选择显式分辨率、参考和音频变体。'),
  ];
}

function sparseModelExamples(model, family) {
  const metadata = { aspect_ratio: '16:9' };
  let name = '基础生成';
  let prompt = '电影级自然光，主体运动自然，镜头平稳推进。';
  if (family === 'happyhorse' && (model.includes('-i2v-') || model.includes('-r2v-'))) {
    name = '图片参考模式';
    prompt = '保持参考图主体身份和构图，人物自然转身。';
    metadata.reference_images = [{ url: sourceImage, role: model.includes('-i2v-') ? 'first_frame' : 'reference_image' }];
  } else if ((family === 'happyhorse' && model.includes('-video-edit-')) || (family === 'wan' && (model.includes('-r2v') || model.includes('videoedit')))) {
    name = '视频参考模式';
    prompt = '保留参考视频的运动节奏，将画面调整为清晨柔和光线。';
    metadata.reference_videos = [{ url: sourceVideo, role: 'reference_video' }];
  } else if (family === 'wan' && model.includes('-i2v')) {
    name = '图片参考模式';
    prompt = '参考图中的人物微笑并向镜头走来。';
    metadata.reference_images = [{ url: sourceImage, role: 'first_frame' }];
  }
  const requests = [];
  if ((family === 'happyhorse' && model.includes('-video-edit-')) || (family === 'wan' && (model.includes('-r2v') || model.includes('videoedit')))) {
    requests.push(createSubmitRequest('图片参考模式（适配器允许）', {
      model,
      prompt: '参考图片中的主体完成自然运动，保持主体身份和构图。',
      duration: 5,
      metadata: { aspect_ratio: '16:9', reference_images: [{ url: sourceImage, role: 'reference_image' }] },
    }, '当前 new-api 适配器允许用图片满足“至少一个参考输入”；原始详情页未明确该输入类型，是否被上游接受以实际响应为准。'));
  }
  requests.push(createSubmitRequest(name, { model, prompt, duration: 5, metadata }, '使用该模型当前已知合法输入模式。原始详情页未提供完整扩展参数表，参见模型目录说明。'));
  return requests;
}

function h3Examples(model) {
  return [
    createSubmitRequest('文生视频', {
      model,
      prompt: '夜晚都市屋顶，人物迎着晚风缓慢走向镜头，电影级写实。',
      duration: 10,
      metadata: { aspect_ratio: '16:9', fps: 24, seed: 42 },
    }, '不传任何参考素材的文生视频模式。'),
    createSubmitRequest('首尾帧生视频', {
      model,
      prompt: '保持人物身份和服装一致，从首帧平稳过渡到尾帧。',
      duration: 8,
      metadata: {
        aspect_ratio: '9:16',
        fps: 24,
        reference_images: [
          { url: sourceImage, role: 'first_frame' },
          { url: '{{last_frame_image_url}}', role: 'last_frame' },
        ],
      },
    }, '首尾帧模式不能再混入参考图、参考视频或参考音频。'),
    createSubmitRequest('参考图片生视频', {
      model,
      prompt: '保持参考人物的脸型、发型和服装，镜头自然推进。',
      duration: 8,
      metadata: {
        aspect_ratio: '21:9',
        fps: 24,
        reference_images: [{ url: sourceImage, role: 'reference_image' }],
      },
    }, '参考生模式；最多可传 9 张且所有图片都应标注为 `reference_image`。'),
    createSubmitRequest('参考视频和音频', {
      model,
      prompt: '沿用参考视频的运镜节奏和参考音频的情绪，保持人物身份稳定。',
      duration: 12,
      metadata: {
        aspect_ratio: '16:9',
        fps: 24,
        reference_images: [{ url: sourceImage, role: 'reference_image' }],
        reference_videos: [{ url: sourceVideo, role: 'reference_video' }],
        reference_audios: [{ url: sourceAudio, role: 'reference_audio' }],
        extra: { reference_video_audio: false },
      },
    }, '参考生模式；视频和音频各最多 3 条，`reference_video_audio: false` 表示不采用参考视频原声。'),
  ];
}

function modelDefinition(entry) {
  const model = entry.model_name;
  if (model.startsWith('doubao-seedance-')) {
    const isReference = model.includes('-video-');
    return {
      description: isReference ? seedanceReferenceDescription(entry) : seedanceStandardDescription(entry),
      items: isReference ? seedanceReferenceExamples(model) : seedanceStandardExamples(model),
    };
  }
  if (model.startsWith('kling-')) {
    return {
      description: klingDescription(entry),
      items: model === 'kling-v3-omni' ? klingAliasExamples(model) : klingExamples(model),
    };
  }
  if (model.startsWith('happyhorse-')) {
    return { description: sparseModelDescription(entry, 'happyhorse'), items: sparseModelExamples(model, 'happyhorse') };
  }
  if (model.startsWith('wan')) {
    return { description: sparseModelDescription(entry, 'wan'), items: sparseModelExamples(model, 'wan') };
  }
  if (model.startsWith('zzdh-Minimax-h3-')) {
    return { description: h3Description(entry), items: h3Examples(model) };
  }
  throw new Error(`No collection profile for ${model}`);
}

const assignedModels = new Set();
const modelFolders = folders.map((folder) => {
  const folderModels = expectedModels.filter((model) => folder.matches(model));
  folderModels.forEach((model) => assignedModels.add(model));
  return {
    name: folder.name,
    item: folderModels.map((model) => {
      const entry = detailByModel.get(model);
      const definition = modelDefinition(entry);
      return { name: model, description: definition.description, item: definition.items };
    }),
  };
});

const unassignedModels = expectedModels.filter((model) => !assignedModels.has(model));
if (unassignedModels.length > 0) {
  throw new Error(`Models without an Apifox folder: ${unassignedModels.join(', ')}`);
}

const collection = {
  info: {
    name: '国产模型',
    description: markdown([
      '# 国产模型异步任务接口',
      '此集合仅展示 new-api 对 ZZDH 视频模型开放的公共请求。每个模型目录都包含原始详情来源、字段说明、已知限制和可直接修改的合法模式示例。',
      `将 \`base_url\` 设置为 \`https://goswitch.online\` 或 \`https://goswitcher.com\`，并填写 \`api_token\`。提交、查询和下载均要求 Bearer Token。`,
      '所有模型均为异步任务：提交后用 `task_id` 查询状态，完成后再下载内容。',
    ]),
    schema: 'https://schema.getpostman.com/json/collection/v2.1.0/collection.json',
  },
  auth: {
    type: 'bearer',
    bearer: [{ key: 'token', value: '{{api_token}}', type: 'string' }],
  },
  variable: [
    { key: 'base_url', value: 'https://goswitch.online' },
    { key: 'api_token', value: '' },
    { key: 'task_id', value: '' },
    { key: 'image_url', value: 'https://cdn.example.com/reference-image.png' },
    { key: 'last_frame_image_url', value: 'https://cdn.example.com/last-frame-image.png' },
    { key: 'video_url', value: 'https://cdn.example.com/reference-video.mp4' },
    { key: 'audio_url', value: 'https://cdn.example.com/reference-audio.mp3' },
    { key: 'callback_url', value: 'https://example.com/zzdh/callback' },
  ],
  item: [
    ...modelFolders,
    {
      name: '查询任务',
      description: '提交成功后使用 `task_id` 查询状态。必须携带创建该任务的同一用户 API Token。',
      item: [{
        name: '查询视频任务',
        request: {
          method: 'GET',
          header: [
            { key: 'Authorization', value: 'Bearer {{api_token}}', type: 'text' },
            { key: 'Accept', value: 'application/json', type: 'text' },
          ],
          url: '{{base_url}}/v1/videos/{{task_id}}',
          description: '使用创建任务时返回的 `task_id` 查询状态。请求 token 必须属于该任务的创建用户。',
        },
        response: [],
      }],
    },
    {
      name: '下载任务',
      description: '仅能下载当前用户已完成任务的视频内容；必须携带创建该任务的同一用户 API Token。',
      item: [{
        name: '下载已完成视频',
        request: {
          method: 'GET',
          header: [
            { key: 'Authorization', value: 'Bearer {{api_token}}', type: 'text' },
            { key: 'Accept', value: 'video/mp4, application/octet-stream', type: 'text' },
          ],
          url: '{{base_url}}/v1/videos/{{task_id}}/content',
          description: '下载已完成任务的内容。请求 token 必须属于该任务的创建用户。',
        },
        response: [],
      }],
    },
  ],
};

await Bun.write(outputPath, `${JSON.stringify(collection, null, 2)}\n`);
console.log(`Wrote ${outputPath}: ${expectedModels.length} model folders, ${collection.item.length} top-level folders.`);
