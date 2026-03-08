# Centralized Gateway Architecture for Multi-Agent Teams

## Executive Summary

This document describes a centralized architecture where the **Gateway serves as the default Controller** for multi-agent teams. All agents communicate through the Gateway, which provides centralized team management, authentication, and agent lifecycle control via both CLI and Web UI.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                          GATEWAY SERVER                              │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                    GATEWAY CONTROLLER                          │  │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────────────┐  │  │
│  │  │ Team Manager │ │ Auth Service │ │ Agent Registry       │  │  │
│  │  │ - Create     │ │ - Team Keys  │ │ - Connected Agents   │  │  │
│  │  │ - List       │ │ - Challenge  │ │ - Capabilities       │  │  │
│  │  │ - Revoke     │ │ - Sessions   │ │ - Health Status      │  │  │
│  │  └──────────────┘ └──────────────┘ └──────────────────────┘  │  │
│  │                                                                │  │
│  │  ┌─────────────────────────────────────────────────────────┐   │  │
│  │  │             WEB UI (Team Management)                     │   │  │
│  │  │  Teams | Agents | Keys | Logs | Tasks | Settings        │   │  │
│  │  └─────────────────────────────────────────────────────────┘   │  │
│  └───────────────────────────────────────────────────────────────┘  │
└───────────────────────────────┬─────────────────────────────────────┘
                                │ Agent Protocol (TCP/WebSocket)
            ┌───────────────────┼───────────────────┐
            ▼                   ▼                   ▼
┌───────────────┐      ┌───────────────┐   ┌───────────────┐
│ Worker Agent  │      │ Worker Agent  │   │ Worker Agent  │
│ (Frontend)    │      │ (Backend)     │   │ (DevOps)      │
└───────────────┘      └───────────────┘   └───────────────┘
```

## Key Design Principles

1. **Gateway as Controller**: The gateway is the default controller agent
2. **Centralized Team Management**: Teams, keys, and agents managed via Gateway
3. **All Agents Connect to Gateway**: No direct agent-to-agent communication
4. **Web UI for Administration**: Visual team and agent management
5. **CLI Integration**: `/team` commands and skills for team operations
6. **Docker-Native**: First-class Docker support with role/key configuration

## Configuration Schema

### Extended Agent Config with Teams

```json
{
  "_comment": "PicoClaw Multi-Agent Configuration with Centralized Gateway",
  
  "agents": {
    "defaults": {
      "workspace": "~/.picoclaw/workspace",
      "model_name": "kimi-coding"
    },
    "list": [
      {
        "id": "gateway-controller",
        "name": "Gateway Controller",
        "role": "controller",
        "default": true,
        "multi_agent": {
          "enabled": true,
          "mode": "gateway",
          "max_teams": 10,
          "max_agents_per_team": 20
        }
      }
    ]
  },
  
  "teams": {
    "enabled": true,
    "management": {
      "allow_cli_creation": true,
      "allow_web_creation": true,
      "default_team": "default",
      "key_rotation_days": 90
    },
    "list": [
      {
        "id": "dev-team-alpha",
        "name": "Development Team Alpha",
        "description": "Primary development team for Project X",
        "team_key": "${TEAM_KEY_DEV_ALPHA}",
        "created_at": "2026-03-08T10:00:00Z",
        "created_by": "admin",
        "status": "active",
        "settings": {
          "auto_accept_agents": false,
          "require_approval": true,
          "max_agents": 10,
          "agent_timeout": 300
        },
        "roles": {
          "frontend": {
            "description": "Frontend developers",
            "capabilities": ["react", "vue", "css"],
            "system_prompt_addendum": "You are a frontend specialist..."
          },
          "backend": {
            "description": "Backend developers", 
            "capabilities": ["go", "python", "sql"],
            "system_prompt_addendum": "You are a backend specialist..."
          },
          "qa": {
            "description": "QA engineers",
            "capabilities": ["testing", "cypress"],
            "system_prompt_addendum": "You are a QA specialist..."
          }
        }
      },
      {
        "id": "ops-team",
        "name": "Operations Team",
        "description": "Infrastructure and DevOps team",
        "team_key": "${TEAM_KEY_OPS}",
        "status": "active",
        "settings": {
          "auto_accept_agents": true,
          "max_agents": 5
        },
        "roles": {
          "kubernetes": {
            "capabilities": ["k8s", "helm", "docker"]
          },
          "monitoring": {
            "capabilities": ["prometheus", "grafana"]
          }
        }
      }
    ]
  },
  
  "gateway": {
    "host": "0.0.0.0",
    "port": 18790,
    "agent_port": 18791,
    "multi_agent": {
      "enabled": true,
      "default_controller": true,
      "discovery": {
        "enabled": true,
        "methods": ["mdns", "static"]
      },
      "web_ui": {
        "enabled": true,
        "path": "/admin",
        "require_auth": true,
        "features": {
          "team_management": true,
          "agent_monitoring": true,
          "task_orchestration": true,
          "logs_viewer": true
        }
      }
    }
  },
  
  "agent_network": {
    "mode": "gateway_client",
    "gateway_address": "localhost:18791",
    "auto_connect": true,
    "reconnect_interval": 10,
    "heartbeat_interval": 30
  }
}
```

## Gateway Web UI Design

### Dashboard Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  🦞 PicoClaw Gateway                                    [User] [⚙️] │
├──────────┬──────────────────────────────────────────────────────────┤
│          │                                                          │
│ Teams    │  ┌────────────────────────────────────────────────────┐  │
│ ├── Alpha│  │ Active Teams (3)                    [+ Create Team] │  │
│ ├── Ops  │  ├────────────────────────────────────────────────────┤  │
│ └── Beta │  │                                                    │  │
│          │  │ 🟢 Dev Team Alpha           5 agents   [Manage]   │  │
│ Agents   │  │    Key: pk_team_a1b2...***   2 online              │  │
│ ├── 12   │  │                                                    │  │
│ └── 3 🔴 │  │ 🟢 Operations Team          3 agents   [Manage]   │  │
│          │  │    Key: pk_team_c3d4...***   3 online              │  │
│ Tasks    │  │                                                    │  │
│ └── 5 ⚡ │  │ 🟡 Beta Testers             2 agents   [Manage]   │  │
│          │  │    Key: pk_team_e5f6...***   0 online              │  │
│ Logs     │  │                                                    │  │
│          │  └────────────────────────────────────────────────────┘  │
│ Settings │                                                          │
│          │  Recent Activity                                         │
│          │  ─────────────────                                       │
│          │  14:32 - frontend-dev-01 joined Dev Team Alpha          │
│          │  14:28 - Task "Build API" completed by backend-dev-01   │
│          │  14:15 - Team key rotated for Operations Team           │
│          │                                                          │
└──────────┴──────────────────────────────────────────────────────────┘
```

