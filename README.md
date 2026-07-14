# Phone Tracker

Personal Android phone tracking system. Track your own device with GPS, battery, events, and remote commands via a WebSocket-connected Go backend with TUI + Web dashboards.

> **⚠️ Legal**: This tool is designed for tracking YOUR OWN devices only. Unauthorized tracking of others is illegal.

## Architecture

```
┌─────────────────────┐     WSS JSON      ┌───────────────────────┐
│  Flutter App         │◄─────────────────►│  Go Server             │
│  (Android)           │                   │                        │
│                      │                   │ ┌─ WebSocket Hub      │
│  ┌─────────────────┐ │  GPS + battery    │ ├─ SQLite (users,     │
│  │ Memory Match    │ │  + alerts + cmds  │ │  locations, alerts, │
│  │ (camouflage)    │ │                   │ │  commands, tokens)  │
│  └─────────────────┘ │                   │ ├─ JWT auth          │
│  ┌─────────────────┐ │                   │ └─ License mgmt      │
│  │ Hidden panel    │ │                   │                        │
│  │ (PIN-protected) │ │                   ├─ Bubble Tea TUI       │
│  └─────────────────┘ │                   └─ Web dashboard        │
└─────────────────────┘                     (HTML/JS)              │
                                           └───────────────────────┘
```

## Repository Structure

```
phone-tracker/
├── app/                          # Flutter Android app
│   ├── lib/
│   │   ├── models/               # Data classes
│   │   ├── services/             # WebSocket, GPS, battery, alerts, buffer
│   │   ├── providers/            # Riverpod state management
│   │   ├── screens/              # Lock screen, dashboard
│   │   └── widgets/              # Map, history, alerts, commands
│   └── android/                  # Native Android (Kotlin)
├── cmd/
│   ├── server/                   # Go server entry point
│   └── dashboard/                # TUI dashboard entry point
├── internal/
│   ├── domain/                   # Core entities
│   ├── ports/                    # Interfaces
│   ├── adapters/                 # SQLite, WebSocket, HTTP
│   └── service/                  # Business logic
├── openspec/
│   └── changes/phonetracker/     # SDD documentation
└── web/
    └── static/                   # Web dashboard (future)
```

## Quick Start

### Prerequisites

- Go 1.23+
- Flutter SDK 3.x (for building the Android app)
- Android device with USB debugging (for deployment)

### 1. Clone & Build Server

```bash
git clone https://github.com/Est3banj/phone-tracker.git
cd phone-tracker

# Build the server
go build ./cmd/server/

# Run (generate a JWT_SECRET first)
export JWT_SECRET=$(openssl rand -hex 32)
./server
```

Server starts on `http://0.0.0.0:8080` by default. Config via environment variables:

| Variable | Default | Description |
|---|---|---|
| `PORT` | 8080 | Server port |
| `HOST` | 0.0.0.0 | Bind address |
| `JWT_SECRET` | *required* | Secret key for JWT tokens |
| `DATABASE_PATH` | phonetracker.db | SQLite database path |
| `ACCESS_TOKEN_TTL` | 15m | JWT access token lifetime |
| `REFRESH_TOKEN_TTL` | 168h | Refresh token lifetime |

### 2. Create Super Admin

```bash
# The server exposes an API to register users
# Example (replace URL/credentials):
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"supersecret","role":"super_admin"}'
```

### 3. Build & Install Android App

```bash
cd app
flutter build apk
# The APK will be at: app/build/app/outputs/flutter-apk/app-release.apk
```

Install the APK on your Android device. On first launch:
1. Grant all requested permissions (GPS, camera, phone state, notifications)
2. Tap the logo 5 times to access the hidden settings panel
3. Configure the server URL and device token
4. Set a PIN to protect the settings panel

### 4. Connect & Monitor

Once the app is running and connected, launch the TUI dashboard:

```bash
go build ./cmd/dashboard/
./dashboard --server http://localhost:8080
```

Or access the web dashboard at `http://localhost:8080`.

## SDD Status

This project is built using Spec-Driven Development. All artifacts are in `openspec/changes/phonetracker/`:

| Phase | Status | Description |
|---|---|---|
| Proposal | ✅ | Product scope and approach |
| Specs | ✅ | 5 domain specs with requirements & scenarios |
| Design | ✅ | Full architecture design |
| Tasks | ✅ | 22 tasks across 6 phases |
| Foundation | ✅ | Go backend: SQLite, auth, license, WS hub |
| Communication | ✅ | Flutter: WS client, GPS, battery, alerts |
| Remote Commands | ⬜ | Lock, wipe, capture photo, alarm |
| Dashboards | ⬜ | Bubble Tea TUI + Web |
| App Facade | ⬜ | Memory Match game + hidden panel |
| Deploy | ⬜ | TLS, systemd, guides |

## License

Private — for personal use only.
