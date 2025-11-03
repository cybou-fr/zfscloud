# Cross-compile ZFSCloud Server for Linux
Write-Host "Building ZFSCloud Server for Linux..." -ForegroundColor Green

$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

Write-Host "Compiling server binary..." -ForegroundColor Yellow
go build -o bin/zfscloud-server ./server/cmd

if ($LASTEXITCODE -eq 0) {
    $fileSize = (Get-Item bin/zfscloud-server).Length / 1MB
    Write-Host "Server binary built successfully!" -ForegroundColor Green
    Write-Host "Size: $([math]::Round($fileSize, 2)) MB" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Transfer to VM: scp bin/zfscloud-server user@vm-ip:~/" -ForegroundColor Yellow
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}

Remove-Item Env:\GOOS
Remove-Item Env:\GOARCH
Remove-Item Env:\CGO_ENABLED
Write-Host "Environment reset." -ForegroundColor Green
