#!/bin/bash

# Config UI Integration Test Runner
# This script runs only the Config UI integration tests

set -e

TIMEOUT_MINUTES=${1:-2}

echo "🚀 Config UI Integration Test Runner"
echo "===================================="
echo "Timeout: $TIMEOUT_MINUTES minutes"

# Set overall timeout
START_TIME=$(date +%s)
TIMEOUT_TIME=$((START_TIME + TIMEOUT_MINUTES * 60))

# Function to check if timeout has been exceeded
timeout_exceeded() {
    CURRENT_TIME=$(date +%s)
    [ $CURRENT_TIME -gt $TIMEOUT_TIME ]
}

# Function to check timeout and exit if exceeded
check_timeout() {
    OPERATION=${1:-"operation"}
    if timeout_exceeded; then
        echo "✗ Timeout exceeded during $OPERATION"
        exit 1
    fi
}

# Function to validate file existence
validate_file() {
    FILE_PATH=$1
    DESCRIPTION=${2:-"file"}
    if [ ! -f "$FILE_PATH" ]; then
        echo "✗ $DESCRIPTION not found: $FILE_PATH"
        return 1
    fi
    echo "✓ $DESCRIPTION found: $FILE_PATH"
    return 0
}

# Function to generate certificates
generate_certificates() {
    echo "✓ Generating certificates..."

    CERT_DIR="./integration/testdata"

    # Remove old certificates if they exist
    rm -f "$CERT_DIR/localhost-cert.pem"
    rm -f "$CERT_DIR/localhost-key.pem"

    # Generate new certificates using OpenSSL in Docker
    if ! docker run --rm -v "$(pwd)/integration/testdata:/certs" alpine/openssl req -x509 -newkey rsa:2048 -nodes -keyout /certs/localhost-key.pem -out /certs/localhost-cert.pem -days 365 -subj "/C=ES/ST=Development/L=LocalDev/O=Development/CN=localhost" -addext "subjectAltName=DNS:localhost,DNS:*.localhost" >/dev/null 2>&1; then
        echo "✗ Failed to generate certificates"
        return 1
    fi

    echo "✓ Certificates generated successfully in $CERT_DIR"
    return 0
}

# Function to validate configuration
validate_configuration() {
    echo "✓ Validating configuration..."

    CONFIG_FILE="./integration/testdata/config.json"
    if ! validate_file "$CONFIG_FILE" "Integration test config"; then
        return 1
    fi

    # Generate certificates
    if ! generate_certificates; then
        return 1
    fi

    return 0
}

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    echo "✗ Docker is not running. Please start Docker first."
    exit 1
fi

check_timeout "initialization"

# Validate configuration and generate certificates
if ! validate_configuration; then
    exit 1
fi

check_timeout "configuration validation"

# Build the application
check_timeout "Docker build"
echo "✓ Building application..."
if ! docker build -f build/package/Dockerfile -t go-reverseproxy-ssl:test .; then
    echo "✗ Failed to build Docker image"
    exit 1
fi

check_timeout "application build"

# Start all services with Docker Compose
echo "✓ Starting all services with Docker Compose..."
if ! docker-compose -f integration/docker-compose.test.yml up -d; then
    echo "✗ Failed to start services"
    exit 1
fi

# Wait for reverse proxy to be ready
echo "✓ Waiting for reverse proxy to be ready..."
TIMEOUT=30
COUNTER=0
while true; do
    if curl -k -s --max-time 5 https://localhost/ >/dev/null 2>&1; then
        echo "✓ Reverse proxy is ready!"
        break
    fi

    if [ $COUNTER -ge $TIMEOUT ]; then
        echo "✗ Service failed to start within $TIMEOUT seconds"
        docker-compose -f integration/docker-compose.test.yml logs reverse-proxy
        docker-compose -f integration/docker-compose.test.yml down
        exit 1
    fi

    COUNTER=$((COUNTER + 1))
    sleep 1
done

check_timeout "reverse proxy ready"

# Wait for config UI to be ready
echo "✓ Waiting for config UI to be ready..."
COUNTER=0
while true; do
    if curl -s --max-time 5 http://localhost:8081/ >/dev/null 2>&1; then
        echo "✓ Config UI is ready!"
        break
    fi

    if [ $COUNTER -ge $TIMEOUT ]; then
        echo "✗ Config UI failed to start within $TIMEOUT seconds"
        docker-compose -f integration/docker-compose.test.yml logs reverse-proxy
        docker-compose -f integration/docker-compose.test.yml down
        exit 1
    fi

    COUNTER=$((COUNTER + 1))
    sleep 1
done

check_timeout "config UI ready"

echo "✓ Services are ready! Running Config UI integration tests..."
echo "═══════════════════════════════════════════════════════════════"

# Run Config UI integration tests
check_timeout "Config UI integration tests"
export INTEGRATION_TEST=true
go test -v ./integration/config_ui_integration_test.go
EXIT_CODE=$?

echo "═══════════════════════════════════════════════════════════════"
if [ $EXIT_CODE -eq 0 ]; then
    echo "✓ All Config UI integration tests passed!"
else
    echo "✗ Some Config UI integration tests failed"
fi

# Cleanup
echo "✓ Cleaning up..."
docker-compose -f integration/docker-compose.test.yml down

# Remove generated certificates
rm -f "./integration/testdata/localhost-cert.pem"
rm -f "./integration/testdata/localhost-key.pem"

if [ $EXIT_CODE -eq 0 ]; then
    echo "✓ Config UI integration test run completed successfully"
else
    echo "✗ Config UI integration test run failed"
    exit $EXIT_CODE
fi