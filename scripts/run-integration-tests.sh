#!/bin/bash

# Integration Test Runner for Go Reverse Proxy SSL
# This script helps run integration tests locally on Linux/macOS
#
# For Windows, use: .\scripts\run-integration-tests.ps1
#
# Prerequisites:
# - Docker and Docker Compose running
# - Go 1.25+ installed
# - curl command available

set -e

# Configuration
TIMEOUT_MINUTES=${1:-10}
TIMEOUT_SECONDS=$((TIMEOUT_MINUTES * 60))
START_TIME=$(date +%s)

echo "🚀 Go Reverse Proxy SSL - Integration Test Runner"
echo "================================================="
echo "⏱️  Timeout: $TIMEOUT_MINUTES minutes"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_info() {
    echo -e "${CYAN}ℹ${NC} $1"
}

# Function to check if timeout has been exceeded
check_timeout() {
    local current_time=$(date +%s)
    local elapsed=$((current_time - START_TIME))
    if [ $elapsed -ge $TIMEOUT_SECONDS ]; then
        print_error "Timeout exceeded during $1"
        cleanup
        exit 1
    fi
}

# Function to validate file existence
validate_file() {
    local filepath=$1
    local description=$2
    if [ ! -f "$filepath" ]; then
        print_error "$description not found: $filepath"
        return 1
    fi
    print_success "$description found: $filepath"
    return 0
}

# Function to generate certificates
generate_certificates() {
    print_success "Generating certificates..."
    
    local cert_dir="./integration/testdata"
    
    # Remove old certificates if they exist
    rm -f "$cert_dir/localhost-cert.pem" "$cert_dir/localhost-key.pem"
    
    # Generate new certificates using OpenSSL in Docker
    if docker run --rm \
        -v "$(pwd)/integration/testdata:/certs" \
        alpine/openssl req -x509 -newkey rsa:2048 -nodes \
        -keyout /certs/localhost-key.pem \
        -out /certs/localhost-cert.pem \
        -days 365 \
        -subj "/C=ES/ST=Development/L=LocalDev/O=Development/CN=localhost" \
        -addext "subjectAltName=DNS:localhost,DNS:*.localhost" 2>/dev/null; then
        
        # Fix permissions (make readable by all users)
        chmod 644 "$cert_dir/localhost-key.pem" 2>/dev/null || true
        chmod 644 "$cert_dir/localhost-cert.pem" 2>/dev/null || true
        
        print_success "Certificates generated successfully in $cert_dir"
        return 0
    else
        print_error "Failed to generate certificates"
        return 1
    fi
}

# Function to validate configuration
validate_configuration() {
    print_success "Validating configuration..."
    
    local config_file="./integration/testdata/config.json"
    if ! validate_file "$config_file" "Integration test config"; then
        return 1
    fi
    
    # Generate certificates
    if ! generate_certificates; then
        return 1
    fi
    
    return 0
}

# Cleanup function
cleanup() {
    print_success "Cleaning up..."
    docker-compose -f integration/docker-compose.test.yml down 2>/dev/null || true
    
    # Remove generated certificates
    rm -f "./integration/testdata/localhost-cert.pem" 2>/dev/null || true
    rm -f "./integration/testdata/localhost-key.pem" 2>/dev/null || true
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker first."
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
print_success "Building application..."
if ! docker build -f build/package/Dockerfile -t go-reverseproxy-ssl:test .; then
    print_error "Failed to build Docker image"
    exit 1
fi

check_timeout "application build"

# Start all services with Docker Compose
print_success "Starting all services with Docker Compose..."
if ! docker-compose -f integration/docker-compose.test.yml up -d; then
    print_error "Failed to start services"
    exit 1
fi

# Wait for reverse proxy to be ready
print_success "Waiting for reverse proxy to be ready..."
timeout=30
counter=0
while true; do
    if curl -k --max-time 5 https://localhost/ > /dev/null 2>&1; then
        print_success "Reverse proxy is ready!"
        break
    fi

    if [ $counter -ge $timeout ]; then
        print_error "Service failed to start within $timeout seconds"
        docker-compose -f integration/docker-compose.test.yml logs reverse-proxy
        exit 1
    fi

    counter=$((counter + 1))
    sleep 1
done

check_timeout "reverse proxy ready"

print_success "Service is ready! Running integration tests..."
echo -e "${CYAN}================================================================${NC}"

# Run integration tests
check_timeout "integration tests"
export INTEGRATION_TEST=true
if go test -v ./integration/...; then
    exit_code=0
else
    exit_code=1
fi

echo -e "${CYAN}================================================================${NC}"
if [ $exit_code -eq 0 ]; then
    print_success "All integration tests passed!"
    print_success "Integration test run completed successfully"
else
    print_error "Some integration tests failed"
    print_error "Integration test run failed"
fi

exit $exit_code