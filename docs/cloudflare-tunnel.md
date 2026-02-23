# Cloudflare Tunnel Deployment

## Overview

Cloudflare Tunnels natively support WebSocket connections — no special configuration is required. Tunnels create a secure outbound connection from your server to Cloudflare's network, eliminating the need to open inbound ports or configure firewalls.

## Prerequisites

- A Cloudflare account with a domain
- `cloudflared` CLI installed on the server
- The chat application running (e.g., via Docker Compose)

## Configuration

### 1. Set allowed origins

The backend validates WebSocket and CORS origins. Update `ALLOWED_ORIGINS` to include your tunnel domain.

```bash
export ALLOWED_ORIGINS=https://chat.yourdomain.com
```

Or in `docker-compose.yml`:

```yaml
environment:
  ALLOWED_ORIGINS: "https://chat.yourdomain.com"
```

### 2. Create the tunnel

```bash
cloudflared tunnel create websocket-chat
cloudflared tunnel route dns websocket-chat chat.yourdomain.com
```

### 3. Configure the tunnel

Create a `config.yml` for `cloudflared`:

```yaml
tunnel: websocket-chat
credentials-file: /path/to/.cloudflared/<tunnel-id>.json

ingress:
  - hostname: chat.yourdomain.com
    service: http://localhost:5173
  - service: http_status:404
```

This routes traffic to the nginx frontend, which handles load balancing to backend instances.

### 4. Run the tunnel

```bash
cloudflared tunnel run websocket-chat
```

## WebSocket Compatibility Notes

- Cloudflare Tunnels handle WebSocket upgrade headers automatically
- The backend's ping interval (54 seconds) is well within Cloudflare's 100-second idle timeout
- No additional proxy or WebSocket-specific settings are needed
- Both `ws://` (upgraded to `wss://` by Cloudflare) and `wss://` are supported

## TUI Client with Cloudflare Tunnel

```bash
./bin/tui --server wss://chat.yourdomain.com
```

## Troubleshooting

- **WebSocket connections drop:** Check that `ALLOWED_ORIGINS` includes the tunnel domain
- **Idle disconnects:** Cloudflare's idle timeout is 100 seconds; the backend sends pings every 54 seconds to keep connections alive
- **Debug logging:** Run `cloudflared tunnel --loglevel debug run websocket-chat`
