# SECURITY AND DEPLOYMENT

Production deployment and security hardening guide for the Freebuff Go auth backend (`server/`).

This document covers:
- Required environment variables and their meanings
- Reverse proxy configurations (Nginx, Caddy) with HTTPS enforcement, rate limiting, and WebSocket/header forwarding
- Docker production image and systemd service for automatic restart
- TLS 1.3, CORS, and database backup strategies

---

## 1. Environment Variables

The server reads all configuration from environment variables. No secrets are hardcoded. Copy `.env.example` to `.env` and override the values for your environment. Export the variables before starting the server, or source the file in the systemd unit / Docker entrypoint.

### Required (no defaults that are safe for production)

| Variable | Example | Description |
| --- | --- | --- |
| `AUTH_LISTEN_ADDR` | `:8080` | Host:port the HTTP server binds to. Use `127.0.0.1:8080` when a reverse proxy is on the same host; use `:8080` only if the proxy routes external traffic to a different interface. |
| `AUTH_DB_PATH` | `/var/lib/freebuff/auth.db` | Filesystem path of the SQLite database. Must be on a persistent volume with appropriate permissions (read/write for the server user). **Never store the database on a tmpfs or ephemeral volume in production.** |
| `AUTH_ADMIN_TOKEN` | *(long random value)* | Admin token gating `/api/admin/*` and `/api/server/start`. When **unset**, all admin endpoints are disabled and return 404. Generate a long random value (e.g. `openssl rand -base64 32`) and keep it secret. **Must be set in production if you plan to use server-start or HWID ban management.** |

### Recommended (override non-secret defaults)

| Variable | Default | Description |
| --- | --- | --- |
| `AUTH_ACCESS_TOKEN_TTL` | `15m` | Lifetime of a short-lived access token (Go duration string, e.g. `15m`, `30m`). Shorter lifetimes reduce the window of abuse for a stolen token. |
| `AUTH_REFRESH_TOKEN_TTL` | `720h` (30 days) | Lifetime of a refresh token / session. Capped at 30 days by the config loader regardless of what you set. Longer sessions improve UX but increase exposure if a refresh token is stolen. |
| `AUTH_BCRYPT_COST` | `12` | bcrypt cost for password hashing. Range 4–31. Increase for high-security environments, but benchmark first: cost 12 is already expensive on CPUs without hardware acceleration. |
| `AUTH_PTERO_SCRIPT` | `scripts/ptero_control.py` | Path of the server-control script invoked by `POST /api/server/start`. The script is executed as a single argv element (no shell). Use an absolute path in production. |

### Optional (update/manifest signing)

| Variable | Description |
| --- | --- |
| `AUTH_LAUNCHER_VERSION` | Current launcher version reported in the update manifest. Default `0.1.0`. |
| `AUTH_UPDATE_BASE_URL` | Base URL from which launcher installers are served (e.g. `https://updates.freebuff.example`). |
| `AUTH_UPDATE_SHA256` | SHA-256 of the currently shipped installer. Used by the manifest endpoint. |
| `AUTH_UPDATE_MANDATORY` | Set to `1` to force clients to update before launching. Default off. |
| `AUTH_UPDATE_SIGNING_KEY` | HMAC-SHA256 signing key for the update manifest. **Must be a strong random value.** Never commit it. |

### Sample `.env`

```sh
# Bind to loopback only; nginx proxies external traffic.
AUTH_LISTEN_ADDR=127.0.0.1:8080

# Persistent database volume.
AUTH_DB_PATH=/var/lib/freebuff/auth.db

# Admin token (generate once, keep secret).
AUTH_ADMIN_TOKEN="K8s5sGx9v3k1Qm2R7tYwZ4bF6hJnLpDcA0sUwXyE1vG2iN4oM7qR8tYwZ0bA3cS5"

# Short access tokens; 30-day refresh sessions.
AUTH_ACCESS_TOKEN_TTL=15m
AUTH_REFRESH_TOKEN_TTL=720h

# bcrypt cost 12 is a reasonable default.
AUTH_BCRYPT_COST=12

# Only enable server-start if you use it.
AUTH_PTERO_SCRIPT=/opt/freebuff/scripts/ptero_control.py

# Update manifest signing key (generate once, keep secret).
AUTH_UPDATE_SIGNING_KEY="SpQr5tY6wM3zX9vAbC1dEfG4hJ7kLnO0pQrS2uV5wX8yZ1bN3mO6cA9dE2fG4hJ7"
```

