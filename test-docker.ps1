# Quick testing script for client-server interaction

Write-Host "=== ZFSCloud Client-Server Test ===" -ForegroundColor Green
Write-Host ""

# Check if server binary exists
if (-not (Test-Path "bin/zfscloud-server")) {
    Write-Host "Server binary not found. Building for Linux..." -ForegroundColor Yellow
    .\build-docker.ps1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Build failed!" -ForegroundColor Red
        exit 1
    }
}

# Check if client binary exists
if (-not (Test-Path "bin/zfscloud-client.exe")) {
    Write-Host "Client binary not found. Building..." -ForegroundColor Yellow
    go build -o bin/zfscloud-client.exe ./client/cmd
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Client build failed!" -ForegroundColor Red
        exit 1
    }
}

Write-Host "✓ Binaries ready" -ForegroundColor Green
Write-Host ""

# Start Docker containers
Write-Host "Starting Docker containers..." -ForegroundColor Yellow
Push-Location docker
docker-compose up -d zfscloud-server
Pop-Location

if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Failed to start Docker containers!" -ForegroundColor Red
    exit 1
}

Write-Host "✓ Server started" -ForegroundColor Green
Write-Host ""

# Wait for server to be ready
Write-Host "Waiting for server to initialize..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Show server logs
Write-Host ""
Write-Host "=== Server Logs ===" -ForegroundColor Cyan
Push-Location docker
docker-compose logs --tail=20 zfscloud-server
Pop-Location

Write-Host ""
Write-Host "=== Test Instructions ===" -ForegroundColor Green
Write-Host ""
Write-Host "1. Run the client:" -ForegroundColor Cyan
Write-Host "   .\bin\zfscloud-client.exe" -ForegroundColor White
Write-Host ""
Write-Host "2. Register:" -ForegroundColor Cyan
Write-Host "   > register" -ForegroundColor White
Write-Host ""
Write-Host "3. In another terminal, approve the registration:" -ForegroundColor Cyan
Write-Host "   cd docker" -ForegroundColor White
Write-Host "   docker exec zfscloud-server sqlite3 /var/lib/zfscloud/zfscloud.db \"SELECT registration_id FROM registrations WHERE status='pending';\"" -ForegroundColor White
Write-Host "   # Copy the registration ID, then:" -ForegroundColor DarkGray
Write-Host "   docker-compose run --rm zfscloud-admin /scripts/approve-registration.sh <registration-id> 10" -ForegroundColor White
Write-Host ""
Write-Host "4. Back in client, send heartbeat:" -ForegroundColor Cyan
Write-Host "   > heartbeat" -ForegroundColor White
Write-Host ""
Write-Host "5. Upload a file:" -ForegroundColor Cyan
Write-Host "   echo 'Test data' > test.txt" -ForegroundColor White
Write-Host "   > upload test.txt /test.txt" -ForegroundColor White
Write-Host ""
Write-Host "6. List files:" -ForegroundColor Cyan
Write-Host "   > list" -ForegroundColor White
Write-Host ""
Write-Host "=== Useful Commands ===" -ForegroundColor Green
Write-Host ""
Write-Host "View server logs:" -ForegroundColor Cyan
Write-Host "  cd docker && docker-compose logs -f zfscloud-server" -ForegroundColor White
Write-Host ""
Write-Host "Check ZFS status:" -ForegroundColor Cyan
Write-Host "  docker exec zfscloud-server zpool status tank" -ForegroundColor White
Write-Host "  docker exec zfscloud-server zfs list -r tank" -ForegroundColor White
Write-Host ""
Write-Host "Stop server:" -ForegroundColor Cyan
Write-Host "  cd docker && docker-compose down" -ForegroundColor White
Write-Host ""
