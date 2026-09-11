# Install plx from the latest GitHub release, then run `plx setup`.
# Usage: irm https://raw.githubusercontent.com/xldplx/plx/main/install.ps1 | iex
$ErrorActionPreference = "Stop"

$Repo = if ($env:PLX_REPO) { $env:PLX_REPO } else { "xldplx/plx" }
$BinDir = if ($env:PLX_INSTALL_DIR) { $env:PLX_INSTALL_DIR } else {
    Join-Path $env:LOCALAPPDATA "plx\bin"
}

function Get-PlxArch {
    switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64" { return "x86_64" }
        "ARM64" { return "arm64" }
        default {
            throw "plx install: unsupported architecture: $($env:PROCESSOR_ARCHITECTURE)"
        }
    }
}

$goarch = Get-PlxArch
$asset = "plx_Windows_${goarch}.zip"

Write-Host "plx install: resolving latest release for $Repo…"
# Prefer the releases/latest redirect so we do not need the GitHub API.
$latest = Invoke-WebRequest -Uri "https://github.com/$Repo/releases/latest" -MaximumRedirection 0 -ErrorAction SilentlyContinue
if ($null -eq $latest -or -not $latest.Headers.Location) {
    # PowerShell may follow redirects; fall back to AbsoluteUri.
    $resp = Invoke-WebRequest -Uri "https://github.com/$Repo/releases/latest" -UseBasicParsing
    $tag = ($resp.BaseResponse.ResponseUri.AbsoluteUri -split "/")[-1]
} else {
    $tag = ($latest.Headers.Location -split "/")[-1]
}
if (-not $tag -or $tag -eq "latest") {
    throw "plx install: could not determine latest release tag"
}

$downloadUrl = "https://github.com/$Repo/releases/download/$tag/$asset"
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("plx-install-" + [guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
    $zipPath = Join-Path $tmp $asset
    Write-Host "plx install: downloading $asset ($tag)…"
    Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -UseBasicParsing

    Expand-Archive -Path $zipPath -DestinationPath $tmp -Force
    $binary = Join-Path $tmp "plx.exe"
    if (-not (Test-Path $binary)) {
        throw "plx install: archive did not contain plx.exe"
    }

    New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
    $dest = Join-Path $BinDir "plx.exe"
    Copy-Item -Path $binary -Destination $dest -Force

    $pathParts = @($env:PATH -split ";" | Where-Object { $_ -ne "" })
    if ($pathParts -notcontains $BinDir) {
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if (-not $userPath) { $userPath = "" }
        $userParts = @($userPath -split ";" | Where-Object { $_ -ne "" })
        if ($userParts -notcontains $BinDir) {
            [Environment]::SetEnvironmentVariable("Path", (($userParts + $BinDir) -join ";"), "User")
        }
        $env:PATH = "$BinDir;$env:PATH"
        Write-Host "plx install: added $BinDir to your user PATH"
    }

    Write-Host "plx install: installed $dest ($tag)"
    Write-Host "plx install: running plx setup…"
    & $dest setup
    Write-Host "plx install: done"
}
finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
