# TUI Client

Terminal-based chat client that connects to the WebSocket server. Built with Go using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework.

## Overview

The TUI is a native terminal application — it runs directly on the developer's machine, not inside Docker. It connects to the chat server over WebSocket, authenticates via REST endpoints, and presents a full-screen terminal UI for real-time messaging.

### Tech Stack

| Dependency | Purpose |
|---|---|
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | TUI framework (Elm Architecture) |
| [Lip Gloss](https://github.com/charmbracelet/lipgloss) | Terminal styling (colors, borders, layout) |
| [coder/websocket](https://github.com/coder/websocket) | WebSocket client library |

### Source Layout

```
cmd/tui/main.go          Entry point, CLI flags, program bootstrap
tui/
  app.go                  Root model — screen state machine
  client/
    auth.go               REST client for login/register/me
    token.go              JWT persistence at ~/.config/chat/token
    ws.go                 WebSocket connection manager
  components/
    textinput.go          Styled text input with cursor, masking
    messagelist.go        Scrollable message viewport
    statusbar.go          Bottom bar — connection status + help
    header.go             Top bar — title + username
    dialog.go             Centered modal with multiple fields
    menu.go               Centered overlay menu with selectable items
  screens/
    login.go              Login dialog
    register.go           Registration dialog
    chat.go               Main chat view (header + messages + input + status)
  styles/
    theme.go              Color palette and Lip Gloss style definitions
```

## Elm Architecture

Bubble Tea uses the **Model-Update-View** pattern (the Elm Architecture):

1. **Model** — A struct holding the entire application state.
2. **Update** — A pure function `(Model, Msg) → (Model, Cmd)` that handles messages and returns the new state plus optional side-effect commands.
3. **View** — A function `Model → string` that renders the current state as a string for the terminal.

### Message Flow

```
 ┌────────────────────────────────────────┐
 │              tea.Program               │
 │                                        │
 │   ┌──────┐   Msg   ┌────────┐         │
 │   │ View │◄────────│ Update │         │
 │   └──┬───┘         └───▲────┘         │
 │      │    render        │ dispatch     │
 │      ▼                  │              │
 │   terminal         tea.Msg             │
 │                    (keyboard,          │
 │                     window resize,     │
 │                     custom messages)   │
 └────────────────────────────────────────┘
```

- **`tea.Msg`** — Any Go value. The framework delivers key presses, window size changes, and custom messages (like `LoginSuccessMsg` or `IncomingChatMsg`) to the `Update` function.
- **`tea.Cmd`** — A function `() → tea.Msg` that performs async work (HTTP call, WebSocket read) and returns a message when done. The framework runs commands in goroutines and feeds results back into `Update`.
- **`tea.Program.Send()`** — Injects a message from outside the Elm loop (used by the WebSocket read goroutine to push incoming chat messages).

## Screen Flow

```
                  ┌──────────────────┐
                  │   App.Init()     │
                  │   Load stored    │
                  │   JWT token      │
                  └────────┬─────────┘
                           │
                    token found?
                   ┌───yes─┴──no───┐
                   │               │
                   ▼               ▼
           ┌──────────────┐  ┌──────────────┐
           │  GET /api/   │  │    Login      │
           │  auth/me     │  │    Screen     │
           └──────┬───────┘  └───┬──────┬───┘
                  │              │      │
            valid?               │   Ctrl+R
           ┌─yes──┴──no──┐      │      │
           │             │      │      ▼
           │             ▼      │  ┌──────────────┐
           │       Login Screen │  │   Register   │
           │                    │  │   Screen     │
           │                    │  └──────┬───────┘
           │                    │         │
           │              LoginSuccessMsg │ RegisterSuccessMsg
           │                    │         │
           ▼                    ▼         ▼
      ┌──────────────────────────────────────┐
      │            Chat Screen               │
      │   WebSocket connected, messages      │
      │   flowing in real time               │
      └──────────────────────────────────────┘
              │
              │  Close code 4001 (JWT rejected)
              ▼
        ForceReloginMsg → delete token → Login Screen
```

**Auto-login**: On startup, `App.Init()` checks `~/.config/chat/token`. If a valid JWT exists and `GET /api/auth/me` succeeds, the app jumps directly to the Chat screen — no login dialog.

**Ctrl+R toggle**: On the Login screen, `Ctrl+R` switches to Register. On Register, `Ctrl+R` switches back to Login.

## Component Hierarchy

### TextInput (`tui/components/textinput.go`)

A custom text input with horizontal scrolling and cursor management.

| Field | Type | Purpose |
|---|---|---|
| `value` | `[]rune` | Current text content |
| `placeholder` | `string` | Placeholder text when empty |
| `focused` | `bool` | Whether input accepts keystrokes |
| `cursor` | `int` | Cursor position in the rune slice |
| `width` | `int` | Visible width (defaults to 40) |
| `mask` | `rune` | If non-zero, masks characters (e.g. `*` for passwords) |
| `offset` | `int` | Horizontal scroll offset for long text |

**Key handling** (when focused):

| Key | Action |
|---|---|
| Printable characters | Insert at cursor position |
| Backspace | Delete character before cursor |
| Delete | Delete character after cursor |
| Left / Right | Move cursor |
| Home / Ctrl+A | Move cursor to start |
| End / Ctrl+E | Move cursor to end |
| Ctrl+U | Delete from start to cursor |
| Ctrl+K | Delete from cursor to end |

**View**: Renders a bordered box using `InputFocusedStyle` or `InputBlurredStyle`. The cursor character is rendered with reverse video. When masked, all characters display as the mask rune.

### MessageList (`tui/components/messagelist.go`)

A scrollable viewport that displays chat messages, join/leave events, and errors.

| Field | Type | Purpose |
|---|---|---|
| `messages` | `[]ChatMessage` | All received messages |
| `width`, `height` | `int` | Viewport dimensions |
| `scrollPos` | `int` | Index of first visible line |
| `lines` | `[]string` | Pre-rendered lines (rebuilt on each message) |
| `ownUserID` | `string` | Current user's ID for styling own messages |
| `autoScroll` | `bool` | Whether to follow new messages |

**Message rendering** by type:
- `chat_message` → `"15:04 username: content"` (own messages in cyan, others in indigo)
- `user_joined` → `"  --> username joined the chat"` (muted italic)
- `user_left` → `"  <-- username left the chat"` (muted italic)
- `error` → `"  [error] message"` (red)

**Scrolling**:
- Up/Down moves one line at a time
- PgUp/PgDown moves half a viewport
- Scrolling up disables auto-scroll; scrolling to the bottom re-enables it

**View**: Content is wrapped in a `PanelBorder` style (rounded border, muted color).

### StatusBar (`tui/components/statusbar.go`)

Bottom bar showing connection state, username, and help text.

| Field | Type | Purpose |
|---|---|---|
| `width` | `int` | Bar width |
| `status` | `ConnectionStatus` | `StatusOffline` / `StatusConnecting` / `StatusOnline` |
| `username` | `string` | Current user's display name |

**View**: Left side shows a colored dot (`●`) plus connection state label and username. Right side shows `"esc menu | pgup/pgdn scroll | ctrl+c quit"`. Background is gray (`#374151`).

Connection indicators:
- `StatusOnline` → green dot, "connected"
- `StatusConnecting` → yellow dot, "connecting"
- `StatusOffline` → red dot, "disconnected"

### Header (`tui/components/header.go`)

Top bar with the app title and the logged-in user.

| Field | Type | Purpose |
|---|---|---|
| `width` | `int` | Bar width |
| `title` | `string` | Always `"WebSocket Chat"` |
| `username` | `string` | Current user |

**View**: Purple background (`#7C3AED`), bold white title on the left, lighter username on the right.

### Dialog (`tui/components/dialog.go`)

A centered modal dialog with labeled input fields, error display, and hint text. Used by both the Login and Register screens.

| Field | Type | Purpose |
|---|---|---|
| `title` | `string` | Dialog heading |
| `fields` | `[]DialogField` | Each field has a `Label` and a `TextInput` |
| `focusIndex` | `int` | Which field has focus |
| `errMsg` | `string` | Error text displayed below fields |
| `hint` | `string` | Hint text at the bottom |
| `width`, `height` | `int` | For centering calculation |

**Key handling**:

| Key | Action |
|---|---|
| Tab | Focus next field |
| Shift+Tab | Focus previous field |
| Enter | Emit `DialogSubmitMsg` with all field values |
| Esc | Emit `DialogCancelMsg` |
| Other | Delegated to the focused `TextInput` |

**View**: Rendered inside a rounded purple border (`DialogStyle`), centered on screen using `lipgloss.Place()`. Title is bold purple, errors are red italic, hints are muted italic.

### Menu (`tui/components/menu.go`)

A centered overlay menu with keyboard-navigable items.

| Field | Type | Purpose |
|---|---|---|
| `title` | `string` | Menu heading |
| `items` | `[]string` | Selectable options |
| `cursor` | `int` | Currently highlighted item index |
| `width` | `int` | Menu box width (default 30) |

**Key handling:**

| Key | Action |
|---|---|
| Up / Down | Move cursor (wraps around) |
| Enter | Emit `MenuSelectMsg{Index, Label}` |
| Esc | Emit `MenuCloseMsg{}` |

**View:** Rendered inside a rounded purple border (reuses `DialogStyle`), centered using `lipgloss.Place()`. Selected item is highlighted in cyan bold with a `▸` prefix. Unselected items are white.

## Screen Details

### Login Screen (`tui/screens/login.go`)

A centered dialog with Username and Password fields.

**Model fields**:

| Field | Type | Purpose |
|---|---|---|
| `dialog` | `components.Dialog` | The login form (2 fields) |
| `serverURL` | `string` | Server URL for REST calls |
| `loading` | `bool` | Prevents double-submit |

**Configuration**: Password field (index 1) is masked with `*`. Hint text reads `"Ctrl+R: switch to Register"`.

**Messages produced**:
- `LoginSuccessMsg{Token, UserID, Username}` — on successful login; handled by `App` to save token and switch to chat
- `SwitchToRegisterMsg{}` — on `Ctrl+R`; handled by `App` to switch screens

**Async flow**: On `DialogSubmitMsg`, validates that both fields are non-empty, then fires `client.Login()` in a `tea.Cmd`. The result arrives as `loginResultMsg`. On error, the dialog shows the error message. On success, emits `LoginSuccessMsg`.

### Register Screen (`tui/screens/register.go`)

A centered dialog with Username, Password, and Confirm Password fields.

**Model fields**: Same structure as `LoginScreen` with `RegisterScreen` naming.

**Configuration**: Password (index 1) and Confirm Password (index 2) are masked with `*`. Hint text reads `"Ctrl+R: switch to Login"`.

**Validation**: Checks that all fields are filled and that passwords match before calling `client.Register()`.

**Messages produced**:
- `RegisterSuccessMsg{Token, UserID, Username}` — on successful registration
- `SwitchToLoginMsg{}` — on `Ctrl+R`

### Chat Screen (`tui/screens/chat.go`)

The main chat view, composed of four sub-components stacked vertically.

**Model fields**:

| Field | Type | Purpose |
|---|---|---|
| `header` | `components.Header` | Top bar |
| `messages` | `components.MessageList` | Scrollable message area |
| `input` | `components.TextInput` | Message input (always focused) |
| `statusBar` | `components.StatusBar` | Bottom bar |
| `wsClient` | `*client.WSClient` | WebSocket client for sending messages |

**Layout** (`SetSize`):
- Header: 1 row
- Status bar: 1 row
- Input (with border): 3 rows
- Messages: remaining height (`h - 1 - 1 - 3`)

**Key bindings**:

| Key | Action |
|---|---|
| Enter | Send message via WebSocket, clear input |
| Esc | Open menu (Resume / Logout) |
| Up / Down | Scroll messages one line |
| PgUp / PgDown | Scroll messages half page |
| Other keys | Delegated to text input |

**Incoming message handling**:
- `client.IncomingChatMsg` → adds message to `MessageList`
- `client.ConnectedMsg` → sets status to Online
- `client.DisconnectedMsg` → sets status to Offline
- `client.ConnectionErrorMsg` → sets status to Offline

**View**: Vertical stack of `header + messages + input + statusBar`, joined by newlines.

### Esc Menu

When the user presses `Esc` on the chat screen, a centered overlay menu appears with two options:

| Option | Action |
|---|---|
| **Resume** | Closes the menu, returns to the chat |
| **Logout** | Disconnects WebSocket, deletes saved token, returns to login screen |

**Menu navigation:**

| Key | Action |
|---|---|
| Up / Down | Move selection (wraps around) |
| Enter | Select the highlighted option |
| Esc | Close the menu (same as Resume) |

**Implementation:** The menu is a `Menu` component (`tui/components/menu.go`) rendered as a centered overlay using `lipgloss.Place()`. When `showMenu` is `true` on the `ChatScreen`, all key input is delegated to the menu, and the menu's `View()` replaces the normal chat view. Non-key messages (like incoming chat messages) continue to be processed while the menu is open.

**Message types:**
- `MenuSelectMsg{Index, Label}` — emitted on Enter, carries the selected item
- `MenuCloseMsg{}` — emitted when pressing Esc in the menu

## Theme System

Defined in `tui/styles/theme.go`. All styles are package-level `lipgloss.Style` variables that components reference directly.

### Color Palette

| Name | Hex | Usage |
|---|---|---|
| `Primary` | `#7C3AED` | Purple — headers, borders, dialog titles, input labels |
| `Secondary` | `#6366F1` | Indigo — other users' message names |
| `Accent` | `#06B6D4` | Cyan — own message names, focused input border |
| `Error` | `#EF4444` | Red — error messages, disconnected indicator |
| `Success` | `#22C55E` | Green — connected indicator |
| `Muted` | `#6B7280` | Gray — timestamps, hints, help text, blurred borders |
| `White` | `#F9FAFB` | Off-white — message content, usernames in status bar |
| `Dark` | `#1F2937` | Dark gray — (defined, available for use) |
| `DarkBg` | `#111827` | Very dark — (defined, available for use) |
| `InputBg` | `#1E293B` | Slate — (defined, available for use) |

### Style Groups

**Header styles**: `HeaderStyle` (purple background, white bold text, padded), `HeaderTitleStyle` (bold white), `HeaderInfoStyle` (light gray `#D1D5DB`).

**Message styles**: `OwnMessageStyle` (cyan bold), `OtherMessageStyle` (indigo bold), `SystemMessageStyle` (muted italic), `TimestampStyle` (muted), `MessageContentStyle` (white).

**StatusBar styles**: `StatusBarStyle` (gray `#374151` background), `StatusConnected` (green), `StatusConnecting` (yellow `#EAB308`), `StatusDisconnected` (red), `StatusHelpStyle` (muted).

**Input styles**: `InputStyle` (rounded border, primary), `InputFocusedStyle` (rounded border, cyan), `InputBlurredStyle` (rounded border, muted), `InputPlaceholderStyle` (muted), `InputLabelStyle` (primary bold).

**Dialog styles**: `DialogStyle` (rounded border, primary, padded, 50-wide), `DialogTitleStyle` (primary bold centered), `DialogErrorStyle` (red italic), `DialogHintStyle` (muted italic centered).

**Panel**: `PanelBorder` (rounded border, muted foreground).

## Client Layer

### REST Auth Client (`tui/client/auth.go`)

Provides three functions for server communication over HTTP:

**`Login(serverURL, username, password) → (*AuthResponse, error)`**
- POST to `/api/auth/login` with JSON `{"username", "password"}`
- Returns `AuthResponse{Token, User{ID, Username}}`

**`Register(serverURL, username, password) → (*AuthResponse, error)`**
- POST to `/api/auth/register` with JSON `{"username", "password"}`
- Accepts both 200 and 201 status codes as success

**`GetMe(serverURL, token) → (*AuthUser, error)`**
- GET to `/api/auth/me` with `Authorization: Bearer <token>` header
- Used for token validation on startup

**URL conversion**: The `httpURL()` helper converts the WebSocket server URL to HTTP:
- `ws://host:port` → `http://host:port`
- `wss://host:port` → `https://host:port`
- Trailing slashes are trimmed

### Token Storage (`tui/client/token.go`)

JWT tokens are persisted at `~/.config/chat/token` (follows XDG-ish convention under the user's home directory).

| Function | Description |
|---|---|
| `SaveToken(token)` | Creates `~/.config/chat/` directory (mode `0700`) and writes token (mode `0600`) |
| `LoadToken()` | Reads the stored token file |
| `DeleteToken()` | Removes the token file |

**Auto-login flow**: On startup, `App.Init()` calls `LoadToken()`. If a token exists, it validates with `GetMe()`. Valid → skip to chat. Invalid → show login screen.

**Force re-login**: When the WebSocket server rejects a JWT (close code 4001), `App` handles `ForceReloginMsg` by calling `DeleteToken()` and resetting to the Login screen.

### WebSocket Client (`tui/client/ws.go`)

Manages the WebSocket connection lifecycle in a background goroutine.

**`WSClient` struct**:

| Field | Type | Purpose |
|---|---|---|
| `serverURL` | `string` | Server URL |
| `token` | `string` | JWT for authentication |
| `conn` | `*websocket.Conn` | Active connection (nil when disconnected) |
| `program` | `*tea.Program` | Reference for `Send()` — injects messages into the Elm loop |
| `cancel` | `context.CancelFunc` | Cancels the read context |
| `mu` | `sync.Mutex` | Protects `conn` and `closed` |
| `closed` | `bool` | Shutdown flag |

**Connection URL**: Built as `{serverURL}/ws?token={jwt}`.

**Background goroutine**: `Connect()` launches `connectWithBackoff()` in a goroutine. This function:
1. Calls `dial()` to establish a WebSocket connection (10-second timeout)
2. On success, sends `ConnectedMsg` to the program and enters `readLoop()`
3. `readLoop()` continuously reads JSON frames and sends `IncomingChatMsg` to the program
4. On connection loss, sends `DisconnectedMsg` and retries

**Reconnection with exponential backoff**: After each failed attempt, the delay is `min(2^attempt seconds, 30 seconds)`. The attempt counter resets to 0 on a successful connection.

**Close code 4001**: If the server closes the WebSocket with status 4001, the client sends `ForceReloginMsg` to the program and stops reconnecting. This indicates the JWT is invalid/expired.

**Thread safety**: All access to `conn` and `closed` is guarded by `mu sync.Mutex`.

**Sending messages**: `Send(msg)` marshals a `protocol.ClientMessage` to JSON and writes it to the WebSocket with a 5-second timeout.

**Message types emitted** (via `tea.Program.Send()`):

| Message | When |
|---|---|
| `ConnectedMsg{}` | WebSocket connection established |
| `DisconnectedMsg{Err}` | Connection lost or dial failed |
| `IncomingChatMsg{Msg}` | Server message received |
| `ForceReloginMsg{}` | Server rejected JWT (close 4001) |

## App Root (`tui/app.go`)

The `App` struct is the root Bubble Tea model. It acts as a screen state machine and owns the WebSocket lifecycle.

### State Machine

The `currentScreen` field (`screen` type) selects which screen receives `Update` and `View` calls:

```
const (
    screenLogin    screen = iota   // 0
    screenRegister                 // 1
    screenChat                     // 2
)
```

### Startup (`Init`)

Returns a `tea.Cmd` that:
1. Loads the saved JWT from disk (`client.LoadToken()`)
2. If found, validates it via `client.GetMe()`
3. Returns `tokenCheckResult{valid: true/false, ...}`

### Message Routing

`App.Update()` handles global messages first, then delegates to the current screen:

| Message | Handler |
|---|---|
| `tea.KeyMsg "ctrl+c"` | Close WebSocket, quit |
| `tea.WindowSizeMsg` | Propagate to all screens |
| `tokenCheckResult` (valid) | Store token/user, connect WS, switch to chat |
| `LoginSuccessMsg` | Save token, store user, connect WS, switch to chat |
| `RegisterSuccessMsg` | Save token, store user, connect WS, switch to chat |
| `SwitchToRegisterMsg` | Set `currentScreen = screenRegister` |
| `SwitchToLoginMsg` | Set `currentScreen = screenLogin` |
| `ForceReloginMsg` | Delete token, reset to login screen |
| Everything else | Delegate to current screen's `Update()` |

### WebSocket Lifecycle

`connectAndSwitchToChat()`:
1. Closes any existing `wsClient`
2. Creates a new `WSClient` with the current token and `tea.Program` reference
3. Creates a new `ChatScreen` with the user info and WS client
4. Calls `wsClient.Connect()` to start the background read goroutine
5. Sets `currentScreen = screenChat`

### Program Reference

`cmd/tui/main.go` creates the `App`, then creates the `tea.Program`, then calls `app.SetProgram(p)`. This two-step initialization is needed because the `WSClient` needs the program reference to inject messages, but the program needs the `App` model at construction time.

## Building and Running

### Build

```bash
make build-tui        # Compiles to ./bin/tui
```

### Run

```bash
# Direct with make (defaults to ws://localhost:8080)
make run-tui

# Built binary with custom server
./bin/tui --server ws://localhost:8080

# Connect to Docker Compose setup (server behind nginx on port 5173)
./bin/tui --server ws://localhost:5173
```

### CLI Flags

| Flag | Default | Description |
|---|---|---|
| `--server` | `ws://localhost:8080` | WebSocket server URL |

The server URL is used both for WebSocket connections (`ws://` or `wss://`) and REST API calls (automatically converted to `http://` or `https://`).

## Key Bindings

### Login / Register Screens

| Key | Action |
|---|---|
| Tab | Next field |
| Shift+Tab | Previous field |
| Enter | Submit form |
| Ctrl+R | Toggle between Login and Register |
| Ctrl+C | Quit |
| Esc | Cancel dialog |

### Chat Screen

| Key | Action |
|---|---|
| Enter | Send message |
| Esc | Open menu (Resume / Logout) |
| Up | Scroll messages up one line |
| Down | Scroll messages down one line |
| PgUp | Scroll messages up half page |
| PgDown | Scroll messages down half page |
| Ctrl+C | Quit |

### Text Input (applies in all screens)

| Key | Action |
|---|---|
| Left / Right | Move cursor |
| Home / Ctrl+A | Jump to start |
| End / Ctrl+E | Jump to end |
| Backspace | Delete before cursor |
| Delete | Delete after cursor |
| Ctrl+U | Delete from start to cursor |
| Ctrl+K | Delete from cursor to end |
