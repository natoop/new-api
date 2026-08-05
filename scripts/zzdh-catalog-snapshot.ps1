param(
    [string]$BaseUrl = 'https://zizidonghua.com',
    [string]$Stamp = (Get-Date -Format 'yyyyMMdd'),
    [int]$TimeoutSec = 12,
    [switch]$IncludeContractDetails,
    [switch]$MergeContractDetailBatches,
    [int]$DetailOffset = 0,
    [ValidateRange(1, 20)]
    [int]$DetailLimit = 10
)

$ErrorActionPreference = 'Stop'
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$base = $BaseUrl.TrimEnd('/')
$catalogUrl = "$base/api/api-docs/models"
$fetchedAt = (Get-Date).ToUniversalTime().ToString('o')

function Get-DetailUrl([string]$ModelName) {
    return "$base/api/api-docs/model/$([Uri]::EscapeDataString($ModelName))"
}

function Get-Classification($Model) {
    $name = [string]$Model.model_name
    $endpoint = [string]($Model.supported_endpoint_types -join ',')

    if ($name -like 'doubao-seedance*') {
        if ($name -match '-video') { return 'V1 generation / reference-video billing' }
        return 'V1 generation / output-seconds billing'
    }
    if ($name -like '*视频超分*') { return 'V1 upscale / input-seconds billing' }
    if ($name -like 'kling-3.0-omni-*') { return 'V8 generation / confirmed variant' }
    if ($name -eq 'kling-v3-omni') { return 'V8 generation / ambiguous generic variant' }
    if ($name -like 'zzdh-Minimax-h3-*') { return 'V8 generation / endpoint metadata conflict' }
    if ($endpoint -match 'openai-video') {
        if ($name -match 'image|t2i') { return 'openai-video metadata anomaly / image candidate' }
        return 'video candidate / generic documentation only'
    }
    if ($endpoint -match 'audio|music') { return 'non-video / audio or music adapter' }
    if ($endpoint -match 'image-generation') { return 'non-video / image adapter' }
    if ($endpoint -match 'embeddings') { return 'non-video / embedding adapter' }
    return 'non-video / ordinary OpenAI adapter'
}

function Get-ExpectedPaths($Model) {
    $name = [string]$Model.model_name
    if ($name -like 'doubao-seedance*' -or $name -like '*视频超分*') {
        return 'POST /v1/video/generations; GET /v1/videos/{task_id}'
    }
    if ($name -like 'kling*' -or $name -like 'zzdh-Minimax-h3-*') {
        return 'POST /v8/videos/generations; GET /v8/videos/generations/{task_id}'
    }
    return ''
}

function ConvertTo-Cell([object]$Value) {
    if ($null -eq $Value) { return '' }
    return ([string]$Value).Replace('|', '\|').Replace("`r", ' ').Replace("`n", ' ')
}

$catalogResponse = Invoke-RestMethod -UseBasicParsing $catalogUrl -TimeoutSec $TimeoutSec
$models = @($catalogResponse.data)
if ($models.Count -eq 0) { throw "ZZDH model catalogue returned no models: $catalogUrl" }

$catalogSnapshot = [ordered]@{
    source_url = $catalogUrl
    fetched_at = $fetchedAt
    model_count = $models.Count
    endpoint_map = $catalogResponse.endpoint_map
    data = $models
}
$catalogPath = Join-Path $repoRoot "ZZDH_MODEL_CATALOG_$Stamp.json"
$catalogSnapshot | ConvertTo-Json -Depth 20 | Set-Content -Path $catalogPath -Encoding UTF8

$videoModels = @($models | Where-Object { $_.supported_endpoint_types -contains 'openai-video' })
$contractModels = @($models | Where-Object {
    $_.model_name -like 'doubao-seedance*' -or
    $_.model_name -like '*视频超分*' -or
    $_.model_name -like 'kling*' -or
    $_.model_name -like 'zzdh-Minimax-h3-*'
} | Sort-Object model_name)