### Team Management Page

```
┌─────────────────────────────────────────────────────────────────────┐
│ Team: Development Team Alpha                    [Edit] [Rotate Key] │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│ Team Key                                                            │
│ ┌─────────────────────────────────────────────────────────────────┐ │
│ │ pk_team_aHR0cHM6Ly9naXRodWIuY29tL3NpX1c5V1J0...    [Copy] [🔄] │ │
│ └─────────────────────────────────────────────────────────────────┘ │
│                                                                     │
│ Connected Agents (5/10 max)                           [+ Add Agent] │
│ ┌─────────────────────────────────────────────────────────────────┐ │
│ │ Agent ID        │ Role      │ Status │ Tasks │ Last Seen │ Act │ │
│ ├─────────────────────────────────────────────────────────────────┤ │
│ │ frontend-dev-01 │ frontend  │ 🟢 Online  │ 2/2   │ 10s ago  │ ⚙️ │ │
│ │ backend-dev-01  │ backend   │ 🟢 Online  │ 1/3   │ 5s ago   │ ⚙️ │ │
│ │ backend-dev-02  │ backend   │ 🟢 Online  │ 0/3   │ 15s ago  │ ⚙️ │ │
│ │ qa-engineer-01  │ qa        │ 🟡 Busy    │ 1/2   │ 30s ago  │ ⚙️ │ │
│ │ devops-01       │ devops    │ ⚫ Offline │ -     │ 2h ago   │ ⚙️ │ │
│ └─────────────────────────────────────────────────────────────────┘ │
│                                                                     │
│ Role Definitions                                                    │
│ ┌─────────────────────────────────────────────────────────────────┐ │
│ │ Frontend Developer                                              │ │
│ │ ├── Capabilities: react, vue, typescript, css                  │ │
│ │ ├── Max Tasks: 2                                                │ │
│ │ └── Tools: read_file, write_file, exec                         │ │
│ │                                                                 │ │
│ │ Backend Developer                                               │ │
│ │ ├── Capabilities: go, python, postgresql, redis                │ │
│ │ ├── Max Tasks: 3                                                │ │
│ │ └── Tools: read_file, write_file, exec, web_search             │ │
│ └─────────────────────────────────────────────────────────────────┘ │
│                                                                     │
│ Danger Zone                                                         │
│ ┌─────────────────────────────────────────────────────────────────┐ │
│ │ [Disband Team]  [Evict All Agents]  [Reset All Keys]           │ │
│ └─────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

## CLI Commands

### Team Management Commands

```bash
# Create a new team
picoclaw team create --name "Dev Team Beta" --description "Secondary dev team"

