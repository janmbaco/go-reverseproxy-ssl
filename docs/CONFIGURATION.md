# Configuration Reference

Complete reference for configuring Go ReverseProxy SSL v3.0.

## Configuration File Structure

The proxy uses a JSON configuration file with the following top-level structure:

```json
{
  "web_virtual_hosts": [ /* HTTP/HTTPS backends */ ],
  "grpc_web_virtual_hosts": [ /* gRPC-Web backends */ ],
  "default_host": "www.example.com",
  "reverse_proxy_port": ":443",
  "default_server_cert": "/path/to/default-cert.pem",
  "default_server_key": "/path/to/default-key.pem",
  "cert_dir": "./certs",
  "log_console_level": 4,
  "log_file_level": 4,
  "logs_dir": "./logs",
  "config_ui_port": ":8081"
}
```

> **Note**: The following fields are **deprecated** and ignored: `grpc_virtual_hosts`, `grpc_json_virtual_hosts`, `ssh_virtual_hosts`

## Global Configuration Options

| Key | Type | Default | Description | Since |
|-----|------|---------|-------------|-------|
| `default_host` | `string` | `"localhost"` | Fallback domain when no virtual host matches | v1.0 |
| `reverse_proxy_port` | `string` | `":443"` | Port for HTTPS listener (`:443`, `:8443`, etc.) | v1.0 |
| `default_server_cert` | `string` | `""` | Path to default TLS certificate (optional) | v2.0 |
| `default_server_key` | `string` | `""` | Path to default TLS private key (optional) | v2.0 |
| `cert_dir` | `string` | `"./certs"` | Directory for Let's Encrypt certificates | v1.0 |
| `log_console_level` | `int` | `4` | Console log level: 0=Off, 1=Fatal, 2=Error, 3=Warning, 4=Info, 5=Trace | v1.0 |
| `log_file_level` | `int` | `4` | File log level (same scale as console) | v1.0 |
| `logs_dir` | `string` | `"./logs"` | Directory for log files | v1.0 |
| `config_ui_port` | `string` | `":8081"` | Port for web-based configuration UI | v3.0 |

## Virtual Host Types

### WebVirtualHost (HTTP/HTTPS backends)

For standard HTTP/HTTPS web applications and APIs.

```json
{
  "web_virtual_hosts": [
    {
      "id": "auto-generated-uuid",
      "from": "api.example.com",
      "scheme": "http",
      "host_name": "localhost",
      "port": 8080,
      "path": "",
      "server_certificate": {
        "certificate_path": "/path/to/cert.pem",
        "private_key_path": "/path/to/key.pem",
        "ca_certificates": ["/path/to/ca.pem"]
      },
      "client_certificate": {
        "certificate_path": "/path/to/client-cert.pem",
        "private_key_path": "/path/to/client-key.pem",
        "ca_certificates": ["/path/to/client-ca.pem"]
      },
      "response_headers": {
        "X-Custom-Header": "value"
      },
      "need_pk_from_client": false
    }
  ]
}
```

**Fields:**

