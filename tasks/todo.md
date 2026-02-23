# WebSocket Chat — Implementation Progress

## Phase 1: Scaffolding
- [ ] Initialize Go module, create directory structure
- [ ] Create config.go
- [ ] Write docker-compose.yml (Postgres + Valkey)
- [ ] Write migration SQL
- [ ] Implement database.go and migrations.go
- [ ] Create shared protocol types (pkg/protocol/message.go)
- [ ] Create Makefile
- [ ] Verify: docker compose up, migrations run

## Phase 2: Authentication
- [ ] user/model.go and user/repository.go
- [ ] auth/password.go (bcrypt)
- [ ] auth/jwt.go
- [ ] auth/service.go
- [ ] auth/handler.go
- [ ] auth/middleware.go
- [ ] Wire routes in server/server.go with CORS
- [ ] Verify: curl register → login → /me

## Phase 3: WebSocket Chat (Single Instance)
- [ ] chat/message.go
- [ ] chat/client.go (readPump, writePump)
- [ ] chat/hub.go
- [ ] chat/handler.go (WS upgrade + JWT)
- [ ] Wire /ws route
- [ ] Verify: connect and send/receive messages

## Phase 4: Multi-Instance via Redis
- [ ] chat/redis.go (RedisAdapter)
- [ ] Modify Hub to use Redis
- [ ] Add backend-2 to docker-compose
- [ ] Verify: cross-instance messaging

## Phase 5: TUI Abstractions
- [ ] tui/styles/theme.go
- [ ] tui/components/textinput.go
- [ ] tui/components/messagelist.go
- [ ] tui/components/statusbar.go
- [ ] tui/components/header.go
- [ ] tui/components/dialog.go

## Phase 6: TUI Application
- [ ] tui/client/auth.go
- [ ] tui/client/token.go
- [ ] tui/client/ws.go
- [ ] tui/screens/login.go
- [ ] tui/screens/register.go
- [ ] tui/screens/chat.go
- [ ] tui/app.go
- [ ] cmd/tui/main.go
- [ ] Verify: login, send/receive messages

## Phase 7: Frontend
- [ ] Scaffold Vite + React + TypeScript + Tailwind
- [ ] api/auth.ts
- [ ] AuthContext + useAuth
- [ ] LoginForm + RegisterForm
- [ ] useWebSocket hook
- [ ] ChatContext
- [ ] Chat components (MessageList, MessageItem, MessageInput, etc.)
- [ ] ChatRoom + App.tsx
- [ ] Verify: end-to-end in browser

## Phase 8: Production Readiness
- [ ] backend/Dockerfile
- [ ] frontend/Dockerfile + nginx.conf
- [ ] Update docker-compose with frontend
- [ ] Graceful shutdown + /healthz
- [ ] Verify: docker compose up --build

## Phase 9: Testing
- [ ] Backend unit tests
- [ ] Backend integration tests
- [ ] Frontend component tests
- [ ] Manual test checklist
