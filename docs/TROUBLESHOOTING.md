# Troubleshooting Guide

Common issues and solutions for Go ReverseProxy SSL v3.0.

## Let's Encrypt Certificate Issues

### Certificate Not Issued

**Symptoms**:
```
Error: no certificate configured for server name: www.example.com
```

**Root Causes**:
- DNS not pointing to your server
- Port 80 blocked or not accessible from internet
- Domain not configured in virtual hosts
- Firewall blocking ACME challenge

**Diagnostic Steps**:

1. **Verify DNS resolution**:
   ```bash
   dig www.example.com
   nslookup www.example.com
   ```
   Expected: Your server's public IP address

2. **Test port 80 accessibility**:
   ```bash
   curl http://YOUR_PUBLIC_IP/.well-known/acme-challenge/test
   ```
   Expected: Connection successful (even if 404)

3. **Check domain in config**:
   ```bash
   cat config.json | jq '.web_virtual_hosts[].from'
   ```
   Expected: Your domain listed

4. **Verify firewall rules**:
   ```bash
   # UFW
   sudo ufw status | grep 80
   
   # firewalld
   sudo firewall-cmd --list-ports
   
   # iptables
   sudo iptables -L -n | grep 80
   ```
   Expected: Port 80 allowed

5. **Check Docker port mapping**:
   ```bash
   docker ps | grep reverseproxy
   ```
   Expected: `0.0.0.0:80->80/tcp`

**Solutions**:

- **DNS issues**: Wait for DNS propagation (up to 48 hours, usually 5-10 minutes)
- **Firewall**: 
  ```bash
  sudo ufw allow 80/tcp
  sudo ufw allow 443/tcp
  ```
- **Docker mapping**: Add `-p 80:80` to docker run command
- **Cloud providers**: Check security groups (AWS, GCP, Azure)

**Check Logs**:
```bash
docker logs reverseproxy | grep -i "acme\|certificate\|challenge"
```

### Certificate Renewal Fails

**Symptoms**:
```
Error: failed to renew certificate for www.example.com
```

