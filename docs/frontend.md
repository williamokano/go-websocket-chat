# Frontend Documentation

## Overview

The frontend is a React single-page application (SPA) that provides a real-time chat interface. Users authenticate via username/password, then communicate over a WebSocket connection. The UI updates in real time as messages arrive and users join or leave.

### Tech Stack

| Technology | Version | Purpose |
|---|---|---|
| React | 19.x | UI framework |
| TypeScript | 5.9 | Type safety |
| Vite | 7.x | Build tool and dev server |
| Tailwind CSS | 4.x | Utility-first styling via `@tailwindcss/vite` plugin |
| nginx | alpine | Production static file server and reverse proxy |
| Node.js | 22 (alpine) | Build environment (Docker) |

No component library is used. All styling is done with Tailwind utility classes directly in JSX.

## Architecture

### Component Tree

```
<StrictMode>
  <AuthProvider>              ← provides auth state to entire app
    <AuthGate>                ← switches between auth forms and chat
      ├── (isLoading)         → Loading spinner
      ├── (isAuthenticated)
      │   └── <ChatProvider>  ← provides WebSocket/chat state
      │       └── <ChatRoom>
      │           ├── <header>
      │           │   ├── <ConnectionStatus />
      │           │   └── <UserInfo />
      │           ├── <MessageList>
      │           │   └── <MessageItem /> (per message)
      │           └── <MessageInput />
      └── (!isAuthenticated)
          ├── <LoginForm />
          └── <RegisterForm />
```

### Provider Hierarchy

1. **`AuthProvider`** wraps the entire app (mounted in `main.tsx`). Manages user identity, JWT token, and auth operations.
2. **`ChatProvider`** wraps only the authenticated chat view (mounted inside `AuthGate` when `isAuthenticated` is true). Manages WebSocket connection, message history, and online count.

This separation ensures the WebSocket connection is only created when a user is authenticated, and is torn down on logout.

### Data Flow

```
User action → Component → Hook/Context → API/WebSocket → Server
Server event → WebSocket → Hook → Context → Component re-render
```

## Type Definitions

All shared types are defined in `src/types/index.ts`:

### Domain Types

```typescript
interface User {
  id: string;
  username: string;
}

type ConnectionStatus = "connected" | "connecting" | "disconnected";
```

### Auth Types

```typescript
interface LoginRequest   { username: string; password: string; }
interface RegisterRequest { username: string; password: string; }
interface AuthResponse    { token: string; user: User; }
```

### Message Types

```typescript
// Client → Server
interface ClientMessage {
  type: "send_message";
  content: string;
}

// Server → Client (discriminated union)
type ServerMessage = ChatMessage | UserJoinedMessage | UserLeftMessage | ErrorMessage;

interface ChatMessage {
  type: "chat_message";
  id: string;
  content: string;
  sender: User;
  timestamp: string;
}

interface UserJoinedMessage {
  type: "user_joined";
  user: User;
  online_count: number;
  timestamp: string;
}

interface UserLeftMessage {
  type: "user_left";
  user: User;
  online_count: number;
  timestamp: string;
}

interface ErrorMessage {
  type: "error";
  message: string;
}
```

## Auth Flow

### Token Storage

The JWT token is stored in `localStorage` under the key `"chat_token"`.

### Startup Sequence

```
App mounts
  → AuthProvider reads "chat_token" from localStorage
    → If no token: set isLoading=false, show LoginForm
    → If token exists: call GET /api/auth/me with Bearer header
      → Success: set user + token in state, show ChatRoom
      → Failure: remove token from localStorage, show LoginForm
```

### Login Flow

```
User submits LoginForm
  → POST /api/auth/login { username, password }
    → Success: store token in localStorage, set user + token in state
      → isAuthenticated becomes true
      → AuthGate renders ChatProvider → ChatRoom
      → ChatProvider creates WebSocket connection
    → Failure: display error message in form
```

### Register Flow

```
User submits RegisterForm
  → Client-side validation:
    → Password must be >= 6 characters
    → Password and confirm must match
  → POST /api/auth/register { username, password }
    → Success: store token in localStorage, set user + token in state
      → Same flow as login success
    → Failure: display error message in form
```

