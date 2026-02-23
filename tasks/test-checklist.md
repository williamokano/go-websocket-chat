# Manual Test Checklist

## Web Client
1. `docker compose up --build` — all services start healthy
2. Open http://localhost:5173 — register "alice" → redirected to chat
3. New tab — register "bob" → redirected to chat
4. Alice sends "hello" → appears in both tabs
5. Third tab — login as "alice" → all three tabs show messages
6. Close Bob's tab → "user left" notification in Alice's tabs
7. `docker compose stop backend-1` → Alice reconnects automatically via nginx to backend-2
8. `docker compose start backend-1` → send from Alice → Bob receives it

## TUI Client
9. `make build-tui && ./bin/tui --server ws://localhost:8080` → TUI launches, shows login
10. Login as "alice" → chat screen appears, sees messages from web
11. Send from TUI → appears in browser tabs
12. Send from browser → appears in TUI
13. Kill server → TUI shows "disconnected", auto-reconnects when server returns
14. Open second terminal as "alice" → both receive messages
15. ctrl+c → TUI exits cleanly
16. Restart TUI → auto-login with stored token

## Cross-Instance
17. Connect client A to backend-1 (port 8080)
18. Connect client B to backend-2 (port 8081)
19. Send from A → B receives via Redis pub/sub
20. Send from B → A receives via Redis pub/sub

## Edge Cases
21. Send empty message → error displayed
22. Very long message (4000+ chars) → handled gracefully
23. Rapid-fire messages → all delivered in order
24. Multiple tabs same user → all tabs update
