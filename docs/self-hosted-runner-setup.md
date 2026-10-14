# Self-Hosted GitHub Actions Runner — telepítés és konfiguráció

## 1. Előfeltételek

- **GitHub fiók**, ami admin jogosultságot kapott a `Ddking-1/minecraftserverluncher` repóban.
- **Hozzáférés a GitHub API-hoz** egy `repo` scope-os Personal Access Token-gel (vagy `gh` CLI-vel, ha telepítve van).
- **Windows** (ajánlott), de Linux/macOS is működik – a script PowerShell-ben van, de lentebb Bash változat is van.
- **Nincsenek erősen korlátozó szerveziós szabályok** a self-hosted runner regisztrálására (lásd: Organization Settings → Actions → Runners).

## 2. Regisztrációs token előállítása (KÉT mód)

### 2a. GitHub web felület (ajánlott, egylépéses)

1. Nyisd meg: `https://github.com/Ddking-1/minecraftserverluncher/settings/actions/runners`
2. Kattints a **New self-hosted runner** gombra.
3. Válaszd a **Windows** / **x64** opciót.
4. A felület megjelenít egy `TOKEN` értéket és a letöltési + konfig parancsokat.
5. Másold ki a `TOKEN` értéket – ez azonosítja a runnert 1 órára (expirál).

### 2b. GitHub API + curl (ha nincs GUI vagy script）

A regisztrációs tokennek **admin access** kell a repóhoz.

**Personal Access Token (PAT) generálása:**
1. `https://github.com/settings/tokens` (classic PAT) – vagy
2. `https://github.com/settings/tokens?type=beta` (fine-grained PAT → Repository access → Administration: write)

**Token generálása az API-n (Windows PowerShell):**

```powershell
# Behelyettesítve a saját PAT-eddel:
$headers = @{
    "Authorization" = "token PLACEHOLDER_PAT"
    "Accept"        = "application/vnd.github+json"
}
$response = Invoke-RestMethod -Method Post `
    -Uri "https://api.github.com/repos/Ddking-1/minecraftserverluncher/actions/runners/registration-token" `
    -Headers $headers
$token = $response.token
Write-Output "Regisztrációs token: $token"
# Token expires in ~1 hour!
```

Ha `gh` elérhető:
```bash
gh auth login   # ha még nem
gh api --method POST repos/Ddking-1/minecraftserverluncher/actions/runners/registration-token --jq .token
```

**Fontos:** a regisztrációs token **csak 1 órán** belül használható. Ha letelik, újragenerálás kell.

## 3. Runner letöltése és kibontása (PowerShell — Windows)

Ez a script letölti a hivatalos `actions/runner` NuGet-szerű csomagot, kibontja és előkészíti.

