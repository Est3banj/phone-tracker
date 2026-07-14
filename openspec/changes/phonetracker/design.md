# Design: Phone Tracker — Personal Android Monitoring System

## Technical Approach

Three-component system in hexagonal architecture: Flutter Android app (foreground service + event collectors + WebSocket client) → Go backend (pub/sub hub + SQLite + command dispatcher) → Bubble Tea TUI (real-time dashboard). Communication via WSS JSON. Domain logic isolated from adapters so each component can evolve independently.

## Architecture Decisions

### Decision: WebSocket pub/sub hub with per-device channels

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Single global broadcast | Simple but leaks data across devices | ❌ |
| **Per-device channel with Hub** | Isolated, clean disconnect, device-level routing | ✅ |
| MQTT broker | Overkill for 1 device, infra burden | ❌ |

### Decision: SQLite with WAL mode

| Option | Tradeoff | Decision |
|--------|----------|----------|
| **SQLite with WAL** | Pure Go (modernc.org/sqlite) or CGO (mattn/go-sqlite3) | ✅ |
| PostgreSQL | Overkill for MVP, adds server dependency | ❌ |

### Decision: Riverpod for Flutter state management

| Option | Tradeoff | Decision |
|--------|----------|----------|
| **Riverpod** | Compile-safe, no BuildContext, testable providers | ✅ |
| BLoC | Heavy boilerplate | ❌ |
| Provider | Deprecated pattern | ❌ |

### Decision: ASCII bounding-box map (no tiles)

| Option | Tradeoff | Decision |
|--------|----------|----------|
| **ASCII grid with `@` marker** | Simple, no libs, terminal-safe | ✅ |
| Sixel/Kitty graphics | Terminal-dependent, complex | ❌ |

## Project Structure

```
phone-tracker/
├── cmd/
│   ├── server/             # Go backend binary
│   └── dashboard/          # Bubble Tea TUI binary
├── internal/
│   ├── domain/             # Core types
│   ├── ports/              # Interfaces
│   ├── service/            # Business logic
│   ├── adapters/
│   │   ├── handler/        # HTTP + WebSocket
│   │   └── repository/     # SQLite implementations
│   └── config/
├── app/                    # Flutter app
│   ├── lib/
│   │   ├── models/
│   │   ├── services/
│   │   ├── providers/
│   │   ├── screens/
│   │   └── widgets/
│   └── android/
└── openspec/               # SDD artifacts
```

## WebSocket Protocol

**Envelope**: `{ "type": "<type>", "ts": "<ISO8601>" }`

**Types**: `location`, `event`, `ping` (30s), `pong`, `command`, `ack`, `result`, `cmd_status`

**Auth**: Token as WS query param. Validated against devices.token_hash.

## Data Flow

### Location Report

```
Phone GPS → Flutter LocationService (60s tick)
              ├── Online? → WebSocket → Go Hub → Repository → SQLite
              │                                        └── Broadcast → Dashboard WS
              └── Offline? → Local SQLite queue → Flush on reconnect (FIFO)
```

### Remote Command

```
TUI CommandPanel → WebSocket → Go Hub → DeviceConn.Write()
                  │                          └── Phone receives → executes → ack
                  └── Offline? → Store pending → Send on device reconnect
```

## Security

| Layer | Mechanism |
|-------|-----------|
| Transport | WSS with TLS |
| Device auth | Token as WS query param |
| Token rotation | HTTP POST `/api/rotate-token` |
| App access | PIN/biometric via local_auth |
| Storage | flutter_secure_storage |
