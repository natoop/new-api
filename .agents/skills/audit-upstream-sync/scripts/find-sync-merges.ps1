[CmdletBinding()]
param(
  [Parameter(Mandatory)]
  [string]$OldHead,

  [Parameter(Mandatory)]
  [string]$OldUpstream,

  [string]$SourceRef = 'upstream/main',

  [string]$HeadRef = 'HEAD'
)

$ErrorActionPreference = 'Stop'

function Resolve-GitCommit([string]$Revision, [string]$Label) {
  $resolved = @(& git rev-parse --verify "${Revision}^{commit}")
  $code = $LASTEXITCODE
  if ($code -ne 0 -or $resolved.Count -ne 1) {
    throw "git rev-parse failed ($code) for ${Label}: $Revision"
  }
  return $resolved[0].Trim()
}

function Test-GitAncestor([string]$Older, [string]$Newer, [string]$Label) {
  & git merge-base --is-ancestor -- $Older $Newer
  $code = $LASTEXITCODE
  switch ($code) {
    0 { return $true }
    1 { return $false }
    default {
      throw "git merge-base --is-ancestor failed ($code) for ${Label}: $Older -> $Newer"
    }
  }
}

function Get-GitPatchId([string]$From, [string]$To, [string]$Label) {
  $diff = @(& git diff $From $To)
  $diffCode = $LASTEXITCODE
  if ($diffCode -ne 0) {
    throw "git diff failed ($diffCode) for ${Label}: $From -> $To"
  }

  $patchIdLine = @($diff | & git patch-id --stable)
  $code = $LASTEXITCODE
  if ($code -ne 0) {
    throw "git patch-id failed ($code) for ${Label}: $From -> $To"
  }
  if ($patchIdLine.Count -eq 0) {
    return 'EMPTY'
  }
  if ($patchIdLine.Count -ne 1) {
    throw "git patch-id returned $($patchIdLine.Count) rows for ${Label}: $From -> $To"
  }
  return (($patchIdLine[0].Trim() -split '\s+')[0])
}

$oldHeadCommit = Resolve-GitCommit $OldHead 'saved_head'
$oldUpstreamCommit = Resolve-GitCommit $OldUpstream 'saved_synced_upstream'
$sourceTip = Resolve-GitCommit $SourceRef 'source_ref'
$headTip = Resolve-GitCommit $HeadRef 'head_ref'

if (-not (Test-GitAncestor $oldHeadCommit $headTip 'saved_head')) {
  throw "saved_head is not an ancestor: $oldHeadCommit -> $headTip"
}
if (-not (Test-GitAncestor $oldUpstreamCommit $sourceTip 'saved_synced_upstream')) {
  throw "saved_synced_upstream is not an ancestor: $oldUpstreamCommit -> $sourceTip"
}

$sourceFirstParentCommits = @(& git rev-list --first-parent $sourceTip)
$code = $LASTEXITCODE
if ($code -ne 0) {
  throw "git rev-list --first-parent failed ($code): $SourceRef"
}

$sourceFirstParentSet =
  [System.Collections.Generic.HashSet[string]]::new(
    [System.StringComparer]::OrdinalIgnoreCase
  )
$sourceFirstParentCommits |
  ForEach-Object { [void]$sourceFirstParentSet.Add($_) }

if (-not $sourceFirstParentSet.Contains($oldUpstreamCommit)) {
  throw "saved_synced_upstream is not on the first-parent line of ${SourceRef}: $oldUpstreamCommit"
}

$candidates = @(
  & git rev-list --merges --topo-order --reverse `
    "$oldHeadCommit..$headTip"
)
$code = $LASTEXITCODE
if ($code -ne 0) {
  throw "git rev-list candidate discovery failed ($code): $oldHeadCommit..$headTip"
}

$headFirstParentCommits = @(& git rev-list --first-parent $headTip)
$code = $LASTEXITCODE
if ($code -ne 0) {
  throw "git rev-list HEAD first-parent failed ($code): $headTip"
}
$headFirstParentSet =
  [System.Collections.Generic.HashSet[string]]::new(
    [System.StringComparer]::OrdinalIgnoreCase
  )
$headFirstParentCommits |
  ForEach-Object { [void]$headFirstParentSet.Add($_) }

$headFirstParentMerges = @(
  & git rev-list --first-parent --merges --reverse `
    "$oldHeadCommit..$headTip"
)
$code = $LASTEXITCODE
if ($code -ne 0) {
  throw "git rev-list HEAD first-parent merges failed ($code): $oldHeadCommit..$headTip"
}

$currentUpstream = $oldUpstreamCommit
$results = [System.Collections.Generic.List[object]]::new()

