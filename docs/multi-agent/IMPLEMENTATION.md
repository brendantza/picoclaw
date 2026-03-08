# Multi-Agent Implementation Roadmap

## Current State Analysis

PicoClaw already has foundational multi-agent capabilities:

1. **AgentRegistry** (`pkg/agent/registry.go`) - Manages multiple agent instances
2. **SubagentManager** (`pkg/tools/subagent.go`) - Spawns tasks within same process
3. **SpawnTool** (`pkg/tools/spawn.go`) - Async task delegation
4. **Message Bus** (`pkg/bus/bus.go`) - Internal message passing
5. **RouteResolver** (`pkg/routing/route.go`) - Routes messages to agents

## Implementation Plan

### Phase 1: Core Infrastructure (Week 1-2)

#### 1.1 Extend Message Bus for Inter-Agent Communication

**File:** `pkg/bus/agent_bus.go` (new)

```go
package bus

// AgentBus extends MessageBus with agent-specific messaging
type AgentBus struct {
    *MessageBus
    agentMessages chan AgentMessage
    subscribers   map[string][]chan AgentMessage
}

type AgentMessage struct {
    ID       string
    Type     AgentMessageType
    From     string
    To       string
    TaskID   string
    Payload  []byte
    Priority int
}
```

#### 1.2 Create Agent Network Layer

**File:** `pkg/agent/network/` (new package)

```
pkg/agent/network/
├── transport.go       // Transport interface
├── transport_unix.go  // Unix domain socket implementation
├── transport_tcp.go   // TCP implementation
├── discovery.go       // Discovery interface
├── discovery_local.go // Local discovery
├── discovery_mdns.go  // mDNS discovery
└── registry.go        // Network registry
```

#### 1.3 Define Multi-Agent Configuration

**File:** `pkg/config/multi_agent.go` (new)

```go
package config

type MultiAgentConfig struct {
    Enabled     bool                   `json:"enabled"`
    Role        string                 `json:"role"` // "controller", "worker", "autonomous"
    Network     AgentNetworkConfig     `json:"network"`
    Orchestrate OrchestrationConfig    `json:"orchestrate,omitempty"`
    Capabilities AgentCapabilities     `json:"capabilities,omitempty"`
}

type AgentNetworkConfig struct {
    Mode      string `json:"mode"`      // "local", "network"
    Transport string `json:"transport"` // "unix", "tcp", "ws"
    Discovery DiscoveryConfig `json:"discovery"`
    BindAddr  string `json:"bind_address,omitempty"`
}
```

### Phase 2: Controller-Worker Pattern (Week 3-4)

#### 2.1 Controller Agent Runtime

**File:** `pkg/agent/controller.go` (new)

```go
package agent

// ControllerAgent manages worker agents
type ControllerAgent struct {
    *AgentInstance
    orchestrator *Orchestrator
    workerPool   *WorkerPool
    taskManager  *TaskManager
}

// Orchestrator handles task distribution
type Orchestrator struct {
    strategy    DistributionStrategy
    workerPool  *WorkerPool
    taskPlanner *TaskPlanner
}

func (o *Orchestrator) DistributeTask(ctx context.Context, task Task) error {
    // 1. Analyze task
    // 2. Find capable workers
    // 3. Distribute subtasks
    // 4. Monitor progress
}
```

#### 2.2 Worker Agent Runtime

**File:** `pkg/agent/worker.go` (new)

```go
package agent

// WorkerAgent executes tasks for controller
type WorkerAgent struct {
    *AgentInstance
    controllerAddr string
    heartbeatInterval time.Duration
    capabilities   []Capability
}

func (w *WorkerAgent) RegisterWithController() error {
    // Send discovery announcement
    // Wait for task assignments
    // Execute and report back
}
```

#### 2.3 New Multi-Agent Tools

**File:** `pkg/tools/multi_agent.go` (new)

```go
// SendAgentMessageTool - Send direct message to agent
type SendAgentMessageTool struct{}

// DiscoverAgentsTool - Find available agents
type DiscoverAgentsTool struct{}

// DelegateTaskTool - Delegate task to specific agent
type DelegateTaskTool struct{}

// BroadcastTaskTool - Send task to all capable agents
type BroadcastTaskTool struct{}
```

### Phase 3: Advanced Orchestration (Week 5-6)

#### 3.1 Task Planning

**File:** `pkg/agent/planner.go` (new)

```go
package agent

// TaskPlanner breaks down complex tasks
type TaskPlanner struct {
    analyzer *TaskAnalyzer
    splitter *TaskSplitter
    dependencyResolver *DependencyResolver
}

func (p *TaskPlanner) Plan(task Task, workers []Worker) (*ExecutionPlan, error) {
    // Break down task
    // Identify dependencies
    // Assign to workers
    // Create execution graph
}
```

