<#
.SYNOPSIS
    Setup script for cloud-ide-mount on Windows.
.DESCRIPTION
    Checks and installs dependencies: Go, Git, GitHub CLI, rclone.
    Generates SSH key and guides authentication.
#>

$ErrorActionPreference = "Stop"
$RepoRoot = Split-Path -Parent $PSScriptRoot
$SshDir = "$env:USERPROFILE\.ssh"
$SshKey = "$SshDir\codespaces.auto"

Write-Host "╔══════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  cloud-ide-mount — Windows Setup         ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# ─── Helper ───────────────────────────────────────────────────────
function Check-Command($name, $required) {
    $cmd = Get-Command $name -ErrorAction SilentlyContinue
    if ($cmd) {
        Write-Host "  ✅ $name found: $($cmd.Source)" -ForegroundColor Green
        return $true
    }
    if ($required) {
        Write-Host "  ❌ $name NOT found." -ForegroundColor Red
        return $false
    }
    Write-Host "  ⚠️  $name NOT found (optional)." -ForegroundColor Yellow
    return $false
}

function Check-Version($name, $cmd, $minVer) {
    try {
        $verStr = & $cmd 2>&1 | Select-Object -First 1
        $ver = [System.Version]($verStr -replace '[^0-9.]')
        if ($ver -ge $minVer) {
            Write-Host "  ✅ $name $ver (≥ $minVer)" -ForegroundColor Green
            return $true
        }
        Write-Host "  ⚠️  $name $ver (< $minVer, please upgrade)" -ForegroundColor Yellow
        return $false
    } catch {
        Write-Host "  ⚠️  Cannot detect $name version" -ForegroundColor Yellow
        return $false
    }
}

# ─── Check prerequisites ─────────────────────────────────────────
Write-Host "  ─── Prerequisites ───" -ForegroundColor Magenta
$allGood = $true

# Git
if (-not (Check-Command "git" $true)) { $allGood = $false }

# Go
$goOk = Check-Command "go" $true
if ($goOk) { Check-Version "Go" "go version" [System.Version]"1.21.0" }

# gh (GitHub CLI)
$ghOk = Check-Command "gh" $true
if (-not $ghOk) {
    Write-Host "  Installing GitHub CLI via winget..." -ForegroundColor Yellow
    try {
        winget install --id GitHub.cli --silent --accept-package-agreements 2>&1 | Out-Null
        $env:Path = [Environment]::GetEnvironmentVariable("Path", "User") + ";$env:Path"
        $ghOk = Check-Command "gh" $true
        if (-not $ghOk) { $allGood = $false }
    } catch {
        Write-Host "  ❌ Failed to install gh. Install manually: winget install GitHub.cli" -ForegroundColor Red
        $allGood = $false
    }
}

# rclone
$rcloneOk = Check-Command "rclone" $true
if (-not $rcloneOk) {
    Write-Host "  Installing rclone via winget..." -ForegroundColor Yellow
    try {
        winget install --id Rclone.Rclone --silent --accept-package-agreements 2>&1 | Out-Null
        $env:Path = [Environment]::GetEnvironmentVariable("Path", "User") + ";$env:Path"
        $rcloneOk = Check-Command "rclone" $true
        if (-not $rcloneOk) { $allGood = $false }
    } catch {
        Write-Host "  ❌ Failed to install rclone. Install manually: winget install Rclone.Rclone" -ForegroundColor Red
        $allGood = $false
    }
}

if (-not $allGood) {
    Write-Host ""
    Write-Host "  ❌ Some required dependencies are missing. Fix above and re-run." -ForegroundColor Red
    exit 1
}

# ─── SSH key ──────────────────────────────────────────────────────
Write-Host ""
Write-Host "  ─── SSH Key ───" -ForegroundColor Magenta
if (-not (Test-Path $SshKey)) {
    if (-not (Test-Path $SshDir)) { New-Item -ItemType Directory -Path $SshDir -Force | Out-Null }
    ssh-keygen -t rsa -b 4096 -f $SshKey -N '""' -C "codespaces-auto" 2>&1 | Out-Null
    Write-Host "  ✅ SSH key generated: $SshKey" -ForegroundColor Green
} else {
    Write-Host "  ✅ SSH key exists: $SshKey" -ForegroundColor Green
}

# ─── gh auth ──────────────────────────────────────────────────────
Write-Host ""
Write-Host "  ─── GitHub CLI Authentication ───" -ForegroundColor Magenta
$authed = gh auth status 2>&1 | Select-String "Logged in"
if (-not $authed) {
    Write-Host "  You need to log in to GitHub CLI."
    Write-Host "  Run: gh auth login" -ForegroundColor Yellow
    Write-Host "  Then re-run this script."
} else {
    Write-Host "  ✅ GitHub CLI authenticated." -ForegroundColor Green
}

# ─── Build ────────────────────────────────────────────────────────
Write-Host ""
Write-Host "  ─── Build ───" -ForegroundColor Magenta
try {
    Push-Location $RepoRoot
    go mod download 2>&1 | Out-Null
    go build -o cloud-ide-mount.exe . 2>&1 | Out-Null
    Write-Host "  ✅ Build successful: $RepoRoot\cloud-ide-mount.exe" -ForegroundColor Green
} catch {
    Write-Host "  ❌ Build failed: $_" -ForegroundColor Red
    $allGood = $false
} finally {
    Pop-Location
}

# ─── Tests ────────────────────────────────────────────────────────
Write-Host ""
Write-Host "  ─── Tests ───" -ForegroundColor Magenta
try {
    Push-Location $RepoRoot
    $testOutput = go test -race -count=1 ./... 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✅ All tests pass." -ForegroundColor Green
    } else {
        Write-Host "  ❌ Some tests failed: " -ForegroundColor Red
        Write-Host $testOutput
    }
} catch {
    Write-Host "  ❌ Test error: $_" -ForegroundColor Red
} finally {
    Pop-Location
}

# ─── Summary ─────────────────────────────────────────────────────
Write-Host ""
Write-Host "╔══════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  Setup complete!                         ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""
Write-Host "  Next steps:" -ForegroundColor Magenta
Write-Host "    cs-mount list         # List codespaces"
Write-Host "    cs-mount mount        # Mount a codespace"
Write-Host "    cs-mount --help       # See all commands"
Write-Host ""
