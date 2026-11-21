# ============================================
# Build Stage
# ============================================
FROM golang:1.27.4-alpine AS builder

WORKDIR /build

# Instalar dependencias de compilación
RUN apk add --no-cache git ca-certificates

# Copiar archivos del módulo Go
COPY go.mod go.sum ./
COPY cmd/ cmd/
COPY internal/ internal/

# Descargar y verificar dependencias
RUN go mod download && go mod verify

# Compilar con optimizaciones para Linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -o go-reverseproxy-ssl \
    ./cmd/reverseproxy

# ============================================
# Runtime Stage
# ============================================
FROM cgr.dev/chainguard/wolfi-base@sha256:be019c1c74fe69204a42c4e10bc4d886a49bf5403ffa74374ef7baa408fc9a26 AS runtime-packages
# Install signed runtime packages into a fresh root, retaining the APK inventory.
RUN mkdir -p /runtime/etc/apk \
    && cp -R /etc/apk/keys /etc/apk/repositories /runtime/etc/apk/ \
    && apk --root /runtime --initdb --no-cache add wolfi-baselayout busybox ca-certificates-bundle su-exec tzdata \
    && test ! -e /runtime/usr/lib/libcrypto.so.3 \
    && test ! -e /runtime/usr/lib/libssl.so.3

FROM scratch
COPY --from=runtime-packages /runtime/ /

LABEL maintainer="janmbaco"
LABEL description="Go ReverseProxy SSL with Let's Encrypt support"

# Crear usuario no-root
RUN addgroup -g 1000 reverseproxy && \
    adduser -D -u 1000 -G reverseproxy reverseproxy

WORKDIR /app

# Copiar binario compilado
COPY --from=builder /build/go-reverseproxy-ssl .

# Copiar entrypoint script
COPY build/package/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Crear directorios necesarios con permisos correctos
RUN mkdir -p /app/certs /app/logs /app/config /tmp && \
    chown -R reverseproxy:reverseproxy /app && \
    chmod 1777 /tmp

# NO cambiar a usuario no-root aquí, lo haremos en el entrypoint
# USER reverseproxy

EXPOSE 80 443

# Healthcheck
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["--config", "/app/config/config.json"]
