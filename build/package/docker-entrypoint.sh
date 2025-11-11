#!/bin/sh
set -e

CONFIG_FILE="${CONFIG_FILE:-/app/config/config.json}"

# Helper function to parse BACKEND (format: "host:port" or "host:port:scheme")
parse_backend() {
    local backend="$1"
    local default_scheme="${2:-http}"
    
    # Split by colon
    local host=$(echo "$backend" | cut -d':' -f1)
    local port=$(echo "$backend" | cut -d':' -f2)
    local scheme=$(echo "$backend" | cut -d':' -f3)
    
    # Use default scheme if not provided
    [ -z "$scheme" ] && scheme="$default_scheme"
    
    # Return JSON object
    echo "\"host_name\": \"$host\", \"port\": $port, \"scheme\": \"$scheme\""
}

# Helper function to add certificate configuration to virtual host
add_certificate_config() {
    local domain="$1"
    local cert_var_prefix="$2"
    local cert_path_var="${cert_var_prefix}_CERT"
    local key_path_var="${cert_var_prefix}_KEY"
    
    eval cert_path=\$$cert_path_var
    eval key_path=\$$key_path_var
    
    if [ -n "$cert_path" ] && [ -n "$key_path" ]; then
        echo ",
      \"server_certificate\": {
        \"certificate_path\": \"$cert_path\",
        \"private_key_path\": \"$key_path\"
      }"
        echo "  - Certificate for $domain: $cert_path" >&2
    fi
}

# If config file doesn't exist, generate from environment variables
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Config file not found at $CONFIG_FILE"
    echo "Generating configuration from environment variables..."
    
    # Default values
    LOG_LEVEL="${LOG_LEVEL:-4}"
    CERT_DIR="${CERT_DIR:-/app/certs}"
    LOGS_DIR="${LOGS_DIR:-/app/logs}"
    CONFIG_UI_PORT="${CONFIG_UI_PORT:-:8081}"
    
    # Default certificates (optional)
    DEFAULT_CERT_JSON=""
    if [ -n "$DEFAULT_SERVER_CERT" ] && [ -n "$DEFAULT_SERVER_KEY" ]; then
        DEFAULT_CERT_JSON=",
  \"default_server_cert\": \"$DEFAULT_SERVER_CERT\",
  \"default_server_key\": \"$DEFAULT_SERVER_KEY\""
        echo "Using default certificates: $DEFAULT_SERVER_CERT"
    fi
    
    # Create config directory if it doesn't exist
    mkdir -p "$(dirname "$CONFIG_FILE")"
    
    # Build virtual hosts array
    VH_JSON=""
    DEFAULT_HOST=""
    
    # Check for multi-domain configuration (WWW, API, APP)
    if [ -n "$WWW_DOMAIN" ] && [ -n "$WWW_BACKEND" ]; then
        WWW_PARSED=$(parse_backend "$WWW_BACKEND")
        WWW_CERT=$(add_certificate_config "$WWW_DOMAIN" "WWW")
        VH_JSON="$VH_JSON    {
      \"from\": \"$WWW_DOMAIN\",
      $WWW_PARSED$WWW_CERT
    }"
        DEFAULT_HOST="$WWW_DOMAIN"
    fi
    
    if [ -n "$API_DOMAIN" ] && [ -n "$API_BACKEND" ]; then
        [ -n "$VH_JSON" ] && VH_JSON="$VH_JSON,"
        API_PARSED=$(parse_backend "$API_BACKEND")
        API_CERT=$(add_certificate_config "$API_DOMAIN" "API")
        VH_JSON="$VH_JSON
    {
      \"from\": \"$API_DOMAIN\",
      $API_PARSED$API_CERT
    }"
        [ -z "$DEFAULT_HOST" ] && DEFAULT_HOST="$API_DOMAIN"
    fi
    
    if [ -n "$APP_DOMAIN" ] && [ -n "$APP_BACKEND" ]; then
        [ -n "$VH_JSON" ] && VH_JSON="$VH_JSON,"
        APP_PARSED=$(parse_backend "$APP_BACKEND")
        APP_CERT=$(add_certificate_config "$APP_DOMAIN" "APP")
        VH_JSON="$VH_JSON
    {
      \"from\": \"$APP_DOMAIN\",
      $APP_PARSED$APP_CERT
    }"
        [ -z "$DEFAULT_HOST" ] && DEFAULT_HOST="$APP_DOMAIN"
    fi
    
    # Fallback to single DOMAIN configuration
    if [ -z "$VH_JSON" ]; then
        DOMAIN="${DOMAIN:-localhost}"
        BACKEND_HOST="${BACKEND_HOST:-localhost}"
        BACKEND_PORT="${BACKEND_PORT:-8080}"
        BACKEND_SCHEME="${BACKEND_SCHEME:-http}"
        
        VH_JSON="    {
      \"from\": \"$DOMAIN\",
      \"scheme\": \"$BACKEND_SCHEME\",
      \"host_name\": \"$BACKEND_HOST\",
      \"port\": $BACKEND_PORT
    }"
        DEFAULT_HOST="$DOMAIN"
        
        echo "Using single domain: $DOMAIN -> $BACKEND_SCHEME://$BACKEND_HOST:$BACKEND_PORT"
    else
        echo "Using multi-domain configuration:"
        [ -n "$WWW_DOMAIN" ] && echo "  - WWW: $WWW_DOMAIN -> $WWW_BACKEND"
        [ -n "$API_DOMAIN" ] && echo "  - API: $API_DOMAIN -> $API_BACKEND"
        [ -n "$APP_DOMAIN" ] && echo "  - APP: $APP_DOMAIN -> $APP_BACKEND"
    fi
    
    # Generate config.json
    cat > "$CONFIG_FILE" <<EOF
{
  "web_virtual_hosts": [
$VH_JSON
  ],
  "default_host": "$DEFAULT_HOST",
  "reverse_proxy_port": ":443",
  "config_ui_port": "$CONFIG_UI_PORT",
  "log_console_level": $LOG_LEVEL,
  "log_file_level": $LOG_LEVEL,
  "logs_dir": "$LOGS_DIR",
  "cert_dir": "$CERT_DIR"$DEFAULT_CERT_JSON
}
EOF
    
    echo "Configuration generated successfully at $CONFIG_FILE"
else
    echo "Using existing configuration at $CONFIG_FILE"
fi

# Copy mounted certificate files from /tmp/mounted-certs to /app/certs with correct permissions
if [ -d "/tmp/mounted-certs" ]; then
    echo "Copying certificates from /tmp/mounted-certs to /app/certs..."
    for cert_file in /tmp/mounted-certs/*.pem; do
        if [ -f "$cert_file" ]; then
            filename=$(basename "$cert_file")
            dest_file="/app/certs/$filename"
            cat "$cert_file" > "$dest_file"
            chmod 644 "$dest_file"
            chown reverseproxy:reverseproxy "$dest_file"
            echo "  ✓ Copied $filename"
        fi
    done
fi

# Execute the main application as reverseproxy user
exec su-exec reverseproxy /app/go-reverseproxy-ssl "$@"
