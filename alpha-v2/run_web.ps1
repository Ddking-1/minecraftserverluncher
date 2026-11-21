$ErrorActionPreference = "Stop"
$base = "C:\GITHUB\minecraftserverluncher\_local_test"
$backendDir = Join-Path $base "backend"
$logDir = Join-Path $base "logs"
New-Item -ItemType Directory -Force -Path $logDir | Out-Null

$env:AUTH_LISTEN_ADDR = ":8080"
$env:AUTH_DB_PATH = Join-Path $backendDir "auth.db"
$env:AUTH_ADMIN_TOKEN = "local-dev-admin-token"
$env:AUTH_SKIN_STORAGE_DIR = Join-Path $backendDir "storage/skins"
$env:AUTH_PTERO_SCRIPT = ""
$env:AUTH_TURNSTILE_SITE_KEY = ""
$env:AUTH_TURNSTILE_SECRET = ""
$env:AUTH_REQUIRE_DEVICE_ID = "0"
$env:AUTH_FEATURE_INVITES = "1"
$env:AUTH_FEATURE_CAPTCHA = "0"

Write-Host "Starting backend on :8080 ..."
$backendExe = Join-Path $backendDir "auth-server.exe"
$backendProc = Start-Process -FilePath $backendExe -WorkingDirectory $backendDir -RedirectStandardOutput (Join-Path $logDir "backend.log") -RedirectStandardError (Join-Path $logDir "backend.err.log") -PassThru -NoNewWindow

$ready = $false
for ($i=0; $i -lt 30; $i++) {
    Start-Sleep -Seconds 1
    try { $r = Invoke-WebRequest -Uri "http://127.0.0.1:8080/readyz" -UseBasicParsing -TimeoutSec 2; if ($r.StatusCode -eq 200) { $ready=$true; break } } catch {}
}
if (-not $ready) { Write-Host "Backend not ready; check logs." } else { Write-Host "Backend ready." }

Start-Process "http://127.0.0.1:8080/"
Start-Process "http://127.0.0.1:8080/admin"

Write-Host "Press Enter to stop backend."
[void][System.Console]::ReadLine()
if (-not $backendProc.HasExited) { Stop-Process -Id $backendProc.Id -Force }
Write-Host "Stopped."