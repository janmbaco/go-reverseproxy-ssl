# Upgrade Guide: v2.x → v3.0.0

> **⚠️ MAJOR VERSION UPGRADE**  
> Version 3.0 introduces significant breaking changes with a complete architectural refactoring. This guide provides step-by-step migration instructions.

---

## Overview

**go-reverseproxy-ssl v3.0** is a **major breaking release** that restructures the entire codebase following **Clean Architecture** principles. The changes improve modularity, testability, and maintainability, but require code and configuration updates for existing users.

### Key Changes Summary

| Area | v2.x | v3.0 | Impact |
|------|------|------|--------|
| **Architecture** | Flat structure (`src/configs`, `src/hosts`) | Clean layers (Domain/Application/Infrastructure/Presentation) | 🔴 Breaking |
| **Module Path** | `github.com/janmbaco/go-reverseproxy-ssl` | `github.com/janmbaco/go-reverseproxy-ssl/v3` | 🔴 Breaking |
| **Import Paths** | `src/configs`, `src/hosts`, `src/grpcutil`, `src/sshutil` | `src/domain`, `src/application`, `src/infrastructure`, `src/presentation` | 🔴 Breaking |
| **Configuration** | Same JSON structure | **New field**: `config_ui_port` | 🟡 Minor |
| **New Features** | None | Web-based Configuration UI | 🟢 Addition |
| **Dependencies** | `go-infrastructure v1.2.5` | `go-infrastructure v2.0` | 🔴 Breaking |
| **Virtual Host API** | Direct struct access | Interface-based (`IVirtualHost`) | 🔴 Breaking |
| **Certificate Management** | `src/configs/certs` | `src/infrastructure/certs` | 🔴 Breaking |
| **Entry Point** | `main.go` with manual setup | `infrastructure.NewServerBootstrapper()` | 🔴 Breaking |

---

## Prerequisites

Before upgrading, ensure:

1. **Backup your configuration**: `cp config.json config.json.backup`
2. **Backup your certificates**: `tar -czf certs-backup.tar.gz certs/`
3. **Review this guide completely** before starting
4. **Test in a non-production environment** first
5. **Go 1.25+** installed (if building from source)

---

## Breaking Changes in Detail

### 1. Module Path Change (Go Modules)

**What changed**: Major version bump requires `/v3` suffix in module path per Go modules specification.

#### v2.x Module & Imports

```go
// go.mod
module github.com/janmbaco/go-reverseproxy-ssl

require (
    github.com/janmbaco/go-infrastructure/ v1.2.5
)
```

```go
// main.go
import (
    "github.com/janmbaco/go-reverseproxy-ssl/src/configs"
    "github.com/janmbaco/go-reverseproxy-ssl/src/hosts"
)
```

#### v3.0 Module & Imports

```go
// go.mod
module github.com/janmbaco/go-reverseproxy-ssl/v3

require (
    github.com/janmbaco/go-infrastructure v2.0.0
)
```

```go
// main.go
import (
    "github.com/janmbaco/go-reverseproxy-ssl/v3/src/domain"
    "github.com/janmbaco/go-reverseproxy-ssl/v3/src/infrastructure"
)
```

**Migration Action**:
```bash
# Update go.mod
go get github.com/janmbaco/go-reverseproxy-ssl/v3
go get github.com/janmbaco/go-infrastructure

# Update all imports in your code
find . -name '*.go' -exec sed -i 's|github.com/janmbaco/go-reverseproxy-ssl|github.com/janmbaco/go-reverseproxy-ssl/v3|g' {} +
find . -name '*.go' -exec sed -i 's|github.com/janmbaco/go-infrastructure/|github.com/janmbaco/go-infrastructure|g' {} +

# Tidy dependencies
go mod tidy
```

---

### 2. Package Structure Reorganization

**What changed**: Complete restructure from flat package layout to Clean Architecture layers.