```powershell
# run_install_runner.ps1
param(
    # Ha üresen hagyod, bekérjük interaktíven. Ha adva van, azt használjuk.
    [string]$Token = "",
    [string]$RunnerUrl = "https://github.com/Ddking-1/minecraftserverluncher"
)

$ErrorActionPreference = "Stop"

# 1) Helyszín
$InstallRoot = Join-Path $HOME "actions-runner"
if (-not (Test-Path $InstallRoot)) { New-Item -Path $InstallRoot -ItemType Directory | Out-Null }

Set-Location $InstallRoot

# 2) Legújabb ajánlott verzió lekérdezése a GitHub API-ból
Write-Host "Lekérdezés: actions/runner legújabb ajánlott verzió..."
$releasesApi = "https://api.github.com/repos/actions/runner/releases/latest"
$release = Invoke-RestMethod -Uri $releasesApi -Headers @{
    "Accept" = "application/vnd.github+json"
}
$tag = $release.tag_name               # pl. "v2.312.0"
$asset = $release.assets | Where-Object { $_.name -match "^actions-runner-win-x64\.zip$" } | Select-Object -First 1
if (-not $asset) { throw "Nem található actions-runner-win-x64.zip a releaseben" }
$downloadUrl = $asset.browser_download_url

Write-Host "Verzió: $tag"
Write-Host "Letöltés: $downloadUrl"

# 3) Letöltés
$zipPath = Join-Path $InstallRoot "runner.zip"
Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -UseBasicParsing

# 4) Kibontás (PowerShell 5+)
Write-Host "Kibontás..."
Expand-Archive -Path $zipPath -DestinationPath $InstallRoot -Force
Remove-Item $zipPath -Force

# 5) Konfig folder és parancsok
$runnerFolderName = "actions-runner"  # majd a zip tartalmazza; ha másik, a Expand-Archive után megnézzük
$dir = Get-ChildItem -Path $InstallRoot -Directory | Where-Object { $_.Name -ne "runner.zip" } | Select-Object -Last 1
if (-not $dir) { throw "Nem található kibontott mappa" }
Set-Location $dir.FullName

# 6) Token bekérése, ha nincs megadva
if ([string]::IsNullOrEmpty($Token)) {
    $Token = Read-Host "Add meg a regisztrációs token-t (github.com/Ddking-1/minecraftserverluncher/settings/actions/runners)"
}

# 7) Konfig
Write-Host "Konfigurálás: $RunnerUrl / token: ..."
.\config.cmd --url $RunnerUrl --token $Token --name "minecraftserverluncher-windows-runner" --labels "self-hosted,windows,private-ci" --work _work

Write-Host "Kész! Indítás: .\run.cmd"
```

**Futtatás:**
```powershell
powershell -ExecutionPolicy Bypass -File run_install_runner.ps1
```

A script:
- Letölti a legfrissebb release és a `actions-runner-win-x64.zip` filet
- Kibontja a `$HOME/actions-runner` mappába
- Konfigolja a saját gépedhez a repó URL és a token használatával
- Elnevezi a `minecraftserverluncher-windows-runner` nevet, választható labelleimmel

## 4. Runner telepítés Bash/Linux verzió (ha Linuxon szeretnéd futtatni)

Ha Linuxod van (pl. egy VPS vagy Intel NUC), itt a megfelelő parancssor:

```bash
#!/usr/bin/env bash
set -euo pipefail

RUNNER_ROOT="$HOME/actions-runner"
RUNNER_URL="https://github.com/Ddking-1/minecraftserverluncher"

mkdir -p "$RUNNER_ROOT"
cd "$RUNNER_ROOT"

# Letöltés (legújabb verzió)
LATEST=$(curl -s "https://api.github.com/repos/actions/runner/releases/latest" \
  | jq -r '.tag_name')
ARCHIVE="actions-runner-linux-x64-${LATEST#v}.tar.gz"
DOWNLOAD_URL="https://github.com/actions/runner/releases/download/${LATEST}/${ARCHIVE}"

echo "Verzió: $LATEST"
curl -Lo "$ARCHIVE" "$DOWNLOAD_URL"
tar xzf "$ARCHIVE"
rm -f "$ARCHIVE"

cd actions-runner*

# Token bekérése
if [ -z "${TOKEN:-}" ]; then
  echo -n "Regisztrációs token: "
  read -r TOKEN
fi

./config.sh --url "$RUNNER_URL" --token "$TOKEN" \
  --name "minecraftserverluncher-linux-runner" \
  --labels "self-hosted,linux,private-ci" \
  --work _work

echo "Kész. Indítás: ./run.sh"
```

## 5. Indítás és ellenőrzés

### Windows
```powershell
cd $HOME\actions-runner\actions-runner-*
.\run.cmd
```

### Linux
```bash
cd $HOME/actions-runner/actions-runner-*
./run.sh
```

## 5. Indítás és ellenőrzés (folytatva)

### Windows – parancssori indítás
```powershell
cd $HOME\\actions-runner\\actions-runner-*
.\\run.cmd
```

### Windows – háttérszolgáltatás (service) telepítése és indítása
A GitHub Actions runner Windows szolgáltatásként is telepíthető, hogy a rendszer újraindítása után is automatikusan elindul.

