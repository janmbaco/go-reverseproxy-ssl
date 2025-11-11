# Go ReverseProxy SSL

[![Go Report Card](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)](https://goreportcard.com/report/github.com/janmbaco/go-reverseproxy-ssl/v3)
[![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-GPL%20v3.0-green.svg)](LICENSE)
[![Release](https://img.shields.io/badge/release-v3.0.0-blue.svg)](https://github.com/janmbaco/go-reverseproxy-ssl/releases/tag/v3.0.0)

**Add automatic HTTPS to any web app, API, or gRPC service in 2 minutes.** No Nginx config, no manual certificates — just point your domain and run.

## What It Does

You have a web app running on `localhost:3000`. You want:
- ✓ **Automatic HTTPS** with Let's Encrypt certificates
- ✓ **No configuration hell** (just a simple JSON file)
- ✓ **Multiple apps/APIs** on different subdomains
- ✓ **gRPC and SSH** support with the same certificates
- ✓ **Web UI** to manage virtual hosts

This reverse proxy handles all of that automatically.

```
Your App (localhost:3000) + This Proxy = https://yourdomain.com (automatic cert)
```

### When to Use This

**✅ Primary use case:**
- VPS/dedicated server (DigitalOcean, Linode, Hetzner, OVH)
- Cloud VM (AWS EC2, Google Compute Engine, Azure Virtual Machines)
- Self-hosted server at home or office
- Docker Compose setup with multiple services
- Full control over ports 80 and 443

**✅ Advanced use case (centralized reverse proxy):**
- Unified entry point for multiple PaaS backends (Heroku, Vercel, Azure App Service, etc.)
- Custom domain management with Let's Encrypt for all services
- Route `api.yourdomain.com` → Heroku app, `web.yourdomain.com` → Vercel, etc.

---

## Quick Start

### Prerequisites

1. **Your own server** (VPS or cloud VM with root access)
2. **Domain name** pointed to your server's IP (e.g., `example.com` → `203.0.113.10`)
3. **Ports 80 and 443** open in your firewall
4. **Docker** (recommended) or Go 1.25+

### Example 1: Single Web Application

You have a Node.js API running on port 3000 in your DigitalOcean droplet. Want it accessible via `https://myapp.com` with automatic SSL:

**Using Docker with environment variables:**

```bash
docker run -d \
  --name reverseproxy \
  -p 80:80 -p 443:443 -p 8081:8081 \
  -e DOMAIN=myapp.com \
  -e BACKEND_HOST=localhost \
  -e BACKEND_PORT=3000 \
  -e BACKEND_SCHEME=http \
  -v reverseproxy-certs:/app/certs \
  --network host \
  janmbaco/go-reverseproxy-ssl:v3
```

**That's it!** Visit `https://myapp.com` — certificate is automatically issued by Let's Encrypt.

**How it works:**
1. DNS `myapp.com` → Your VPS IP (203.0.113.10)
2. Proxy listens on ports 80/443
3. Let's Encrypt validates via HTTP-01 challenge (port 80)
4. Certificate issued and cached in `/app/certs`
5. Proxy forwards HTTPS requests to `http://localhost:3000`

**Platform-specific notes:**
- **DigitalOcean/Linode Droplet**: Use `--network host` or backend service IP
- **AWS EC2 / Azure VM**: Ensure ports 80/443 in security groups/NSG
- **Docker Compose**: Replace `localhost` with service name (see Example 3)

---

### Example 2: Multiple Subdomains (ENV vars only)

You have three services and want separate subdomains with automatic HTTPS — **no config.json needed**:
- React app on port 3000 → `www.myapp.com`
- Express API on port 8080 → `api.myapp.com`
- Admin dashboard on port 5000 → `app.myapp.com`

**Using Docker with environment variables:**

```bash
docker run -d \
  --name reverseproxy \
  -p 80:80 -p 443:443 -p 8081:8081 \
  -e WWW_DOMAIN=www.myapp.com \
  -e WWW_BACKEND=localhost:3000 \
  -e API_DOMAIN=api.myapp.com \
  -e API_BACKEND=localhost:8080 \
  -e APP_DOMAIN=app.myapp.com \
  -e APP_BACKEND=localhost:5000 \
  -v reverseproxy-certs:/app/certs \
  --network host \
  janmbaco/go-reverseproxy-ssl:v3
```

**Result:**
- `https://www.myapp.com` → React app (automatic cert)
- `https://api.myapp.com` → Express API (automatic cert)
- `https://app.myapp.com` → Admin dashboard (automatic cert)

**Backend format:**
- Basic: `host:port` (defaults to HTTP)
- With scheme: `host:port:https` (for HTTPS backends)

**Examples:**
```bash
-e WWW_BACKEND=frontend-service:3000           # HTTP backend
-e API_BACKEND=api.herokuapp.com:443:https    # HTTPS backend
-e APP_BACKEND=172.17.0.1:5000                # Docker bridge IP
```

---

### Example 3: Multiple Services (Docker Compose)

Same setup as Example 2, but using Docker Compose with custom `config.json` for advanced features:

You have:
- Frontend on port 3000 → `www.myapp.com`
- API on port 8080 → `api.myapp.com`
- Admin panel on port 5000 → `admin.myapp.com`

**docker-compose.yml:**

```yaml
version: "3.9"

services:
  reverseproxy:
    image: janmbaco/go-reverseproxy-ssl:v3
    ports:
      - "80:80"
      - "443:443"
      - "8081:8081"
    volumes:
      - ./config.json:/app/config/config.json:ro
      - certs:/app/certs
    networks: [app]

  frontend:
    image: my-frontend:latest
    expose: ["3000"]
    networks: [app]

  api:
    image: my-api:latest
    expose: ["8080"]
    networks: [app]

  admin:
    image: my-admin:latest
    expose: ["5000"]
    networks: [app]

networks:
  app:
volumes:
  certs:
```

**config.json:**

```json
{
  "web_virtual_hosts": [
    {
      "from": "www.myapp.com",
      "scheme": "http",
      "host_name": "frontend",
      "port": 3000
    },
    {
      "from": "api.myapp.com",
      "scheme": "http",
      "host_name": "api",
      "port": 8080
    },
    {
      "from": "admin.myapp.com",
      "scheme": "http",
      "host_name": "admin",
      "port": 5000
    }
  ],
  "default_host": "www.myapp.com",
  "reverse_proxy_port": ":443",
  "config_ui_port": ":8081"
}
```

Run: `docker-compose up -d`

**Result:**
- `https://www.myapp.com` → frontend (automatic cert)
- `https://api.myapp.com` → API (automatic cert)
- `https://admin.myapp.com` → admin panel (automatic cert)

---

### Example 4: gRPC-Web Service

Backend gRPC service on port 9090, expose as `https://grpc.myapp.com` for browser access:

**Simple config (transparent mode - proxies all services/methods):**

```json
{
  "grpc_web_virtual_hosts": [
    {
      "from": "grpc.myapp.com",
      "scheme": "http",
      "host_name": "grpc-service",
      "port": 9090,
      "grpc_web_proxy": {
        "is_transparent_server": true,
        "allow_all_origins": true
      }
    }
  ],
  "default_host": "grpc.myapp.com",
  "reverse_proxy_port": ":443"
}
```

**Advanced config (selective service/method proxying):**

```json
{
  "grpc_web_virtual_hosts": [
    {
      "from": "grpc.myapp.com",
      "scheme": "http",
      "host_name": "grpc-service",
      "port": 9090,
      "grpc_web_proxy": {
        "grpc_services": {
          "hello.HelloService": ["SayHello", "ServerStreamingChat"],
          "user.UserService": ["GetUser", "UpdateUser"]
        },
        "is_transparent_server": false,
        "authority": "grpc-service:9090",
        "allow_all_origins": false,
        "allowed_origins": ["https://myapp.com"]
      }
    }
  ],
  "default_host": "grpc.myapp.com",
  "reverse_proxy_port": ":443"
}
```

**JavaScript gRPC-Web client:**

```javascript
const client = new GreeterClient('https://grpc.myapp.com');
```

> **Note**: The `grpc_web_proxy` field is **required** for gRPC-Web virtual hosts. Set `is_transparent_server: true` to proxy all services automatically, or set it to `false` and specify `grpc_services` for fine-grained control.

---

## Web Configuration UI

Instead of editing JSON, use the built-in web interface:

1. **Access UI**: `http://localhost:8081` (or your server IP:8081)
2. **Add virtual host**: Click "New Virtual Host"
3. **Fill form**:
   - Domain: `example.com`
   - Backend Host: `my-app`
   - Backend Port: `3000`
   - Protocol: `http`
4. **Save** → Configuration updated automatically

**UI Features:**
- Visual virtual host management
- Certificate upload for custom certs
- Configuration preview (JSON)
- Live validation

> **Security**: Bind to `127.0.0.1:8081` for localhost-only access, or put behind authentication.

---

## Common Use Cases

### Use Case 1: "I have a React dev server on localhost:3000"

**Minimal config:**

```json
{
  "web_virtual_hosts": [
    {
      "from": "myapp.com",
      "scheme": "http",
      "host_name": "localhost",
      "port": 3000
    }
  ],
  "default_host": "myapp.com",
  "reverse_proxy_port": ":443"
}
```

**Docker command:**

```bash
docker run -d \
  -p 80:80 -p 443:443 \
  -e DOMAIN=myapp.com \
  -e BACKEND_HOST=host.docker.internal \
  -e BACKEND_PORT=3000 \
  -v certs:/app/certs \
  janmbaco/go-reverseproxy-ssl:v3
```

**Verify:**
```bash
curl -I https://myapp.com  # Should return 200 with Let's Encrypt cert
```

---

### Use Case 2: "I have API + Frontend in Docker Compose"

**Scenario**: Express API (port 8080) + Next.js (port 3000)

**docker-compose.yml:**

```yaml
services:
  proxy:
    image: janmbaco/go-reverseproxy-ssl:v3
    ports: ["80:80", "443:443"]
    volumes:
      - ./config.json:/app/config/config.json:ro
      - certs:/app/certs
    networks: [backend]

  api:
    build: ./api
    expose: ["8080"]
    networks: [backend]

  web:
    build: ./web
    expose: ["3000"]
    networks: [backend]

networks:
  backend:
volumes:
  certs:
```

**config.json:**

```json
{
  "web_virtual_hosts": [
    {"from": "api.example.com", "host_name": "api", "port": 8080, "scheme": "http"},
    {"from": "www.example.com", "host_name": "web", "port": 3000, "scheme": "http"}
  ],
  "default_host": "www.example.com",
  "reverse_proxy_port": ":443"
}
```

**Result:**
- Frontend: `https://www.example.com`
- API: `https://api.example.com`
- Both with automatic HTTPS

---

### Use Case 3: "I need custom SSL certificate (not Let's Encrypt)"

**When you need this:**
- Internal/private CA
- Wildcard certificate
- EV (Extended Validation) certificate
- Development with self-signed cert

**config.json:**

```json
{
  "web_virtual_hosts": [
    {
      "from": "secure.internal.com",
      "scheme": "http",
      "host_name": "backend",
      "port": 8080,
      "server_certificate": {
        "certificate_path": "/certs/internal.pem",
        "private_key_path": "/certs/internal-key.pem"
      }
    }
  ],
  "default_host": "secure.internal.com",
  "reverse_proxy_port": ":443"
}
```

**Docker mount:**

```bash
docker run -d \
  -p 80:80 -p 443:443 \
  -v ./certs:/certs:ro \
  -v ./config.json:/app/config/config.json:ro \
  janmbaco/go-reverseproxy-ssl:v3
```

**Alternative: Using environment variables with custom certificates:**

```bash
docker run -d \
  --name reverseproxy \
  -p 80:80 -p 443:443 -p 8081:8081 \
  -e DOMAIN=secure.internal.com \
  -e BACKEND_HOST=backend \
  -e BACKEND_PORT=8080 \
  -e BACKEND_SCHEME=http \
  -e CERT_FILE=/certs/internal.pem \
  -e KEY_FILE=/certs/internal-key.pem \
  -v ./certs:/certs:ro \
  -v certs-storage:/app/certs \
  janmbaco/go-reverseproxy-ssl:v3
```

> **Note**: When using custom certificates with ENV vars, mount your certificate directory with `-v` and specify paths with `CERT_FILE` and `KEY_FILE` environment variables.

---

### Use Case 4: "I want path-based routing (example.com/api vs example.com/web)"

**Scenario**: Route `/api/*` to API service, `/` to web

**Limitation**: This proxy routes by **domain** (not path). Use a single domain with backend path rewriting:

**Workaround 1 - Use subdomains** (recommended):
- `api.example.com` → API service
- `www.example.com` → Web service

**Workaround 2 - Backend handles routing**:
- Single virtual host → Nginx/Traefik backend that does path routing

**Future**: Path-based routing planned for v3.1

---

## Configuration

### Minimal Configuration

```json
{
  "web_virtual_hosts": [
    {
      "from": "example.com",
      "scheme": "http",
      "host_name": "localhost",
      "port": 8080
    }
  ],
  "default_host": "example.com",
  "reverse_proxy_port": ":443"
}
```

### Full Configuration Reference

See [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for complete reference including:
- All virtual host types (Web, gRPC-Web)
- Custom headers
- Client certificate authentication (mTLS)
- Backend mTLS
- Logging configuration
- Advanced options

> **Note**: `grpc_virtual_hosts`, `grpc_json_virtual_hosts`, and `ssh_virtual_hosts` are deprecated and ignored.

---

## DNS Setup

### Required DNS Records

Point your domain to your server's public IP:

```
# A record
example.com        A    203.0.113.10
api.example.com    A    203.0.113.10
admin.example.com  A    203.0.113.10

# Optional: IPv6
example.com        AAAA 2001:db8::1
```

### Wildcard Subdomain (optional)

```
*.example.com      A    203.0.113.10
```

Then any subdomain works: `foo.example.com`, `bar.example.com`, etc.

---

## Firewall Configuration

### Required Ports

| Port | Protocol | Purpose |
|------|----------|---------|
| 80 | TCP | HTTP (redirects to HTTPS, Let's Encrypt challenges) |
| 443 | TCP | HTTPS (main traffic) |
| 8081 | TCP | Web Configuration UI (optional, can be localhost-only) |

### Allow Ports

**Linux (UFW):**
```bash
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw reload
```

**Linux (firewalld):**
```bash
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

**AWS Security Group:**
```
HTTP   TCP  80   0.0.0.0/0
HTTPS  TCP  443  0.0.0.0/0
```

**Docker Host:**
```bash
# Ports automatically mapped with -p flag
docker run -p 80:80 -p 443:443 ...
```

---

## FAQ

### How does Let's Encrypt work?

1. You request `https://example.com`
2. Proxy checks for certificate in `/app/certs/example.com/`
3. If not found, requests certificate from Let's Encrypt
4. Let's Encrypt validates ownership via HTTP-01 challenge (port 80)
5. Certificate issued and cached
6. Automatically renewed 30 days before expiration

**Requirements**: DNS must point to your server, port 80 accessible from internet.

---

### Can I use my own certificates?

Yes! Set `server_certificate` in virtual host config:

```json
{
  "from": "example.com",
  "server_certificate": {
    "certificate_path": "/certs/example.pem",
    "private_key_path": "/certs/example-key.pem"
  }
}
```

Or set default certificate for all domains:

```json
{
  "default_server_cert": "/certs/default.pem",
  "default_server_key": "/certs/default-key.pem"
}
```

---

### How do I handle multiple backends for the same domain?

**Short answer**: Not directly supported.

**Alternatives**:
1. **Use different subdomains**: `app1.example.com`, `app2.example.com`
2. **Backend load balancer**: Proxy → Nginx/HAProxy → Multiple backends
3. **Wait for v3.1**: Load balancing feature planned

---

### What about Docker networking?

**Docker Desktop (Mac/Windows)**:
- Use `host.docker.internal` to access host machine

**Docker Linux**:
- Use `172.17.0.1` (default bridge gateway)
- Or use Docker Compose with service names

**Docker Compose** (recommended):
```yaml
services:
  proxy:
    networks: [backend]
  app:
    networks: [backend]
```

Config:
```json
{
  "host_name": "app",  // Service name, not localhost
  "port": 3000
}
```

---

### How do I check logs?

**Docker logs:**
```bash
docker logs -f reverseproxy
```

**Log files** (inside container):
```
/app/logs/YYYY-MM-DD.log
```

**Log levels** (in config.json):
```json
{
  "log_console_level": 4,  // 1=Fatal, 2=Error, 3=Warning, 4=Info, 5=Trace
  "log_file_level": 4
}
```

---

### Certificate not working?

See [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) for detailed troubleshooting, including:
- Let's Encrypt validation failures
- DNS propagation issues
- Firewall blocking port 80
- Docker networking problems

---

### Can I use this for gRPC-Web (from browser)?

Yes! Use `grpc_web_virtual_hosts`:

```json
{
  "grpc_web_virtual_hosts": [
    {
      "from": "grpcweb.example.com",
      "scheme": "http",
      "host_name": "grpc-backend",
      "port": 9090
    }
  ],
  "default_host": "grpcweb.example.com",
  "reverse_proxy_port": ":443"
}
```

Then use gRPC-Web client in JavaScript/TypeScript.

---

## Installation Options

### Docker (Recommended)

**Quick start with ENV vars:**
```bash
docker run -d \
  --name reverseproxy \
  -p 80:80 -p 443:443 -p 8081:8081 \
  -e DOMAIN=example.com \
  -e BACKEND_HOST=host.docker.internal \
  -e BACKEND_PORT=3000 \
  -v certs:/app/certs \
  janmbaco/go-reverseproxy-ssl:v3
```

**With custom config file:**
```bash
docker run -d \
  --name reverseproxy \
  -p 80:80 -p 443:443 \
  -v ./configs/config.json:/app/config/config.json:ro \
  -v certs:/app/certs \
  janmbaco/go-reverseproxy-ssl:v3
```

---

### Docker Compose

See `docker-compose.quickstart.yml` for complete example.

---

### Build from Source

**Requirements**: Go 1.25+

```bash
git clone https://github.com/janmbaco/go-reverseproxy-ssl.git
cd go-reverseproxy-ssl
go build -o bin/reverseproxy ./cmd/reverseproxy
./bin/reverseproxy --config configs/config.json
```

---

### Go Module

```bash
go get github.com/janmbaco/go-reverseproxy-ssl/v3
```

```go
package main

import "github.com/janmbaco/go-reverseproxy-ssl/v3/src/infrastructure"

func main() {
    bootstrapper := infrastructure.NewServerBootstrapper("config.json")
    bootstrapper.Start()
}
```

---

## Features

- ✓ **Automatic HTTPS** with Let's Encrypt (HTTP-01 challenge)
- ✓ **HTTP → HTTPS** automatic redirect
- ✓ **Multiple protocols**: HTTP/1.1, HTTP/2, gRPC-Web
- ✓ **Web Configuration UI** for managing virtual hosts
- ✓ **Custom certificates** support (BYO cert)
- ✓ **Client certificate authentication** (mTLS)
- ✓ **Custom response headers**
- ✓ **Lightweight**: ~20MB Docker image (Alpine-based)
- ✓ **Zero-downtime** certificate renewal
- ✓ **Structured logging** with configurable levels

---

## Documentation

- **[CONFIGURATION.md](docs/CONFIGURATION.md)** - Complete configuration reference
- **[TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)** - Common issues and solutions
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Internal architecture details
- **[UPGRADE-3.0.md](UPGRADE-3.0.md)** - Migration guide from v2.x

---

## What's New in v3.0

- **Clean Architecture**: Complete restructure for better maintainability
- **Web Configuration UI**: Manage virtual hosts without editing JSON
- **Improved DI**: Powered by go-infrastructure v2.0
- **Breaking Changes**: See [UPGRADE-3.0.md](UPGRADE-3.0.md) for migration guide

---

## License

GNU General Public License v3.0 - see [LICENSE](LICENSE) file.

**Key points**:
- ✓ Free to use, modify, and distribute
- ✓ Source code must be provided with distributions
- ✓ Derivative works must also be GPL v3.0
- ✓ No warranty


---

## Testing

### Unit Tests

Run all unit tests (no external dependencies required):

```bash
go test ./internal/... -v
```

Unit tests cover:
- Virtual host routing logic
- Configuration parsing
- Certificate management
- HTTP redirector functionality
- Server state management

### Integration Tests

Integration tests require Docker and test end-to-end functionality with real HTTPS certificates and backend services.

**Automated execution (recommended):**

```bash
# Windows (PowerShell)
.\scripts\run-integration-tests.ps1

# Linux/macOS (Bash)
./scripts/run-integration-tests.sh
```

**Config UI specific tests:**

```bash
# Windows (PowerShell) - Config UI only
.\scripts\run-config-ui-integration-tests.ps1

# Linux/macOS (Bash) - Config UI only
./scripts/run-config-ui-integration-tests.sh
```

The script automatically:
1. Generates temporary self-signed certificates using `alpine/openssl`
2. Builds the Docker image from `build/package/Dockerfile`
3. Starts all services with Docker Compose (`docker-compose.test.yml`)
4. Waits for reverse proxy to be ready (health check on port 443)
5. Waits for config UI to be ready (health check on port 8081)
6. Runs integration tests with verbose output
7. Cleans up all containers and temporary certificates

**What gets tested:**
- ✅ HTTP-to-HTTPS reverse proxy with multiple backends
- ✅ gRPC-Web proxy with streaming support
- ✅ TLS certificate loading and permissions
- ✅ Docker networking between services
- ✅ Virtual host routing by domain
- ✅ **Config UI HTML pages and API endpoints**
- ✅ **Virtual host CRUD operations via API**
- ✅ **Certificate upload functionality**

**Test services (Docker Compose):**
- `reverse-proxy` - Main application under test (ports 443 + 8081 for config UI)
- `backend-8080` through `backend-8083` - Nginx backends for web virtual hosts
- `grpc-backend` - gRPC server for gRPC-Web tests

**Manual execution:**
```bash
# Build test image
docker build -f build/package/Dockerfile -t go-reverseproxy-ssl:test .

# Generate certificates
docker run --rm -v ${PWD}/integration/testdata:/certs alpine/openssl req -x509 \
  -newkey rsa:2048 -nodes -days 365 \
  -keyout /certs/localhost-key.pem -out /certs/localhost-cert.pem \
  -subj "/CN=localhost"

# Start all services
docker-compose -f docker-compose.test.yml up -d

# Run tests
INTEGRATION_TEST=true go test -v ./integration/...

# Cleanup
docker-compose -f docker-compose.test.yml down
```

**CI/CD:**
- Unit tests run automatically on every push/PR
- Integration tests can be triggered manually or on schedule
- Use `INTEGRATION_TEST=true` environment variable to enable integration tests

### Test Coverage

- **Unit Tests**: 60+ tests covering all business logic
- **Integration Tests**: End-to-end HTTP/HTTPS/gRPC-Web proxy testing
- **Config UI Tests**: Complete web interface and API testing
- **Certificate Tests**: TLS certificate loading, permissions, and validation
- **CI Coverage**: Automated testing with GitHub Actions

---

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## Support

- **Issues**: [GitHub Issues](https://github.com/janmbaco/go-reverseproxy-ssl/issues)
- **Discussions**: [GitHub Discussions](https://github.com/janmbaco/go-reverseproxy-ssl/discussions)

---

**Version**: v3.0.0  
**Go Version**: 1.25+  
**License**: GPL v3.0
