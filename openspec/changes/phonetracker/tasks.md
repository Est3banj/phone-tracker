# Tasks: Phone Tracker

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High
Est. changed lines: 6000–7500

## Phase 1: Foundation — Backend Core ✅

- [x] 1.1 Go skeleton: go.mod, cmd/server/main.go, cmd/dashboard/main.go, config
- [x] 1.2 SQLite schema + repos (users, devices, locations, events, commands, tokens)
- [x] 1.3 Domain types + ports: location, event, command, user, device, token
- [x] 1.4 Auth: register, login, JWT/token rotate, logout, device token gen
- [x] 1.5 License + role middleware; WebSocket Hub with auth/license intercept

## Phase 2: Phone Communication ✅

- [x] 2.1 Flutter skeleton: pubspec.yaml, main.dart, app.dart, Riverpod providers
- [x] 2.2 WS client: state machine, backoff, offline SQLite buffer, FIFO flush
- [x] 2.3 Location service: configurable GPS, accuracy filter, battery in reports
- [x] 2.4 Events: SIM change, battery low (<15%), WiFi disconnect, BOOT_COMPLETED

## Phase 3: Remote Commands ⬜

- [ ] 3.1 Server dispatch: pending queue, 60s timeout, status transitions
- [ ] 3.2 Flutter exec: lock_device, wipe_device, capture_photo, trigger_alarm
- [ ] 3.3 Full flow: ack → execute → result → repo write + broadcast

## Phase 4: Dashboards ⬜

- [ ] 4.1 TUI: login model, 4-tab (Map/History/Alerts/Commands), WS connect
- [ ] 4.2 TUI map (ASCII grid @ marker) + history table paginated 50/page
- [ ] 4.3 TUI alerts (color-coded) + command dispatch with status tracking
- [ ] 4.4 Web dashboard: index/dashboard/admin.html, CSS, JS via embed.FS

## Phase 5: App Facade ⬜

- [ ] 5.1 Memory Match: 4×4 grid, flip, match, timer, score (Riverpod)
- [ ] 5.2 Secret gate: 5× tap → hidden settings (server URL, token, PIN)
- [ ] 5.3 Permission justifier: map game features → Android permissions

## Phase 6: Polish & Deploy ⬜

- [ ] 6.1 TLS: self-signed cert, WSS enforce, cert pinning on Flutter
- [ ] 6.2 Deployment guide + systemd service for Go binary
- [ ] 6.3 User guide: APK install, battery whitelist per OEM, add device token
