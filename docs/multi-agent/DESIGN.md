# Multi-Agent Architecture Design for PicoClaw

## Executive Summary

This document proposes a comprehensive multi-agent system for PicoClaw that enables multiple specialized agents to collaborate, communicate, and solve complex problems through a controller-worker pattern.

## Use Cases

### 1. Software Development Team
```
User → Controller Agent → Task Distribution
                    ├── Frontend Dev Agent (React/Vue)
                    ├── Backend Dev Agent (API/DB)
                    ├── QA Agent (Testing)
                    └── DevOps Agent (CI/CD)
```

### 2. IT Operations Center
```
User → Network Admin Agent → Specialized Agents
                    ├── Kubernetes Agent
                    ├── Database Agent
                    ├── Security Agent
                    └── Monitoring Agent
```

### 3. Troubleshooting Chain
```
User → Triage Agent → Diagnosis Agents
                    ├── Network Diagnostics
                    ├── System Logs Analysis
                    └── Application Debugging
```

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    CONTROLLER AGENT                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │  Task Router │  │  Orchestrator│  │  Result Aggregator   │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
└──────────┬─────────────────────────────────────────────────────┘
           │ Agent Communication Bus (ACB)
           ▼
┌─────────────────────────────────────────────────────────────────┐
│                     WORKER AGENTS POOL                           │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐   │
│  │  Dev-1  │ │  QA-1   │ │  Ops-1  │ │ Net-1   │ │ K8s-1   │   │
│  │ Frontend│ │ Testing │ │Deploy   │ │Network  │ │K8s Mgmt │   │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Communication Mechanisms

### 1. Agent Communication Bus (ACB)

Extension of the existing bus system with agent-specific message types:

```go
package bus

// AgentMessage types for inter-agent communication
type AgentMessageType string

const (
    AgentMsgTask       AgentMessageType = "task"       // Task assignment
    AgentMsgResult     AgentMessageType = "result"     // Task completion
    AgentMsgQuery      AgentMessageType = "query"      // Information request
    AgentMsgBroadcast  AgentMessageType = "broadcast"  // Broadcast to all
    AgentMsgDiscovery  AgentMessageType = "discovery"  // Agent discovery
    AgentMsgHeartbeat  AgentMessageType = "heartbeat"  // Health check
    AgentMsgNegotiate  AgentMessageType = "negotiate"  // Capability negotiation
)

type AgentMessage struct {
    ID          string           `json:"id"`
    Type        AgentMessageType `json:"type"`
    From        string           `json:"from"`        // Source agent ID
    To          string           `json:"to"`          // Target agent ID ("*" for broadcast)
    TaskID      string           `json:"task_id,omitempty"`
    Payload     json.RawMessage  `json:"payload"`
    Priority    int              `json:"priority"`    // 0-10, higher = more urgent
    Timestamp   time.Time        `json:"timestamp"`
    TTL         time.Duration    `json:"ttl"`         // Time-to-live for message
    ReplyTo     string           `json:"reply_to,omitempty"` // For async responses
}

// TaskPayload for task assignment
type TaskPayload struct {
    Description string            `json:"description"`
    Requirements []string         `json:"requirements"`
    Context     map[string]string `json:"context"`
    Deadline    *time.Time        `json:"deadline,omitempty"`
    Dependencies []string         `json:"dependencies"` // Other task IDs
}

// ResultPayload for task completion
type ResultPayload struct {
    TaskID      string      `json:"task_id"`
    Status      TaskStatus  `json:"status"` // pending, running, completed, failed
    Output      string      `json:"output"`
    Artifacts   []Artifact  `json:"artifacts"` // Files, URLs, etc.
    Metrics     TaskMetrics `json:"metrics"`   // Duration, tokens used, etc.
}

// DiscoveryPayload for agent capability advertisement
type DiscoveryPayload struct {
    AgentID      string            `json:"agent_id"`
    Capabilities []Capability      `json:"capabilities"`
    Load         float64           `json:"load"`      // 0.0 - 1.0
    Status       AgentStatus       `json:"status"`    // active, busy, offline
    Metadata     map[string]string `json:"metadata"`
}

type Capability struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Skills      []string `json:"skills"`      // e.g., ["python", "kubernetes", "react"]
    Tools       []string `json:"tools"`       // Available tools
}
```

### 2. Communication Patterns

#### A. Direct Message (Point-to-Point)
```
Controller → Worker A
           ← Result from Worker A
```