#### v2.x Package Structure

```
src/
├── configs/
│   ├── certs/
│   │   ├── certificatedefs.go
│   │   └── certmanager.go
│   └── config.go
├── hosts/
│   ├── clientcertificatehost.go
│   ├── virtualhost.go
│   ├── webvirtualhost.go
│   ├── grpcvirtualhost.go
│   ├── grpcwebvirtualhost.go
│   ├── grpcjsonvirtualhost.go
│   ├── sshvirtualhost.go
│   ├── ioc/
│   │   └── register.go
│   └── resolver/
│       ├── virtualhost_resolver.go
│       └── virtualhost_resolver_error.go
├── grpcutil/
│   ├── grpcproxy.go
│   ├── grpcwebproxy.go
│   ├── jsoncodec.go
│   └── providers.go
└── sshutil/
    ├── proxy.go
    └── proxy_error.go
```

#### v3.0 Package Structure

```
src/
├── domain/                    # Business entities and interfaces (NEW)
│   ├── config.go
│   ├── interfaces.go
│   ├── virtualhost.go
│   ├── clientcertificatehost.go
│   ├── webvirtualhost.go
│   ├── grpcvirtualhost.go
│   ├── grpcwebvirtualhost.go
│   ├── grpcjsonvirtualhost.go
│   └── sshvirtualhost.go
├── application/               # Use cases and orchestration (NEW)
│   ├── reverse_proxy_configurator.go
│   ├── server_state.go
│   ├── http_utils.go
│   └── hosts_resolver/
│       └── virtualhost_resolver.go
├── infrastructure/            # External concerns (REORGANIZED)
│   ├── bootstrapper.go
│   ├── http_redirector.go
│   ├── certs/
│   │   ├── certificatedefs.go
│   │   └── certmanager.go
│   ├── grpcutil/
│   │   ├── grpc_proxy.go
│   │   ├── grpc_web_proxy.go
│   │   └── json_codec.go
│   └── sshutil/
│       └── ssh_proxy.go
└── presentation/              # UI and API endpoints (NEW)
    ├── config_ui.go
    ├── recovery_handler.go
    ├── static/
    │   ├── css/
    │   └── js/
    └── templates/
        ├── layouts/
        └── pages/
```

**Migration Action**: Update all import paths in your code.

---

### 3. Import Path Mapping

| v2.x Import | v3.0 Import | Notes |
|-------------|-------------|-------|
| `src/configs` | `v3/src/domain` | Config struct moved to domain layer |
| `src/configs/certs` | `v3/src/infrastructure/certs` | Certificate handling is infrastructure |
| `src/hosts` | `v3/src/domain` | Virtual host entities are domain models |
| `src/hosts/resolver` | `v3/src/application/hosts_resolver` | Resolution logic is application layer |
| `src/hosts/ioc` | *(removed)* | Replaced by `go-infrastructure/v2` DI modules |
| `src/grpcutil` | `v3/src/infrastructure/grpcutil` | gRPC utilities are infrastructure |
| `src/sshutil` | `v3/src/infrastructure/sshutil` | SSH utilities are infrastructure |
| *(none)* | `v3/src/presentation` | **NEW**: ConfigUI web interface |

**Example Migration**:

```go
// ❌ v2.x
import (
    "github.com/janmbaco/go-reverseproxy-ssl/src/configs"
    "github.com/janmbaco/go-reverseproxy-ssl/src/configs/certs"
    "github.com/janmbaco/go-reverseproxy-ssl/src/hosts"
    "github.com/janmbaco/go-reverseproxy-ssl/src/hosts/resolver"
)

func main() {
    config := &configs.Config{...}
    vhResolver := resolver.NewVirtualHostResolver(...)
}
```

