# Multi-Agent Quick Start Guide

## Overview

This guide shows you how to set up and test a multi-agent PicoClaw configuration with a software development team.

## Prerequisites

- PicoClaw installed and working
- Kimi Coding API key configured
- Go 1.23+ (for building from source)

## Step 1: Configure the Multi-Agent Setup

1. Copy the example config:
```bash
cp docs/multi-agent/EXAMPLE_CONFIG.json ~/.picoclaw/config.json
```

2. Edit and add your API key:
```bash
nano ~/.picoclaw/config.json
# Replace YOUR_KIMI_API_KEY_HERE with your actual key
```

## Step 2: Create Workspaces

Each agent needs its own workspace:

```bash
mkdir -p ~/.picoclaw/workspace-frontend
mkdir -p ~/.picoclaw/workspace-backend
mkdir -p ~/.picoclaw/workspace-qa
mkdir -p ~/.picoclaw/workspace-devops
```

## Step 3: Test the Setup

### Check Status
```bash
picoclaw status
```

You should see:
- dev-lead (controller)
- frontend-dev (worker)
- backend-dev (worker)
- qa-engineer (worker)
- devops (worker)

### Test Basic Agent Communication

Start the controller agent interactively:

```bash
picoclaw agent
```

Try these commands:

```
# Ask the controller to list available workers
List available team members and their capabilities

# Delegate a task
Delegate the frontend UI task to the frontend developer

# Get status of all workers
What is the status of all team members?
```

## Step 4: Example Workflows

### Example 1: Build a Web Application

**User Request:** "Create a user authentication system for a web app"

**Expected Flow:**
1. Controller analyzes the request
2. Controller delegates tasks:
   - Backend Dev: Create auth API endpoints
   - Frontend Dev: Build login/signup UI
   - QA Engineer: Write auth tests
   - DevOps: Create Docker setup
3. Workers execute in parallel
4. Controller aggregates results
5. Final report to user

### Example 2: Troubleshoot an Issue

**User Request:** "Debug why the API is returning 500 errors"

**Expected Flow:**
1. Controller delegates to Backend Dev
2. Backend Dev investigates logs/code
3. If infrastructure suspected, Backend Dev asks Controller to involve DevOps
4. Root cause identified and fix proposed

### Example 3: Code Review Pipeline

**User Request:** "Review this code for best practices"

**Expected Flow:**
1. Controller distributes to multiple workers:
   - Backend Dev: Review Go/Python code
   - Frontend Dev: Review UI components
   - QA Engineer: Check test coverage
2. Controller aggregates all reviews
3. Unified feedback to user

## Available Tools (Once Implemented)

### For Controller Agents:

| Tool | Purpose | Example |
|------|---------|---------|
| `discover_agents` | Find available workers | "Find all backend developers" |
| `send_agent_message` | Send direct message | "Message to backend-dev: create API" |
| `delegate_task` | Assign task with tracking | "Delegate DB design to backend" |
| `broadcast_task` | Send to all capable agents | "All devs: review this PR" |
| `orchestrate_task` | Complex multi-step workflow | "Build full auth system" |

### For All Agents:

| Tool | Purpose |
|------|---------|
| `report_result` | Send completion status |
| `request_help` | Ask another agent for assistance |
| `query_agent` | Get info from specific agent |

## Configuration Options

### Static Worker Configuration

Pre-define workers in config:

```json
{
  "multi_agent": {
    "enabled": true,
    "discovery_method": "static",
    "workers": ["frontend-dev", "backend-dev"]
  }
}
```

### Dynamic Discovery

Auto-discover workers on local network:

```json
{
  "multi_agent": {
    "enabled": true,
    "discovery_method": "auto",
    "discovery_interval": 60
  }
}
```

### Manual Registration

Workers register with controller at startup:

```json
{
  "multi_agent": {
    "enabled": true,
    "role": "worker",
    "register_with": ["dev-lead"],
    "heartbeat_interval": 30
  }
}
```

## Monitoring

### Check Worker Status

```bash
# View all agents and their status
picoclaw status --verbose

# Check specific worker
picoclaw agent --id backend-dev status
```

### View Communication Logs

```bash
# Enable debug logging
picoclaw agent --debug

# View message bus activity
tail -f ~/.picoclaw/logs/agent-bus.log
```

## Troubleshooting

### Workers Not Found

1. Check config has correct agent IDs
2. Verify workspaces exist
3. Check logs for discovery errors

### Messages Not Delivered

1. Check if controller is running
2. Verify network transport (unix/tcp)
3. Check firewall settings

### Task Stuck

1. Check worker heartbeat status
2. Review worker logs
3. Verify task was properly assigned

## Advanced: Custom Agent Types

Create specialized agents for your domain:

```json
{
  "id": "security-auditor",
  "name": "Security Specialist",
  "role": "worker",
  "specialization": "security",
  "system_prompt": "You are a security specialist...",
  "capabilities": {
    "can_handle": ["security_audit", "vulnerability_scan", "compliance_check"]
  }
}
```

## Performance Tuning

### For High-Volume Tasks

```json
{
  "multi_agent": {
    "coordination_strategy": "parallel",
    "max_workers": 10,
    "task_timeout": 600
  }
}
```

### For Sequential Dependencies

```json
{
  "multi_agent": {
    "coordination_strategy": "sequential",
    "dependency_resolution": "explicit"
  }
}
```

## Next Steps

1. **Try the examples** above
2. **Customize the agents** for your workflow
3. **Add custom capabilities** to agents
4. **Integrate with your CI/CD** via DevOps agent
5. **Scale horizontally** with network discovery

## Getting Help

- View logs: `~/.picoclaw/logs/`
- Check agent workspaces: `~/.picoclaw/workspace-*/`
- Debug mode: `picoclaw agent --debug`

