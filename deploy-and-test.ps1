# ZFSCloud Deployment and Testing Script
# This script deploys the server in Docker and tests with the GUI client

param(
    [switch]$Clean,      # Clean existing Docker containers and volumes
    [switch]$NoBuild,    # Skip building binaries
    [switch]$NoServer,   # Skip server deployment
    [switch]$NoClient,   # Skip client launch
    [switch]$Verbose     # Show detailed output
)

$ErrorActionPreference = "Stop"

# Colors
function Write-Step { param($msg) Write-Host "`n==> $msg" -ForegroundColor Cyan }
function Write-Success { param($msg) Write-Host "[OK] $msg" -ForegroundColor Green }
function Write-Error { param($msg) Write-Host "[ERROR] $msg" -ForegroundColor Red }
function Write-Info { param($msg) Write-Host "  $msg" -ForegroundColor Yellow }

Write-Host "===============================================================" -ForegroundColor Cyan
Write-Host "         ZFSCloud Deployment & Testing Script" -ForegroundColor Cyan
Write-Host "===============================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "This script will:" -ForegroundColor Yellow
Write-Host "  1. Build server (Linux) and client (Windows) binaries" -ForegroundColor Yellow
Write-Host "  2. Deploy ZFS-enabled server in Docker" -ForegroundColor Yellow
Write-Host "  3. Wait for server to be ready" -ForegroundColor Yellow
Write-Host "  4. Launch GUI client for testing" -ForegroundColor Yellow
Write-Host ""

# Check prerequisites
Write-Step "Checking prerequisites..."

# Check Docker
try {
    $dockerVersion = docker --version
    Write-Success "Docker found: $dockerVersion"
} catch {
    Write-Error "Docker not found! Please install Docker Desktop."
    Write-Info "Download from: https://www.docker.com/products/docker-desktop"
    exit 1
}

# Check Docker running
try {
    docker ps | Out-Null
    Write-Success "Docker is running"
} catch {
    Write-Error "Docker is not running! Please start Docker Desktop."
    exit 1
}

# Check docker-compose
try {
    $composeVersion = docker-compose --version
    Write-Success "Docker Compose found: $composeVersion"
} catch {
    Write-Error "Docker Compose not found!"
    exit 1
}

# Check Go
if (-not $NoBuild) {
    try {
        $goVersion = go version
        Write-Success "Go found: $goVersion"
    } catch {
        Write-Error "Go not found! Please install Go 1.21+."
        exit 1
    }
}

# Clean existing deployment
if ($Clean) {
    Write-Step "Cleaning existing deployment..."
    
    Set-Location docker
    
    Write-Info "Stopping containers..."
    docker-compose down -v 2>$null | Out-Null
    
    Write-Info "Removing images..."
    docker rmi zfscloud-server zfscloud-admin 2>$null | Out-Null
    
    Write-Success "Cleanup complete"
    Set-Location ..
}

# Build binaries
if (-not $NoBuild) {
    Write-Step "Building binaries..."
    
    # Build server for Linux
    Write-Info "Building server for Linux (Docker)..."
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "0"
    
    go build -o bin/zfscloud-server ./server/cmd
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Server build failed!"
        exit 1
    }
    
    $serverInfo = Get-Item bin/zfscloud-server
    Write-Success "Server built: $([math]::Round($serverInfo.Length/1MB, 2)) MB"
    
    # Build GUI client for Windows
    Write-Info "Building GUI client for Windows..."
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    
    go build -o bin/zfscloud-gui.exe ./client/cmd
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Client build failed!"
        exit 1
    }
    
    $clientInfo = Get-Item bin/zfscloud-gui.exe
    Write-Success "GUI client built: $([math]::Round($clientInfo.Length/1MB, 2)) MB"
} else {
    Write-Info "Skipping build (using existing binaries)"
}

# Deploy server
if (-not $NoServer) {
    Write-Step "Deploying server in Docker..."
    
    Set-Location docker
    
    # Build and start container
    Write-Info "Building Docker image..."
    docker-compose build zfscloud-server
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Docker build failed!"
        Set-Location ..
        exit 1
    }
    
    Write-Info "Starting server container..."
    docker-compose up -d zfscloud-server
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to start server!"
        Set-Location ..
        exit 1
    }
    
    Write-Success "Server container started"
    
    # Wait for server to be ready
    Write-Info "Waiting for server to initialize..."
    
    $maxAttempts = 30
    $attempt = 0
    $ready = $false
    
    while ($attempt -lt $maxAttempts -and -not $ready) {
        Start-Sleep -Seconds 2
        $attempt++
        
        # Check if container is running
        $containerStatus = docker inspect -f '{{.State.Running}}' zfscloud-server 2>$null
        
        if ($containerStatus -eq "true") {
            # Check for "Server initialized" in logs
            $logs = docker logs zfscloud-server 2>&1
            if ($logs -match "Server initialized successfully|Server listening on") {
                $ready = $true
                Write-Success "Server is ready! (took $($attempt * 2) seconds)"
            } else {
                Write-Host "." -NoNewline
            }
        } else {
            Write-Error "Container stopped unexpectedly!"
            Write-Info "Logs:"
            docker logs zfscloud-server
            Set-Location ..
            exit 1
        }
    }
    
    if (-not $ready) {
        Write-Error "Server failed to start within $($maxAttempts * 2) seconds"
        Write-Info "Check logs with: docker logs zfscloud-server"
        Set-Location ..
        exit 1
    }
    
    # Show server status
    Write-Step "Server Status"
    Write-Info "Container status:"
    docker ps --filter "name=zfscloud-server" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    
    Write-Info "`nZFS pool status:"
    docker exec zfscloud-server zpool status tank 2>&1 | ForEach-Object { Write-Host "  $_" }
    
    Write-Info "`nServer logs (last 10 lines):"
    docker logs --tail 10 zfscloud-server 2>&1 | ForEach-Object { Write-Host "  $_" }
    
    Set-Location ..
} else {
    Write-Info "Skipping server deployment"
}