- `id` (string): Auto-generated UUID (managed by ConfigUI)
- `from` (string, required): Public domain name (e.g., `api.example.com`)
- `scheme` (string, required): Backend protocol (`http` or `https`)
- `host_name` (string, required): Backend hostname (`localhost`, `192.168.1.10`, or service name in Docker)
- `port` (int, required): Backend service port (e.g., `8080`)
- `path` (string, optional): Backend path prefix (e.g., `/api/v1`)
- `server_certificate` (object, optional): Custom TLS certificate for this domain (omit for Let's Encrypt)
  - `certificate_path` (string): Path to PEM certificate file
  - `private_key_path` (string): Path to PEM private key file
  - `ca_certificates` (array[string]): Paths to CA certificate files for client cert validation
- `client_certificate` (object, optional): Client certificate for mTLS to backend
  - `certificate_path` (string): Path to client certificate
  - `private_key_path` (string): Path to client private key
  - `ca_certificates` (array[string]): CA certificates to trust from backend
- `response_headers` (object, optional): Custom HTTP headers to add to all responses
- `need_pk_from_client` (bool, optional): If `true`, requires client certificate and adds `X-Forwarded-PrivateKey` header

### GrpcVirtualHost (Native gRPC) **DEPRECATED**

> **Deprecated**: This virtual host type is no longer supported. Use `grpc_web_virtual_hosts` for gRPC-Web support.

For gRPC services with native protocol support.

```json
{
  "grpc_virtual_hosts": [
    {
      "from": "grpc.example.com",
      "scheme": "http",
      "host_name": "localhost",
      "port": 9090,
      "server_certificate": null
    }
  ]
}
```

**Fields:**
- `from` (string, required): Public domain for gRPC service
- `scheme` (string, required): Backend protocol (`http` or `https`)
- `host_name` (string, required): Backend hostname
- `port` (int, required): Backend gRPC port
- `server_certificate` (object, optional): Custom certificate (same structure as WebVirtualHost)

### GrpcWebVirtualHost (gRPC-Web for browsers)

For gRPC services accessible from web browsers using the gRPC-Web protocol.

**Example 1: Transparent server (proxies all services/methods):**

```json
{
  "grpc_web_virtual_hosts": [
    {
      "from": "grpcweb.example.com",
      "scheme": "http",
      "host_name": "grpc-backend",
      "port": 9091,
      "grpc_web_proxy": {
        "is_transparent_server": true,
        "authority": "grpc-backend:9091",
        "allow_all_origins": true
      }
    }
  ]
}
```

**Example 2: Selective service/method proxying:**

```json
{
  "grpc_web_virtual_hosts": [
    {
      "from": "grpcweb.example.com",
      "scheme": "http",
      "host_name": "grpc-backend",
      "port": 9091,
      "grpc_web_proxy": {
        "grpc_services": {
          "hello.HelloService": ["SayHello", "ServerStreamingChat"],
          "user.UserService": ["GetUser", "UpdateUser"]
        },
        "is_transparent_server": false,
        "authority": "grpc-backend:9091",
        "allow_all_origins": false,
        "allowed_origins": ["https://myapp.com"],
        "use_web_sockets": false,
        "allowed_headers": ["X-Custom-Header"]
      }
    }
  ]
}
```

**Fields:**

- `from` (string, **required**): Public domain for gRPC-Web service
- `scheme` (string, **required**): Backend protocol (`http` or `https`)
- `host_name` (string, **required**): Backend hostname
- `port` (int, **required**): Backend gRPC port
- `grpc_web_proxy` (object, **required**): Complete gRPC-Web configuration
  - `is_transparent_server` (bool, optional): If `true`, proxies **all** gRPC services and methods automatically without needing to specify `grpc_services` (default: false)
  - `grpc_services` (object, optional): Map of service names to method arrays for selective proxying. **Only used when `is_transparent_server` is false**. Format: `{"ServiceName": ["Method1", "Method2"]}`. Methods not listed will be rejected with an error.
  - `authority` (string, optional): Override the `:authority` header sent to backend (useful for routing)
  - `allow_all_origins` (bool, optional): Allow requests from any origin for CORS (default: false)
  - `allowed_origins` (array[string], optional): List of allowed origins for CORS when `allow_all_origins` is false
  - `use_web_sockets` (bool, optional): Enable WebSocket transport for streaming (default: false)
  - `allowed_headers` (array[string], optional): Additional allowed request headers for CORS

**Note**: The `grpc_web_proxy` field is **mandatory** for gRPC-Web virtual hosts. Use `is_transparent_server: true` for simple setups where you want to proxy all gRPC services, or set it to `false` and specify `grpc_services` for fine-grained control over which services and methods are exposed.

### GrpcJSONVirtualHost (JSON-to-gRPC transcoding) **DEPRECATED**

> **Deprecated**: This virtual host type is no longer supported.

For exposing gRPC services via JSON/REST-like interface.

```json
{
  "grpc_json_virtual_hosts": [
    {
      "from": "api-json.example.com",
      "scheme": "http",
      "host_name": "localhost",
      "port": 9092
    }
  ]
}
```

**Fields:**
- Same as GrpcVirtualHost
- Automatically transcodes JSON requests to gRPC

### SSHVirtualHost (SSH tunneling) **DEPRECATED**

> **Deprecated**: This virtual host type is no longer supported.

For secure SSH connections over HTTPS.

```json
{
  "ssh_virtual_hosts": [
    {
      "from": "ssh.example.com",
      "scheme": "tcp",
      "host_name": "localhost",
      "port": 22,
      "client_certificate": {
        "ca_certificates": ["/path/to/authorized-keys.pem"]
      }
    }
  ]
}
```

**Fields:**
- `from` (string, required): Public domain for SSH access
- `scheme` (string, required): Must be `"tcp"`
- `host_name` (string, required): SSH server hostname
- `port` (int, required): SSH server port (typically `22`)
- `client_certificate` (object, optional): CA certificates for client authentication

## Complete Examples

### Example 1: Single Web Application

```json
{
  "web_virtual_hosts": [
    {
      "from": "www.example.com",
      "scheme": "http",
      "host_name": "localhost",
      "port": 3000
    }
  ],
  "default_host": "www.example.com",
  "reverse_proxy_port": ":443",
  "log_console_level": 4,
  "log_file_level": 4,
  "logs_dir": "./logs",
  "cert_dir": "./certs",
  "config_ui_port": ":8081"
}
```

### Example 2: Multi-Service Setup (Web + API + Admin)

```json
{
  "web_virtual_hosts": [
    {
      "from": "www.example.com",
      "scheme": "http",
      "host_name": "web-app",
      "port": 3000
    },
    {
      "from": "api.example.com",
      "scheme": "http",
      "host_name": "api-service",
      "port": 8080
    },
    {
      "from": "admin.example.com",
      "scheme": "http",
      "host_name": "admin-panel",
      "port": 5000
    }
  ],
  "default_host": "www.example.com",
  "reverse_proxy_port": ":443",
  "log_console_level": 3,
  "log_file_level": 4,
  "logs_dir": "/var/log/reverseproxy",
  "cert_dir": "/var/certs",
  "config_ui_port": ":8081"
}
```

### Example 3: Custom Certificates

```json
{
  "web_virtual_hosts": [
    {
      "from": "secure.example.com",
      "scheme": "http",
      "host_name": "localhost",
      "port": 8080,
      "server_certificate": {
        "certificate_path": "/certs/secure.example.com.pem",
        "private_key_path": "/certs/secure.example.com-key.pem"
      }
    }
  ],
  "default_host": "secure.example.com",
  "reverse_proxy_port": ":443",
  "default_server_cert": "/certs/default-cert.pem",
  "default_server_key": "/certs/default-key.pem"
}
```

### Example 4: Client Certificate Authentication (mTLS)

```json
{
  "web_virtual_hosts": [
    {
      "from": "admin.example.com",
      "scheme": "http",
      "host_name": "admin-backend",
      "port": 8080,
      "need_pk_from_client": true,
      "server_certificate": {
        "ca_certificates": ["/certs/client-ca.pem"]
      }
    }
  ],
  "default_host": "admin.example.com",
  "reverse_proxy_port": ":443"
}
```

### Example 5: Path-Based Routing

```json
{
  "web_virtual_hosts": [
    {
      "from": "api.example.com/v1",
      "scheme": "http",
      "host_name": "api-v1",
      "port": 8080,
      "path": "/api"
    },
    {
      "from": "api.example.com/v2",
      "scheme": "http",
      "host_name": "api-v2",
      "port": 8081,
      "path": "/api"
    }
  ],
  "default_host": "api.example.com",
  "reverse_proxy_port": ":443"
}
```

Requests:
- `https://api.example.com/v1/users` → `http://api-v1:8080/api/users`
- `https://api.example.com/v2/users` → `http://api-v2:8081/api/users`

### Example 6: Custom Response Headers

```json
{
  "web_virtual_hosts": [
    {
      "from": "api.example.com",
      "scheme": "http",
      "host_name": "api-backend",
      "port": 8080,
      "response_headers": {
        "X-Frame-Options": "DENY",
        "X-Content-Type-Options": "nosniff",
        "Strict-Transport-Security": "max-age=31536000; includeSubDomains"
      }
    }
  ],
  "default_host": "api.example.com",
  "reverse_proxy_port": ":443"
}
```

### Example 7: Mixed Protocol Setup (HTTP + gRPC-Web)

```json
{
  "web_virtual_hosts": [
    {
      "from": "www.example.com",
      "scheme": "http",
      "host_name": "web-app",
      "port": 3000
    },
    {
      "from": "api.example.com",
      "scheme": "http",
      "host_name": "rest-api",
      "port": 8080
    }
  ],
  "grpc_web_virtual_hosts": [
    {
      "from": "grpcweb.example.com",
      "scheme": "http",
      "host_name": "grpc-service",
      "port": 9090,
      "grpc_web_proxy": {
        "is_transparent_server": true,
        "authority": "grpc-service:9090",
        "allow_all_origins": true
      }
    }
  ],
  "default_host": "www.example.com",
  "reverse_proxy_port": ":443"
}
```

**Note**: With `is_transparent_server: true`, all gRPC services and methods are automatically proxied without needing to list them in `grpc_services`.

### Example 8: Docker Compose with Custom Certificates

```yaml
# docker-compose.yml
services:
  reverseproxy:
    image: janmbaco/go-reverseproxy-ssl:v3
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./config.json:/app/config/config.json:ro
      - ./certs:/tmp/mounted-certs:ro
      - reverseproxy-certs:/app/certs
    networks: [app]
  
  backend:
    image: my-backend:latest
    expose: ["8080"]
    networks: [app]

networks:
  app:
volumes:
  reverseproxy-certs:
```

## config.json
```json
{
  "web_virtual_hosts": [
    {
      "from": "example.com",
      "scheme": "http",
      "host_name": "backend",
      "port": 8080
    }
  ],
  "default_host": "example.com",
  "reverse_proxy_port": ":443",
  "default_server_cert": "/app/certs/example.com-cert.pem",
  "default_server_key": "/app/certs/example.com-key.pem"
}
```

The entrypoint will copy certificates from `/tmp/mounted-certs/*.pem` to `/app/certs/` with correct permissions on startup.

## Log Levels

| Level | Value | Description | Use Case |
|-------|-------|-------------|----------|
| Off | 0 | No logging | Not recommended |
| Fatal | 1 | Critical errors only | Minimal logging |
| Error | 2 | Errors | Production default |
| Warning | 3 | Warnings + Errors | Production with debugging |
| Info | 4 | Informational | Development default |
| Trace | 5 | Verbose debugging | Development/troubleshooting |

**Example**: Set console to Info (4) and file to Trace (5) for detailed file logs without console spam:

```json
{
  "log_console_level": 4,
  "log_file_level": 5
}
```

## Docker-Specific Configuration

### Using Environment Variables

The Docker image supports automatic configuration generation via environment variables when using the `docker-entrypoint.sh` script:

**Simple single-backend setup:**

```bash
docker run -d \
  -p 80:80 -p 443:443 \
  -e DOMAIN=example.com \
  -e BACKEND_HOST=my-app \
  -e BACKEND_PORT=3000 \
  -e BACKEND_SCHEME=http \
  -v certs:/app/certs \
  janmbaco/go-reverseproxy-ssl:v3
```

**Multi-subdomain setup:**

```bash
docker run -d \
  -p 80:80 -p 443:443 \
  -e WWW_DOMAIN=www.example.com \
  -e WWW_BACKEND=frontend:3000 \
  -e API_DOMAIN=api.example.com \
  -e API_BACKEND=backend:8080:https \
  -e APP_DOMAIN=app.example.com \
  -e APP_BACKEND=admin:5000 \
  -v certs:/app/certs \
  janmbaco/go-reverseproxy-ssl:v3
```

**Environment variable format:**
- `<PREFIX>_DOMAIN` - Public domain name
- `<PREFIX>_BACKEND` - Backend in format `host:port` or `host:port:scheme`
- Supported prefixes: `DOMAIN` (default), `WWW`, `API`, `APP`, `ADMIN`, etc.

**Custom certificates with environment variables:**

```bash
docker run -d \
  -p 80:80 -p 443:443 \
  -e DOMAIN=example.com \
  -e BACKEND_HOST=my-app \
  -e BACKEND_PORT=3000 \
  -e DEFAULT_SERVER_CERT=/tmp/mounted-certs/cert.pem \
  -e DEFAULT_SERVER_KEY=/tmp/mounted-certs/key.pem \
  -v ./certs:/tmp/mounted-certs:ro \
  janmbaco/go-reverseproxy-ssl:v3
```

**Note**: When using environment variables, the entrypoint script generates `/app/config/config.json` automatically. Mounted certificates should be placed in `/tmp/mounted-certs` (read-only), and the entrypoint copies them to `/app/certs` with correct permissions.

### Using Custom config.json

For advanced configurations, mount your own config file:

```bash
docker run -d \
  -p 80:80 -p 443:443 \
  -v ./config.json:/app/config/config.json:ro \
  -v certs:/app/certs \
  janmbaco/go-reverseproxy-ssl:v3 \
  --config /app/config/config.json
```

### Certificate Management in Docker

**Let's Encrypt certificates (automatic):**

Certificates are stored in `/app/certs` inside the container. Use a named volume to persist them:

```bash
docker run -d \
  -v reverseproxy-certs:/app/certs \
  janmbaco/go-reverseproxy-ssl:v3
```

**Custom certificates:**

Mount certificates to `/tmp/mounted-certs` (read-only), and the entrypoint will copy them to `/app/certs` with correct permissions:

```bash
docker run -d \
  -v ./my-certs:/tmp/mounted-certs:ro \
  -v ./config.json:/app/config/config.json:ro \
  janmbaco/go-reverseproxy-ssl:v3
```

**config.json for custom certificates:**

```json
{
  "web_virtual_hosts": [
    {
      "from": "example.com",
      "scheme": "http",
      "host_name": "backend",
      "port": 8080,
      "server_certificate": {
        "certificate_path": "/app/certs/example.com-cert.pem",
        "private_key_path": "/app/certs/example.com-key.pem"
      }
    }
  ],
  "default_server_cert": "/app/certs/default-cert.pem",
  "default_server_key": "/app/certs/default-key.pem"
}
```

**How certificate copying works:**
1. Mount your certificates to `/tmp/mounted-certs` (any `.pem` files)
2. Entrypoint script copies them to `/app/certs` on startup
3. Sets permissions to `644` and ownership to `reverseproxy:reverseproxy`
4. Application runs as non-root user `reverseproxy` via `su-exec`

### Backend Hostname in Docker

- **Docker Desktop (Mac/Windows)**: Use `host.docker.internal` to access host machine
- **Docker Linux**: Use `172.17.0.1` (default bridge gateway) or service name in Docker Compose
- **Docker Compose**: Use service name (e.g., `my-app`) for internal network communication

**Docker Compose example:**

```yaml
services:
  reverseproxy:
    image: janmbaco/go-reverseproxy-ssl:v3
    networks: [backend]
  
  my-app:
    image: my-app:latest
    networks: [backend]

networks:
  backend:
```

**config.json:**

```json
{
  "web_virtual_hosts": [
    {
      "from": "example.com",
      "host_name": "my-app",  // Service name, not localhost
      "port": 3000,
      "scheme": "http"
    }
  ]
}
```

### Running as Non-Root User

The Docker image uses a multi-stage security approach:

1. **Build stage**: Uses `golang:1.25-alpine` to compile the binary
2. **Runtime stage**: Uses `alpine:latest` with minimal packages
3. **User creation**: Creates `reverseproxy` user (UID 1000, GID 1000)
4. **Entrypoint execution**: 
   - Runs as `root` initially to handle certificate permissions
   - Copies certificates from `/tmp/mounted-certs` to `/app/certs`
   - Uses `su-exec` to switch to `reverseproxy` user before running application

**Security benefits:**
- Application runs as non-root user
- Certificate files have correct permissions (644)
- Read-only mounted certificates don't need write access
- Minimal attack surface with Alpine base image (~20MB)

## Configuration Validation

The proxy validates configuration on startup. Common errors:

- **Missing required fields**: `from`, `host_name`, `port`, `scheme`
- **Invalid port numbers**: Must be 1-65535
- **Invalid scheme**: Must be `http`, `https`, or `tcp` (SSH only)
- **Duplicate domains**: Each `from` must be unique across all virtual host types
- **Invalid certificate paths**: Files must exist and be readable
- **Invalid log levels**: Must be 0-5

**Note**: Deprecated fields (`grpc_virtual_hosts`, `grpc_json_virtual_hosts`, `ssh_virtual_hosts`) are ignored but do not cause validation errors for backward compatibility.

## See Also

- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Common issues and solutions
- [ARCHITECTURE.md](ARCHITECTURE.md) - Internal architecture details
- [UPGRADE-3.0.md](../UPGRADE-3.0.md) - Migration guide from v2.x
