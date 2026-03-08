# Multi-Agent Quick Reference

## Centralized Gateway Commands

### Gateway (Controller) Commands

```bash
# Start gateway with team management enabled
picoclaw gateway --teams-enabled

# Create a team
picoclaw team create --name "Dev Team Alpha" --max-agents 10

# List teams
picoclaw team list

# Show team details
picoclaw team show dev-team-alpha

# Rotate team key
picoclaw team rotate-key dev-team-alpha

# Evict agent from team
picoclaw team evict dev-team-alpha --agent worker-01

# Wipe agent (remove all team keys)
picoclaw team wipe-agent worker-01

# Add role to team
picoclaw team role add dev-team-alpha \
  --id "frontend" \
  --capabilities "react,vue,css"
```

### Worker Agent Commands

```bash
# Start worker and join team
picoclaw agent \
  --join dev-team-alpha \
  --role frontend \
  --gateway localhost:18791

# Or via environment
export PICOCLAW_TEAM_ID=dev-team-alpha
export PICOCLAW_AGENT_ROLE=frontend
export PICOCLAW_GATEWAY_ADDRESS=localhost:18791
export PICOCLAW_TEAM_KEY=pk_team_...
picoclaw agent --auto-connect
```

### Interactive Commands (within agent)

```
/team list              # List available teams
/team status            # Show current team status
/team activate <name>   # Switch to team
/team leave             # Leave current team
/team evict <agent>     # Evict agent (controller only)
/team wipe <agent>      # Wipe agent (controller only)
```

### Docker Quick Start

```bash
# Generate team key
export TEAM_KEY=$(go run scripts/gen-team-key.go)

# Start gateway
docker run -d \
  -p 18790:18790 \
  -p 18791:18791 \
  -v gateway-data:/data \
  picoclaw gateway

# Create team via API
curl -X POST http://localhost:18790/api/v1/teams \
  -H "Content-Type: application/json" \
  -d '{"name":"Dev Team","max_agents":5}'

# Start worker
docker run -d \
  -e PICOCLAW_TEAM_ID=dev-team \
  -e PICOCLAW_AGENT_ROLE=backend \
  -e PICOCLAW_TEAM_KEY=$TEAM_KEY \
  -e PICOCLAW_GATEWAY_ADDRESS=host.docker.internal:18791 \
  picoclaw agent --mode=worker
```

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `PICOCLAW_TEAM_ID` | Team to join | `dev-team-alpha` |
| `PICOCLAW_AGENT_ROLE` | Agent role in team | `frontend`, `backend` |
| `PICOCLAW_TEAM_KEY` | Team authentication key | `pk_team_...` |
| `PICOCLAW_TEAM_KEY_FILE` | Path to key file | `/run/secrets/team_key` |
| `PICOCLAW_GATEWAY_ADDRESS` | Gateway connection address | `gateway:18791` |
| `PICOCLAW_AUTO_CONNECT` | Auto-connect to gateway | `true` |

## Configuration Snippets

### Minimal Worker Config

```json
{
  "agent_network": {
    "mode": "gateway_client",
    "gateway_address": "gateway:18791",
    "team_id": "dev-team-alpha",
    "role": "frontend",
    "auto_connect": true
  }
}
```

### Gateway with Teams

```json
{
  "gateway": {
    "port": 18790,
    "agent_port": 18791,
    "multi_agent": {
      "enabled": true,
      "web_ui": { "enabled": true, "path": "/admin" }
    }
  },
  "teams": { "enabled": true }
}
```

## Web UI URLs

| URL | Description |
|-----|-------------|
| `http://gateway:18790/` | Gateway status |
| `http://gateway:18790/admin` | Team management UI |
| `http://gateway:18790/admin/teams` | Teams list |
| `http://gateway:18790/admin/agents` | Connected agents |

## Troubleshooting

```bash
# Check agent connection
picoclaw agent --status

# View team membership
picoclaw team status

# Test gateway connectivity
curl http://gateway:18790/health

# View logs
docker logs picoclaw-gateway
picoclaw agent --debug

# Regenerate team key
picoclaw team rotate-key <team-id> --force
```