# List all teams
picoclaw team list
# Output:
# ID                NAME                AGENTS    STATUS    KEY
# dev-team-alpha    Dev Team Alpha      5/10      active    pk_team_a1b2...***
# ops-team          Operations Team     3/5       active    pk_team_c3d4...***

# Show team details
picoclaw team show dev-team-alpha

# Generate/regenerate team key
picoclaw team rotate-key dev-team-alpha

# Add role to team
picoclaw team role add dev-team-alpha --id "security" \
  --capabilities "audit,scan,compliance" \
  --max-tasks 2

# Remove agent from team
picoclaw team evict dev-team-alpha --agent frontend-dev-01

# Wipe agent (remove all team keys from agent)
picoclaw team wipe-agent frontend-dev-01

# Activate/deactivate team
picoclaw team activate dev-team-alpha
picoclaw team deactivate dev-team-alpha
```

### Agent Commands with Team Context

```bash
# Join a team as worker
picoclaw agent --join dev-team-alpha --role frontend

# List teams available on gateway
picoclaw agent --discover-teams

# Show current team membership
picoclaw agent --team-status

# Leave current team
picoclaw agent --leave-team
```

### Interactive Commands (within agent session)

```
User: /team list
Controller: Available teams:
  1. Dev Team Alpha (5 agents, 3 online)
  2. Operations Team (3 agents, 3 online)
  
User: /team activate dev-team-alpha
Controller: Activated team "Dev Team Alpha". 
            Available workers: frontend-dev-01, backend-dev-01, backend-dev-02, qa-engineer-01

User: /team status
Controller: Team: Dev Team Alpha
            ├─ frontend-dev-01: 🟢 Online, 1/2 tasks
            ├─ backend-dev-01: 🟢 Online, 0/3 tasks  
            ├─ backend-dev-02: 🟢 Online, 2/3 tasks
            └─ qa-engineer-01: 🟡 Busy, 1/2 tasks

User: /team evict frontend-dev-01
Controller: Evicted frontend-dev-01 from team. They will no longer receive tasks.

User: Create a login page
Controller: I'll delegate this to the frontend team.
           [Delegating to frontend-dev-01...]
```

## Skill-Based Team Management

Alternative to `/team` commands: use a dedicated skill.

### Skill: `team_manager`

```go
// pkg/skills/team_manager/skill.go

type TeamManagerSkill struct {
    teamService *teams.Service
}

func (s *TeamManagerSkill) Name() string { return "team_manager" }

func (s *TeamManagerSkill) Description() string {
    return "Manage multi-agent teams: create teams, assign agents, rotate keys, monitor status"
}

func (s *TeamManagerSkill) AvailableTools() []tools.Tool {
    return []tools.Tool{
        &CreateTeamTool{teamService: s.teamService},
        &ListTeamsTool{teamService: s.teamService},
        &AddAgentToTeamTool{teamService: s.teamService},
        &EvictAgentTool{teamService: s.teamService},
        &RotateTeamKeyTool{teamService: s.teamService},
        &GetTeamStatusTool{teamService: s.teamService},
    }
}
```

### Usage via Skill

```
User: @team_manager create a new team called "Security Team" for security audits
→ Tool: create_team(name="Security Team", description="...")
→ Result: Created team "security-team" with key pk_team_xxx...

User: @team_manager add my current agent to the security team as a security auditor
→ Tool: add_agent_to_team(team_id="security-team", agent_id="current", role="auditor")
→ Result: Added agent to team

User: @team_manager show me the status of all teams
→ Tool: list_teams()
→ Result: [Team list with status]

User: @team_manager rotate the key for the security team
→ Tool: rotate_team_key(team_id="security-team")
→ Result: New key generated. Agents must reconnect with new key.
```

## Docker Integration

### Dockerfile with Role Definition

```dockerfile
FROM picoclaw:latest

# Set team and role via environment
ENV PICOCLAW_TEAM_ID=dev-team-alpha
ENV PICOCLAW_AGENT_ROLE=frontend
ENV PICOCLAW_TEAM_KEY_FILE=/run/secrets/team_key
ENV PICOCLAW_GATEWAY_ADDRESS=gateway:18791

