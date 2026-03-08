# PicoClaw Controller Node Setup

This guide explains how to set up a PicoClaw controller node with team management capabilities.

## Architecture

The controller runs the **launcher** which provides:
- **Web UI** (port 18800): Configuration editor, team management, logs
- **Gateway control**: Start/stop button to launch the gateway subprocess
- **Team API** (port 18790, when gateway is running): Agent join, heartbeat, team management

```
┌─────────────────────────────────────────────────────────────┐
│                    Controller Node                           │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                  Launcher (Port 18800)                │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │  │
│  │  │   Web UI     │  │   Teams API  │  │  Process   │ │  │
│  │  │  (React)     │  │  (Internal)  │  │  Control   │ │  │
│  │  └──────────────┘  └──────────────┘  └────────────┘ │  │
│  └──────────────────────────────────────────────────────┘  │
│                         │                                    │
│                         │ Click "Start"                      │
│                         ▼                                    │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Gateway Subprocess (Port 18790)          │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌─────────┐ │  │
│  │  │ Channels │ │   Cron   │ │  Health  │ │  Teams  │ │  │
│  │  │ Telegram │ │  Service │ │  Server  │ │   API   │ │  │
│  │  │ Discord  │ │          │ │          │ │         │ │  │
│  │  │  etc.    │ │          │ │          │ │         │ │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └─────────┘ │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Quick Start

### 1. Configure Environment

Copy the example file and edit:

```bash
cp .env.example .env
nano .env  # Edit with your API keys
```

Required settings:
```env
PICOCLAW_PROVIDERS_KIMI_CODING_API_KEY=sk-your-api-key-here
```

### 2. Start the Controller

```bash
docker compose up -d
```

### 3. Access the Web UI

Open http://localhost:18800 in your browser.

### 4. Start the Gateway

Click the **"Start"** button in the Web UI to launch the gateway.

### 5. Create a Team

Navigate to **Teams** → **Create Team** in the Web UI.

Or via API:
```bash
curl -X POST http://localhost:18790/api/teams \
  -H "Content-Type: application/json" \
  -d '{"name":"my-team","description":"My development team"}'
```

### 6. Join Agents

**Discover teams:**
```bash
picoclaw team discover
```

**Join with team key:**
```bash
picoclaw team join my-team-abc123 \
  --key pk_team_xxxxx \
  --gateway http://controller-ip:18790 \
  --role backend
```

## API Endpoints

When the gateway is running, these endpoints are available:

### Team Management

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/teams` | GET | List all teams |
| `/api/teams` | POST | Create new team |
| `/api/teams/{id}` | GET | Get team details |
| `/api/teams/{id}` | PUT | Update team |
| `/api/teams/{id}` | DELETE | Delete team |
| `/api/teams/{id}/rotate-key` | POST | Rotate team key |
| `/api/teams/{id}/agents` | GET | List team agents |

### Agent Operations

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/teams/join` | POST | Agent join request |
| `/api/teams/heartbeat` | POST | Agent heartbeat |
| `/api/teams/{id}/agents/{agentId}/evict` | POST | Evict agent |
| `/api/teams/{id}/agents/{agentId}/wipe` | POST | Wipe agent |

### Health

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Gateway health check |
| `/ready` | GET | Gateway readiness |

## Docker Compose Configuration

```yaml
services:
  picoclaw:
    image: 'ghcr.io/brendantza/picoclaw:latest'
    command: ["picoclaw-launcher", "-public", "/home/picoclaw/.picoclaw/config.json"]
    ports:
      - '18800:18800'  # Launcher Web UI
      - '18790:18790'  # Gateway API
    environment:
      - PICOCLAW_PROVIDERS_KIMI_CODING_API_KEY=sk-...
    volumes:
      - 'picoclaw-data:/home/picoclaw/.picoclaw'
```

## Security Considerations

1. **Team Keys**: Never commit team keys to version control. Use environment variables.

2. **Network**: The controller exposes:
   - Port 18800 (launcher Web UI)
   - Port 18790 (gateway API)
   
   Ensure these are properly firewalled for production.

3. **HTTPS**: For production, put the controller behind a reverse proxy with HTTPS.

## Troubleshooting

### Gateway won't start
- Check logs in the Web UI (Logs tab)
- Verify API key is set correctly
- Check port 18790 is not in use: `lsof -i :18790`

### Agents can't connect
- Ensure gateway is running (check Web UI status)
- Verify team key is correct
- Check firewall rules between agent and controller

### Team UI not working
- Ensure gateway is running (team API is part of gateway)
- Check browser console for errors

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PICOCLAW_PROVIDERS_KIMI_CODING_API_KEY` | Kimi API key | (required) |
| `PICOCLAW_TEAM_ID` | Team to auto-join | (none) |
| `PICOCLAW_TEAM_KEY` | Team authentication | (none) |
| `PICOCLAW_AGENT_ROLE` | Role in team | `worker` |

## See Also

- [TEAM_AUTH.md](TEAM_AUTH.md) - Authentication protocol details
- [LAUNCHER_UI_EXTENSION.md](LAUNCHER_UI_EXTENSION.md) - UI documentation