foreach ($merge in $candidates) {
  $parentLine = @(& git show -s --format='%P' $merge)
  $code = $LASTEXITCODE
  if ($code -ne 0 -or $parentLine.Count -ne 1) {
    throw "git show parents failed ($code): $merge"
  }

  $parents = @(
    ($parentLine[0].Trim() -split '\s+') |
      Where-Object { $_ }
  )
  if ($parents.Count -ne 2) {
    Write-Warning "Skip ${merge}: expected exactly two parents, got $($parents.Count)"
    continue
  }

  $preMergeHead = $parents[0]
  $newUpstream = $parents[1]

  if (Test-GitAncestor $merge $sourceTip "source history candidate $merge") {
    Write-Verbose "Skip ${merge}: merge is already part of source history"
    continue
  }
  if (-not $sourceFirstParentSet.Contains($newUpstream)) {
    Write-Verbose "Skip ${merge}: second parent is not on the source first-parent line"
    continue
  }
  if ($newUpstream -eq $currentUpstream) {
    Write-Verbose "Skip ${merge}: upstream anchor does not advance"
    continue
  }
  if (-not (Test-GitAncestor $currentUpstream $newUpstream "candidate $merge")) {
    Write-Verbose "Skip ${merge}: upstream anchor does not move forward"
    continue
  }

  $integrationMerge = $merge
  $integrationBase = $preMergeHead
  if (-not $headFirstParentSet.Contains($merge)) {
    foreach ($carrier in $headFirstParentMerges) {
      if (-not (Test-GitAncestor $merge $carrier "carrier $carrier")) {
        continue
      }

      $carrierParentLine = @(& git show -s --format='%P' $carrier)
      $code = $LASTEXITCODE
      if ($code -ne 0 -or $carrierParentLine.Count -ne 1) {
        throw "git show carrier parents failed ($code): $carrier"
      }
      $carrierParents = @(
        ($carrierParentLine[0].Trim() -split '\s+') |
          Where-Object { $_ }
      )
      if ($carrierParents.Count -lt 2) {
        continue
      }

      $carrierFirstParent = $carrierParents[0]
      if (Test-GitAncestor $merge $carrierFirstParent "carrier first parent $carrier") {
        continue
      }

      $integrationMerge = $carrier
      $integrationBase = $carrierFirstParent
      break
    }

    if ($integrationMerge -eq $merge) {
      throw "could not find first-parent carrier merge for nested sync: $merge"
    }
  }

  $result = [pscustomobject]@{
    Merge            = $merge
    PreMergeHead     = $preMergeHead
    UpstreamFrom     = $currentUpstream
    UpstreamTo       = $newUpstream
    IntegrationMerge = $integrationMerge
    IntegrationBase  = $integrationBase
    SyncPatchId      = ''
    UpstreamPatchId  = ''
    SyncPatchMatches = $false
    IntegrationUpstreamFrom = ''
    IntegrationUpstreamTo = ''
    IntegrationPatchId = ''
    IntegrationExpectedPatchId = ''
    IntegrationPatchMatches = $false
  }
  $result.SyncPatchId = Get-GitPatchId $preMergeHead $merge "sync merge $merge"
  $result.UpstreamPatchId = Get-GitPatchId $currentUpstream $newUpstream "upstream range $merge"
  $result.SyncPatchMatches = $result.SyncPatchId -eq $result.UpstreamPatchId
  if (-not $result.SyncPatchMatches) {
    Write-Warning "Sync merge ${merge} contains conflict resolution or local adaptation"
  }
  $results.Add($result)

  $currentUpstream = $newUpstream
}

$integrationGroups = @($results | Group-Object IntegrationMerge)
foreach ($group in $integrationGroups) {
  $batches = @($group.Group)
  $first = $batches[0]
  $last = $batches[-1]
  $integrationPatchId = Get-GitPatchId `
    $first.IntegrationBase `
    $first.IntegrationMerge `
    "integration $($first.IntegrationMerge)"
  $integrationExpectedPatchId = Get-GitPatchId `
    $first.UpstreamFrom `
    $last.UpstreamTo `
    "integration upstream range $($first.IntegrationMerge)"
  $integrationMatches = $integrationPatchId -eq $integrationExpectedPatchId

  if (-not $integrationMatches) {
    Write-Warning "Integration merge $($first.IntegrationMerge) contains conflict resolution, local adaptation, or additional changes"
  }

  foreach ($batch in $batches) {
    $batch.IntegrationUpstreamFrom = $first.UpstreamFrom
    $batch.IntegrationUpstreamTo = $last.UpstreamTo
    $batch.IntegrationPatchId = $integrationPatchId
    $batch.IntegrationExpectedPatchId = $integrationExpectedPatchId
    $batch.IntegrationPatchMatches = $integrationMatches
  }
}

$results
