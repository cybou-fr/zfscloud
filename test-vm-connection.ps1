# Quick Test Script - Test Client-Server Connection
# Run this after server is deployed to VM

param(
    [Parameter(Mandatory=$false)]
    [string]$ServerIP = "192.168.56.10",
    
    [Parameter(Mandatory=$false)]
    [int]$ServerPort = 9090
)

Write-Host "ZFSCloud Client-Server Connection Test" -ForegroundColor Cyan
Write-Host "=======================================" -ForegroundColor Cyan
Write-Host ""

# Test 1: Network connectivity
Write-Host "[1/4] Testing network connectivity..." -ForegroundColor Yellow
$pingResult = Test-Connection -ComputerName $ServerIP -Count 2 -Quiet

if ($pingResult) {
    Write-Host "  ✓ VM is reachable at $ServerIP" -ForegroundColor Green
} else {
    Write-Host "  ✗ Cannot ping VM at $ServerIP" -ForegroundColor Red
    Write-Host "  Check VirtualBox network settings" -ForegroundColor Yellow
    exit 1
}

# Test 2: Port connectivity
Write-Host "`n[2/4] Testing port $ServerPort..." -ForegroundColor Yellow
$tcpResult = Test-NetConnection -ComputerName $ServerIP -Port $ServerPort -WarningAction SilentlyContinue

if ($tcpResult.TcpTestSucceeded) {
    Write-Host "  ✓ Port $ServerPort is open and listening" -ForegroundColor Green
} else {
    Write-Host "  ✗ Port $ServerPort is not accessible" -ForegroundColor Red
    Write-Host "  Ensure server is running: sudo ./bin/zfscloud-server" -ForegroundColor Yellow
    exit 1
}

# Test 3: Check if CLI exists
Write-Host "`n[3/4] Checking client binary..." -ForegroundColor Yellow
if (Test-Path "bin/zfscloud-cli.exe") {
    Write-Host "  ✓ CLI client found" -ForegroundColor Green
} else {
    Write-Host "  ✗ CLI client not found" -ForegroundColor Red
    Write-Host "  Run: go build -o bin/zfscloud-cli.exe ./client/cmd" -ForegroundColor Yellow
    exit 1
}

# Test 4: Create test config
Write-Host "`n[4/4] Creating test configuration..." -ForegroundColor Yellow
$testConfig = @"
server:
  address: "${ServerIP}:${ServerPort}"
  
client:
  id: "test-client-$(Get-Random -Maximum 9999)"
  hostname: "$env:COMPUTERNAME"
  
paths:
  backup_dirs:
    - "C:\\Users\\$env:USERNAME\\Documents\\TestBackup"
  exclude_patterns:
    - "*.tmp"
    - "*.log"

storage:
  chunk_size: 4194304

security:
  tls:
    enabled: false
"@

$testConfig | Out-File -FilePath "test-vm-config.yaml" -Encoding UTF8
Write-Host "  ✓ Test config created: test-vm-config.yaml" -ForegroundColor Green

# Create test backup directory
$testDir = "C:\Users\$env:USERNAME\Documents\TestBackup"
if (-not (Test-Path $testDir)) {
    New-Item -Path $testDir -ItemType Directory -Force | Out-Null
    Write-Host "  ✓ Created test directory: $testDir" -ForegroundColor Green
}

# Create test file
$testFile = Join-Path $testDir "test-$(Get-Date -Format 'yyyyMMdd-HHmmss').txt"
"Test backup file created at $(Get-Date)" | Out-File -FilePath $testFile -Encoding UTF8
Write-Host "  ✓ Created test file: $testFile" -ForegroundColor Green

# Summary
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "✓ Connection test PASSED!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "`nReady to test! Try these commands:" -ForegroundColor Yellow
Write-Host ""
Write-Host "  # Test connection" -ForegroundColor White
Write-Host "  .\bin\zfscloud-cli.exe --config test-vm-config.yaml status" -ForegroundColor Cyan
Write-Host ""
Write-Host "  # Run backup" -ForegroundColor White
Write-Host "  .\bin\zfscloud-cli.exe --config test-vm-config.yaml backup" -ForegroundColor Cyan
Write-Host ""
Write-Host "  # List backups" -ForegroundColor White
Write-Host "  .\bin\zfscloud-cli.exe --config test-vm-config.yaml list" -ForegroundColor Cyan
Write-Host ""
Write-Host "  # Or launch GUI" -ForegroundColor White
Write-Host "  .\bin\zfscloud-gui.exe" -ForegroundColor Cyan
Write-Host "  (Configure server: $ServerIP`:$ServerPort)" -ForegroundColor Gray
Write-Host ""
