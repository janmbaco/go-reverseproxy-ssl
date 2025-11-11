# Docker Development Files

This directory contains Docker configuration files for **development and testing** purposes.

## Files

### `Dockerfile.dev`
Development Dockerfile with debugging capabilities:
- Includes **Delve** debugger
- Generates self-signed certificates for `localhost` and `test.local`
- Exposes debug port `40000`
- Used for local development and troubleshooting

**Usage:**
```bash
# Build debug image
docker build -f docker/Dockerfile.dev -t go-reverseproxy-ssl:debug .

# Run with debugging enabled
docker run -d \
  -p 80:80 -p 443:443 -p 40000:40000 \
  --name reverseproxy-debug \
  -v $(pwd)/config.json:/app/config/config.json \
  go-reverseproxy-ssl:debug
```

---

### `docker-compose.dev.yml`
Complete development environment with test services:
- Reverse proxy with debugging
- 4 test web services (`test-web-dashboard`, `test-web-1`, `test-web-2`, `test-web-3`)
- Pre-configured with self-signed certificates

**Usage:**
```bash
# Start development environment
cd docker
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose -f docker-compose.dev.yml logs -f reverseproxy-debug

# Stop environment
docker-compose -f docker-compose.dev.yml down
```

**Attach debugger:**
1. Start the container
2. In VS Code, use "Attach to Go Debug Container" launch configuration
3. Set breakpoints and debug

---

## Production Files (in root)

For production deployment, use files in the **project root**:

- **`Dockerfile`** - Production build (optimized, no debug tools)
- **`Dockerfile.quickstart`** - Simplified build with ENV var config
- **`docker-compose.quickstart.yml`** - User-friendly example

---

## Testing

The development environment includes test HTML applications in `../integration/testdata/`:
- `dashboard/` - Main dashboard (port 8080)
- `prueba-1/` - Test site 1 (port 8081)
- `prueba-2/` - Test site 2 (port 8082)
- `prueba-3/` - Test site 3 (port 8083)

All accessible via the reverse proxy with automatic HTTPS (self-signed certs).

---

## Notes

- These files are **not intended for production use**
- Self-signed certificates will trigger browser warnings (expected behavior)
- Debug builds are larger and slower than production builds
- For production, always use `Dockerfile` or `Dockerfile.quickstart`
