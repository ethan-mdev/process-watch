param(
    [string]$Version = "dev"
)

$ProjectName = "process-watch"
$OutputDir = "dist"

# Build targets: @("goos/goarch", ...)
$BuildTargets = @(
    "linux/amd64",
    "darwin/amd64",
    "darwin/arm64",
    "windows/amd64"
)

# Create dist directory if it doesn't exist
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

Write-Host "Building $ProjectName v$Version..." -ForegroundColor Green
Write-Host ""

foreach ($target in $BuildTargets) {
    $goos, $goarch = $target.Split("/")
    
    # Determine output filename
    $output = if ($goos -eq "windows") {
        Join-Path $OutputDir "$ProjectName-$goos-$goarch.exe"
    }
    else {
        Join-Path $OutputDir "$ProjectName-$goos-$goarch"
    }
    
    Write-Host "Building for $goos/$goarch -> $output"
    
    $env:GOOS = $goos
    $env:GOARCH = $goarch
    
    & go build -o $output .
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Build failed for $target" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "✓ Build complete! Binaries in $OutputDir/:" -ForegroundColor Green
Get-ChildItem -Path "$OutputDir/$ProjectName-*" -File | ForEach-Object {
    $size = [Math]::Round($_.Length / 1MB, 2)
    Write-Host "  $($_.Name) ($size MB)"
}