# Or via build args
ARG TEAM_KEY
ENV PICOCLAW_TEAM_KEY=${TEAM_KEY}

# Expose agent port if needed
EXPOSE 5001

CMD ["picoclaw", "agent", "--mode=worker", "--auto-connect"]
```

### Docker Compose Example

```yaml
version: '3.8'

services:
  # Gateway (Controller)
  gateway:
    image: picoclaw:latest
    command: gateway
    ports:
      - "18790:18790"  # Web UI / API
      - "18791:18791"  # Agent protocol
    volumes:
      - gateway-data:/data
      - ./config:/config
    environment:
      - PICOCLAW_CONFIG=/config/gateway.yaml
      - PICOCLAW_TEAMS_ENABLED=true
    networks:
      - picoclaw

  # Frontend Worker
  frontend-dev:
    image: picoclaw-worker:latest
    environment:
      - PICOCLAW_AGENT_ID=frontend-dev-01
      - PICOCLAW_AGENT_ROLE=frontend
      - PICOCLAW_TEAM_ID=dev-team-alpha
      - PICOCLAW_GATEWAY_ADDRESS=gateway:18791
    secrets:
      - team_key_dev_alpha
    volumes:
      - frontend-workspace:/workspace
    networks:
      - picoclaw
    depends_on:
      - gateway

  # Backend Worker  
  backend-dev:
    image: picoclaw-worker:latest
    environment:
      - PICOCLAW_AGENT_ID=backend-dev-01
      - PICOCLAW_AGENT_ROLE=backend
      - PICOCLAW_TEAM_ID=dev-team-alpha
      - PICOCLAW_GATEWAY_ADDRESS=gateway:18791
    secrets:
      - team_key_dev_alpha
    volumes:
      - backend-workspace:/workspace
    networks:
      - picoclaw
    depends_on:
      - gateway

  # QA Worker
  qa-engineer:
    image: picoclaw-worker:latest
    environment:
      - PICOCLAW_AGENT_ID=qa-engineer-01
      - PICOCLAW_AGENT_ROLE=qa
      - PICOCLAW_TEAM_ID=dev-team-alpha
      - PICOCLAW_GATEWAY_ADDRESS=gateway:18791
    secrets:
      - team_key_dev_alpha
    volumes:
      - qa-workspace:/workspace
    networks:
      - picoclaw
    depends_on:
      - gateway

secrets:
  team_key_dev_alpha:
    file: ./secrets/team_key_dev_alpha.txt

volumes:
  gateway-data:
  frontend-workspace:
  backend-workspace:
  qa-workspace:

networks:
  picoclaw:
    driver: bridge
```

### Kubernetes Deployment

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: team-keys
type: Opaque
stringData:
  dev-alpha-key: "pk_team_..."
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: picoclaw-workers
spec:
  replicas: 3
  selector:
    matchLabels:
      app: picoclaw-worker
  template:
    metadata:
      labels:
        app: picoclaw-worker
        team: dev-alpha
        role: backend
    spec:
      containers:
      - name: worker
        image: picoclaw:latest
        env:
        - name: PICOCLAW_MODE
          value: "worker"
        - name: PICOCLAW_TEAM_ID
          value: "dev-team-alpha"
        - name: PICOCLAW_AGENT_ROLE
          value: "backend"
        - name: PICOCLAW_GATEWAY_ADDRESS
          value: "picoclaw-gateway:18791"
        - name: PICOCLAW_TEAM_KEY
          valueFrom:
            secretKeyRef:
              name: team-keys
              key: dev-alpha-key
```

## Agent Lifecycle Management

### Connection Flow

```
┌──────────────┐              ┌──────────────────────┐
│ Worker Agent │              │ Gateway (Controller) │
└──────┬───────┘              └──────────┬───────────┘
       │                                 │
       │ 1. TCP Connect to Gateway       │
       │ ───────────────────────────────>│
       │                                 │
       │ 2. Team Auth Challenge          │
       │<────────────────────────────────│
       │                                 │
       │ 3. Auth Response (signed)       │
       │ ───────────────────────────────>│
       │                                 │
       │ 4. Session Established          │
       │<════════════════════════════════│ (encrypted)
       │                                 │
       │ 5. Register with Team+Role      │
       │ ───────────────────────────────>│
       │                                 │
       │ 6. Join Team (if approved)      │
       │<────────────────────────────────│
       │                                 │
       │ 7. Heartbeat (every 30s)        │
       │◄───────────────────────────────>│
       │                                 │
```

