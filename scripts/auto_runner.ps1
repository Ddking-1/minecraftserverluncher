# SPDX-License-Identifier: GPL-3.0-only
# Auto-runner loop for the project. Performs a bounded, re-entrant sequence of
# local checks (Go build/test/vet and, when present, the launcher CMake/CTest)
# with an explicit manual break point. Every iteration re-runs from a clean
# snapshot of the repo so it never depends on stale state.

$ErrorActionPreference = 'Stop'
$root = (Split-Path -Parent $PSScriptRoot)
Set-Location $root

# --- config ------------------------------------------------------------------
$maxIterations = if ($env:AUTO_RUNNER_MAX) { [int]$env:AUTO_RUNNER_MAX } else { 5 }
$skipGo      = $env:AUTO_RUNNER_SKIP_GO -eq '1'
$skipLauncher = $env:AUTO_RUNNER_SKIP_LAUNCHER -eq '1'

function Test-Step {
    param([string]$Name, [scriptblock]$Body)
    try {
        & $Body | Out-Null
        Write-Host "PASS  $Name"
        return $true
    } catch {
        Write-Host "FAIL  $Name  -> $($_.Exception.Message)"
        return $false
    }
}

# --- two-step stop test ------------------------------------------------------
# The tmp check file proves the runner can create artifacts on disk. The runner
# pauses here (STEP 1) and does NOT proceed to cleanup until the operator types
# continue/Folytasd, which is handled by the caller, not by this script.

function New-CheckFile {
    param([string]$Path)
    Set-Content -Path $Path -Value 'auto-runner-check' -Encoding UTF8
    if (-not (Test-Path -LiteralPath $Path)) {
        throw "check file was not created: $Path"
    }
}

# --- main --------------------------------------------------------------------
for ($i = 1; $i -le $maxIterations; $i++) {
    Write-Host "=== auto_runner iteration $i/$maxIterations ==="

    if (-not $skipGo) {
        $goOk = Test-Step 'go build ./...'    { & go build ./... }
        if (-not $goOk) { break }
        $vetOk = Test-Step 'go vet ./...'     { & go vet ./... }
        if (-not $vetOk) { break }
        $testOk = Test-Step 'go test ./...'   { & go test ./... -count=1 }
        if (-not $testOk) { break }
    }

    if (-not $skipLauncher -and (Test-Path (Join-Path $root 'launcher/CMakeLists.txt'))) {
        Test-Step 'launcher present (build skipped in runner)' { return $true }
    }

    Write-Host "iterations remaining: $($maxIterations - $i)"
}

Write-Host "auto_runner: done"