**Common Causes**:
- Certificate volume not persistent
- Port 80 temporarily unavailable
- Rate limit hit (Let's Encrypt: 5 renewals per week)

**Solutions**:

1. **Ensure persistent volume**:
   ```yaml
   volumes:
     - reverseproxy-certs:/app/certs  # Named volume (recommended)
     # OR
     - ./certs:/app/certs              # Host mount
   ```

2. **Check rate limits**:
   - Let's Encrypt allows 50 certificates per domain per week
   - 5 duplicate certificates per week
   - Solution: Wait 7 days or use staging environment for testing

3. **Force renewal** (if needed):
   ```bash
   docker exec reverseproxy rm -rf /app/certs/example.com
   docker restart reverseproxy
   ```

---

## Configuration Issues

### ConfigUI Not Accessible

**Symptoms**:
```
curl: (7) Failed to connect to localhost port 8081
```

**Diagnostic Steps**:

1. **Check config**:
   ```bash
   cat config.json | jq '.config_ui_port'
   ```
   Expected: `":8081"` or your custom port

2. **Verify Docker port mapping**:
   ```bash
   docker ps | grep reverseproxy
   ```
   Expected: `0.0.0.0:8081->8081/tcp`

3. **Check if service started**:
   ```bash
   docker logs reverseproxy | grep -i "configui\|8081"
   ```
   Expected: `ConfigUI server listening on :8081`

4. **Test from inside container**:
   ```bash
   docker exec reverseproxy wget -O- http://localhost:8081
   ```

**Solutions**:

- **Port not mapped**: Add `-p 8081:8081` to docker run
- **Config missing**: Set `"config_ui_port": ":8081"` in config.json
- **Template errors**: Check logs for template parsing errors
- **Firewall**: If accessing remotely, allow port 8081

### Invalid Configuration Errors

**Symptoms**:
```
Error: invalid configuration: missing required field 'from'
```

**Common Mistakes**:

1. **Missing required fields**:
   ```json
   // ❌ Wrong
   {
     "web_virtual_hosts": [
       {"host_name": "localhost", "port": 8080}
     ]
   }
   
   // ✅ Correct
   {
     "web_virtual_hosts": [
       {
         "from": "example.com",
         "scheme": "http",
         "host_name": "localhost",
         "port": 8080
       }
     ]
   }
   ```

2. **Invalid port numbers**:
   ```json
   // ❌ Wrong
   {"port": "8080"}  // String instead of number
   {"port": 70000}   // Out of range (1-65535)
   
   // ✅ Correct
   {"port": 8080}
   ```

3. **Invalid scheme**:
   ```json
   // ✅ Valid schemes
   {"scheme": "http"}   // For HTTP backends (most common)
   {"scheme": "https"}  // For HTTPS backends (when backend has TLS)
   
   // ❌ Invalid
   {"scheme": "tcp"}    // Only for SSH virtual hosts
   ```

**Validation Tool**:
```bash
# Validate config before starting
docker run --rm -v ./config.json:/config.json \
  janmbaco/go-reverseproxy-ssl:v3 \
  --config /config.json --validate
```

> **Note**: Configuration validation also runs automatically when changes are made through the ConfigUI or when the configuration file is modified while the server is running. Invalid configurations will prevent the server from restarting and will be logged as errors.

---

## Networking Issues

### Proxy Not Forwarding Requests (502 Bad Gateway)

**Symptoms**:
```
curl https://example.com
< HTTP/2 502
```

**Diagnostic Steps**:

1. **Check backend is running**:
   ```bash
   # From host
   curl http://localhost:8080
   
   # From Docker
   docker exec reverseproxy wget -O- http://localhost:8080
   ```

2. **Verify hostname resolution**:
   ```bash
   # Test DNS/hostname
   docker exec reverseproxy ping backend-service
   ```

3. **Check Docker network**:
   ```bash
   docker network inspect bridge
   # Or for Docker Compose
   docker network inspect <project>_app-network
   ```

4. **Test backend connectivity from proxy**:
   ```bash
   docker exec reverseproxy telnet backend-service 8080
   ```

**Solutions by Environment**:

**Docker Compose**:
```yaml
services:
  reverseproxy:
    networks: [app-network]
  backend:
    networks: [app-network]

networks:
  app-network:
```

Config:
```json
{
  "host_name": "backend",  // Service name, not localhost
  "port": 8080
}
```

**Docker Desktop (Mac/Windows)**:
```json
{
  "host_name": "host.docker.internal",  // Special hostname for host access
  "port": 8080
}
```

**Docker Linux**:
```json
{
  "host_name": "172.17.0.1",  // Default bridge gateway
  "port": 8080
}
```

**Standalone Docker**:
```bash
docker run --network host ...  # Use host network (Linux only)
```

### Connection Timeout

**Symptoms**:
```
Error: dial tcp: i/o timeout
```

**Common Causes**:
- Backend service not listening on expected port
- Firewall between proxy and backend
- Backend accepting connections too slowly

**Solutions**:

1. **Verify backend listening**:
   ```bash
   docker exec backend netstat -tulpn | grep 8080
   ```

2. **Check backend logs**:
   ```bash
   docker logs backend
   ```

3. **Increase timeout** (if backend is slow):
   ```go
   // Custom build with longer timeout
   // Modify src/infrastructure/http_server.go
   ```

---

## Client Certificate (mTLS) Issues

### Client Certificate Rejected (401 Unauthorized)

**Symptoms**:
```
curl https://example.com
< HTTP/2 401
< Unauthorized
```

**Diagnostic Steps**:

1. **Verify client cert requirement**:
   ```bash
   cat config.json | jq '.web_virtual_hosts[] | select(.from=="example.com") | .need_pk_from_client'
   ```
   Expected: `true`

2. **Check CA certificate configured**:
   ```bash
   cat config.json | jq '.web_virtual_hosts[] | select(.from=="example.com") | .server_certificate.ca_certificates'
   ```
   Expected: `["/path/to/ca.pem"]`

3. **Verify client cert validity**:
   ```bash
   openssl x509 -in client.pem -noout -dates
   openssl verify -CAfile ca.pem client.pem
   ```
   Expected: Not expired, verification OK

4. **Test with curl**:
   ```bash
   curl -v --cert client.pem --key client-key.pem https://example.com
   ```

**Common Issues**:

1. **CA mismatch**:
   - Client cert not signed by configured CA
   - Solution: Use correct CA or generate new client cert

2. **Certificate expired**:
   ```bash
   openssl x509 -in client.pem -noout -dates
   ```
   Solution: Generate new certificate

3. **Certificate format wrong**:
   - Must be PEM format (Base64)
   - Check: `openssl x509 -in client.pem -text`

4. **Private key mismatch**:
   ```bash
   # Verify key matches cert
   openssl x509 -noout -modulus -in client.pem | openssl md5
   openssl rsa -noout -modulus -in client-key.pem | openssl md5
   ```
   Expected: Same hash

**Generate Valid Client Certificate**:
```bash
# Generate CA (if needed)
openssl genrsa -out ca-key.pem 4096
openssl req -new -x509 -days 3650 -key ca-key.pem -out ca.pem

# Generate client cert
openssl genrsa -out client-key.pem 4096
openssl req -new -key client-key.pem -out client.csr
openssl x509 -req -days 365 -in client.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out client.pem

# Verify
openssl verify -CAfile ca.pem client.pem
```

---

## Performance Issues

### High Memory Usage

**Symptoms**:
```
docker stats reverseproxy
# MEM USAGE: 500MB+
```

**Common Causes**:
- Too many concurrent connections
- Certificate cache not being released
- Log files accumulating in memory

**Solutions**:

1. **Limit log verbosity**:
   ```json
   {
     "log_console_level": 2,  // Error only
     "log_file_level": 3      // Warning
   }
   ```

2. **Rotate logs**:
   ```bash
   # Add log rotation (cron or logrotate)
   find /app/logs -name "*.log" -mtime +7 -delete
   ```

3. **Set Docker memory limits**:
   ```bash
   docker run --memory="256m" ...
   ```

### Slow Response Times

**Symptoms**:
- Requests taking >5 seconds
- Timeouts

**Diagnostic Steps**:

1. **Test backend directly**:
   ```bash
   time curl http://backend:8080
   ```

2. **Check proxy logs**:
   ```bash
   docker logs reverseproxy | grep -i "slow\|timeout"
   ```

3. **Monitor Docker stats**:
   ```bash
   docker stats reverseproxy backend
   ```

**Common Solutions**:
- Backend is slow: Optimize backend application
- Network latency: Use same Docker network
- DNS resolution: Use IP instead of hostname for testing

---

## Certificate File Issues

### Permission Denied Errors

**Symptoms**:
```
Error: open /app/certs/example.com/cert.pem: permission denied
```

**Solutions**:

1. **Fix file permissions**:
   ```bash
   sudo chown -R 1000:1000 ./certs
   sudo chmod -R 755 ./certs
   ```

2. **Use named volumes**:
   ```yaml
   volumes:
     - certs:/app/certs  # Docker manages permissions
   ```

3. **Check SELinux** (if on RHEL/CentOS):
   ```bash
   sudo chcon -Rt svirt_sandbox_file_t ./certs
   ```

### Custom Certificate Not Working

**Symptoms**:
- Still using Let's Encrypt despite custom cert configured
- Certificate validation errors

**Checklist**:

1. **Verify paths**:
   ```bash
   docker exec reverseproxy ls -la /path/to/cert.pem
   ```

2. **Check certificate format** (must be PEM):
   ```bash
   openssl x509 -in cert.pem -text -noout
   ```

3. **Verify private key**:
   ```bash
   openssl rsa -in key.pem -check
   ```

4. **Test cert/key match**:
   ```bash
   openssl x509 -noout -modulus -in cert.pem | openssl md5
   openssl rsa -noout -modulus -in key.pem | openssl md5
   ```

---

## Docker-Specific Issues

### Container Exits Immediately

**Symptoms**:
```
docker ps -a
# STATUS: Exited (1) 2 seconds ago
```

**Check Logs**:
```bash
docker logs reverseproxy
```

**Common Errors**:

1. **Config file not found**:
   ```
   Error: config file not found: /app/config/config.json
   ```
   Solution: Mount config file correctly:
   ```bash
   docker run -v $(pwd)/config.json:/app/config/config.json:ro ...
   ```

2. **Invalid JSON**:
   ```
   Error: invalid character '}' looking for beginning of object key string
   ```
   Solution: Validate JSON:
   ```bash
   cat config.json | jq .
   ```

3. **Port already in use**:
   ```
   Error: bind: address already in use
   ```
   Solution: Stop conflicting service or use different port:
   ```bash
   # Find process using port 443
   sudo lsof -i :443
   sudo netstat -tulpn | grep :443
   ```

### Volume Not Persisting

**Symptoms**:
- Certificates regenerated on every restart
- Logs disappear

**Solutions**:

1. **Use named volumes**:
   ```yaml
   volumes:
     - certs:/app/certs      # Named volume (persists)
     # NOT: /app/certs        # Anonymous volume (deleted on remove)
   ```

2. **Verify volume exists**:
   ```bash
   docker volume ls | grep certs
   docker volume inspect certs
   ```

3. **Remove with `-v` flag** (to delete volumes):
   ```bash
   docker-compose down -v  # Deletes volumes too
   ```

---

## Getting More Help

1. **Enable trace logging**:
   ```json
   {
     "log_console_level": 5,
     "log_file_level": 5
   }
   ```

2. **Collect diagnostic info**:
   ```bash
   docker logs reverseproxy > logs.txt
   docker inspect reverseproxy > inspect.json
   docker exec reverseproxy cat /app/config/config.json > config.json
   ```

3. **Check GitHub Issues**:
   - [github.com/janmbaco/go-reverseproxy-ssl/issues](https://github.com/janmbaco/go-reverseproxy-ssl/issues)

4. **Report a bug**:
   - Include: Go version, Docker version, OS, config.json (sanitized), full logs
