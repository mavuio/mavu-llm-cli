# Run `llm os --serve` as a macOS service

This guide shows how to keep the OpenCode sessions API running permanently on a Mac using `launchd`.

## 1) Build the binary

From the repo root:

```bash
go build -o llm
```

## 2) Create a LaunchAgent

Create `~/Library/LaunchAgents/com.mavu.opencode-sessions.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key>
    <string>com.mavu.opencode-sessions</string>

    <key>ProgramArguments</key>
    <array>
      <string>/Users/manfred/Documents/www/mavu-llm-cli/llm</string>
      <string>os</string>
      <string>--serve</string>
      <string>--listen</string>
      <string>192.168.102.10:8787</string>
      <string>--storage-path</string>
      <string>/Users/manfred/.local/share/opencode/opencode.db</string>
    </array>

    <key>EnvironmentVariables</key>
    <dict>
      <key>MAVU_SESSIONS_API_TOKEN</key>
      <string>awiyjYa9JDC3xzhYvTdd</string>
    </dict>

    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>

    <key>WorkingDirectory</key>
    <string>/Users/manfred/Documents/www/mavu-llm-cli</string>

    <key>StandardOutPath</key>
    <string>/tmp/mavu-opencode-sessions.out.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/mavu-opencode-sessions.err.log</string>
  </dict>
</plist>
```

## 3) Load and start the service

```bash
launchctl bootout gui/$(id -u) ~/Library/LaunchAgents/com.mavu.opencode-sessions.plist 2>/dev/null; launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.mavu.opencode-sessions.plist; launchctl kickstart -k gui/$(id -u)/com.mavu.opencode-sessions
```

## 4) Check status

```bash
launchctl print gui/$(id -u)/com.mavu.opencode-sessions
```

You should see `state = running`.

## 5) Verify API

```bash
curl -H "Authorization: Bearer awiyjYa9JDC3xzhYvTdd" http://192.168.102.10:8787/health
```

Expected response:

```json
{"ok":true}
```

## Common operations

Restart after rebuilding `llm`:

```bash
launchctl kickstart -k gui/$(id -u)/com.mavu.opencode-sessions
```

Stop/unload:

```bash
launchctl bootout gui/$(id -u) ~/Library/LaunchAgents/com.mavu.opencode-sessions.plist
```

View logs:

```bash
tail -f /tmp/mavu-opencode-sessions.out.log /tmp/mavu-opencode-sessions.err.log
```

## Security notes

- Use a strong token and rotate it if shared.
- Prefer `127.0.0.1` for local-only access.
- If binding to a LAN IP (like `192.168.102.10`), restrict network access with firewall/router rules.