#### Service telepítése
```powershell
cd $HOME\\actions-runner\\actions-runner*
.\\svc.cmd install
```

#### Service indítása
```powershell
.\\svc.cmd start
```

#### Service státusz ellenőrzése
```powershell
Get-Service actions.runner*
```

#### Service konfiguráció frissítése (token lejárt esetén)
```powershell
.\\config.cmd --url https://github.com/Ddking-1/minecraftserverluncher --token <új-token> --name "minecraftserverluncher-windows-runner" --labels "self-hosted,windows,private-ci" --work _work
```

### Linux – parancssori indítás
```bash
cd $HOME/actions-runner/actions-runner-*
./run.sh
```

### Linux – háttérszolgáltatás (service) telepítése és indítása
```bash
cd $HOME/actions-runner/actions-runner*
./svc.sh install
./svc.sh start
```

A futtató folyamat confirmálja:
```
√ Connected to GitHub
√ Runner registered! ...
√ Runner set up successfully!
√ Active
```

Ezután a GitHub repó **Settings → Actions → Runners** menüpontban látható lesz a runner-státusz (online / idle / busyt).

## 6. Teszt futtatás elindítása

Miután a runner online:

```bash
git commit --allow-empty -m "ci: verify self-hosted runner available"
git push origin master
```

Várakozás:
- 1-2 perc alatt a workflow elindul

A projektben minden job most már `runs-on: [self-hosted, linux]` vagy `runs-on: [self-hosted, windows]` formátumú. Ez azt jelenti, hogy **self-hosted runner szükséges**, és a runnernek **mindkét címke** (a `self-hosted` és a megfelelő OS címke) kell, hogy fejlészen illeszkedjen.

**Fontos:**
- A `runs-on: [self-hosted, linux]` jobokhoz Linuxos self-hosted runner kell (vagy egy Windows runner, ha rendel egy `linux` címkét is — de nem ajánlott).
- A `runs-on: [self-hosted, windows]` jobokhoz Windowsos self-hosted runner kell.
- Ha van két géped (vagy egy két operációs rendszerű VM), érdemes lehet két külön runner regisztrálni: egyet Linux-on, egyet Windows-on, mindkettővel a megfelelő címkékkel.
- Ha csak egy géppel áladsz a fejeden, és az pl. Windows, akkor érdemes lehet a Linux-os jobokat vagy átírni `runs-on: [self-hosted, windows]`-re (kivéve, ha a jobok Linux-specifikus parancsokat tartalmaznak), vagy Linux runner regisztrálni.

A `ci.yml`, `backend.yml`, `launcher.yml` fájlokban lévő módosításokat a repository-ban már elvégeztük: minden job `runs-on: [self-hosted, linux]` vagy `runs-on: [self-hosted, windows]`. Ezek után a GitHub Actions **csak akkor fogja elkezdeni** a jobot, ha van online self-hosted runner, amely megfelel a címkeelt.

## 7. Ellenőrzés GitHub API-n

A runner állapotának lekérdezése:

```bash
curl -s -H "Authorization: token PLACEHOLDER_PAT" \
  https://api.github.com/repos/Ddking-1/minecraftserverluncher/actions/runners | jq .
```

Vagy `gh`:
```bash
gh api repos/Ddking-1/minecraftserverluncher/actions/runners
```

## 8. Biztonsági megjegyzés

- A self-hosted runner **nem biztonságos forkokból érkező PR-ek** futtatására, ha nem akarsz, mert a fork workflows untrusted kódot futtathatnak a gépeden.
- Ajánlott: engedélyezd a **Fork workflows** policy-t szervezet szinten, vagy korlátozd a workflow triggert `push` és `pull_request` (de ne `pull_request_target`).
- A `ci.yml` jelenleg `pull_request`Triggers is tartalmazza – ez jó, mert a fork PR ugyanúgy lefut a runner-on (ha van hozzá jogosultság), de a kód a gépeden fut.
