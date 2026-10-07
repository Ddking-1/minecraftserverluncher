# PowerShell local checks script for minecraftserverlauncher.
# Usage: .\scripts\run_all_checks.ps1

$ErrorActionPreference = 'Stop'
Write-Host "==> Go: vet"
Push-Location -Path "server" -ErrorAction SilentlyContinue
try {
    go vet ./...
} finally {
    Pop-Location
}

Write-Host "==> Go: test"
Push-Location -Path "server" -ErrorAction SilentlyContinue
try {
    go test ./... -count=1 -timeout 120s -v
} finally {
    Pop-Location
}

Write-Host "==> Go: race"
Push-Location -Path "server" -ErrorAction SilentlyContinue
try {
    go test ./... -count=1 -timeout 120s -v -race
} catch {
    Write-Warning "Race detector may not be available in this environment: $_"
} finally {
    Pop-Location
}

Write-Host "==> Launcher: configure + build + ctest (offscreen)"
if (-not (Test-Path "launcher/CMakeLists.txt")) {
    Write-Error "launcher/CMakeLists.txt not found; aborting launcher steps"
    exit 1
}

$buildDir = "build-check"
if (Test-Path $buildDir) {
    Remove-Item -Recurse -Force $buildDir
}

cmake -B $buildDir -S launcher -G "Visual Studio 17 2022" -A x64
cmake --build $buildDir --config Release --target MinecraftLauncher, test_hwid, test_settings, test_update_manifest, test_skin_model, test_backend_contract, test_settings_store, test_mod_scanner, test_skin_cache, test_rotating_logger, test_update_checker, test_file_integrity, test_editor_config, test_skin_client, test_download_asset, test_download_queue, test_java_detector, test_game_paths, test_game_launcher, test_editor_settings_widget, test_login_dialog, test_skin_upload_widget, test_update_manager
$env:QT_QPA_PLATFORM = "offscreen"
ctest --test-dir $buildDir --output-on-failure -C Release

Write-Host "==> Done"
