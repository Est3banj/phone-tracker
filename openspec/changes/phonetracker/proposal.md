# Proposal: Phone Tracker — Personal Android Monitoring System

## Intent

Build a personal phone tracking system that continues reporting location after the device powers off. Google Find My Device stops at power-off; this solves that gap by running as a persistent foreground service reporting via WebSocket to a Go backend with a TUI dashboard.

## Scope

### In Scope
- Flutter Android app: foreground service, GPS reporting, battery status, event alerts, remote command execution, app PIN lock
- Go backend: WebSocket hub, SQLite persistence, command dispatch, health endpoint
- Bubble Tea TUI: real-time map, location history table, alerts feed, command panel
- Local network MVP with TLS (self-signed); ready for VPS deployment

### Out of Scope
- Root-only features (hide icon, survive force-stop, system app registration)
- Peer-to-peer / mesh offline network
- iOS support
- Multi-user auth or team features
- Push notifications (FCM)

## Capabilities

### New Capabilities
- `phone-reporting`: periodic GPS, battery, and event data ingestion from Android device
- `phone-alerts`: SIM change, battery low, WiFi disconnect, power-on detection and storage
- `remote-commands`: lock, wipe, capture photo, trigger alarm — dispatched via WebSocket
- `tui-dashboard`: real-time Bubble Tea terminal UI with map, history table, alerts feed, command panel
- `app-access-control`: PIN/biometric lock for the Flutter app UI

### Modified Capabilities
None — greenfield project.

## Approach

Three-component system: Flutter app (foreground service + location/event collectors + WebSocket client) → Go backend (WebSocket hub + SQLite + command dispatcher) → Bubble Tea TUI (real-time dashboard). Communication via WSS JSON. Go service layer decouples domain logic from adapters (hexagonal architecture).

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Force-stop kills foreground service | High | WorkManager heartbeat fallback; user education; battery opt-out request |
| OEM battery killing (Xiaomi, Samsung, Huawei) | High | Guide user to whitelist app per manufacturer |
| Android 14+ foreground service restrictions | Medium | Declare foregroundServiceType="location"; handle FGS exceptions |
| WebSocket disconnection | Medium | Exponential backoff reconnect; offline SQLite queue |
| Device Admin not enabled (lock/wipe) | Medium | Guided setup at install time; graceful fallback |

## Success Criteria

- Phone sends GPS location every configurable interval; server stores and dashboard displays it
- Battery, SIM change, WiFi disconnect, and power-on events appear in alerts feed <10s
- Remote commands reach phone and execute within 15s
- App survives screen-off, app switch, and Doze mode (not force-stop)
- App unlocks via PIN/biometric on launch; foreground service continues after UI close
- Dashboard shows real-time updates, location history, and alert log
