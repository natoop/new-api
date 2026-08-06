const stamp = process.argv[2] ?? new Date().toISOString().slice(0, 10).replaceAll('-', '');
const sourcePath = 'ZZDH_VIDEO_API_APIFOX.openapi.yaml';
const outputPath = `ZZDH_TARGET_MODEL_DETAILS_${stamp}.json`;

if (await Bun.file(outputPath).exists()) {
  throw new Error(`${outputPath} already exists; preserve evidence snapshots and use a new date stamp.`);
}

const source = Bun.YAML.parse(await Bun.file(sourcePath).text());
const models = source?.components?.schemas?.ZZDHVideoModel?.enum;
if (!Array.isArray(models) || models.length === 0) {
  throw new Error(`Unable to read ZZDH models from ${sourcePath}`);
}

const baseURL = 'https://zizidonghua.com';
const fetchedAt = new Date().toISOString();
const concurrency = 5;
let nextIndex = 0;

async function fetchModelDetail(modelName) {
  const detailURL = `${baseURL}/api/api-docs/model/${encodeURIComponent(modelName)}`;
  try {
    const response = await fetch(detailURL, { headers: { Accept: 'application/json' } });
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    const payload = await response.json();
    if (!payload?.success || !payload?.data) {
      throw new Error('response does not contain detail data');
    }
    return { model_name: modelName, detail_url: detailURL, detail: payload.data, fetch_error: '' };
  } catch (error) {
    return {
      model_name: modelName,
      detail_url: detailURL,
      detail: null,
      fetch_error: error instanceof Error ? error.message : String(error),
    };
  }
}

const results = new Array(models.length);
await Promise.all(Array.from({ length: Math.min(concurrency, models.length) }, async () => {
  while (nextIndex < models.length) {
    const index = nextIndex++;
    results[index] = await fetchModelDetail(models[index]);
  }
}));

const failures = results.filter((entry) => entry.fetch_error !== '');
const snapshot = {
  source_url: `${baseURL}/api/api-docs/model/{model_name}`,
  fetched_at: fetchedAt,
  model_count: models.length,
  data: results,
};

await Bun.write(outputPath, `${JSON.stringify(snapshot, null, 2)}\n`);
console.log(`Wrote ${outputPath}: ${models.length - failures.length}/${models.length} details fetched.`);
if (failures.length > 0) {
  console.error(`Failed models: ${failures.map((entry) => entry.model_name).join(', ')}`);
  process.exitCode = 1;
}