#### B. Publish-Subscribe (Broadcast)
```
Controller → [Worker A, Worker B, Worker C] (broadcast discovery)
           ← [Capabilities A, Capabilities B, Capabilities C]
```

#### C. Request-Reply (Query)
```
Worker A → Query to Worker B
         ← Response from Worker B
```

#### D. Pipeline (Chained Tasks)
```
User Input → Parser Agent → Analyzer Agent → Generator Agent → Output
```

### 3. Communication Transport Options

| Transport | Use Case | Pros | Cons |
|-----------|----------|------|------|
| **In-Memory Bus** | Single process, multiple agents | Fast, simple | Limited to single instance |
| **Unix Domain Socket** | Same machine, separate processes | Fast, secure | Same machine only |
| **TCP Socket** | Network distributed | Scalable, remote | Network overhead, security |
| **WebSocket** | Browser-based agents | Real-time bidirectional | Requires HTTP server |
| **Message Queue** (Redis/RabbitMQ) | Enterprise scale | Reliable, persistent | Infrastructure required |
| **gRPC** | Microservices architecture | Efficient, typed | Complexity |

## Agent Configuration

### 1. Agent Roles Definition

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.picoclaw/workspace",
      "model_name": "kimi-coding"
    },
    "list": [
      {
        "id": "controller",
        "name": "Dev Team Lead",
        "role": "controller",
        "model": {
          "primary": "kimi-coding",
          "fallbacks": ["claude-sonnet", "gpt-4"]
        },
        "system_prompt": "You are a technical lead coordinating a development team...",
        "skills": ["orchestration", "planning", "code_review"],
        "multi_agent": {
          "enabled": true,
          "discovery_method": "auto",
          "max_workers": 5,
          "coordination_strategy": "parallel"
        }
      },
      {
        "id": "frontend-dev",
        "name": "Frontend Developer",
        "role": "worker",
        "specialization": "frontend",
        "model": {
          "primary": "kimi-coding"
        },
        "system_prompt": "You are a frontend developer expert in React, Vue, and CSS...",
        "skills": ["react", "vue", "typescript", "css", "ui_design"],
        "capabilities": {
          "can_handle": ["ui_components", "styling", "frontend_architecture"],
          "max_concurrent_tasks": 2
        },
        "multi_agent": {
          "enabled": true,
          "register_with": ["controller"],
          "heartbeat_interval": 30
        }
      },
      {
        "id": "backend-dev",
        "name": "Backend Developer",
        "role": "worker",
        "specialization": "backend",
        "system_prompt": "You are a backend developer expert in Go, Python, and databases...",
        "skills": ["go", "python", "postgresql", "redis", "api_design"],
        "capabilities": {
          "can_handle": ["api_endpoints", "database_schema", "business_logic"],
          "max_concurrent_tasks": 3
        },
        "multi_agent": {
          "enabled": true,
          "register_with": ["controller"],
          "heartbeat_interval": 30
        }
      },
      {
        "id": "qa-engineer",
        "name": "QA Engineer",
        "role": "worker",
        "specialization": "testing",
        "system_prompt": "You are a QA engineer focused on test automation...",
        "skills": ["testing", "cypress", "jest", "go_test"],
        "capabilities": {
          "can_handle": ["unit_tests", "integration_tests", "e2e_tests"],
          "max_concurrent_tasks": 2
        },
        "multi_agent": {
          "enabled": true,
          "register_with": ["controller"],
          "heartbeat_interval": 30
        }
      },
      {
        "id": "devops",
        "name": "DevOps Engineer",
        "role": "worker",
        "specialization": "infrastructure",
        "system_prompt": "You are a DevOps engineer expert in Kubernetes and CI/CD...",
        "skills": ["kubernetes", "docker", "github_actions", "terraform"],
        "capabilities": {
          "can_handle": ["deployment", "infrastructure", "ci_cd"],
          "max_concurrent_tasks": 2
        },
        "multi_agent": {
          "enabled": true,
          "register_with": ["controller"],
          "heartbeat_interval": 30
        }
      }
    ]
  }
}
```

### 2. Agent Network Configuration

```json
{
  "agent_network": {
    "mode": "local",
    "transport": "unix_socket",
    "discovery": {
      "enabled": true,
      "method": "multicast",
      "interval": 60,
      "timeout": 10
    },
    "communication": {
      "encryption": "tls",
      "auth_method": "token",
      "max_message_size": "10MB",
      "compression": true
    },
    "controller": {
      "bind_address": "/tmp/picoclaw-controller.sock",
      "max_workers": 10,
      "task_timeout": 300,
      "retry_policy": {
        "max_retries": 3,
        "backoff": "exponential"
      }
    },
    "workers": {
      "auto_register": true,
      "heartbeat_interval": 30,
      "max_missed_heartbeats": 3
    }
  }
}
```

## Implementation Architecture

### 1. Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                  AGENT NETWORK LAYER                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   Router    │  │  Discovery  │  │  Connection Manager │ │
│  │  (routes    │  │  (finds     │  │  (maintains conns)  │ │
│  │   messages) │  │   agents)   │  │                     │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                 AGENT COMMUNICATION BUS                      │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Message Queue • Priority Queue • Broadcast Handler │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   AGENT RUNTIME                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  Controller  │  │    Worker    │  │  Autonomous Agent│  │
│  │   Runtime    │  │   Runtime    │  │     Runtime      │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 2. New Tools for Multi-Agent

```go
// tools/agent_communication.go