### Eviction/Wipe Process

```
Controller decides to evict agent:

1. Send eviction notice to agent
   ┌─────────────┐           ┌─────────────┐
   │  Controller │ ────────> │    Agent    │
   └─────────────┘  EVICT    └─────────────┘

2. Agent acknowledges and wipes local team key
   ┌─────────────┐           ┌─────────────┐
   │  Controller │ <──────── │    Agent    │
   └─────────────┘  ACK      └─────────────┘
                    (wipes key)

3. Connection closed
4. Agent removed from team roster
5. Agent must re-authenticate with new key to rejoin
```

### Wipe Command

```bash
# Controller initiates wipe
picoclaw team wipe-agent frontend-dev-01 --reason "Security rotation"

# Or via interactive command
User: /team wipe frontend-dev-01
Controller: ⚠️  This will remove frontend-dev-01 from the team and wipe their team key.
             They will need the new team key to rejoin.
             Reason for wipe: [security rotation | compromised key | team change]
             Confirm? [y/N]: y
             
             Wiping frontend-dev-01... ✓
             Agent has been disconnected and team key wiped.
             
             Would you like to:
             1. Generate new team key for remaining agents
             2. Keep current team key
             
             Select: 1
             
             New team key generated: pk_team_xxx...
             Agents will auto-reconnect with new key.
```

## Implementation Plan

### Phase 1: Gateway Team Management Core
- [ ] Extend config with `teams` section
- [ ] Create `pkg/teams/service.go` for team management
- [ ] Add team CRUD operations
- [ ] Team key generation and storage

### Phase 2: Gateway Web UI
- [ ] Extend gateway UI with team pages
- [ ] Team list/create/edit views
- [ ] Agent roster management
- [ ] Key rotation UI

### Phase 3: Agent Protocol
- [ ] Extend agent protocol for team registration
- [ ] Gateway as controller by default
- [ ] Agent team join/leave
- [ ] Role advertisement

### Phase 4: CLI Integration
- [ ] `picoclaw team` commands
- [ ] `/team` interactive commands
- [ ] Agent team status

### Phase 5: Docker/K8s Support
- [ ] Environment variable config
- [ ] Docker Compose examples
- [ ] Kubernetes manifests
- [ ] Helm chart

### Phase 6: Skills (Optional)
- [ ] `team_manager` skill
- [ ] Natural language team management

## API Endpoints (Gateway)

```
# Team Management
GET    /api/v1/teams                    # List teams
POST   /api/v1/teams                    # Create team
GET    /api/v1/teams/:id                # Get team details
PUT    /api/v1/teams/:id                # Update team
DELETE /api/v1/teams/:id                # Disband team
POST   /api/v1/teams/:id/rotate-key     # Rotate team key

# Agent Management within Team
GET    /api/v1/teams/:id/agents         # List team agents
POST   /api/v1/teams/:id/agents         # Add agent to team
DELETE /api/v1/teams/:id/agents/:aid    # Evict agent from team
POST   /api/v1/teams/:id/agents/:aid/wipe  # Wipe agent

# Role Management
GET    /api/v1/teams/:id/roles          # List roles
POST   /api/v1/teams/:id/roles          # Create role
PUT    /api/v1/teams/:id/roles/:rid     # Update role
DELETE /api/v1/teams/:id/roles/:rid     # Delete role

# Agent Protocol (WebSocket/TCP)
WS     /agent/v1/connect                # Agent connection endpoint
```

## Security Considerations

1. **Team Key Storage**
   - Gateway stores team keys encrypted at rest
   - Keys never logged in plaintext
   - Memory-only during rotation

2. **Agent Authentication**
   - Challenge-response for each connection
   - Session keys rotated periodically
   - Automatic eviction on failed auth

3. **Access Control**
   - Web UI requires authentication
   - Role-based access to team management
   - Audit log for all team operations

4. **Network Security**
   - TLS for all agent-gateway communication
   - Optional mTLS for agent certificates
   - Network policies for K8s deployments

## Summary

This centralized gateway architecture provides:

- ✅ **Single Control Point**: Gateway manages all teams and agents
- ✅ **Visual Management**: Web UI for team administration
- ✅ **CLI Integration**: Commands and interactive /team commands
- ✅ **Docker-Native**: First-class container support
- ✅ **Secure**: Team keys with challenge-response auth
- ✅ **Lifecycle Management**: Evict, wipe, rotate keys
- ✅ **Scalable**: Multiple teams, many agents per team

