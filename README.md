# WebSocket Chat

Ephemeral real-time chat application with three client options: a React web app, a TUI terminal client, and any WebSocket-capable tool. Features JWT-based authentication, horizontal scaling via Redis Pub/Sub, and PostgreSQL for user storage.

## Architecture

```
Browser ──┐                              ┌── Backend 1 ──┐
          ├── nginx (load balancer) ────┤               ├── Valkey (Redis)
Browser ──┘                              └── Backend 2 ──┘       │
                                                                  │
TUI ────────────────── WebSocket ──────────── Backend ────────────┘
                                                  │
                                              PostgreSQL
```

## Tech Stack

| Component | Technology |
|---|---|
| Backend | Go 1.24, chi router, coder/websocket |
| Database | PostgreSQL 17, pgx v5 connection pool |
| Cache/PubSub | Valkey 8 (Redis-compatible) |
| Auth | JWT (HS256), bcrypt password hashing |
| Frontend | React 19, TypeScript 5.9, Vite 7, Tailwind CSS 4 |
| TUI | Go, Bubble Tea, Lip Gloss |
| Proxy | nginx with ip_hash load balancing |
| Containers | Docker Compose |

## Quick Start

```bash
docker compose up --build
```

Open http://localhost:5173

## Database Configuration

- **Connection string format:** `postgres://user:pass@host:port/db?sslmode=disable`
- **Connection pooling:** Uses pgxpool from jackc/pgx for efficient connection management
- **Automatic migrations:** Migrations run on first backend start when `RUN_MIGRATE=true`
- **Migration tracking:** Applied migrations are tracked in the `schema_migrations` table
- **Multi-instance safety:** Only enable `RUN_MIGRATE=true` on one instance to avoid migration races. In the default Docker Compose setup, only `backend-1` runs migrations.

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP listen port |
| `DATABASE_URL` | `postgres://chat:chat@localhost:5432/chat?sslmode=disable` | PostgreSQL connection string |
| `REDIS_URL` | `redis://localhost:6379` | Valkey/Redis connection string |
| `JWT_SECRET` | `dev-secret-change-in-production` | HMAC secret for JWT signing |
| `RUN_MIGRATE` | `true` | Run database migrations on startup |
| `ALLOWED_ORIGINS` | `http://localhost:5173,http://localhost:3000` | Comma-separated list of allowed CORS and WebSocket origins |

## TUI Client

```bash
make build-tui
./bin/tui --server ws://localhost:8080
# Or connect through nginx:
./bin/tui --server ws://localhost:5173
```

## Development

**Prerequisites:** Go 1.24+, Node.js 22+, PostgreSQL, Valkey/Redis

```bash
# Backend
make run-server

# Frontend
cd frontend && npm install && npm run dev

# TUI
make run-tui
```

## Docker

| Service | Port | Description |
|---|---|---|
| postgres | 5432 | PostgreSQL 17 |
| valkey | 6379 | Valkey 8 (Redis-compatible) |
| backend-1 | 8080 | Backend (runs migrations) |
| backend-2 | 8081 | Backend replica |
| frontend | 5173 | React app via nginx |

Two backend instances demonstrate horizontal scaling via Redis Pub/Sub. Messages sent to one backend are broadcast to clients connected to the other. The nginx reverse proxy in the frontend container uses ip_hash load balancing to route WebSocket connections consistently.

## CI/CD

GitHub Actions runs Go tests and frontend tests on all pushes and pull requests to `main`. On push to `main`, it builds and pushes Docker images to Docker Hub:

- `<username>/websocket-chat-backend`
- `<username>/websocket-chat-frontend`

**Required secrets:** `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`

## Cloudflare Tunnel

For production deployment behind Cloudflare Tunnels, see [docs/cloudflare-tunnel.md](docs/cloudflare-tunnel.md).

## Project Structure

```
├── cmd/
│   ├── server/          # Backend entry point
│   └── tui/             # TUI entry point
├── internal/
│   ├── auth/            # JWT auth, bcrypt, middleware
│   ├── chat/            # WebSocket hub, clients, Redis pub/sub
│   ├── config/          # Environment config
│   ├── database/        # Connection pool, migrations
│   ├── server/          # HTTP router setup
│   └── user/            # User model & repository
├── tui/
│   ├── client/          # REST + WebSocket clients
│   ├── components/      # UI components
│   ├── screens/         # Login, Register, Chat screens
│   └── styles/          # Color theme
├── pkg/protocol/        # Shared message types
├── frontend/            # React SPA
├── migrations/          # SQL migrations
├── docs/                # Documentation
├── docker-compose.yml
├── Dockerfile.backend
└── Makefile
```

## Documentation

- [Backend](docs/backend.md)
- [Frontend](docs/frontend.md)
- [TUI Client](docs/tui.md)
- [Cloudflare Tunnel](docs/cloudflare-tunnel.md)