// AgentMessageTool - Send messages to other agents
type AgentMessageTool struct {
    bus *bus.AgentBus
}

func (t *AgentMessageTool) Name() string { return "send_agent_message" }
func (t *AgentMessageTool) Description() string { 
    return "Send a message to another agent. Use for task assignment, queries, or coordination." 
}

// AgentQueryTool - Query agent capabilities
type AgentQueryTool struct {
    registry *agent.AgentRegistry
}

func (t *AgentQueryTool) Name() string { return "discover_agents" }
func (t *AgentQueryTool) Description() string {
    return "Discover available agents and their capabilities. Use to find the right agent for a task."
}

// AgentSpawnTool - Spawn a new agent instance
type AgentSpawnTool struct {
    orchestrator *agent.Orchestrator
}

func (t *AgentSpawnTool) Name() string { return "spawn_agent" }
func (t *AgentSpawnTool) Description() string {
    return "Spawn a new agent instance dynamically. Use when existing agents cannot handle the workload."
}

// TaskOrchestrationTool - Complex task distribution
type TaskOrchestrationTool struct {
    orchestrator *agent.Orchestrator
}

func (t *TaskOrchestrationTool) Name() string { return "orchestrate_task" }
func (t *TaskOrchestrationTool) Description() string {
    return "Break down complex tasks and distribute to multiple agents. Manages dependencies and aggregation."
}
```

### 3. Controller Agent Logic

```go
package agent

// ControllerAgent extends AgentInstance with orchestration capabilities
type ControllerAgent struct {
    *AgentInstance
    orchestrator *Orchestrator
    workers      map[string]*WorkerInfo
    taskQueue    *TaskQueue
}

type Orchestrator struct {
    registry      *AgentRegistry
    bus           *bus.AgentBus
    strategy      OrchestrationStrategy
    taskPlanner   *TaskPlanner
    resultAggregator *ResultAggregator
}

// OrchestrationStrategy defines how tasks are distributed
type OrchestrationStrategy int

const (
    StrategyRoundRobin OrchestrationStrategy = iota
    StrategyLoadBalanced
    StrategyCapabilityMatch
    StrategyParallel
    StrategySequential
    StrategyAdaptive
)

// DistributeTask analyzes and distributes tasks to workers
func (o *Orchestrator) DistributeTask(ctx context.Context, task Task) (*TaskPlan, error) {
    // 1. Analyze task requirements
    requirements := o.taskPlanner.AnalyzeRequirements(task)
    
    // 2. Find capable workers
    candidates := o.findCapableWorkers(requirements)
    
    // 3. Apply selection strategy
    selected := o.selectWorkers(candidates, requirements, o.strategy)
    
    // 4. Create task plan with dependencies
    plan := o.taskPlanner.CreatePlan(task, selected)
    
    // 5. Distribute subtasks
    for _, subtask := range plan.Subtasks {
        if err := o.assignToWorker(ctx, subtask); err != nil {
            return nil, err
        }
    }
    
    return plan, nil
}

// HandleResult processes worker results
func (o *Orchestrator) HandleResult(result TaskResult) error {
    // 1. Store result
    o.taskPlanner.UpdateTaskStatus(result.TaskID, result.Status)
    
    // 2. Check if all subtasks complete
    if o.taskPlanner.IsComplete(result.ParentTaskID) {
        // 3. Aggregate results
        aggregated := o.resultAggregator.Aggregate(result.ParentTaskID)
        
        // 4. Notify controller agent
        return o.notifyCompletion(result.ParentTaskID, aggregated)
    }
    
    // 5. Handle partial results - trigger dependent tasks
    return o.triggerDependentTasks(result.TaskID)
}
```

## Auto-Discovery Mechanisms

### 1. Local Discovery (Same Machine)

```go
// Uses Unix domain sockets with multicast on local network

