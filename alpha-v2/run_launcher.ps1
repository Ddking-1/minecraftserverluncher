$ErrorActionPreference = "Stop"
$base = "C:\GITHUB\minecraftserverluncher\_local_test"
$backendDir = Join-Path $base "backend"
$mcDir = Join-Path $base "minecraft-server"
$velDir = Join-Path $base "velocity"
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

$procs = @()

# Backend
Write-Host "[1/3] Starting backend..."
$backendExe = Join-Path $backendDir "auth-server.exe"
$backendProc = Start-Process -FilePath $backendExe -WorkingDirectory $backendDir -RedirectStandardOutput (Join-Path $logDir "backend.log") -RedirectStandardError (Join-Path $logDir "backend.err.log") -PassThru -NoNewWindow
$procs += $backendProc
$ready = $false
for ($i=0; $i -lt 30; $i++) { Start-Sleep -Seconds 1; try { $r = Invoke-WebRequest -Uri "http://127.0.0.1:8080/readyz" -UseBasicParsing -TimeoutSec 2; if ($r.StatusCode -eq 200) { $ready=$true; break } } catch {} }
if ($ready) { Write-Host "  backend ready." } else { Write-Host "  backend NOT ready." }

# Minecraft server (max 2GB RAM, optimized)
Write-Host "[2/3] Starting Minecraft server (max 2GB RAM)..."
$javaExe = "C:\Program Files\Microsoft\jdk-21.0.12.8-hotspot\bin\java.exe"
if (-not (Test-Path $javaExe)) { $javaExe = "C:\Program Files\Microsoft\jdk-25.0.4.7-hotspot\bin\java.exe" }
$mcArgs = @("-Xms512M","-Xmx2G","-XX:+UseG1GC","-XX:MaxGCPauseMillis=100","-XX:+ParallelRefProcEnabled","-jar","server.jar","nogui")
$mcProc = Start-Process -FilePath $javaExe -ArgumentList $mcArgs -WorkingDirectory $mcDir -RedirectStandardOutput (Join-Path $logDir "minecraft.log") -RedirectStandardError (Join-Path $logDir "minecraft.err.log") -PassThru -NoNewWindow
$procs += $mcProc
$mcReady = $false
for ($i=0; $i -lt 120; $i++) { Start-Sleep -Seconds 1; if (Test-Path (Join-Path $logDir "minecraft.log")) { $c = Get-Content (Join-Path $logDir "minecraft.log") -Raw -ErrorAction SilentlyContinue; if ($c -match "Done \(") { $mcReady=$true; break } }; if ($mcProc.HasExited) { break } }
if ($mcReady) { Write-Host "  MC server ready." } else { Write-Host "  MC server not ready yet." }

# Velocity
Write-Host "[3/3] Starting Velocity..."
$velProc = Start-Process -FilePath $javaExe -ArgumentList "-Xms256M","-Xmx1G","-jar","velocity.jar" -WorkingDirectory $velDir -RedirectStandardOutput (Join-Path $logDir "velocity.log") -RedirectStandardError (Join-Path $logDir "velocity.err.log") -PassThru -NoNewWindow
$procs += $velProc
$velReady = $false
for ($i=0; $i -lt 60; $i++) { Start-Sleep -Seconds 1; if (Test-Path (Join-Path $logDir "velocity.log")) { $c = Get-Content (Join-Path $logDir "velocity.log") -Raw -ErrorAction SilentlyContinue; if ($c -match "Done \(") { $velReady=$true; break } }; if ($velProc.HasExited) { break } }
if ($velReady) { Write-Host "  Velocity ready." } else { Write-Host "  Velocity not ready yet." }

# Open launcher
$launcherExe = Join-Path $base "launcher\MinecraftLauncher.exe"
if (Test-Path $launcherExe) { Start-Process -FilePath $launcherExe }
Start-Process "http://127.0.0.1:8080/admin"

Write-Host ""
Write-Host "All services started. Press Enter to stop."
[void][System.Console]::ReadLine()
foreach ($p in $procs) { try { if (-not $p.HasExited) { Stop-Process -Id $p.Id -Force } } catch {} }
Write-Host "Stopped."