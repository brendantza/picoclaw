# Dev Agent Setup

This guide explains how to run a local PicoClaw agent that joins your team.

## Quick Start

### 1. Using Docker Compose (Recommended)

```bash
# Set your API key (if agent needs direct model access)
export PICOCLAW_PROVIDERS_KIMI_CODING_API_KEY=sk-...

# Start the agent
docker compose -f docker-compose.dev-agent.yml up -d

# View logs
docker compose -f docker-compose.dev-agent.yml logs -f
```

### 2. Configuration

Edit `docker-compose.dev-agent.yml` to set your team details:

```yaml
environment:
  - TEAM_ID=your-team-id           # Get this from controller UI
  - TEAM_KEY=pk_team_xxx           # Team key from controller
  - GATEWAY_ADDRESS=http://192.168.6.122:18790  # Controller address
  - AGENT_ROLE=worker              # Role: worker, frontend, backend, etc.
  - AGENT_ID=my-dev-agent          # Unique agent ID
```

### 3. Verify Connection

Check if agent joined the team:

```bash
# From the agent container
docker exec picoclaw-dev-agent picoclaw team list

# Or check team status
docker exec picoclaw-dev-agent picoclaw team status dev-team
```

## Manual Join (if auto-join fails)

```bash
# Enter container
docker exec -it picoclaw-dev-agent sh

# Join manually
picoclaw team join dev-team \
  --key pk_team_MK2-SdlgjeX0Mwt0BrkHKPY1DPs2ZJ6tj9xfkMs5NUM= \
  --gateway http://192.168.6.122:18790 \
  --role worker

# Check status
picoclaw team status
```

## Troubleshooting

### Agent can't connect to gateway
- Verify gateway is running: `curl http://192.168.6.122:18790/health`
- Check firewall between agent and gateway
- Ensure correct `GATEWAY_ADDRESS` (use IP, not hostname)

### Team join fails
- Verify team key is correct
- Check if agent already joined: `picoclaw team list`
- Try leaving first: `picoclaw team leave dev-team`

### Permission errors
The container runs as `picoclaw` user (UID 1000). Data is stored in:
- `/home/picoclaw/.picoclaw/teams/` - Team memberships

## Network Options

### Option 1: Bridge Network (Default)
Works when agent and gateway are on same Docker network or accessible via IP.

### Option 2: Host Network (for mDNS discovery)
```yaml
services:
  picoclaw-agent:
    network_mode: host
```

### Option 3: External Network (if gateway is in another compose)
```yaml
networks:
  picoclaw-agent-network:
    external: true
    name: picoclaw_picoclaw-network  # From controller compose
```

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `TEAM_ID` | Team ID to join | Yes |
| `TEAM_KEY` | Team authentication key | Yes |
| `GATEWAY_ADDRESS` | Controller gateway URL | Yes |
| `AGENT_ROLE` | Role in team (worker, frontend, etc.) | Yes |
| `AGENT_ID` | Unique agent identifier | Yes |
| `PICOCLAW_PROVIDERS_KIMI_CODING_API_KEY` | API key for agent's own AI | Optional |

## Stop Agent

```bash
docker compose -f docker-compose.dev-agent.yml down

# Remove data
docker compose -f docker-compose.dev-agent.yml down -v
```