type LocalDiscovery struct {
    multicastAddr string
    port          int
}

func (d *LocalDiscovery) Discover() ([]AgentInfo, error) {
    // 1. Send multicast discovery request
    // 2. Listen for responses (with timeout)
    // 3. Parse agent advertisements
    // 4. Return list of available agents
}
```

### 2. Network Discovery (Distributed)

```go
type NetworkDiscovery struct {
    registryURL string  // Optional central registry
    useConsul   bool    // Use HashiCorp Consul
    useMDNS     bool    // Use multicast DNS
}

// mDNS (Bonjour/Avahi) for local network discovery
func (d *NetworkDiscovery) discoverMDNS() ([]AgentInfo, error) {
    // Browse for _picoclaw-agent._tcp services
}

// Consul for enterprise discovery
func (d *NetworkDiscovery) discoverConsul() ([]AgentInfo, error) {
    // Query Consul service catalog
}
```

### 3. Manual Registration

```json
{
  "agent_network": {
    "static_peers": [
      {
        "agent_id": "frontend-dev-01",
        "address": "192.168.1.100:5001",
        "capabilities": ["react", "vue"]
      },
      {
        "agent_id": "backend-dev-01",
        "address": "192.168.1.101:5001",
        "capabilities": ["go", "python"]
      }
    ]
  }
}
```

## Communication Flow Example

### Software Development Workflow

```
1. User Request: "Create a web app with user authentication"

2. Controller Agent (Analyze):
   ├─ Break down into subtasks:
   │  ├── Design database schema
   │  ├── Create backend API
   │  ├── Build frontend UI
   │  └── Write tests
   └─ Query: discover_agents(capabilities=["backend", "frontend", "qa"])

3. Controller Agent (Distribute):
   ├─ send_agent_message(to="backend-dev", task="Create auth API")
   ├─ send_agent_message(to="frontend-dev", task="Build login UI")
   └─ send_agent_message(to="qa-engineer", task="Write auth tests")

4. Workers Execute (Parallel):
   ├─ Backend Dev → Creates API endpoints
   ├─ Frontend Dev → Creates React components
   └─ QA Engineer → Writes Cypress tests

5. Workers Report:
   ├─ Each sends: send_agent_message(to="controller", result=...)
   └─ Controller tracks completion

6. Controller Aggregates:
   ├─ Combines all outputs
   ├─ Validates integration points
   └─ Reports to user

7. User Review → Iterate if needed
```

## Security Considerations

### 1. Authentication
- Token-based auth between agents
- mTLS for network communication
- Capability-based access control

### 2. Isolation
- Sandboxed execution for worker agents
- Workspace separation
- Resource limits (CPU, memory, disk)

### 3. Audit Logging
- All inter-agent messages logged
- Task assignments tracked
- Results stored for accountability

## Implementation Phases

### Phase 1: Basic Multi-Agent (MVP)
- [ ] Agent message bus extension
- [ ] Basic controller-worker communication
- [ ] Static agent configuration
- [ ] Simple task distribution

### Phase 2: Auto-Discovery
- [ ] Local network discovery (mDNS)
- [ ] Dynamic agent registration
- [ ] Heartbeat mechanism
- [ ] Capability advertisement

### Phase 3: Advanced Orchestration
- [ ] Task planning and dependency management
- [ ] Result aggregation strategies
- [ ] Parallel execution
- [ ] Error handling and retries

### Phase 4: Enterprise Features
- [ ] Distributed across machines
- [ ] Central registry option
- [ ] Load balancing
- [ ] Metrics and monitoring

## Configuration Example

```json
{
  "agents": {
    "list": [
      {
        "id": "dev-team-lead",
        "role": "controller",
        "multi_agent": {
          "enabled": true,
          "orchestration": {
            "strategy": "parallel",
            "max_workers": 5,
            "task_timeout": 300
          }
        }
      }
    ]
  },
  "agent_network": {
    "enabled": true,
    "mode": "local",
    "discovery": "auto",
    "transports": ["unix_socket", "tcp"]
  }
}
```

## Migration Path

Existing single-agent setups continue to work unchanged. Multi-agent features are opt-in via configuration.

