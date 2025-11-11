# Config UI Integration Test Runner
# This script runs only the Config UI integration tests

param(
    [int]$TimeoutMinutes = 2
)

Write-Host "🚀 Config UI Integration Test Runner" -ForegroundColor Green
Write-Host "====================================" -ForegroundColor Green
Write-Host "Timeout: $TimeoutMinutes minutes" -ForegroundColor Yellow

# Set overall timeout
$startTime = Get-Date
$timeoutTime = $startTime.AddMinutes($TimeoutMinutes)

# Function to check if timeout has been exceeded
function Test-TimeoutExceeded {
    return (Get-Date) -gt $timeoutTime
}

# Function to check timeout and exit if exceeded
function Test-TimeoutAndExit {
    param([string]$operation = "operation")
    if (Test-TimeoutExceeded) {
        Write-Host "✗ Timeout exceeded during $operation" -ForegroundColor Red
        exit 1
    }
}

# Function to validate file existence
function Test-FileExists {
    param([string]$filePath, [string]$description = "file")
    if (-not (Test-Path $filePath)) {
        Write-Host "✗ $description not found: $filePath" -ForegroundColor Red
        return $false
    }
    Write-Host "✓ $description found: $filePath" -ForegroundColor Green
    return $true
}

# Function to generate certificates
function New-Certificates {
    Write-Host "✓ Generating certificates..." -ForegroundColor Green

    $certDir = ".\integration\testdata"

    # Remove old certificates if they exist
    Remove-Item "$certDir\localhost-cert.pem" -ErrorAction SilentlyContinue
    Remove-Item "$certDir\localhost-key.pem" -ErrorAction SilentlyContinue

    # Generate new certificates using OpenSSL in Docker
    try {
        docker run --rm -v "${PWD}/integration/testdata:/certs" alpine/openssl req -x509 -newkey rsa:2048 -nodes -keyout /certs/localhost-key.pem -out /certs/localhost-cert.pem -days 365 -subj "/C=ES/ST=Development/L=LocalDev/O=Development/CN=localhost" -addext "subjectAltName=DNS:localhost,DNS:*.localhost" 2>$null
        if ($LASTEXITCODE -ne 0) {
            Write-Host "✗ Failed to generate certificates" -ForegroundColor Red
            return $false
        }

        # Fix permissions (make readable by all users)
        $keyFile = Join-Path $certDir "localhost-key.pem"
        $certFile = Join-Path $certDir "localhost-cert.pem"

        if (Test-Path $keyFile) {
            icacls $keyFile /grant "Everyone:(R)" /T /C | Out-Null
        }
        if (Test-Path $certFile) {
            icacls $certFile /grant "Everyone:(R)" /T /C | Out-Null
        }

        Write-Host "✓ Certificates generated successfully in $certDir" -ForegroundColor Green
        return $true
    } catch {
        Write-Host "✗ Failed to generate certificates: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

# Function to validate configuration
function Test-Configuration {
    Write-Host "✓ Validating configuration..." -ForegroundColor Green

    $configFile = ".\integration\testdata\config.json"
    if (-not (Test-FileExists -filePath $configFile -description "Integration test config")) {
        return $false
    }

    # Generate certificates
    if (-not (New-Certificates)) {
        return $false
    }

    return $true
}

# Check if Docker is running
try {
    docker info | Out-Null
} catch {
    Write-Host "✗ Docker is not running. Please start Docker first." -ForegroundColor Red
    exit 1
}

Test-TimeoutAndExit "initialization"

# Validate configuration and generate certificates
if (-not (Test-Configuration)) {
    exit 1
}

Test-TimeoutAndExit "configuration validation"

# Build the application
Test-TimeoutAndExit "Docker build"
Write-Host "✓ Building application..." -ForegroundColor Green
docker build -f build/package/Dockerfile -t go-reverseproxy-ssl:test .
if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Failed to build Docker image" -ForegroundColor Red
    exit 1
}

Test-TimeoutAndExit "application build"

# Start all services with Docker Compose
Write-Host "✓ Starting all services with Docker Compose..." -ForegroundColor Green
docker-compose -f integration/docker-compose.test.yml up -d
if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Failed to start services" -ForegroundColor Red
    exit 1
}

# Wait for reverse proxy to be ready
Write-Host "✓ Waiting for reverse proxy to be ready..." -ForegroundColor Green
$timeout = 30
$counter = 0
while ($true) {
    try {
        Invoke-WebRequest -Uri "https://localhost/" -SkipCertificateCheck -TimeoutSec 5 | Out-Null
        Write-Host "✓ Reverse proxy is ready!" -ForegroundColor Green
        break
    } catch {
        if ($counter -ge $timeout) {
            Write-Host "✗ Service failed to start within $timeout seconds" -ForegroundColor Red
            docker-compose -f integration/docker-compose.test.yml logs reverse-proxy
            docker-compose -f integration/docker-compose.test.yml down
            exit 1
        }
        $counter++
        Start-Sleep -Seconds 1
    }
}

Test-TimeoutAndExit "reverse proxy ready"

# Wait for config UI to be ready
Write-Host "✓ Waiting for config UI to be ready..." -ForegroundColor Green
$counter = 0
while ($true) {
    try {
        Invoke-WebRequest -Uri "http://localhost:8081/" -TimeoutSec 5 | Out-Null
        Write-Host "✓ Config UI is ready!" -ForegroundColor Green
        break
    } catch {
        if ($counter -ge $timeout) {
            Write-Host "✗ Config UI failed to start within $timeout seconds" -ForegroundColor Red
            docker-compose -f integration/docker-compose.test.yml logs reverse-proxy
            docker-compose -f integration/docker-compose.test.yml down
            exit 1
        }
        $counter++
        Start-Sleep -Seconds 1
    }
}

Test-TimeoutAndExit "config UI ready"

Write-Host "✓ Services are ready! Running Config UI integration tests..." -ForegroundColor Green
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan

# Run Config UI integration tests
Test-TimeoutAndExit "Config UI integration tests"
$env:INTEGRATION_TEST = "true"
go test -v ./integration/config_ui_integration_test.go
$exitCode = $LASTEXITCODE

Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
if ($exitCode -eq 0) {
    Write-Host "✓ All Config UI integration tests passed!" -ForegroundColor Green
} else {
    Write-Host "✗ Some Config UI integration tests failed" -ForegroundColor Red
}

# Cleanup
Write-Host "✓ Cleaning up..." -ForegroundColor Green
docker-compose -f integration/docker-compose.test.yml down

# Remove generated certificates
Remove-Item "$PWD/integration/testdata/localhost-cert.pem" -ErrorAction SilentlyContinue
Remove-Item "$PWD/integration/testdata/localhost-key.pem" -ErrorAction SilentlyContinue

if ($exitCode -eq 0) {
    Write-Host "✓ Config UI integration test run completed successfully" -ForegroundColor Green
} else {
    Write-Host "✗ Config UI integration test run failed" -ForegroundColor Red
    exit $exitCode
}