### Logout Flow

```
User clicks "Sign out"
  → Remove "chat_token" from localStorage
  → Clear user and token from state
  → isAuthenticated becomes false
  → ChatProvider unmounts → WebSocket closes
  → AuthGate renders LoginForm
```

### JWT Expiry

When the server detects an expired or invalid JWT on the WebSocket connection, it closes the socket with close code **4001**. The `useWebSocket` hook detects this code, disables reconnection, and calls `onLogout()` which triggers the full logout flow.

## API Client

Located in `src/api/auth.ts`. All functions use relative URLs (e.g., `/api/auth/login`) which are proxied by nginx in production or Vite dev server config in development.

| Function | Method | Endpoint | Auth | Returns |
|---|---|---|---|---|
| `login(username, password)` | POST | `/api/auth/login` | None | `AuthResponse` |
| `register(username, password)` | POST | `/api/auth/register` | None | `AuthResponse` |
| `getMe(token)` | GET | `/api/auth/me` | `Bearer {token}` | `User` |

Error handling: on non-OK responses, the functions attempt to parse a JSON body with an `error` field. If parsing fails, a generic error message is thrown.

## WebSocket Integration

### Connection Setup

The `useWebSocket` hook (`src/hooks/useWebSocket.ts`) manages the WebSocket lifecycle.

**URL derivation:**
```
Protocol: window.location.protocol === "https:" ? "wss:" : "ws:"
URL:      {proto}//{window.location.host}/ws?token={jwt}
```

The JWT token is passed as a query parameter `token`.

### Reconnection Strategy

Exponential backoff with the following parameters:

| Parameter | Value |
|---|---|
| Initial delay | 1,000 ms |
| Backoff multiplier | 2x |
| Maximum delay | 30,000 ms |
| Reset on | Successful connection (`onopen`) |

**Sequence:** 1s, 2s, 4s, 8s, 16s, 30s, 30s, 30s, ...

Reconnection is disabled when:
- Close code is **4001** (invalid JWT) — triggers logout instead
- The component unmounts (cleanup in `useEffect`)

### Close Code 4001 Handling

When the WebSocket receives close code 4001:
1. `shouldReconnectRef` is set to `false` (prevents reconnection)
2. `onLogout()` is called (clears auth state, returns to login)

### Message Handling

**Receiving:** All incoming WebSocket messages are parsed as JSON and appended to the `messages` state array. The `ServerMessage` discriminated union type is used.

**Sending:** The `sendMessage` function creates a `ClientMessage` (`{ type: "send_message", content }`) and sends it as JSON. Sending is silently ignored if the socket is not in `OPEN` state.

### Connection Status

The hook exposes a `connectionStatus` state with three possible values:
- `"connected"` — WebSocket is open
- `"connecting"` — WebSocket is being established
- `"disconnected"` — WebSocket is closed (may be reconnecting)

## Contexts

### AuthContext

**File:** `src/contexts/AuthContext.tsx`

Wraps the entire app. Provides:

| Field | Type | Description |
|---|---|---|
| `user` | `User \| null` | The authenticated user, or null |
| `token` | `string \| null` | The JWT token, or null |
| `isAuthenticated` | `boolean` | `true` when both `user` and `token` are non-null |
| `isLoading` | `boolean` | `true` during initial token validation on mount |
| `login` | `(username, password) => Promise<void>` | Authenticates and stores token |
| `register` | `(username, password) => Promise<void>` | Creates account and stores token |
| `logout` | `() => void` | Clears token and user state |

### ChatContext

**File:** `src/contexts/ChatContext.tsx`

Wraps only the authenticated chat view. Provides:

| Field | Type | Description |
|---|---|---|
| `messages` | `ServerMessage[]` | All messages received during the session |
| `sendMessage` | `(content: string) => void` | Sends a chat message over the WebSocket |
| `connectionStatus` | `ConnectionStatus` | Current WebSocket connection state |
| `onlineCount` | `number` | Derived from the most recent `user_joined` or `user_left` message |

