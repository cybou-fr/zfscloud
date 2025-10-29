# Build script for Docker deployment
# This builds the server binary for Linux to run in Docker

Write-Host "Building ZFSCloud Server for Linux (Docker)..." -ForegroundColor Green

# Set environment for Linux build
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "1"

# Build server
Write-Host "Building server binary..." -ForegroundColor Yellow
go build -o bin/zfscloud-server ./server/cmd

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Server built successfully: bin/zfscloud-server" -ForegroundColor Green
    
    # Show file info
    if (Test-Path "bin/zfscloud-server") {
        $fileInfo = Get-Item "bin/zfscloud-server"
        Write-Host "  Size: $($fileInfo.Length / 1MB) MB" -ForegroundColor Cyan
    }
} else {
    Write-Host "✗ Server build failed!" -ForegroundColor Red
    exit 1
}

# Reset environment
$env:GOOS = "windows"
$env:GOARCH = "amd64"

Write-Host "`nNow you can run:" -ForegroundColor Green
Write-Host "  cd docker" -ForegroundColor Cyan
Write-Host "  docker-compose up -d" -ForegroundColor Cyan
