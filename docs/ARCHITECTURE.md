# Architecture Overview

Internal architecture and design decisions for Go ReverseProxy SSL v3.0.

> **Note**: Diagrams in this document use Mermaid syntax. For best viewing, use a Markdown viewer that supports Mermaid diagrams (GitHub, GitLab, VS Code with Mermaid extension, etc.).

## Clean Architecture

Version 3.0 follows **Clean Architecture** principles with clear separation of concerns:

```
┌────────────────────────────────────────────────────────┐
│                   Presentation Layer                   │
│  (ConfigUI, HTTP Handlers, API Endpoints)              │
└───────────────────┬────────────────────────────────────┘
                    │
┌───────────────────▼────────────────────────────────────┐
│                  Application Layer                     │
│  (ReverseProxyConfigurator, ServerState, HostResolver) │
└───────────────────┬────────────────────────────────────┘
                    │
┌───────────────────▼────────────────────────────────────┐
│                    Domain Layer                        │
│  (Config, VirtualHost, Interfaces, Business Logic)     │
└───────────────────┬────────────────────────────────────┘
                    │
┌───────────────────▼────────────────────────────────────┐
│                Infrastructure Layer                    │
│  (CertManager, HTTP Server, Bootstrapper, gRPC Utils)  │
└────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

#### Domain Layer (`internal/domain/`)
- **Pure business logic** with no external dependencies
- **Core entities**: `Config`, `VirtualHost`, `ClientCertificateHost`
- **Interfaces**: Define contracts for infrastructure (e.g., `IVirtualHost`, `ICertificateProvider`)
- **Immutable data structures** where possible
- **No I/O operations**, only data transformations

**Key Files**:
- `config.go` - Configuration data structures
- `virtualhost.go` - Base virtual host interfaces
- `webvirtualhost.go` - Web virtual host definition
- `grpcwebvirtualhost.go` - gRPC-Web virtual host definition
- `interfaces.go` - Domain interfaces and contracts

#### Application Layer (`internal/application/`)
- **Orchestrates business logic** using domain entities
- **Stateful services**: `ReverseProxyConfigurator`, `ServerState`
- **Coordinates** between domain and infrastructure
- **No direct I/O**, delegates to infrastructure

**Key Components**:
- `reverse_proxy_configurator.go` - Configures HTTP reverse proxies for each virtual host
- `server_state.go` - Manages server lifecycle and state
- `hosts_resolver/` - DNS/hostname resolution logic
- `http_utils.go` - HTTP utilities and helpers

#### Infrastructure Layer (`internal/infrastructure/`)
- **External concerns**: File I/O, HTTP server, TLS, database (future)
- **Implements domain interfaces** (e.g., certificate management)
- **Handles real-world I/O** and third-party integrations
- **Dependency Injection** setup and bootstrapping

**Key Components**:
- `server_bootstrapper.go` - Application entry point, DI container setup
- `certificate_manager.go` - Let's Encrypt integration, certificate storage
- `http_server.go` - HTTP/HTTPS server implementation
- `config_loader.go` - JSON configuration file parsing
- `certificates/` - Certificate management utilities
- `grpcutil/` - gRPC proxy utilities

#### Presentation Layer (`internal/presentation/`)
- **User-facing interfaces**: ConfigUI web interface, REST API
- **HTTP handlers** for virtual host management
- **HTML templates** and static assets
- **API endpoints** for configuration management

**Key Components**:
- `config_ui.go` - Web UI server implementation
- `config_handlers.go` - Configuration API handlers
- `recovery_middleware.go` - Error recovery middleware
- `templates/` - HTML templates for web UI
- `static/` - CSS, JavaScript, images

---

## Request Flow

### HTTPS Request Lifecycle

```
1. Client Request
   ↓
2. TLS Termination (Infrastructure Layer)
   ↓
3. Host Resolution (Application Layer)
   ↓
4. Virtual Host Lookup (Domain Layer)
   ↓
5. Proxy Selection (Application Layer)
   ↓
6. Backend Request (Infrastructure Layer)
   ↓
7. Response Handling
   ↓
8. Client Response
```

**Detailed Flow**:

1. **Client Request**: `https://example.com/api/users`