---

## 2. Reverse Proxy Configuration

The Go backend is designed to sit behind a reverse proxy that terminates TLS. The backend itself does **not** implement TLS; it binds to a raw TCP port. The proxy handles:

- HTTPS termination (TLS 1.3 preferred)
- HTTP → HTTPS redirection
- Rate limiting at the edge (supplemental to the backend's in-memory limiter)
- WebSocket upgrade forwarding (if the frontend uses WebSockets through the same origin)
- Forwarding of client IP, protocol, and host headers

### 2.1 Nginx

#### Minimal HTTPS-only configuration

```nginx
# /etc/nginx/sites-available/freebuff-auth
server {
    listen 80;
    listen [::]:80;
    server_name auth.freebuff.example;

    # Redirect all HTTP to HTTPS.
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name auth.freebuff.example;

    # --- TLS configuration (TLS 1.3 preferred, 1.2 allowed for legacy) ---
    ssl_certificate         /etc/ssl/certs/freebuff-auth.pem;
    ssl_certificate_key     /etc/ssl/private/freebuff-auth.key;

    ssl_protocols TLSv1.3 TLSv1.2;
    ssl_ciphers             TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256:TLS_AES_128_GCM_SHA256;
    ssl_prefer_server_ciphers on;

    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:10m;
    ssl_session_tickets off;  # disable tickets for forward secrecy

    # OCSP stapling (requires resolver and cert with OCSP URI).
    ssl_stapling on;
    ssl_stapling_verify on;
    resolver 8.8.8.8 8.8.4.4 valid=300s;
    resolver_timeout 5s;

    # --- Security headers ---
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;
    add_header Referrer-Policy "no-referrer" always;
    add_header Content-Security-Policy "default-src 'none';" always;

    # --- Logging ---
    access_log /var/log/nginx/freebuff-auth-access.log;
    error_log  /var/log/nginx/freebuff-auth-error.log;

    # --- Backend ---
    location / {
        proxy_pass http://127.0.0.1:8080;

        # Pass the original client IP.
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Host $host;

        # Preserve the original Host for any absolute redirects the backend may emit.
        proxy_set_header X-Original-Forwarded-Host $host;

        # Timeouts must be larger than the backend's WriteTimeout (30s).
        proxy_connect_timeout 10s;
        proxy_send_timeout    30s;
        proxy_read_timeout    30s;

        # Disable response buffering for streaming endpoints (if any).
        proxy_buffering off;

        # HTTP/1.1 keep-alive to the backend.
        proxy_http_version 1.1;
        proxy_set_header Connection "";

        # --- Rate limiting (supplemental to backend limiter) ---
        # Define a zone in the http block: limit_req_zone $binary_remote_addr zone=auth:10m rate=10r/m;
        limit_req zone=auth burst=5 nodelay;
        limit_req_status 429;

        # --- CORS (if the API is called from browsers on other origins) ---
        # Enable only for known origins; do not use "*" in production.
        # add_header Access-Control-Allow-Origin "https://launcher.freebuff.example" always;
        # add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
        # add_header Access-Control-Allow-Headers "Authorization, Content-Type, X-Request-Id, X-Admin-Token" always;
        # add_header Access-Control-Max-Age "86400" always;

        # Handle preflight.
        if ($request_method = 'OPTIONS') {
            add_header Access-Control-Allow-Origin "https://launcher.freebuff.example" always;
            add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
            add_header Access-Control-Allow-Headers "Authorization, Content-Type, X-Request-Id, X-Admin-Token" always;
            add_header Access-Control-Max-Age "86400" always;
            add_header Content-Length 0;
            add_header Content-Type text/plain;
            return 204;
        }
    }

    # --- Deny access to hidden files (defense-in-depth) ---
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }
}
```

**Notes on the Nginx config:**
- The backend already sets `Strict-Transport-Security`, `X-Content-Type-Options`, `X-Frame-Options`, and `Referrer-Policy`. Nginx re-adds them so they survive even if the backend is bypassed (e.g. a future change).
- Rate limiting is defined with `limit_req_zone` in the `http` block (not shown above; add it to `/etc/nginx/nginx.conf` inside the `http {}` block).
- The `limit_req zone=auth burst=5 nodelay` allows 10 requests/minute with a burst of 5. Tune to match your auth traffic patterns.
- WebSocket forwarding: if you later add WebSocket endpoints, add `proxy_set_header Upgrade $http_upgrade;` and `proxy_set_header Connection "upgrade";` to the relevant location. The current API has no WebSocket endpoints, so these are omitted.

#### `limit_req_zone` in `http` block (nginx.conf)

```nginx
http {
    # ... other settings ...

    # 10 requests per minute per client IP for auth endpoints.
    limit_req_zone $binary_remote_addr zone=auth:10m rate=10r/m;

    # Optional: separate zone for admin endpoints (stricter).
    limit_req_zone $binary_remote_addr zone=admin:10m rate=2r/m;
}
```

Then in the admin location, use `limit_req zone=admin burst=1 nodelay;`.

### 2.2 Caddy

Caddy is simpler to configure and manages TLS certificates automatically (via Let's Encrypt). Suitable for smaller deployments.

```caddy
# Caddyfile
auth.freebuff.example {
    # Automatic HTTPS with Let's Encrypt.
    tls {
        protocols tls1.3 tls1.2
        ciphers TLS_AES_256_GCM_SHA384 TLS_CHACHA20_POLY1305_SHA256 TLS_AES_128_GCM_SHA256
        curves x25519 secp256r1
    }

    # Security headers.
    header {
        Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
        X-Content-Type-Options "nosniff"
        X-Frame-Options "DENY"
        Referrer-Policy "no-referrer"
        Content-Security-Policy "default-src 'none'"
        -Server  # hide Caddy version
    }

    # Rate limiting: 10 requests per minute per client IP.
    rate_limit {
        zone auth {
            key {remote_host}
            events 10
            window 1m
        }
    }

    # Forward to the backend.
    reverse_proxy 127.0.0.1:8080 {
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-For {remote_host}
        header_up X-Forwarded-Proto {scheme}
        header_up Host {host}

        # Timeouts.
        transport http {
            dial_timeout 10s
            response_header_timeout 30s
            keep_alive_timeout 30s
        }
    }

    # CORS for known origin (enable only if needed).
    # header / {
    #     Access-Control-Allow-Origin "https://launcher.freebuff.example"
    #     Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS"
    #     Access-Control-Allow-Headers "Authorization, Content-Type, X-Request-Id, X-Admin-Token"
    #     Access-Control-Max-Age "86400"
    # }

    # Handle OPTIONS preflight for the above CORS.
    # @cors_preflight method OPTIONS
    # respond @cors_preflight 204
}
```

**Notes on the Caddy config:**
- Caddy automatically obtains and renews Let's Encrypt certificates. Ensure port 80 is open for the HTTP challenge (or use DNS challenge for wildcard/Internal certs).
- Rate limiting is per client IP via `rate_limit` plugin (part of Caddy's standard distribution in recent versions; if your Caddy build lacks it, use the `caddy-rate-limit` community module or switch to Nginx).
- Caddy's default TLS policy already prefers TLS 1.3; the explicit `tls` block hardens it further.
- The `header` directive applies to all responses; add per-location header overrides if needed.

### 2.3 WebSocket and header forwarding notes

The current API does **not** use WebSockets, but if you add real-time endpoints later:

- **Nginx**: add to the proxy location:
  ```nginx
  proxy_set_header Upgrade $http_upgrade;
  proxy_set_header Connection "upgrade";
  proxy_read_timeout 86400s;  # long-lived WebSocket
  ```
- **Caddy**: `reverse_proxy` already forwards `Upgrade` and `Connection` headers by default. For long-lived connections, set `transport http { keep_alive_timeout 86400s }`.

For header forwarding in general:
- Always forward `X-Real-IP` / `X-Forwarded-For` so the backend's (future) IP-based rate limiter sees the real client IP.
- Forward `X-Forwarded-Proto` if the backend ever generates absolute URLs.
- The current backend does not trust `X-Forwarded-For` for anything security-sensitive (the in-memory limiter is per-direct-IP), but future changes may rely on it.

---

## 3. Docker & Systemd

### 3.1 Production Dockerfile

Multi-stage build to produce a minimal image. The backend is a single statically-linked Go binary (using `modernc.org/sqlite`, a pure-Go SQLite implementation, so no CGO is required).

```dockerfile
# Dockerfile
# Build stage.
FROM golang:1.26-alpine AS builder

# Install build dependencies.
RUN apk add --no-cache git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build a static binary with stripped symbols.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -extldflags=-static" -o /out/server ./server

# Runtime stage.
FROM alpine:3.21

# Create a non-root user.
RUN addgroup -S freebuff && adduser -S -G freebuff -h /data freebuff

# Create the data directory for the database.
RUN mkdir -p /data && chown -R freebuff:freebuff /data

WORKDIR /data
USER freebuff

# Copy the binary from the builder.
COPY --from=builder /out/server /usr/local/bin/server
COPY --from=builder /src/server/scripts /opt/freebuff/scripts  # if ptero_control.py is used

# The server binary is the entrypoint.
ENTRYPOINT ["server"]
```

**Notes:**
- The binary is statically linked (`CGO_ENABLED=0` + `-extldflags=-static`), so the runtime image only needs `alpine` (no `libc` or `sqlite` libraries).
- Run as a non-root user (`freebuff`) with a writable `/data` directory for the database.
- If `AUTH_PTERO_SCRIPT` points to `scripts/ptero_control.py`, copy that script into the image at the expected path, or mount it at runtime.
- The image does **not** include a shell (`sh`/`bash`) by default; if you need debugging, add `RUN apk add --no-cache busybox` or use a debug variant.

#### Docker build and run

```sh
# Build.
docker build -t freebuff-auth:latest .

# Run with environment variables (use --env-file or -e).
docker run -d --name freebuff-auth \
  --restart unless-stopped \
  -v /var/lib/freebuff/auth.db:/data/auth.db:ro \  # or rw if the container writes
  -e AUTH_LISTEN_ADDR=127.0.0.1:8080 \
  -e AUTH_DB_PATH=/data/auth.db \
  -e AUTH_ADMIN_TOKEN="your-secret-token" \
  -e AUTH_UPDATE_SIGNING_KEY="your-signing-key" \
  -p 127.0.0.1:8080:8080 \
  freebuff-auth:latest
```

**Important:** If the database is mounted from the host, ensure the file permissions allow the `freebuff` user inside the container to read/write. The `:ro` mount above is read-only; for a writable database, use `:rw` and ensure ownership matches (e.g. `chown 1000:1000 /var/lib/freebuff/auth.db` if the container user is uid 1000).

### 3.2 Systemd service

For a bare-metal or VM deployment without Docker.

```ini
# /etc/systemd/system/freebuff-auth.service
[Unit]
Description=Freebuff Auth Backend
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=freebuff
Group=freebuff

# The server user must own the database directory.
WorkingDirectory=/var/lib/freebuff
ExecStart=/usr/local/bin/server
EnvironmentFile=/etc/freebuff-auth/.env

# Restart on failure; back off if it fails repeatedly.
Restart=on-failure
RestartSec=5s

# Resource limits (tune to your environment).
MemoryMax=256M
MemoryHigh=192M
CPUQuota=50%

# Security hardening (systemd-side).
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/var/lib/freebuff
PrivateTmp=yes
RestrictNamespaces=yes
RestrictSUIDSGID=yes
MemoryDenyWriteExecute=yes

# Logging to journald.
StandardOutput=journal
StandardError=journal
SyslogIdentifier=freebuff-auth

[Install]
WantedBy=multi-user.target
```

**Environment file** (`/etc/freebuff-auth/.env`):

```sh
# /etc/freebuff-auth/.env
AUTH_LISTEN_ADDR=127.0.0.1:8080
AUTH_DB_PATH=/var/lib/freebuff/auth.db
AUTH_ADMIN_TOKEN=your-secret-admin-token
AUTH_ACCESS_TOKEN_TTL=15m
AUTH_REFRESH_TOKEN_TTL=720h
AUTH_BCRYPT_COST=12
AUTH_PTERO_SCRIPT=/opt/freebuff/scripts/ptero_control.py
AUTH_UPDATE_SIGNING_KEY=your-update-signing-key
AUTH_LAUNCHER_VERSION=0.1.0
AUTH_UPDATE_BASE_URL=https://updates.freebuff.example
AUTH_UPDATE_SHA256=e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
AUTH_UPDATE_MANDATORY=0
```

**Notes on the systemd unit:**
- `EnvironmentFile` reads the `.env` file; variables are passed to the process. The server's `config.Load()` reads them from the environment.
- `ProtectSystem=strict` makes the entire filesystem read-only except for paths in `ReadWritePaths`. The database lives under `/var/lib/freebuff`, which is writable.
- `NoNewPrivileges=yes` prevents the server from gaining extra privileges (e.g. via setuid binaries).
- `MemoryMax` and `CPUQuota` prevent a runaway process from starving the host.
- After creating the unit, run:
  ```sh
  sudo systemctl daemon-reload
  sudo systemctl enable --now freebuff-auth
  sudo journalctl -u freebuff-auth -f  # follow logs
  ```

#### Automatic restart and health checking

The server exposes `/healthz` (always 200) and `/readyz` (200 only when the database is reachable). Use a systemd watchdog or a separate watchdog service to restart the server if `/readyz` fails repeatedly.

Simple watchdog approach (separate service):

```ini
# /etc/systemd/system/freebuff-auth-watchdog.service
[Unit]
Description=Freebuff Auth Backend Watchdog
After=freebuff-auth.service
Requires=freebuff-auth.service

[Service]
Type=oneshot
ExecStart=/usr/local/bin/freebuff-watchdog
# Run periodically.
TimerSec=30s
```

Where `freebuff-watchdog` is a small script:

```sh
#!/bin/sh
# /usr/local/bin/freebuff-watchdog
set -e

# Probe /readyz; if it fails 3 times in a row, restart the service.
for i in 1 2 3; do
    if curl -sf http://127.0.0.1:8080/readyz > /dev/null 2>&1; then
        exit 0
    fi
    sleep 1
done

# Readyz is unhealthy; restart the backend.
logger -t freebuff-watchdog "readyz unhealthy, restarting freebuff-auth"
systemctl restart freebuff-auth
```

Enable the timer:

```ini
# /etc/systemd/system/freebuff-auth-watchdog.timer
[Unit]
Description=Run freebuff-auth watchdog every 30s

[Timer]
OnBootSec=30s
OnUnitActiveSec=30s

[Install]
WantedBy=timers.target
```

---

## 4. Security Hardening

### 4.1 TLS 1.3

- **Terminate TLS at the reverse proxy**, not in the Go backend. The backend binds to a raw TCP port.
- Use **TLS 1.3** as the preferred protocol; allow TLS 1.2 only if you must support legacy clients.
- Disable TLS session tickets to ensure forward secrecy is not weakened across restarts.
- Use strong ciphers:
  - `TLS_AES_256_GCM_SHA384`
  - `TLS_CHACHA20_POLY1305_SHA256`
  - `TLS_AES_128_GCM_SHA256`
- Prefer `X25519` for key exchange.
- Enable **HSTS** with `includeSubDomains` and `preload` once you are confident the domain will never need to serve HTTP again.
- Use OCSP stapling if your certificate supports it (reduces latency and avoids OCSP look-ups by clients).

### 4.2 CORS

- **Do not use `Access-Control-Allow-Origin: *`** for authenticated endpoints. Only allow known origins (e.g. your launcher web UI).
- Allow only the methods you actually use: `GET, POST, PUT, DELETE, OPTIONS`.
- Allow only the headers your API consumes: `Authorization`, `Content-Type`, `X-Request-Id`, `X-Admin-Token`.
- Set `Access-Control-Max-Age` to cache preflight results (e.g. 86400 seconds).
- The backend itself does **not** set CORS headers; the reverse proxy should add them if browser-based clients need cross-origin access. If all clients are non-browser (e.g. a native launcher), CORS can be omitted entirely.

### 4.3 Rate limiting

Two layers of rate limiting protect the auth endpoints:

1. **In-memory limiter in the Go backend** (`api/ratelimit.go`): 10 requests/minute per direct IP. Applies to all endpoints via the `s.limit` wrapper. This is **not** IP-aware when behind a proxy (the backend sees the proxy's IP). If you need per-client-IP limiting in the backend, forward `X-Real-IP` and update the limiter to trust that header (only when behind a trusted proxy).
2. **Reverse proxy rate limiting** (Nginx `limit_req_zone` or Caddy `rate_limit`): 10 requests/minute per client IP at the edge. This is the primary per-IP protection.

Tune the limits based on your traffic and abuse patterns. Authentication endpoints are the most sensitive; consider stricter limits for login/register (e.g. 5/minute) and even stricter for admin endpoints (2/minute).

### 4.4 Database backup strategy

SQLite is a single-file database. Back it up consistently.

**Option A: Periodic file copy (simplest)**

```sh
#!/bin/sh
# /usr/local/bin/freebuff-db-backup
set -e

DB=/var/lib/freebuff/auth.db
BACKUP_DIR=/var/backups/freebuff
RETENTION=7  # days

mkdir -p "$BACKUP_DIR"

# Copy the database file. SQLite is crash-safe for copies on the same
# filesystem; for extra safety, use the SQLite backup API or .backup.
cp "$DB" "$BACKUP_DIR/auth-$(date +%Y%m%d-%H%M%S).db"

# Rotate old backups.
find "$BACKUP_DIR" -name 'auth-*.db' -mtime +${RETENTION} -delete
```

Run daily via cron or systemd timer.

**Option B: SQLite online backup (safer under load)**

Use the SQLite `.backup` command or the `sqlite3` CLI:

```sh
sqlite3 /var/lib/freebuff/auth.db ".backup '/var/backups/freebuff/auth-$(date +%Y%m%d-%H%M%S).db'"
```

This uses the SQLite backup API, which is safe to run concurrently with the server.

**Option C: WAL mode and checkpointing**

Enable WAL mode for better concurrency and easier backups:

```sql
sqlite3 /var/lib/freebuff/auth.db "PRAGMA journal_mode=WAL;"
```

With WAL mode:
- The main database file is always consistent.
- Backups can copy the main file while the server is running (the `-wal` and `-shm` files must also be copied or the backup must use the backup API).
- Periodically checkpoint: `PRAGMA wal_checkpoint(TRUNCATE);` during low-traffic windows.

**Backup checklist:**
- Store backups on a different physical volume / host from the live database.
- Encrypt backups at rest if they contain user data (password hashes, device IDs).password hashes are bcrypt outputs and are safe if stolen, but device IDs and account metadata should still be protected.
- Test restore regularly. A backup that cannot be restored is not a backup.
- If you use `AUTH_DB_PATH` on a persistent Docker volume, back up the volume or use `docker exec` to run the backup command inside the container.

### 4.5 Additional hardening

- **Run as non-root.** Both the Docker image (uid 1000 `freebuff` user) and the systemd unit ( `User=freebuff`) do this.
- **Protect the admin token.** Never commit `AUTH_ADMIN_TOKEN` to version control. Generate it once with `openssl rand -base64 32` and store it in a secrets manager or a protected environment file.
- **Rotate secrets periodically.** Refresh token signing keys, admin token, and update-signing key on a schedule. Old tokens signed with the previous key must be handled gracefully (e.g. accept both keys during a transition window). The current backend does not sign tokens with a server-side key (tokens are opaque random values), so key rotation applies to the update-signing key and admin token only.
- **Log securely.** The backend logs request metadata (method, path, status, latency, request ID) but never logs passwords, tokens, or device private keys. Do not add logging of those values.
- **Panic recovery.** The middleware (`withRecovery`) catches panics and returns 500. Do not remove it; it prevents a single handler panic from taking down the whole server.
- **Context timeouts.** All database operations use the request context with a timeout. The HTTP server has `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` set. Do not increase these without understanding the DoS implications.
- **Keep dependencies updated.** The Go backend uses `modernc.org/sqlite` (pure Go SQLite) and `golang.org/x/crypto` (bcrypt). Monitor for security advisories and update promptly.

---

## 5. Deployment checklist

Before going live:

- [ ] Generate and store `AUTH_ADMIN_TOKEN` and `AUTH_UPDATE_SIGNING_KEY` securely.
- [ ] Set `AUTH_LISTEN_ADDR` to `127.0.0.1:8080` (or a restricted interface) so only the reverse proxy can reach the backend directly.
- [ ] Configure the reverse proxy (Nginx or Caddy) with TLS 1.3, HSTS, and rate limiting.
- [ ] Test the full auth flow: register → login → refresh → logout → register again.
- [ ] Test the happy path and error paths: invalid credentials, banned HWID, malformed bearer token, oversized body.
- [ ] Test health probes: `/healthz` (200), `/readyz` (200 when DB is up, 503 when DB is down).
- [ ] Set up database backups and test restore.
- [ ] Configure log rotation for the reverse proxy and the backend (journald handles the backend; rotate Nginx/Caddy logs separately).
- [ ] Set up monitoring/alerts for 5xx rates, database unavailability, and rate-limit hits.
- [ ] Document the emergency rotate procedure for `AUTH_ADMIN_TOKEN` and `AUTH_UPDATE_SIGNING_KEY`.

---

## 6. Quick reference: endpoints and their security posture

| Endpoint | Auth required | Admin required | Rate limited | Notes |
| --- | --- | --- | --- | --- |
| `POST /api/v1/auth/register` | No | No | Yes | Creates account + session. |
| `POST /api/v1/auth/login` | No | No | Yes | Issues session; banned HWID → 403. |
| `POST /api/v1/auth/refresh` | No (refresh token) | No | Yes | Rotation invalidates old refresh token. |
| `POST /api/v1/auth/logout` | No (refresh token) | No | Yes | Idempotent. |
| `POST /api/v1/auth/logout-all` | Yes (access token) | No | Yes | Revokes all sessions for the account. |
| `GET /api/v1/account/me` | Yes | No | Yes | Returns profile. |
| `POST /api/v1/account/password` | Yes | No | Yes | Changes password; revokes all sessions. |
| `GET /api/v1/account/sessions` | Yes | No | Yes | Lists sessions. |
| `POST /api/v1/account/sessions/revoke` | Yes | No | Yes | Revokes one session; 403 if not owned. |
| `POST /api/v1/account/sessions/logout-others` | Yes | No | Yes | Revokes all but current session. |
| `GET /api/v1/skins` | Yes | No | Yes | Lists own skins. |
| `PUT /api/v1/skins/{variant}` | Yes | No | Yes | Upload skin; rate-limited to 2/3h. |
| `GET /api/v1/skins/{account}/{variant}` | Yes | No | Yes | Public skin fetch. |
| `DELETE /api/v1/skins/{variant}` | Yes | No | Yes | Delete own skin. |
| `POST /api/v1/skins/upload/{variant}` | Yes | No | Yes | Async upload (reserved). |
| `GET /api/v1/tasks/{id}` | Yes | No | Yes | Task status. |
| `POST /api/v1/proxy/auth` | No (admin token) | Yes | Yes | Server-side token validation. |
| `POST /api/admin/hwid/ban` | No (admin token) | Yes | Yes | Ban HWID. |
| `POST /api/admin/hwid/unban` | No (admin token) | Yes | Yes | Unban HWID. |
| `GET /api/admin/hwid/bans` | No (admin token) | Yes | Yes | List HWID bans. |
| `GET /api/admin/users` | No (admin token) | Yes | Yes | List accounts. |
| `DELETE /api/admin/users/{id}` | No (admin token) | Yes | Yes | Delete account. |
| `POST /api/server/start` | No (admin token) | Yes | Yes | Start Minecraft server via script. |
| `GET /healthz` | No | No | No | Liveness probe. |
| `GET /readyz` | No | No | No | Readiness probe (DB check). |
| `GET /api/v1/updates/manifest` | No | No | No | Signed update manifest. |
