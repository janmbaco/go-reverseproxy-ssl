# Scripts Directory

This directory contains automation scripts for the Go Reverse Proxy SSL project.

## Integration Test Scripts

### `run-integration-tests.sh` (Linux/macOS)
Bash script to run integration tests with Docker.

**Requirements:**
- Docker running
- Go 1.26+ installed
- curl command available

**Usage:**
```bash
./scripts/run-integration-tests.sh
```

### `run-integration-tests.ps1` (Windows)
PowerShell script to run integration tests with Docker.

**Requirements:**
- Docker running
- Go 1.26+ installed
- PowerShell 5.1+ or PowerShell Core

**Usage:**
```powershell
.\scripts\run-integration-tests.ps1
```

## What the scripts do:

1. **Check Docker**: Verify Docker is running
2. **Build**: Build the application Docker image
3. **Start Service**: Run the reverse proxy in a container
4. **Wait for Ready**: Wait for the service to be healthy
5. **Run Tests**: Execute integration tests with `INTEGRATION_TEST=true`
6. **Cleanup**: Stop and remove the test container

## Manual execution

Run the integration stack from the repository root. It creates fresh configuration and certificate volumes from the checked-in sample.

```bash
docker compose -p reverseproxy-manual -f integration/docker-compose.test.yml up --build -d
INTEGRATION_TEST=true go test -count=1 -v ./integration/...
docker compose -p reverseproxy-manual -f integration/docker-compose.test.yml down --volumes
```