2. **TLS Termination** (`http_server.go`):
   - Retrieve certificate for `example.com` from cache or Let's Encrypt
   - Decrypt TLS connection
   - Extract SNI (Server Name Indication)

3. **Host Resolution** (`reverse_proxy_configurator.go`):
   - Extract hostname from request: `example.com`
   - Check for client certificate if required

4. **Virtual Host Lookup** (Domain Layer):
   - Find matching `VirtualHost` in configuration
   - Return backend target (scheme, host, port, path)

5. **Proxy Selection** (`reverse_proxy_configurator.go`):
   - Retrieve or create `httputil.ReverseProxy` for backend
   - Configure proxy director (URL rewriting, headers)

6. **Backend Request** (`httputil.ReverseProxy`):
   - Forward request to backend: `http://backend-service:8080/api/users`
   - Apply custom headers if configured
   - Add `X-Forwarded-*` headers

7. **Response Handling**:
   - Add custom response headers
   - Apply logging
   - Error handling (502 if backend unreachable)

8. **Client Response**:
   - Send response over encrypted TLS connection

---

## Certificate Management

### Automatic Certificates (Let's Encrypt)

```
┌─────────────┐
│ HTTP :80    │ ← ACME Challenge Request
└──────┬──────┘
       │
       ▼
┌─────────────────────┐
│ Certificate Manager │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐     ┌──────────────┐
│ Let's Encrypt ACME  │ ←→  │ Certificate  │
└─────────────────────┘     │ Cache (Disk) │
                            └──────────────┘
```



**Certificate Acquisition Flow**:

1. **First HTTPS Request** for `example.com`
2. **Cache Miss**: Certificate not found in `/app/certs/example.com/`
3. **ACME Challenge**:
   - HTTP-01 challenge: Serve token on `http://example.com/.well-known/acme-challenge/`
   - Let's Encrypt validates ownership
4. **Certificate Issued**: Saved to `cert_dir/example.com/cert.pem`
5. **Cache Update**: Certificate cached in memory for fast access
6. **TLS Handshake**: Continue with newly issued certificate

**Certificate Renewal**:
- Automatic renewal 30 days before expiration
- Background goroutine checks certificates daily
- Renewal uses same ACME process
- Zero-downtime: New cert loaded without restart

### Custom Certificates

- User-provided PEM certificates
- Bypass Let's Encrypt entirely
- Useful for:
  - Internal/private CAs
  - Wildcard certificates
  - EV (Extended Validation) certificates
  - Development (self-signed)

---

## Dependency Injection

Powered by `github.com/janmbaco/go-infrastructure` DI container.

### Module System

```go
// Example: Certificate Manager Module
type CertManagerModule struct{}

func (m *CertManagerModule) Configure(builder *dependencyinjection.Builder) {
    builder.RegisterSingleton(
        "ICertificateManager",
        func(container *dependencyinjection.Container) interface{} {
            config := container.Resolve("Config").(*domain.Config)
            logger := container.Resolve("ILogger").(logs.ILogger)
            return NewCertificateManager(config.CertDir, logger)
        },
    )
}
```

**Benefits**:
- **Testability**: Easy to mock dependencies
- **Modularity**: Swap implementations without changing consumers
- **Lifecycle Management**: Singleton, transient, scoped lifetimes
- **Type Safety**: Compile-time checks for dependencies

### Dependency Graph

```
ServerBootstrapper
  ├── Config (singleton)
  ├── Logger (singleton)
  ├── CertificateManager (singleton)
  │     └── FileSystem (singleton)
  ├── ReverseProxyConfigurator (singleton)
  │     ├── Config
  │     ├── Logger
  │     └── CertificateManager
  ├── HTTPServer (singleton)
  │     ├── ReverseProxyConfigurator
  │     └── CertificateManager
  └── ConfigUIServer (singleton)
        ├── Config
        └── Logger
```

---

## Concurrency Model

### Goroutine Usage

1. **HTTP Server**: One goroutine per request (Go standard library)
2. **Certificate Renewal**: Background goroutine checking expiration daily
3. **ConfigUI Server**: Separate HTTP server on different port

### Thread Safety