# Create client configuration
Write-Step "Creating client configuration..."

$configPath = "client-config.yaml"

if (Test-Path $configPath) {
    Write-Info "Client config already exists: $configPath"
} else {
    $defaultConfig = @"
client:
  client_id: "test-client-1"
  client_name: "Test Client"
  server_address: "localhost:50051"
  watch_directory: "$($env:USERPROFILE)\Documents"
  upload_interval: "5m"
  private_key_path: "./client-key.pem"

backup:
  compression: true
  deduplication: true
  batch_size: 100
"@
    
    $defaultConfig | Out-File -FilePath $configPath -Encoding UTF8
    Write-Success "Created client config: $configPath"
    Write-Info "Edit this file to customize settings"
}

# Launch GUI client
if (-not $NoClient) {
    Write-Step "Launching GUI client..."
    
    if (-not (Test-Path "bin/zfscloud-gui.exe")) {
        Write-Error "GUI client not found! Run without -NoClient"
        exit 1
    }
    
    Write-Info "Starting zfscloud-gui.exe..."
    Write-Info "The GUI will open in a new window"
    Write-Info ""
    Write-Info "To test the system:"
    Write-Info "  1. Go to Settings"
    Write-Info "  2. Configure server: localhost:50051"
    Write-Info "  3. Click 'Test Connection'"
    Write-Info "  4. Click 'Save Settings'"
    Write-Info "  5. Go to Dashboard"
    Write-Info "  6. Click 'Start Backup'"
    
    # Launch GUI in background
    Start-Process -FilePath "bin\zfscloud-gui.exe" -WorkingDirectory $PWD
    
    Write-Success "GUI client launched!"
} else {
    Write-Info "Skipping client launch"
    Write-Info "Run manually: .\bin\zfscloud-gui.exe"
}

# Show next steps
Write-Step "Deployment Complete!"

Write-Host ""
Write-Host "===============================================================" -ForegroundColor Green
Write-Host "                      Next Steps" -ForegroundColor Green
Write-Host "===============================================================" -ForegroundColor Green
Write-Host ""
Write-Host "Server:" -ForegroundColor Cyan
Write-Host "  - Running at: localhost:50051" -ForegroundColor White
Write-Host "  - View logs: docker logs -f zfscloud-server" -ForegroundColor White
Write-Host "  - Check ZFS: docker exec zfscloud-server zfs list" -ForegroundColor White
Write-Host ""
Write-Host "Client:" -ForegroundColor Cyan
Write-Host "  - GUI should be running" -ForegroundColor White
Write-Host "  - Config file: client-config.yaml" -ForegroundColor White
Write-Host "  - Test connection in Settings tab" -ForegroundColor White
Write-Host ""
Write-Host "Testing Workflow:" -ForegroundColor Cyan
Write-Host "  1. Configure client settings" -ForegroundColor White
Write-Host "  2. Test connection" -ForegroundColor White
Write-Host "  3. Start backup from Dashboard" -ForegroundColor White
Write-Host "  4. Monitor statistics" -ForegroundColor White
Write-Host "  5. Browse files in Files tab" -ForegroundColor White
Write-Host "  6. Test restore functionality" -ForegroundColor White
Write-Host ""
Write-Host "Management:" -ForegroundColor Cyan
Write-Host "  - Stop: docker-compose -f docker/docker-compose.yml down" -ForegroundColor White
Write-Host "  - Clean: .\deploy-and-test.ps1 -Clean" -ForegroundColor White
Write-Host ""

if ($Verbose) {
    Write-Step "Detailed Information"
    
    Write-Info "Server environment:"
    docker exec zfscloud-server printenv | Select-String "ZFSCLOUD|TZ"
    
    Write-Info "`nServer filesystem:"
    docker exec zfscloud-server df -h /tank 2>&1
    
    Write-Info "`nDatabase info:"
    docker exec zfscloud-server sqlite3 /var/lib/zfscloud/zfscloud.db ".tables"
}

Write-Host "`n[OK] All done! Happy testing!`n" -ForegroundColor Cyan