```go
// ✅ v3.0
import (
    "github.com/janmbaco/go-reverseproxy-ssl/v3/src/domain"
    "github.com/janmbaco/go-reverseproxy-ssl/v3/src/application/hosts_resolver"
    "github.com/janmbaco/go-reverseproxy-ssl/v3/src/infrastructure"
    infraCerts "github.com/janmbaco/go-reverseproxy-ssl/v3/src/infrastructure/certs"
)

func main() {
    config := &domain.Config{...}
    // VirtualHostResolver now requires DI container
    container := dependencyinjection.NewBuilder().MustBuild()
    vhResolver := hosts_resolver.NewVirtualHostResolver(container, logger)
}
```

---

### 4. Configuration File Changes

**What changed**: New optional field `config_ui_port` added.

#### v2.x Configuration

```json
{
  "web_virtual_hosts": [...],
  "ssh_virtual_hosts": [...],
  "grpc_virtual_hosts": [...],
  "grpc_json_virtual_hosts": [...],
  "grpc_web_virtual_hosts": [...],
  "default_host": "www.example.com",
  "reverse_proxy_port": ":443",
  "log_console_level": 4,
  "log_file_level": 4,
  "logs_dir": "./logs"
}
```

#### v3.0 Configuration

```json
{
  "web_virtual_hosts": [...],
  "ssh_virtual_hosts": [...],
  "grpc_virtual_hosts": [...],
  "grpc_json_virtual_hosts": [...],
  "grpc_web_virtual_hosts": [...],
  "default_host": "www.example.com",
  "reverse_proxy_port": ":443",
  "default_server_cert": "",           // Optional: default cert path
  "default_server_key": "",            // Optional: default key path
  "cert_dir": "./certs",               // Optional: cert directory
  "log_console_level": 4,
  "log_file_level": 4,
  "logs_dir": "./logs",
  "config_ui_port": ":8081"            // NEW: ConfigUI port
}
```

**Migration Action**:

1. **Add `config_ui_port`** to your `config.json`:
   ```json
   {
     "config_ui_port": ":8081"
   }
   ```

2. **(Optional)** Specify cert directory and defaults if you want to customize:
   ```json
   {
     "cert_dir": "/app/certs",
     "default_server_cert": "/app/certs/default.pem",
     "default_server_key": "/app/certs/default-key.pem"
   }
   ```

3. **No changes required** for virtual host configurations—structure remains the same.

---

### 5. Bootstrapping and Entry Point

**What changed**: Simplified entry point using new `ServerBootstrapper`.

#### v2.x Entry Point (Custom Setup)

```go
// main.go (v2.x - simplified example)
package main

import (
    "github.com/janmbaco/go-reverseproxy-ssl/src/configs"
    "github.com/janmbaco/go-reverseproxy-ssl/src/hosts/resolver"
)

func main() {
    configFile := flag.String("config", "", "config file")
    flag.Parse()
    
    // Manual setup of all components
    container := setupDependencies()
    config := loadConfig(*configFile)
    certManager := setupCertManager(config)
    vhResolver := resolver.NewVirtualHostResolver(...)
    server := setupServer(config, certManager, vhResolver)
    
    server.Start()
}
```

#### v3.0 Entry Point (Bootstrapper)

```go
// main.go (v3.0)
package main

import (
    "flag"
    "os"
    "github.com/janmbaco/go-reverseproxy-ssl/v3/src/infrastructure"
)

func main() {
    configFile := flag.String("config", "", "config file")
    flag.Parse()
    
    if *configFile == "" {
        os.Stderr.WriteString("You must set a config file!\n")
        flag.PrintDefaults()
        return
    }

    // Single bootstrapper handles all setup
    bootstrapper := infrastructure.NewServerBootstrapper(*configFile)
    bootstrapper.Start()
}
```

**Migration Action**: Replace your entire `main.go` with the v3.0 pattern above.

---

### 6. Dependency Injection Changes

**What changed**: IOC container replaced by `go-infrastructure/v2` modular DI system.

#### v2.x IoC Registration (Manual)