- **Config**: Immutable after initial load (v3.0), read-only
- **Certificate Cache**: `sync.RWMutex` for concurrent read/write
- **Proxy Map**: Built once on startup, read-only afterward
- **Logging**: Thread-safe logger implementation



**Future Enhancement (v3.1)**:
- Live configuration reload
- Atomic config swap with `atomic.Value`
- Graceful virtual host updates without restart

---

## Virtual Host Routing

### Routing Algorithm

```go
func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 1. Extract hostname from request
    hostname := extractHostname(r.Host)
    
    // 2. Lookup proxy for hostname
    proxy, exists := s.proxyMap[hostname]
    if !exists {
        // 3. Fallback to default host
        proxy = s.proxyMap[s.config.DefaultHost]
    }
    
    // 4. Forward to proxy
    proxy.ServeHTTP(w, r)
}
```

### Routing Flow Diagram


**Matching Rules** (in order):
1. **Exact domain match**: `example.com`
2. **Subdomain match** (future): `*.example.com`
3. **Path-based match**: `example.com/api` vs `example.com/web`
4. **Default host**: Configured fallback

### Proxy Creation

Each virtual host gets its own `httputil.ReverseProxy`:

```go
func createProxy(vh *VirtualHost) *httputil.ReverseProxy {
    target := &url.URL{
        Scheme: vh.Scheme,
        Host:   fmt.Sprintf("%s:%d", vh.HostName, vh.Port),
    }
    
    proxy := httputil.NewSingleHostReverseProxy(target)
    
    // Custom director
    proxy.Director = func(req *http.Request) {
        req.URL.Scheme = target.Scheme
        req.URL.Host = target.Host
        req.URL.Path = singleJoiningSlash(vh.Path, req.URL.Path)
        
        // Add custom headers
        for k, v := range vh.ResponseHeaders {
            req.Header.Set(k, v)
        }
        
        // X-Forwarded-* headers
        req.Header.Set("X-Forwarded-For", clientIP(req))
        req.Header.Set("X-Forwarded-Proto", "https")
        req.Header.Set("X-Forwarded-Host", req.Host)
    }
    
    return proxy
}
```

---

## Protocol Support

### HTTP/2

- Enabled by default for HTTPS connections
- Negotiated via ALPN (Application-Layer Protocol Negotiation)
- Fallback to HTTP/1.1 if client doesn't support HTTP/2

### gRPC-Web

- Browser-compatible gRPC
- Translates gRPC-Web protocol to native gRPC
- Enables gRPC from JavaScript/TypeScript clients


---

## Error Handling

### Error Propagation

```
Domain Layer:      Return errors (no panics)
    ↓
Application Layer: Wrap errors with context
    ↓
Infrastructure:    Log errors, return HTTP status
    ↓
Presentation:      User-friendly error messages
```


**Example**:
```go
// Domain: Validation error
func (c *Config) Validate() error {
    if c.DefaultHost == "" {
        return errors.New("default_host is required")
    }
    return nil
}

// Application: Add context
func (cfg *Configurator) LoadConfig() error {
    if err := config.Validate(); err != nil {
        return fmt.Errorf("config validation failed: %w", err)
    }
    return nil
}

// Infrastructure: Log and return status
func (s *Server) Start() error {
    if err := s.configurator.LoadConfig(); err != nil {
        s.logger.Error("Failed to load config", err)
        return err
    }
    return nil
}
```

---



### Benchmarks

Typical performance (on AWS t3.medium, 2 vCPU, 4GB RAM):

- **Throughput**: ~10,000 req/sec (simple proxying)
- **Latency**: <5ms added latency (proxy overhead)
- **Memory**: ~50MB base + ~5MB per 1000 concurrent connections
- **TLS Handshake**: <50ms (cached cert), <200ms (new cert from Let's Encrypt)

### Scaling

- **Vertical**: Add more CPU/memory to single instance
- **Horizontal**: Run multiple instances behind L4 load balancer (HAProxy, AWS ALB)
- **Recommendation**: Multiple instances for high availability

---

## See Also

- [CONFIGURATION.md](CONFIGURATION.md) - Complete configuration reference
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Common issues and solutions
- [UPGRADE-3.0.md](../UPGRADE-3.0.md) - Migration guide from v2.x
