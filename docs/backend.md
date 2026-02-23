# Backend Server Documentation

## Overview

The backend is a real-time WebSocket chat server written in Go. It provides JWT-based authentication (register/login), a REST API for user management, and a WebSocket endpoint for real-time messaging. Redis Pub/Sub enables horizontal scaling across multiple server instances so that a message sent to one instance is delivered to clients connected to any instance.

### Tech Stack

| Component        | Library / Tool                                   |
| ---------------- | ------------------------------------------------ |
| HTTP router      | [chi v5](https://github.com/go-chi/chi)          |
| WebSocket        | [coder/websocket](https://github.com/coder/websocket) |
| Database         | PostgreSQL 17 via [pgx v5](https://github.com/jackc/pgx) (connection pool) |
| Migrations       | [golang-migrate v4](https://github.com/golang-migrate/migrate) |
| Redis / Valkey   | [go-redis v9](https://github.com/redis/go-redis) |
| JWT              | [golang-jwt v5](https://github.com/golang-jwt/jwt) |
| Password hashing | [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) (default cost) |
| Config           | [caarlos0/env v11](https://github.com/caarlos0/env) (environment variables) |
| UUID             | [google/uuid](https://github.com/google/uuid)    |

---

## Architecture

### Package Layout

```
cmd/server/main.go        # Entrypoint — config, migrations, DB pool, server creation
internal/
  config/config.go         # Environment-variable-based configuration
  database/
    database.go            # pgxpool connection pool creation
    migrations.go          # golang-migrate wrapper
  user/
    model.go               # User struct (id, username, password, created_at)
    repository.go          # CRUD — Create, FindByUsername, FindByID
  auth/
    password.go            # bcrypt hash / check helpers
    jwt.go                 # JWTService — generate & validate HS256 tokens
    service.go             # AuthService — register, login, get-user (business logic)
    handler.go             # HTTP handlers for /api/auth/*
    middleware.go           # Bearer-token auth middleware for protected routes
  chat/
    message.go             # Factory functions for ServerMessage types
    client.go              # WebSocket client — ReadPump, WritePump, ping/pong
    hub.go                 # Hub — client registry, broadcast fan-out
    redis.go               # RedisAdapter — Pub/Sub publish & subscribe
    handler.go             # HTTP → WebSocket upgrade handler with JWT auth
  server/
    server.go              # Wires everything together; registers routes; runs HTTP server
pkg/
  protocol/message.go      # Shared message types used by server and TUI client
migrations/
  000001_create_users_table.up.sql
  000001_create_users_table.down.sql
```

### Dependency Flow

```
config.Load()
    │
    ├── database.RunMigrations()   (if RUN_MIGRATE=true)
    ├── database.NewPool()
    │       │
    │       └── user.NewRepository(pool)
    │               │
    │               ├── auth.NewJWTService(secret)
    │               │       │
    │               │       ├── auth.NewService(userRepo, jwtService)
    │               │       │       └── auth.NewHandler(authService)
    │               │       ├── auth.NewMiddleware(jwtService)
    │               │       └── chat.NewHandler(hub, jwtService)
    │               │
    │               └── chat.NewHub(redisURL)
    │                       └── chat.NewRedisAdapter(redisURL, hub)
    │
    └── server.New(cfg, pool)  →  chi.Router  →  server.Run(ctx)
```

All dependencies are created eagerly in `server.New()`. There is no dependency injection framework — plain constructor functions are composed manually.

---

## Configuration

All configuration is read from environment variables at startup via `config.Load()`.

| Variable       | Type   | Default                                                | Description                                         |
| -------------- | ------ | ------------------------------------------------------ | --------------------------------------------------- |
| `PORT`         | int    | `8080`                                                 | HTTP listen port                                     |
| `DATABASE_URL` | string | `postgres://chat:chat@localhost:5432/chat?sslmode=disable` | PostgreSQL connection string                         |
| `REDIS_URL`    | string | `redis://localhost:6379`                               | Redis / Valkey connection string                     |
| `JWT_SECRET`   | string | `dev-secret-change-in-production`                      | HMAC secret for signing JWT tokens                   |
| `RUN_MIGRATE`  | bool   | `true`                                                 | Run database migrations on startup                   |
| `ALLOWED_ORIGINS` | string | `http://localhost:5173,http://localhost:3000` | Comma-separated list of allowed CORS and WebSocket origins |

> **Production note:** Always change `JWT_SECRET` and `DATABASE_URL` credentials in production. Set `RUN_MIGRATE=false` on all instances except one to avoid migration races.

### Configurable Origins (`ALLOWED_ORIGINS`)

The `ALLOWED_ORIGINS` environment variable controls both the CORS `Access-Control-Allow-Origin` header and the WebSocket `AcceptOptions.OriginPatterns`. Origins are specified as full URLs (e.g., `http://localhost:5173`) separated by commas.

**How it works:**

1. The raw URLs are passed directly to the CORS middleware (`cors.Handler`).
2. For WebSocket origin checking, `Config.WebSocketOriginPatterns()` strips the scheme from each URL, returning just the `host:port` portion (e.g., `localhost:5173`). This matches the format expected by the `coder/websocket` library's `AcceptOptions.OriginPatterns`.

**Examples:**

```bash
# Default (local development)
ALLOWED_ORIGINS="http://localhost:5173,http://localhost:3000"

# Production with a single domain
ALLOWED_ORIGINS="https://chat.example.com"

# Multiple production domains
ALLOWED_ORIGINS="https://chat.example.com,https://chat-staging.example.com"
```

> **Note:** When deploying behind a reverse proxy (e.g., Cloudflare Tunnel), set `ALLOWED_ORIGINS` to the public-facing URL that users access.

---

## API Reference

### `POST /api/auth/register`

Create a new user account.

**Request:**

```json
{
  "username": "alice",
  "password": "secret123"
}
```

**Validation rules:**
- `username` must be 3–50 characters
- `password` must be at least 6 characters

**Success response** — `201 Created`:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "a1b2c3d4-...",
    "username": "alice",
    "created_at": "2026-02-23T12:00:00Z"
  }
}
```

**Error responses:**
- `400 Bad Request` — invalid JSON body, validation failure, or duplicate username

---

### `POST /api/auth/login`

Authenticate with existing credentials.

**Request:**

```json
{
  "username": "alice",
  "password": "secret123"
}
```

**Success response** — `200 OK`:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "a1b2c3d4-...",
    "username": "alice",
    "created_at": "2026-02-23T12:00:00Z"
  }
}
```

**Error responses:**
- `400 Bad Request` — invalid JSON body
- `401 Unauthorized` — wrong username or password

---

### `GET /api/auth/me`

Get the authenticated user's profile. Requires the auth middleware.

**Request header:**

```
Authorization: Bearer <JWT>
```

**Success response** — `200 OK`:

```json
{
  "id": "a1b2c3d4-...",
  "username": "alice"
}
```

**Error responses:**
- `401 Unauthorized` — missing, malformed, or expired token
- `500 Internal Server Error` — user not found in database

---

### `GET /ws?token=<JWT>`

Upgrade to a WebSocket connection. Authentication is performed via the `token` query parameter (not a header, because the browser WebSocket API does not support custom headers).

**Query parameters:**
- `token` (required) — a valid JWT obtained from `/api/auth/login` or `/api/auth/register`

**Success:** HTTP 101 Switching Protocols — WebSocket connection established.

**Error responses:**
- `401 Unauthorized` — missing or invalid token (plain text response, not JSON)

**Allowed origins:** `localhost:5173`, `localhost:3000`

---

### `GET /healthz`

Health check endpoint.

**Response** — `200 OK`:

```
ok
```

---

## WebSocket Protocol

All WebSocket messages are JSON. The shared type definitions live in `pkg/protocol/message.go`.

### Client → Server

#### `send_message`

Send a chat message to all connected users.

```json
{
  "type": "send_message",
  "content": "Hello, everyone!"
}
```

- `content` must be non-empty; the server returns an error message if it is.
- Unknown message types receive an error response.
- Maximum message size: **4096 bytes** (set via `conn.SetReadLimit`).

### Server → Client

#### `chat_message`

A chat message from another user (or yourself).

```json
{
  "type": "chat_message",
  "id": "msg_a1b2c3d4",
  "content": "Hello, everyone!",
  "sender": {
    "id": "a1b2c3d4-...",
    "username": "alice"
  },
  "timestamp": "2026-02-23T12:00:00Z"
}
```

#### `user_joined`

A user connected to the chat.

```json
{
  "type": "user_joined",
  "user": {
    "id": "a1b2c3d4-...",
    "username": "alice"
  },
  "online_count": 5,
  "timestamp": "2026-02-23T12:00:00Z"
}
```

#### `user_left`

A user disconnected from the chat.

```json
{
  "type": "user_left",
  "user": {
    "id": "a1b2c3d4-...",
    "username": "alice"
  },
  "online_count": 4,
  "timestamp": "2026-02-23T12:00:00Z"
}
```

#### `error`

An error message in response to an invalid client action.

```json
{
  "type": "error",
  "message": "message content cannot be empty"
}
```

Possible error messages:
- `"invalid message format"` — JSON parse failure
- `"message content cannot be empty"` — blank content in `send_message`
- `"unknown message type"` — unrecognized `type` field

### Keep-Alive

The server sends WebSocket **ping** frames every **54 seconds** (`pongWait * 9 / 10`, where `pongWait` is 60s). If a pong is not received within 10 seconds (`writeWait`), the connection is closed.

---

## Redis Pub/Sub

### Channel

```
chat:broadcast
```

### Flow

```
Client A (instance 1)
    │  send_message
    ▼
ReadPump  →  hub.broadcast chan  →  hub.Run()  →  redis.Publish("chat:broadcast", msg)
                                                          │
                                          ┌───────────────┼───────────────┐
                                          ▼               ▼               ▼
                                    Instance 1       Instance 2       Instance N
                                    Subscribe()      Subscribe()      Subscribe()
                                          │               │               │
                                          ▼               ▼               ▼
                                    BroadcastLocal   BroadcastLocal   BroadcastLocal
                                          │               │               │
                                          ▼               ▼               ▼
                                    All local        All local        All local
                                    clients          clients          clients
```

### Design Decision

**All messages go through Redis** — even on a single-instance deployment. The `hub.broadcast` channel feeds into `redis.Publish()`, and the `Subscribe()` goroutine calls `hub.BroadcastLocal()` to fan out to local clients. This keeps a single code path and means adding more instances requires zero code changes.

---

## Database

### Schema

**Table: `users`**

| Column       | Type           | Constraints                        |
| ------------ | -------------- | ---------------------------------- |
| `id`         | `UUID`         | `PRIMARY KEY DEFAULT gen_random_uuid()` |
| `username`   | `VARCHAR(50)`  | `NOT NULL UNIQUE`                  |
| `password`   | `VARCHAR(255)` | `NOT NULL`                         |
| `created_at` | `TIMESTAMPTZ`  | `NOT NULL DEFAULT now()`           |

**Indexes:**
- `idx_users_username` on `username` (in addition to the unique constraint)

**Extensions:**
- `pgcrypto` (for `gen_random_uuid()`)

### Migrations

Migrations use [golang-migrate](https://github.com/golang-migrate/migrate) with the filesystem source.

- **Location:** `migrations/` directory in the project root
- **Naming convention:** `000001_create_users_table.up.sql` / `.down.sql`
- **Execution:** Automatically at startup when `RUN_MIGRATE=true` (the default)
- **Idempotent:** `m.Up()` is called; `migrate.ErrNoChange` is silently ignored
- **Version tracking:** golang-migrate manages a `schema_migrations` table in PostgreSQL

**Down migration** (`000001_create_users_table.down.sql`):

```sql
DROP TABLE IF EXISTS users;
```

### Connection Pooling

The database connection is managed by `pgxpool.Pool` (from `jackc/pgx/v5`). The pool is created in `main.go`, pinged to verify connectivity, and passed to the user repository. The pool is closed via `defer pool.Close()` on server shutdown.

---

## Hub Architecture

The `Hub` is the central message broker for WebSocket connections on a single server instance.

### Data Structures

```go
type Hub struct {
    clients    map[*Client]bool   // registered clients
    broadcast  chan []byte         // inbound messages from clients
    register   chan *Client        // client registration requests
    unregister chan *Client        // client removal requests
    mu         sync.RWMutex       // protects the clients map
    redis      *RedisAdapter      // Pub/Sub adapter
}
```

### Client Lifecycle

1. **WebSocket upgrade** — `chat.Handler.HandleWebSocket` validates the JWT, accepts the WebSocket connection, creates a `Client`, and sends it to `hub.register`.
2. **Registration** — `Hub.Run()` receives the client, adds it to the map, publishes a `user_joined` message via Redis.
3. **ReadPump** — Runs in the HTTP handler goroutine. Reads messages from the WebSocket, parses them, and sends `chat_message` payloads to `hub.broadcast`.
4. **WritePump** — Runs in a separate goroutine. Pulls from `client.send` channel and writes to the WebSocket. Also manages ping/pong keep-alive.
5. **Disconnection** — When `ReadPump` exits (read error or context cancellation), the client is sent to `hub.unregister`. The hub removes it from the map, closes the `send` channel, and publishes a `user_left` message via Redis.

### Broadcast Flow

```
client.ReadPump()
    → json.Unmarshal → NewChatMessage()
    → hub.broadcast <- data
        → hub.Run() receives from broadcast
        → redis.Publish(ctx, data)
            → Redis PUBLISH "chat:broadcast"
            → redis.Subscribe() receives
            → hub.BroadcastLocal(message)
                → for each client: client.send <- message
                    → client.WritePump() writes to WebSocket
```

### Goroutine Model

Per WebSocket connection, two goroutines run:

| Goroutine    | Started by                    | Responsibility                            |
| ------------ | ----------------------------- | ----------------------------------------- |
| `ReadPump`   | `HandleWebSocket` (blocking)  | Read messages, dispatch to hub            |
| `WritePump`  | `HandleWebSocket` (`go`)      | Write messages, send pings                |

Additionally, `Hub.Run()` runs as a single goroutine managing all client state, and `RedisAdapter.Subscribe()` runs as a goroutine receiving published messages.

---

## Running Locally

### Prerequisites

- Go 1.24+
- PostgreSQL 17 (or compatible)
- Valkey 8 / Redis 7+ (Redis-compatible)

### Option 1: Direct

1. Start Postgres and Redis/Valkey locally.

2. Set environment variables (or use the defaults):

   ```bash
   export DATABASE_URL="postgres://chat:chat@localhost:5432/chat?sslmode=disable"
   export REDIS_URL="redis://localhost:6379"
   export JWT_SECRET="some-secret"
   ```

3. Run the server:

   ```bash
   make run-server
   # or directly:
   go run ./cmd/server
   ```

4. Build a binary:

   ```bash
   make build-server
   ./bin/server
   ```

### Option 2: Docker Compose

Docker Compose brings up Postgres, Valkey, two backend instances, and the frontend:

```bash
make docker-up      # docker compose up --build -d
make docker-logs    # docker compose logs -f
make docker-down    # docker compose down
```

**Services:**

| Service      | Host Port | Description                                     |
| ------------ | --------- | ----------------------------------------------- |
| `postgres`   | 5432      | PostgreSQL 17                                    |
| `valkey`     | 6379      | Valkey 8 (Redis-compatible)                      |
| `backend-1`  | 8080      | Backend instance 1 (runs migrations)             |
| `backend-2`  | 8081      | Backend instance 2 (`RUN_MIGRATE=false`)         |
| `frontend`   | 5173      | React frontend served by nginx                   |

The two backend instances demonstrate horizontal scaling via Redis Pub/Sub. `backend-1` runs migrations; `backend-2` waits for `backend-1` to be healthy before starting.

### Dockerfile

The backend uses a **multi-stage Alpine build** (`Dockerfile.backend`):

1. **Builder stage** — `golang:1.24-alpine`, downloads dependencies, compiles a static binary (`CGO_ENABLED=0`).
2. **Runtime stage** — `alpine:3.21`, copies the binary and migrations directory. Exposes port 8080.

---

## Graceful Shutdown

### Signal Handling

`main.go` creates a context using `signal.NotifyContext` that cancels on `SIGINT` (Ctrl+C) or `SIGTERM` (container stop).

### Shutdown Sequence

1. **Signal received** → context is canceled.
2. **`server.Run()`** — detects `ctx.Done()`, calls `srv.Shutdown()` with a **10-second timeout**. This stops accepting new connections and waits for in-flight HTTP requests to complete.
3. **`hub.Run()`** — detects `ctx.Done()`, exits its event loop. The Redis subscriber also exits on context cancellation.
4. **`pool.Close()`** — the deferred call in `main.go` closes all database connections.

### CORS

The server allows cross-origin requests from:
- `http://localhost:5173` (Vite dev server)
- `http://localhost:3000`

Allowed methods: `GET`, `POST`, `OPTIONS`. Credentials are allowed. Preflight responses are cached for 300 seconds.

### Middleware Stack

Applied globally via Chi:
1. `chimiddleware.Logger` — logs each request
2. `chimiddleware.Recoverer` — recovers from panics, returns 500
3. `chimiddleware.RequestID` — assigns a unique request ID header
4. `cors.Handler` — CORS configuration

---

## Deployment Considerations

### Environment Variables for Production

| Variable | Recommendation |
|---|---|
| `JWT_SECRET` | Use a strong, unique secret (at least 32 characters). Never use the default. |
| `DATABASE_URL` | Use a dedicated database user with minimal permissions. Enable `sslmode=require` or `sslmode=verify-full`. |
| `REDIS_URL` | Use authentication (`redis://user:password@host:port`) and TLS if available. |
| `ALLOWED_ORIGINS` | Set to the exact production domain(s). Never use wildcard `*` with credentials. |
| `RUN_MIGRATE` | Enable on exactly one instance. Disable on all others to prevent migration races. |

### Database Connection Pooling

The backend uses `pgxpool.Pool` from the `jackc/pgx/v5` library for connection pooling. Key characteristics:

- **Default pool size:** pgxpool defaults to `max(4, runtime.NumCPU())` connections
- **Connection lifetime:** Connections are recycled automatically
- **Health checking:** The pool pings the database on startup to verify connectivity
- **Cleanup:** `pool.Close()` is deferred in `main.go`, ensuring all connections are returned on shutdown

To tune the pool, append query parameters to `DATABASE_URL`:

```
postgres://user:pass@host:port/db?sslmode=disable&pool_max_conns=10&pool_min_conns=2
```

### Horizontal Scaling

The application supports multiple backend instances out of the box:

1. All chat messages route through Redis Pub/Sub (`chat:broadcast` channel)
2. Each instance subscribes to the channel and broadcasts to its local clients
3. User join/leave events are also published through Redis
4. Stateless JWT authentication means any instance can validate tokens

**Requirements for scaling:**
- All instances must share the same `DATABASE_URL`, `REDIS_URL`, and `JWT_SECRET`
- Only one instance should have `RUN_MIGRATE=true`
- A load balancer (e.g., nginx with `ip_hash`) should distribute WebSocket connections