```go
// v2.x: Manual registration in src/hosts/ioc/register.go
package ioc

import (
    "github.com/janmbaco/go-infrastructure/dependencyinjection"
    "github.com/janmbaco/go-reverseproxy-ssl/src/hosts"
)

func RegisterVirtualHosts(container dependencyinjection.Container) {
    register := container.Register()
    register.AsTransient(
        new(*hosts.WebVirtualHost),
        func() *hosts.WebVirtualHost { return &hosts.WebVirtualHost{} },
        nil,
    )
}
```

#### v3.0 IoC Module (go-infrastructure/v2)

```go
// v3.0: Module-based registration
import (
    "github.com/janmbaco/go-infrastructure/dependencyinjection"
    "github.com/janmbaco/go-infrastructure/logs/ioc"
    logsIoc "github.com/janmbaco/go-infrastructure/logs/ioc"
)

container := dependencyinjection.NewBuilder().
    AddModule(logsIoc.NewLogsModule()).
    AddModule(errorsIoc.NewErrorsModule()).
    AddModule(eventsIoc.NewEventsModule()).
    MustBuild()
```

**Migration Action**: 
- Remove manual IoC registration code
- Use pre-built modules from `go-infrastructure/v2`
- Follow the bootstrapper pattern (no custom DI setup needed)

---

### 7. Virtual Host Provider Functions

**What changed**: Provider functions signature and location changed.

#### v2.x Provider

```go
// v2.x: src/hosts/webvirtualhost.go
package hosts

func WebVirtualHostProvider(host *WebVirtualHost) IVirtualHost {
    return host
}
```

#### v3.0 Provider

```go
// v3.0: src/domain/webvirtualhost.go
package domain

func WebVirtualHostProvider(host *WebVirtualHost, logger Logger) IVirtualHost {
    host.logger = logger
    return host
}
```

**Change**: Logger injection added to all providers.

**Migration Action**: If you were calling providers directly, add `logger` parameter:

```go
// ❌ v2.x
vh := hosts.WebVirtualHostProvider(&webVH)

// ✅ v3.0
vh := domain.WebVirtualHostProvider(&webVH, logger)
```

---

### 8. IVirtualHost Interface Changes

**What changed**: New method `EnsureID()` added to `IVirtualHost`.

#### v2.x IVirtualHost

```go
// v2.x
type IVirtualHost interface {
    http.Handler
    GetID() string
    GetFrom() string
    SetURLToReplace()
    GetHostToReplace() string
    GetURLToReplace() string
    GetURL() string
    GetAuthorizedCAs() []string
    GetServerCertificate() CertificateProvider
    GetHostName() string
}
```

#### v3.0 IVirtualHost

```go
// v3.0
type IVirtualHost interface {
    http.Handler
    GetID() string
    GetFrom() string
    SetURLToReplace()
    GetHostToReplace() string
    GetURLToReplace() string
    GetURL() string
    GetAuthorizedCAs() []string
    GetServerCertificate() CertificateProvider
    GetHostName() string
    EnsureID()  // NEW: Auto-generates UUID if missing
}
```

**Migration Action**: If you implemented custom virtual hosts, add `EnsureID()` method:

```go
import "github.com/google/uuid"

func (vh *CustomVirtualHost) EnsureID() {
    if vh.ID == "" {
        vh.ID = uuid.New().String()
    }
}
```

---

### 9. Certificate Management API

**What changed**: Certificate management moved from `src/configs/certs` to `src/infrastructure/certs`.

#### v2.x Certificate Manager

```go
// v2.x
import "github.com/janmbaco/go-reverseproxy-ssl/src/configs/certs"

certManager := certs.NewCertManager(autocertManager)
certManager.AddCertificate("example.com", certificate)
tlsConfig := certManager.GetTLSConfig()
```

#### v3.0 Certificate Manager

