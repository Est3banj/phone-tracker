# Phone Tracker — Specification

## Domain: phone-reporting

### Requirements
- REQ-REPORT-001: Phone MUST send LocationReport at configurable interval (default 60s, range 15-300s)
- REQ-REPORT-002: Every report MUST include battery level (0-100) and charging state
- REQ-REPORT-003: Offline buffer in local SQLite, flush FIFO on reconnect, drop >7d
- REQ-REPORT-004: Exponential backoff 1s→2s→4s→...→60s max, resets on connect
- REQ-REPORT-005: State machine: DISCONNECTED→CONNECTING→CONNECTED|BACKOFF_WAIT→SHUTDOWN
- REQ-REPORT-006: Server ping every 30s, 90s timeout = disconnect
- REQ-REPORT-007: Discard GPS with accuracy >100m

### Wire Protocol
```
location:  { "type": "location", "lat": f, "lng": f, "alt": f?, "accuracy": f?,
              "speed": f?, "battery": int, "charging": bool }
ping:      { "type": "ping" }
pong:      { "type": "pong" }
```

---

## Domain: phone-alerts

### Requirements
- REQ-ALERT-001: Detect SIM change via TelephonyManager, send within 10s
- REQ-ALERT-002: Battery low at <15%, no re-trigger until >20% then <15%
- REQ-ALERT-003: WiFi disconnect while cellular unavailable, send within 30s
- REQ-ALERT-004: BOOT_COMPLETED → start service + send power_on
- REQ-ALERT-005: Server dedup same-type alerts within 5min window
- REQ-ALERT-006: Events retained 90 days

### Wire Protocol
```
event: { "type": "event", "event_type": "sim_change"|"battery_low"|"wifi_disconnected"|"power_on",
          "payload": { ... } }
```

---

## Domain: remote-commands

### Requirements
- REQ-CMD-001: Forward command to device WS within 1s if connected, store pending otherwise
- REQ-CMD-002: Phone must ack within 15s, send result on completion
- REQ-CMD-003: Actions: lock_device, wipe_device, capture_photo, trigger_alarm
- REQ-CMD-004: No ack within 60s → timed_out
- REQ-CMD-005: Commands retained 90 days

### Wire Protocol
```
command: { "type": "command", "cmd_id": "uuid", "action": "...", "params": {}, "ts": "..." }
ack:     { "type": "ack", "cmd_id": "uuid", "status": "received" }
result:  { "type": "result", "cmd_id": "uuid", "status": "executed"|"failed", "error"?: "..." }
```

Status flow: `pending → sent → received → executed | failed | timed_out`

---

## Domain: tui-dashboard

### Requirements
- REQ-TUI-001: Display new data within 2s of server receipt
- REQ-TUI-002: 4 tabs: Map, History (50/page), Alerts (color-coded), Commands
- REQ-TUI-003: ASCII map with `@` marker, "Waiting for data..." empty state
- REQ-TUI-004: Command dispatch with status tracking
- REQ-TUI-005: Alert colors: sim_change=Red, battery_low=Yellow, wifi=Yellow, power_on=Green
- REQ-TUI-006: History paginated 50 rows, optional date search

---

## Domain: app-access-control

### Requirements
- REQ-ACCESS-001: Lock screen on app foreground with PIN/biometric
- REQ-ACCESS-002: Foreground service continues when UI is locked/closed
- REQ-ACCESS-003: Auto-lock after 5min background inactivity
- REQ-ACCESS-004: Token in flutter_secure_storage, passed as WS query param
- REQ-ACCESS-005: Server validates token on WS upgrade, HTTP 401 if invalid
- REQ-ACCESS-006: Token rotation via POST /api/rotate-token, 5min grace period
- REQ-ACCESS-007: WSS required, plain WS rejected

### Security Model
| Layer | Mechanism |
|-------|-----------|
| Transport | WSS + cert pinning |
| Device auth | Token pt_v1_<32-byte-hex> |
| Rotation | POST /api/rotate-token |
| App access | PIN (6+ digits) or biometric |
| Rate limit | 5 failures → 60s lockout |
| Storage | flutter_secure_storage |