The `onlineCount` is computed by scanning the messages array in reverse for the latest `user_joined` or `user_left` event and reading its `online_count` field. Returns `0` if no such event exists.

## Hooks

### `useAuth()`

**File:** `src/hooks/useAuth.ts`

Convenience hook that reads `AuthContext`. Throws if used outside `AuthProvider`.

**Returns:** `AuthContextValue` (same shape as the context).

### `useWebSocket({ token, onLogout })`

**File:** `src/hooks/useWebSocket.ts`

Manages the full WebSocket lifecycle: connection, reconnection, message accumulation, and sending.

**Parameters:**
- `token: string` — JWT token for authentication
- `onLogout: () => void` — Called when close code 4001 is received

**Returns:**
- `messages: ServerMessage[]` — Accumulated messages
- `sendMessage: (content: string) => void` — Send a chat message
- `connectionStatus: ConnectionStatus` — Current connection state

### `useChat()`

**File:** `src/hooks/useChat.ts`

Convenience hook that reads `ChatContext`. Throws if used outside `ChatProvider`.

**Returns:** `ChatContextValue` (same shape as the context).

## Components

### LoginForm

**File:** `src/components/LoginForm/index.tsx`

| | |
|---|---|
| **Purpose** | Username/password sign-in form |
| **Props** | `onSwitchToRegister: () => void` — callback to switch to register view |
| **Key behavior** | Calls `login()` from `useAuth()`. Displays error messages from failed login attempts. Shows loading state on the submit button. |
| **Used in** | `AuthGate` (in `App.tsx`) when `view === "login"` |

### RegisterForm

**File:** `src/components/RegisterForm/index.tsx`

| | |
|---|---|
| **Purpose** | Account creation form with password confirmation |
| **Props** | `onSwitchToLogin: () => void` — callback to switch to login view |
| **Key behavior** | Client-side validation: password >= 6 chars, password and confirm must match. Calls `register()` from `useAuth()`. Displays error messages. Shows loading state. |
| **Used in** | `AuthGate` (in `App.tsx`) when `view === "register"` |

### ChatRoom

**File:** `src/components/ChatRoom/index.tsx`

| | |
|---|---|
| **Purpose** | Main chat interface — header with status/user info, message list, and input |
| **Props** | None |
| **Key behavior** | Composes `ConnectionStatus`, `UserInfo`, `MessageList`, and `MessageInput`. Reads from both `useAuth()` (for user/logout) and `useChat()` (for messages/sending/status). Uses a full-height flexbox layout (`h-screen`). |
| **Used in** | `ChatProvider` (inside `AuthGate` when authenticated) |

### MessageList

**File:** `src/components/MessageList/index.tsx`

| | |
|---|---|
| **Purpose** | Scrollable list of all messages |
| **Props** | `messages: ServerMessage[]` |
| **Key behavior** | Auto-scrolls to bottom when new messages arrive (via a ref and `scrollIntoView({ behavior: "smooth" })`). Shows "No messages yet. Say hello!" when empty. Uses chat message `id` as React key for `ChatMessage` types, falls back to array index for system messages. |
| **Used in** | `ChatRoom` |

### MessageItem

**File:** `src/components/MessageItem/index.tsx`

| | |
|---|---|
| **Purpose** | Renders a single message based on its type |
| **Props** | `message: ServerMessage` |
| **Key behavior** | Discriminates on `message.type` to render different layouts: |
| | - `chat_message`: sender name, timestamp, and content (preserves whitespace) |
| | - `user_joined`: centered gray text ("{username} joined . {time} . {count} online") |
| | - `user_left`: centered gray text ("{username} left . {time} . {count} online") |
| | - `error`: centered red text with the error message |
| **Used in** | `MessageList` |

Timestamps are formatted using `Date.toLocaleTimeString` with `hour: "2-digit"` and `minute: "2-digit"`.

### MessageInput

**File:** `src/components/MessageInput/index.tsx`