```go
// v3.0
import infraCerts "github.com/janmbaco/go-reverseproxy-ssl/v3/src/infrastructure/certs"

certManager := infraCerts.NewCertManager(autocertManager)
certManager.AddCertificate("example.com", certificate)
tlsConfig := certManager.GetTLSConfig()
```

**Migration Action**: Update import path only (API remains the same).

---

### 10. Error Handling Changes (go-infrastructure v2.0 impact)

**What changed**: `go-infrastructure v2.0` removes panic-based error patterns.

See [go-infrastructure UPGRADE-2.0.md](https://github.com/janmbaco/go-infrastructure/blob/master/MIGRATION-V2.md) for details.

**Key Changes**:
- `ErrorDefer` removed → use `ErrorHandler`
- `RequireNotNil` removed → use `ValidateNotNil` (returns error)
- `CheckNilParameter` removed → use manual `nil` checks

**Migration Example**:

```go
// ❌ v2.x (panic-based)
import "github.com/janmbaco/go-infrastructure/errors/errorschecker"

errorschecker.CheckNilParameter(param, "param")

// ✅ v3.0 (error-based)
import "github.com/janmbaco/go-infrastructure/errors/validation"

if err := validation.ValidateNotNil(map[string]interface{}{"param": param}); err != nil {
    return err
}
```

---

## Step-by-Step Migration Guide

### Step 1: Backup Everything

```bash
# Configuration
cp config.json config.json.v2-backup

# Certificates
tar -czf certs-v2-backup.tar.gz certs/

# Binary (if custom build)
cp go-reverseproxy-ssl go-reverseproxy-ssl-v2-backup
```

---

### Step 2: Update Go Module Dependencies

#### If Using as Standalone Application (Docker/Binary)

No action required—Docker images are versioned.

```bash
# Pull new version
docker pull janmbaco/go-reverseproxy-ssl:v3

# OR build from source
git clone https://github.com/janmbaco/go-reverseproxy-ssl.git
cd go-reverseproxy-ssl
git checkout v3.0.0
go build -o go-reverseproxy-ssl .
```

#### If Importing as Library in Your Code

Update `go.mod`:

```bash
# Update to v3
go get github.com/janmbaco/go-reverseproxy-ssl/v3@v3.0.0
go get github.com/janmbaco/go-infrastructure@v2.0.0

# Clean up
go mod tidy
```

---

### Step 3: Update Configuration File

Add new `config_ui_port` field:

```bash
# Using jq (if available)
jq '. + {"config_ui_port": ":8081"}' config.json > config.json.tmp && mv config.json.tmp config.json

# OR manually edit config.json
vi config.json
# Add: "config_ui_port": ":8081"
```

**Verify Configuration**:
```bash
# Check JSON syntax
jq empty config.json && echo "Valid JSON" || echo "Invalid JSON"

# Validate required fields
jq '.web_virtual_hosts, .default_host, .reverse_proxy_port, .config_ui_port' config.json
```

---

### Step 4: Update Import Paths (If Using as Library)

```bash
# Find all Go files importing old paths
grep -r "github.com/janmbaco/go-reverseproxy-ssl" . --include="*.go"

# Replace v2.x imports with v3
find . -name '*.go' -exec sed -i 's|github.com/janmbaco/go-reverseproxy-ssl/src/configs|github.com/janmbaco/go-reverseproxy-ssl/v3/src/domain|g' {} +
find . -name '*.go' -exec sed -i 's|github.com/janmbaco/go-reverseproxy-ssl/src/hosts|github.com/janmbaco/go-reverseproxy-ssl/v3/src/domain|g' {} +
find . -name '*.go' -exec sed -i 's|github.com/janmbaco/go-reverseproxy-ssl/src/grpcutil|github.com/janmbaco/go-reverseproxy-ssl/v3/src/infrastructure/grpcutil|g' {} +
find . -name '*.go' -exec sed -i 's|github.com/janmbaco/go-reverseproxy-ssl/src/sshutil|github.com/janmbaco/go-reverseproxy-ssl/v3/src/infrastructure/sshutil|g' {} +

# Update go-infrastructure imports
find . -name '*.go' -exec sed -i 's|github.com/janmbaco/go-infrastructure/errors/errorschecker|github.com/janmbaco/go-infrastructure/errors/validation|g' {} +
```

---

### Step 5: Update main.go (If Custom Entry Point)

Replace your `main.go` with the new bootstrapper pattern:

```go
package main

import (
    "flag"
    "os"
    "github.com/janmbaco/go-reverseproxy-ssl/v3/src/infrastructure"
)

func main() {
    var configFile = flag.String("config", "", "config file")
    flag.Parse()
    
    if len(*configFile) == 0 {
        os.Stderr.WriteString("You must set a config file!\n")
        flag.PrintDefaults()
        return
    }

    bootstrapper := infrastructure.NewServerBootstrapper(*configFile)
    bootstrapper.Start()
}
```

---

### Step 6: Update Custom Virtual Hosts (If Any)

If you implemented custom virtual host types, add `EnsureID()` method:

```go
import "github.com/google/uuid"

func (vh *MyCustomVirtualHost) EnsureID() {
    if vh.ID == "" {
        vh.ID = uuid.New().String()
    }
}
```

---

### Step 7: Update Error Handling Patterns

Replace panic-based validation with error returns:

```go
// ❌ OLD (v2.x)
import "github.com/janmbaco/go-infrastructure/errors/errorschecker"

errorschecker.CheckNilParameter(param, "param")

// ✅ NEW (v3.0)
import "github.com/janmbaco/go-infrastructure/errors/validation"

if err := validation.ValidateNotNil(map[string]interface{}{"param": param}); err != nil {
    return fmt.Errorf("validation failed: %w", err)
}
```

---

### Step 8: Build and Test

```bash
# Build
go build -o go-reverseproxy-ssl .

# Test configuration loading
./go-reverseproxy-ssl --config config.json &
sleep 3

# Test HTTPS endpoint
curl -k https://localhost:443

# Test ConfigUI (new in v3.0)
curl http://localhost:8081/api/virtualhosts

# Stop test server
killall go-reverseproxy-ssl
```

---

### Step 9: Deploy v3.0

#### Docker Deployment

```bash
# Stop v2.x container
docker stop reverseproxy
docker rm reverseproxy

# Run v3.0 (note new port mapping for ConfigUI)
docker run -d \
  --name reverseproxy \
  -p 80:80 \
  -p 443:443 \
  -p 8081:8081 \
  -v $(pwd)/config:/app/config:ro \
  -v reverseproxy-certs:/app/certs \
  -v reverseproxy-logs:/app/logs \
  janmbaco/go-reverseproxy-ssl:v3 \
  --config /app/config/config.json
```

#### Docker Compose Deployment

Update `docker-compose.yml`:

```yaml
version: "3.9"

services:
  reverseproxy:
    image: janmbaco/go-reverseproxy-ssl:v3  # Changed from v2
    ports:
      - "80:80"
      - "443:443"
      - "8081:8081"  # NEW: ConfigUI port
    volumes:
      - ./config/config.json:/app/config/config.json:ro
      - reverseproxy-certs:/app/certs
      - reverseproxy-logs:/app/logs
```

```bash
docker-compose pull
docker-compose up -d
```

---

### Step 10: Verify Migration

Run these checks to ensure successful migration:

#### 1. Check Service Health

```bash
# Server is running
docker logs reverseproxy | grep "Server started"

# ConfigUI is accessible
curl -s http://localhost:8081/ | grep "Dashboard"

# Virtual hosts loaded
curl -s http://localhost:8081/api/virtualhosts | jq '.virtualHosts | length'
```

#### 2. Test HTTPS Certificate

```bash
# Let's Encrypt cert is valid
echo | openssl s_client -connect www.example.com:443 -servername www.example.com 2>/dev/null | grep "Verify return code: 0"
```

#### 3. Test Backend Proxying

```bash
# HTTP→HTTPS redirect works
curl -I http://www.example.com | grep "301 Moved Permanently"

# HTTPS proxying works
curl -s https://www.example.com | grep "<title>"
```

#### 4. Check Logs for Errors

```bash
# No critical errors
docker logs reverseproxy 2>&1 | grep -i "error\|fatal\|panic"
```

---

## Rollback Plan

If migration fails, revert to v2.x:

### Rollback Steps

```bash
# 1. Stop v3.0 container
docker stop reverseproxy
docker rm reverseproxy

# 2. Restore v2.x configuration
cp config.json.v2-backup config.json

# 3. Restore certificates (if needed)
tar -xzf certs-v2-backup.tar.gz

# 4. Run v2.x container
docker run -d \
  --name reverseproxy \
  -p 80:80 \
  -p 443:443 \
  -v $(pwd)/config:/app/config:ro \
  -v reverseproxy-certs:/app/certs \
  -v reverseproxy-logs:/app/logs \
  janmbaco/go-reverseproxy-ssl:v2.1.2 \
  --config /app/config/config.json

# 5. Verify v2.x is working
curl -I https://www.example.com
```

### Pin to v2.x in go.mod (If Using as Library)

```bash
go get github.com/janmbaco/go-reverseproxy-ssl@v2.1.2
go get github.com/janmbaco/go-infrastructure/@v1.2.5
go mod tidy
```

---

## Common Migration Issues

### Issue 1: Import Path Not Found

**Error**:
```
build cannot find package "github.com/janmbaco/go-reverseproxy-ssl/src/configs"
```

**Solution**:
Update import to v3 path:
```go
import "github.com/janmbaco/go-reverseproxy-ssl/v3/src/domain"
```

---

### Issue 2: ConfigUI Not Accessible

**Error**:
```
curl: (7) Failed to connect to localhost port 8081
```

**Solution**:
1. Ensure `config_ui_port` is set in `config.json`
2. Verify Docker port mapping: `docker ps | grep 8081`
3. Check logs: `docker logs reverseproxy | grep ConfigUI`

---

### Issue 3: Virtual Hosts Missing ID Field

**Error**:
```
Error: virtual host missing ID
```

**Solution**:
Virtual hosts now auto-generate UUIDs. Call `EnsureID()` on load:
```go
for _, vh := range config.WebVirtualHosts {
    vh.EnsureID()
}
```

---

### Issue 4: Logger Injection Missing

**Error**:
```
panic: runtime error: invalid memory address or nil pointer dereference
```

**Solution**:
Virtual host providers now require logger injection:
```go
// Before (v2.x)
vh := WebVirtualHostProvider(host)

// After (v3.0)
vh := WebVirtualHostProvider(host, logger)
```

---

## Migration Checklist

Use this checklist to track your migration progress:

- [ ] **Prerequisites**
  - [ ] Backup configuration file (`config.json`)
  - [ ] Backup certificates directory
  - [ ] Backup custom code (if any)
  - [ ] Read this guide completely

- [ ] **Dependencies**
  - [ ] Update to Go 1.25+ (if building from source)
  - [ ] Update `go.mod` to v3 module path
  - [ ] Run `go mod tidy`

- [ ] **Configuration**
  - [ ] Add `config_ui_port` field
  - [ ] Validate JSON syntax
  - [ ] Test configuration with dry-run (if available)

- [ ] **Code Changes** (if using as library)
  - [ ] Update all import paths to `/v3`
  - [ ] Replace `src/configs` → `src/domain`
  - [ ] Replace `src/hosts` → `src/domain`
  - [ ] Replace `src/grpcutil` → `src/infrastructure/grpcutil`
  - [ ] Replace `src/sshutil` → `src/infrastructure/sshutil`
  - [ ] Update `main.go` to use `ServerBootstrapper`
  - [ ] Add `EnsureID()` to custom virtual hosts
  - [ ] Replace panic-based validation with error returns

- [ ] **Testing**
  - [ ] Build application without errors
  - [ ] Unit tests pass (`go test ./...`)
  - [ ] Integration tests pass
  - [ ] Manual smoke tests (see Step 8)

- [ ] **Deployment**
  - [ ] Deploy to staging/test environment first
  - [ ] Verify HTTPS certificates work
  - [ ] Verify backend proxying works
  - [ ] Verify ConfigUI is accessible
  - [ ] Check logs for errors

- [ ] **Verification**
  - [ ] All virtual hosts responding correctly
  - [ ] Let's Encrypt certificates renewing
  - [ ] ConfigUI dashboard accessible
  - [ ] Backend services reachable
  - [ ] No errors in logs

- [ ] **Production**
  - [ ] Schedule maintenance window
  - [ ] Deploy to production
  - [ ] Monitor for 24-48 hours
  - [ ] Document changes for team

---

## FAQ

### Q: Can I upgrade from v1.x directly to v3.0?

**A**: No. You must first upgrade to v2.x, then to v3.0. v1.x → v2.x is a minor upgrade; v2.x → v3.0 is major.

---

### Q: Will my existing certificates work with v3.0?

**A**: Yes. Let's Encrypt certificates stored in `certs/` directory are fully compatible. Simply mount the same volume.

---

### Q: Do I need to change my configuration file?

**A**: Mostly no. Only add the new `config_ui_port` field. All virtual host configurations remain backward-compatible.

---

### Q: Can I disable the ConfigUI?

**A**: Not recommended, but you can bind to localhost-only:
```json
{
  "config_ui_port": "127.0.0.1:8081"
}
```

To fully disable, omit the field or use an invalid port (not recommended for troubleshooting).

---

### Q: Is v3.0 stable for production?

**A**: Yes. v3.0.0 is production-ready after extensive refactoring and testing. However, always test in staging first.

---

### Q: What's the performance impact of v3.0?

**A**: Minimal. The refactoring improves code quality without sacrificing performance. ConfigUI adds ~5MB memory overhead.

---

### Q: Can I use v2.x and v3.0 side-by-side?

**A**: Yes, using different module paths:
```go
import (
    v2 "github.com/janmbaco/go-reverseproxy-ssl/src/configs"
    v3 "github.com/janmbaco/go-reverseproxy-ssl/v3/src/domain"
)
```

Not recommended for long-term use.

---

## Additional Resources

- **[README.md](./README.md)** - Full v3.0 documentation with feature list and examples
- **[go-infrastructure MIGRATION-V2.md](https://github.com/janmbaco/go-infrastructure/blob/master/MIGRATION-V2.md)** - go-infrastructure v1.x → v2.0 migration
- **[GitHub Releases](https://github.com/janmbaco/go-reverseproxy-ssl/releases)** - Changelog and release notes
- **[Issues](https://github.com/janmbaco/go-reverseproxy-ssl/issues)** - Report migration problems or ask questions

---

## Support

If you encounter issues during migration:

1. **Check this guide** for common issues
2. **Review logs**: `docker logs reverseproxy -f`
3. **Test configuration**: Validate JSON syntax and required fields
4. **Open an issue**: [GitHub Issues](https://github.com/janmbaco/go-reverseproxy-ssl/issues) with:
   - v2.x version you're upgrading from
   - Error messages and logs
   - Configuration file (sanitized)
   - Steps to reproduce

---

**Version**: 3.0.0  
**Migration Guide Date**: November 2025  
**Estimated Migration Time**: 30-60 minutes (simple deployments), 2-4 hours (custom integrations)  
**Breaking Changes**: Yes (major version)  
**Backward Compatibility**: No (requires migration)