$detailPath = ''
if ($MergeContractDetailBatches) {
    $selected = @{}
    Get-ChildItem -Path $repoRoot -Filter "ZZDH_MODEL_DETAILS_${Stamp}_*.json" -File |
        Sort-Object Name |
        ForEach-Object {
            $batch = Get-Content -Raw $_.FullName | ConvertFrom-Json
            foreach ($entry in @($batch.data)) {
                $name = [string]$entry.model_name
                if ([string]::IsNullOrWhiteSpace($name)) { continue }
                $existing = $selected[$name]
                $hasSuccess = [string]::IsNullOrWhiteSpace([string]$entry.fetch_error)
                $existingFailed = $null -eq $existing -or -not [string]::IsNullOrWhiteSpace([string]$existing.fetch_error)
                if ($null -eq $existing -or ($hasSuccess -and $existingFailed)) {
                    $selected[$name] = $entry
                }
            }
        }
    if ($selected.Count -eq 0) {
        throw "No contract-detail batch files found for $Stamp; refusing to overwrite the merged snapshot."
    }
    $detailPath = Join-Path $repoRoot "ZZDH_MODEL_DETAILS_$Stamp.json"
    [ordered]@{
        source_url = "$base/api/api-docs/model/{model_name}"
        fetched_at = $fetchedAt
        model_count = $selected.Count
        data = @($selected.Values | Sort-Object model_name)
    } | ConvertTo-Json -Depth 30 | Set-Content -Path $detailPath -Encoding UTF8
} elseif ($IncludeContractDetails) {
    $batch = @($contractModels | Select-Object -Skip $DetailOffset -First $DetailLimit)
    $details = foreach ($model in $batch) {
        $detailUrl = Get-DetailUrl ([string]$model.model_name)
        try {
            $response = Invoke-RestMethod -UseBasicParsing $detailUrl -TimeoutSec $TimeoutSec
            [ordered]@{
                model_name = $model.model_name
                detail_url = $detailUrl
                fetched_at = $fetchedAt
                detail = $response.data
                fetch_error = ''
            }
        } catch {
            [ordered]@{
                model_name = $model.model_name
                detail_url = $detailUrl
                fetched_at = $fetchedAt
                detail = $null
                fetch_error = $_.Exception.Message
            }
        }
    }
    $last = $DetailOffset + [Math]::Max($batch.Count - 1, 0)
    $detailPath = Join-Path $repoRoot "ZZDH_MODEL_DETAILS_${Stamp}_${DetailOffset}-${last}.json"
    [ordered]@{
        source_url = "$base/api/api-docs/model/{model_name}"
        fetched_at = $fetchedAt
        offset = $DetailOffset
        limit = $DetailLimit
        data = @($details)
    } | ConvertTo-Json -Depth 30 | Set-Content -Path $detailPath -Encoding UTF8
}

$lines = [System.Collections.Generic.List[string]]::new()
$lines.Add('# ZZDH Model Catalogue Snapshot')
$lines.Add('')
$lines.Add("- Snapshot date: $Stamp")
$lines.Add("- Fetched at (UTC): $fetchedAt")
$lines.Add("- Source: [$catalogUrl]($catalogUrl)")
$lines.Add("- Total models: $($models.Count)")
$lines.Add(('- Models tagged ``openai-video``: {0}' -f $videoModels.Count))
$lines.Add(('- Raw catalogue: ``ZZDH_MODEL_CATALOG_{0}.json``' -f $Stamp))
if ($detailPath) { $lines.Add(('- Contract-detail batch: ``{0}``' -f (Split-Path $detailPath -Leaf))) }
$lines.Add('')
$lines.Add('This file is an evidence snapshot, not runtime configuration. A model is not production-enabled merely because it appears here.')
$lines.Add('')

foreach ($group in @($models | Group-Object { Get-Classification $_ } | Sort-Object Name)) {
    $lines.Add("## $($group.Name) ($($group.Count))")
    $lines.Add('')
    $lines.Add('| Model | Catalog endpoint | Confirmed or expected task paths | Catalogue note |')
    $lines.Add('| --- | --- | --- | --- |')
    foreach ($model in @($group.Group | Sort-Object model_name)) {
        $endpoint = ($model.supported_endpoint_types -join ', ')
        $paths = Get-ExpectedPaths $model
        $lines.Add("| $(ConvertTo-Cell $model.model_name) | $(ConvertTo-Cell $endpoint) | $(ConvertTo-Cell $paths) | $(ConvertTo-Cell $model.description) |")
    }
    $lines.Add('')
}

$markdownPath = Join-Path $repoRoot "ZZDH_MODEL_CATALOG_$Stamp.md"
$lines -join "`r`n" | Set-Content -Path $markdownPath -Encoding UTF8

Write-Output "Wrote $catalogPath"
if ($detailPath) { Write-Output "Wrote $detailPath" }
Write-Output "Wrote $markdownPath"
