param(
  [string]$Version = "dev"
)

$ErrorActionPreference = "Stop"
$distDir = "dist"

if (Test-Path $distDir) { Remove-Item -Recurse -Force $distDir }
New-Item -ItemType Directory -Path $distDir | Out-Null

$targets = @(
  @{ OS = "windows"; Arch = "amd64"; Ext = ".exe" }
  @{ OS = "linux";   Arch = "amd64"; Ext = ""     }
  @{ OS = "darwin";  Arch = "amd64"; Ext = ""     }
  @{ OS = "darwin";  Arch = "arm64"; Ext = ""     }
)

$ldflags = "-s -w -X github.com/yemresalcan/mr-ducksafe/internal/cmd.Version=$Version"

foreach ($t in $targets) {
  $out = Join-Path $distDir "mr-ducksafe-$($t.OS)-$($t.Arch)$($t.Ext)"
  Write-Host "Building $out" -ForegroundColor Cyan
  $env:GOOS = $t.OS
  $env:GOARCH = $t.Arch
  $env:CGO_ENABLED = "0"
  & go build -ldflags $ldflags -o $out ./cmd/mr-ducksafe
  if ($LASTEXITCODE -ne 0) { throw "Build failed for $($t.OS)/$($t.Arch)" }
}

# Build hostundaki env'i temizle
$env:GOOS = ""
$env:GOARCH = ""

Write-Host ""
Write-Host "Done. Artifacts in $distDir/" -ForegroundColor Green
Get-ChildItem $distDir | Select-Object Name, @{N="Size(MB)";E={[math]::Round($_.Length/1MB, 2)}}