#### 3.2 Result Aggregation

**File:** `pkg/agent/aggregator.go` (new)

```go
package agent

// ResultAggregator combines worker outputs
type ResultAggregator struct {
    strategy AggregationStrategy
}

func (a *ResultAggregator) Aggregate(results []TaskResult) (*AggregatedResult, error) {
    // Merge code from multiple workers
    // Resolve conflicts
    // Validate integration
}
```

### Phase 4: Auto-Discovery (Week 7-8)

#### 4.1 mDNS Discovery

**File:** `pkg/agent/discovery/mdns.go` (new)

```go
package discovery

import "github.com/grandcat/zeroconf"

type MDNSDiscovery struct {
    resolver *zeroconf.Resolver
}

func (d *MDNSDiscovery) Browse(service string) ([]ServiceEntry, error) {
    // Browse for _picoclaw-agent._tcp
}

func (d *MDNSDiscovery) Register(agent AgentInfo) error {
    // Register agent as service
}
```

#### 4.2 Heartbeat & Health Monitoring

**File:** `pkg/agent/health/` (new package)

```go
// Heartbeat monitor for agent liveness
type HeartbeatMonitor struct {
    interval time.Duration
    timeout  time.Duration
}

// HealthChecker monitors agent health
type HealthChecker struct {
    metrics MetricsCollector
}
```

## Code Changes by File

### Existing Files to Modify

| File | Changes |
|------|---------|
| `pkg/config/config.go` | Add MultiAgentConfig to AgentConfig |
| `pkg/agent/registry.go` | Add network awareness |
| `pkg/agent/loop.go` | Handle inter-agent messages |
| `pkg/bus/bus.go` | Add AgentMessage support |
| `pkg/tools/registry.go` | Register new multi-agent tools |

### New Files to Create

| File | Purpose |
|------|---------|
| `pkg/bus/agent_bus.go` | Agent message types and bus |
| `pkg/agent/network/transport.go` | Transport abstraction |
| `pkg/agent/network/discovery.go` | Discovery abstraction |
| `pkg/agent/controller.go` | Controller runtime |
| `pkg/agent/worker.go` | Worker runtime |
| `pkg/agent/orchestrator.go` | Task orchestration |
| `pkg/agent/planner.go` | Task planning |
| `pkg/agent/aggregator.go` | Result aggregation |
| `pkg/tools/multi_agent.go` | Multi-agent tools |
| `pkg/agent/discovery/mdns.go` | mDNS discovery |
| `pkg/agent/discovery/consul.go` | Consul integration |

## Minimal Viable Implementation (MVP)

For immediate testing, implement only:

1. **AgentBus** - In-memory agent messaging
2. **SendAgentMessageTool** - Basic message sending
3. **DiscoverAgentsTool** - List available agents
4. **Controller config** - Mark agent as controller
5. **Static workers** - Pre-configured worker agents

### MVP Configuration

```json
{
  "agents": {
    "list": [
      {
        "id": "lead",
        "role": "controller",
        "multi_agent": {
          "enabled": true,
          "workers": ["dev1", "dev2"]
        }
      },
      {
        "id": "dev1",
        "role": "worker",
        "specialization": "backend"
      },
      {
        "id": "dev2",
        "role": "worker",
        "specialization": "frontend"
      }
    ]
  }
}
```

### MVP Tools Available to Controller

- `send_agent_message` - Send message to specific worker
- `discover_workers` - List available workers and capabilities
- `delegate_task` - Assign task to worker and wait for result

## Testing Strategy

### Unit Tests
```
pkg/agent/network/*_test.go
pkg/agent/controller_test.go
pkg/agent/worker_test.go
```

### Integration Tests
```
tests/multi_agent/communication_test.go
tests/multi_agent/orchestration_test.go
tests/multi_agent/discovery_test.go
```

### Manual Test Scenarios

1. **Basic Communication**
   ```
   User → Controller → Message to Worker → Response back
   ```

2. **Task Delegation**
   ```
   User → Controller → Delegate task → Worker executes → Result returned
   ```

3. **Parallel Tasks**
   ```
   User → Controller → Split task → [Worker A, Worker B] → Aggregate results
   ```

## Migration Strategy

1. **Backward Compatible**: Existing configs work unchanged
2. **Opt-in**: Multi-agent features require explicit configuration
3. **Graceful Degradation**: If network fails, fallback to single-agent mode

## Success Metrics

- Controller can send/receive messages to workers
- Tasks are distributed based on capabilities
- Workers report results back to controller
- Discovery finds agents within 5 seconds
- Message latency < 10ms (local)