| | |
|---|---|
| **Purpose** | Auto-resizing textarea with send button |
| **Props** | `onSend: (content: string) => void`, `connectionStatus: ConnectionStatus` |
| **Key behavior** | Enter sends the message; Shift+Enter inserts a newline. Textarea auto-resizes up to 120px max height. Input and button are disabled when not connected. Placeholder changes to "Reconnecting..." when disconnected. Trims whitespace before sending. |
| **Used in** | `ChatRoom` |

### ConnectionStatus

**File:** `src/components/ConnectionStatus/index.tsx`

| | |
|---|---|
| **Purpose** | Colored dot indicator with connection state text |
| **Props** | `status: ConnectionStatus`, `onlineCount: number` |
| **Key behavior** | Shows a colored dot (green/yellow/red) and label ("Connected"/"Connecting..."/"Disconnected"). When connected and `onlineCount > 0`, also displays the online user count. |
| **Used in** | `ChatRoom` header |

### UserInfo

**File:** `src/components/UserInfo/index.tsx`

| | |
|---|---|
| **Purpose** | Displays current username and sign-out button |
| **Props** | `user: User`, `onLogout: () => void` |
| **Key behavior** | Shows the username and a "Sign out" button that calls `onLogout`. |
| **Used in** | `ChatRoom` header |

## Styling

Tailwind CSS v4 is used via the `@tailwindcss/vite` plugin (no PostCSS config needed). The entire CSS entry point is:

```css
/* src/index.css */
@import "tailwindcss";
```

### Design Approach

- **No component library** — all styling uses Tailwind utility classes directly in JSX
- **Color palette** — gray for backgrounds/text, blue for primary actions, red for errors, green/yellow/red for connection status dots
- **Layout** — full-screen flexbox layout for the chat view (`h-screen`, `flex flex-col`); centered card layout for auth forms
- **Responsive** — auth forms use `max-w-sm` with `px-4` padding for mobile; chat view is full-width
- **Interactive states** — hover effects on buttons and message rows, focus rings on inputs, disabled opacity on buttons

## Development

### Prerequisites

- Node.js 22+
- npm

### Setup

```bash
cd frontend
npm install
npm run dev
```

The Vite dev server starts on **port 5173** by default.

### Available Scripts

| Script | Command | Description |
|---|---|---|
| `dev` | `vite` | Start Vite dev server with HMR |
| `build` | `tsc -b && vite build` | Type-check then build for production |
| `lint` | `eslint .` | Run ESLint |
| `preview` | `vite preview` | Preview the production build locally |

### Dev Server Notes

In development, API and WebSocket requests to `/api/*` and `/ws` need to reach the backend. This can be handled by either:
1. Running the backend on the same host and configuring a Vite proxy
2. Running the full Docker Compose stack and accessing via nginx

The current `vite.config.ts` does not include proxy configuration, so the full stack (via Docker Compose) is the simplest way to develop with a working backend.

### Environment & Build Configuration

The frontend has no build-time environment variables — all configuration is derived at runtime from `window.location`:

| Configuration | Source | Example |
|---|---|---|
| API base URL | `window.location.origin` (relative `/api/*` paths) | `/api/auth/login` |
| WebSocket URL | Protocol from `window.location.protocol`, host from `window.location.host` | `wss://chat.example.com/ws?token=...` |

This means the same build artifact works in any environment — no rebuild needed when deploying to different domains.

**Why no `.env` files?**

The React app uses relative URLs for API calls (`/api/auth/login`) and constructs WebSocket URLs from `window.location`. This design means:
- No environment-specific builds
- Works behind any reverse proxy that routes `/api/*` and `/ws` to the backend
- The nginx configuration handles all routing in production
- Vite's dev server proxy (if configured) handles routing in development

### Development Workflow

**Option A: Full stack via Docker Compose (recommended)**

The simplest way to develop with a working backend:

```bash
# From the project root
docker compose up --build -d

# Frontend is available at http://localhost:5173
# Changes require rebuilding: docker compose up --build frontend
```

**Option B: Local frontend with Docker backend**

Run the backend services in Docker but the frontend locally for hot-reload:

