# Integration Tests for Config UI

This directory contains integration tests for the Go Reverse Proxy SSL configuration UI.

## Available Tests

### Config UI Integration Tests (`config_ui_integration_test.go`)

Tests the configuration UI API endpoints and HTML pages:

- **TestConfigUI_HTMLPages**: Verifies that HTML pages load correctly and contain expected elements
- **TestConfigUI_VirtualHostsAPI**: Tests CRUD operations for virtual hosts (GET, POST, DELETE)
- **TestConfigUI_CertificatesAPI**: Tests certificate upload functionality
- **TestConfigUI_ServerStatus**: Verifies that the config UI server is running and responding

## Running Tests

### Run All Integration Tests

```bash
# Using PowerShell (Windows)
.\scripts\run-integration-tests.ps1

# Using Bash (Linux/macOS)
./scripts/run-integration-tests.sh
```

### Run Only Config UI Tests

```bash
# Set environment variable and run specific test file
export INTEGRATION_TEST=true
go test -v ./integration/config_ui_integration_test.go
```

### Run Specific Test Function

```bash
export INTEGRATION_TEST=true
go test -v ./integration/config_ui_integration_test.go -run TestConfigUI_HTMLPages
```

## Test Requirements

- Docker must be running
- Docker Compose must be available
- Environment variable `INTEGRATION_TEST=true` must be set
- Test certificates will be auto-generated

## Test Infrastructure

The tests use Docker Compose to spin up:

- **Backend services**: Nginx containers on ports 8080-8083
- **gRPC backend**: Test gRPC service
- **Reverse proxy**: Main application with config UI on port 8081

## Expected Test Flow

1. **Setup**: Generate certificates and build Docker images
2. **Infrastructure**: Start all backend services and reverse proxy
3. **Wait**: Ensure reverse proxy is ready (HTTPS on port 443)
4. **Test**: Run HTTP requests against config UI (port 8081)
5. **Cleanup**: Stop containers and remove test certificates

## Test Data

Test configuration is loaded from `integration/testdata/config.json` which includes:

- Web virtual hosts for different backend services
- gRPC virtual hosts with transparent and selective proxying
- Config UI running on port 8081
- SSL certificates for localhost testing