# Build script for ZFSCloud
# Usage: .\build.ps1 [server|client|all]

param(
    [string]$Target = "all"
)

$ErrorActionPreference = "Stop"

Write-Host "ZFSCloud Build Script" -ForegroundColor Cyan
Write-Host "=====================" -ForegroundColor Cyan

# Create bin directory if it doesn't exist
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

function Build-Server {
    Write-Host "`nBuilding server..." -ForegroundColor Yellow
    go build -o bin/zfscloud-server.exe ./server/cmd
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Server built successfully: bin/zfscloud-server.exe" -ForegroundColor Green
    } else {
        Write-Host "Server build failed!" -ForegroundColor Red
        exit 1
    }
}

function Build-Client {
    Write-Host "`nBuilding client..." -ForegroundColor Yellow
    go build -o bin/zfscloud-client.exe ./client/cmd
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Client built successfully: bin/zfscloud-client.exe" -ForegroundColor Green
    } else {
        Write-Host "Client build failed!" -ForegroundColor Red
        exit 1
    }
}

function Generate-Proto {
    Write-Host "`nGenerating protobuf files..." -ForegroundColor Yellow
    
    # Check if protoc is installed
    $protocExists = Get-Command protoc -ErrorAction SilentlyContinue
    if (-not $protocExists) {
        Write-Host "Warning: protoc not found. Skipping protobuf generation." -ForegroundColor Yellow
        Write-Host "Install Protocol Buffers compiler to generate proto files." -ForegroundColor Yellow
        return
    }
    
    # Find protoc include path
    $protocPath = (Get-Command protoc).Source
    $protocDir = Split-Path -Parent $protocPath
    $includePath = Join-Path (Split-Path -Parent $protocDir) "include"
    
    if (-not (Test-Path $includePath)) {
        Write-Host "Warning: Protobuf include path not found. Skipping generation." -ForegroundColor Yellow
        return
    }
    
    protoc "-I$includePath" -I. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Protobuf files generated successfully" -ForegroundColor Green
    } else {
        Write-Host "Protobuf generation failed!" -ForegroundColor Red
    }
}

# Main build logic
switch ($Target.ToLower()) {
    "server" {
        Build-Server
    }
    "client" {
        Build-Client
    }
    "all" {
        Generate-Proto
        Build-Server
        Build-Client
    }
    "proto" {
        Generate-Proto
    }
    default {
        Write-Host "Unknown target: $Target" -ForegroundColor Red
        Write-Host "Valid targets: server, client, all, proto" -ForegroundColor Yellow
        exit 1
    }
}

Write-Host "`nBuild complete!" -ForegroundColor Green