```bash
# Start backend services
docker compose up --build -d postgres valkey backend-1 backend-2

# Run frontend dev server (in a separate terminal)
cd frontend
npm install
npm run dev
```

Then configure Vite to proxy API requests to the backend. Add to `vite.config.ts`:

```typescript
export default defineConfig({
  // ... existing config
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
      },
    },
  },
})
```

**Option C: Everything local**

Run Go backend and frontend dev server locally (requires Go, Node.js, PostgreSQL, and Valkey/Redis):

```bash
# Terminal 1: Backend
make run-server

# Terminal 2: Frontend
cd frontend && npm run dev
```

### Testing

```bash
cd frontend
npm test          # Run tests with Vitest
npm run lint      # Run ESLint
npm run build     # Type-check + production build
```

## Production Build

### Build Process

```bash
npm run build
```

This runs TypeScript compilation (`tsc -b`) followed by the Vite production build. Output goes to the `dist/` directory.

### Docker

The `Dockerfile` uses a multi-stage build:

1. **Builder stage** (`node:22-alpine`):
   - Installs dependencies with `npm ci`
   - Runs `npm run build` to produce `dist/`

2. **Runtime stage** (`nginx:alpine`):
   - Copies `dist/` to nginx's HTML root (`/usr/share/nginx/html`)
   - Copies `nginx.conf` as the default server config
   - Exposes port 80

### nginx Configuration

The `nginx.conf` configures:

| Location | Behavior |
|---|---|
| `/api/` | Proxies to the `backend` upstream (load-balanced) |
| `/ws` | Proxies to the `backend` upstream with WebSocket upgrade headers. Sets `proxy_read_timeout 86400` (24h) to keep connections alive. |
| `/` | Serves static files from `/usr/share/nginx/html`. Uses `try_files $uri $uri/ /index.html` for SPA routing. |

The `backend` upstream uses `ip_hash` load balancing across `backend-1:8080` and `backend-2:8080`.

## TypeScript Configuration

The project uses a split tsconfig setup:

- `tsconfig.json` — root config, references `tsconfig.app.json` and `tsconfig.node.json`
- `tsconfig.app.json` — app code config (target ES2020, strict mode, react-jsx, includes `src/`)
- `tsconfig.node.json` — Vite config file compilation

Strict mode is enabled with `noUnusedLocals`, `noUnusedParameters`, and `noFallthroughCasesInSwitch`.

## File Structure

```
frontend/
  docs/
    frontend.md          ← this file
  src/
    api/
      auth.ts            ← HTTP client for auth endpoints
    components/
      ChatRoom/
        index.tsx         ← main chat layout
      ConnectionStatus/
        index.tsx         ← connection indicator dot + label
      LoginForm/
        index.tsx         ← sign-in form
      MessageInput/
        index.tsx         ← auto-resizing textarea + send button
      MessageItem/
        index.tsx         ← single message renderer
      MessageList/
        index.tsx         ← scrollable message container
      RegisterForm/
        index.tsx         ← account creation form
      UserInfo/
        index.tsx         ← username display + sign-out button
    contexts/
      AuthContext.tsx      ← auth state provider
      ChatContext.tsx      ← chat/WebSocket state provider
    hooks/
      useAuth.ts          ← convenience hook for AuthContext
      useChat.ts          ← convenience hook for ChatContext
      useWebSocket.ts     ← WebSocket connection management
    types/
      index.ts            ← shared TypeScript interfaces
    App.tsx               ← AuthGate routing component
    index.css             ← Tailwind CSS entry point
    main.tsx              ← React root + AuthProvider mount
    vite-env.d.ts         ← Vite type declarations
  Dockerfile              ← multi-stage build (node → nginx)
  index.html              ← HTML entry point
  nginx.conf              ← production reverse proxy config
  package.json            ← dependencies and scripts
  tsconfig.json           ← TypeScript project references
  tsconfig.app.json       ← app TypeScript config
  tsconfig.node.json      ← Vite config TypeScript config
  vite.config.ts          ← Vite + React + Tailwind plugins
```
