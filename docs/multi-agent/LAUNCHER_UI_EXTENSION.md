# Extending PicoLauncher UI for Team Management

## Current Launcher UI Structure

```
┌─────────────────────────────────────────────────────────────────────┐
│  🦞 PicoClaw Config                                    [Theme] [EN]  │
├─────────────────────────────────────────────────────────────────────┤
│  Sidebar                    │ Content Area                           │
│  ─────────────────          │ ───────────────────                    │
│  Providers                  │ Panel: Models / Auth / Channels / Logs │
│  ├── Models                 │                                        │
│  └── Auth                   │ Raw JSON Editor                        │
│  Channels                   │                                        │
│  ├── Telegram               │                                        │
│  ├── Discord                │                                        │
│  └── ...                    │                                        │
│  ─────────                  │                                        │
│  Logs                       │                                        │
│  Raw JSON                   │                                        │
└─────────────────────────────────────────────────────────────────────┘
```

## Proposed Extension: Teams Section

```
┌─────────────────────────────────────────────────────────────────────┐
│  🦞 PicoClaw Config                                    [Theme] [EN]  │
├─────────────────────────────────────────────────────────────────────┤
│  Sidebar                    │ Content Area                           │
│  ─────────────────          │ ───────────────────                    │
│  Providers                  │ TEAMS DASHBOARD                        │
│  ├── Models                 │ ┌───────────────────────────────────┐  │
│  └── Auth                   │ │ Active Teams: 3                   │  │
│                             │ │ Connected Agents: 12              │  │
│  TEAMS (NEW)               │ └───────────────────────────────────┘  │
│  ├── Dashboard              │                                        │
│  ├── Dev Team Alpha        │ Team: Dev Team Alpha                   │
│  ├── Operations Team       │ ┌───────────────────────────────────┐  │
│  └── Beta Testers          │ │ Key: pk_team_xxx...*** [Rotate]   │  │
│                             │ │ Agents: 5/10                      │  │
│  Channels                   │ │ Status: 🟢 Active                 │  │
│  ├── Telegram               │ └───────────────────────────────────┘  │
│  └── ...                    │                                        │
│                             │ Connected Agents                       │
│  ─────────                  │ ┌───────────────────────────────────┐  │
│  Logs                       │ │ frontend-dev-01  🟢 Online        │  │
│  Raw JSON                   │ │ backend-dev-01   🟢 Online        │  │
│                             │ │ qa-engineer-01   🟡 Busy          │  │
│                             │ └───────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

## UI Components to Add

### 1. Sidebar Extension

**File:** `cmd/picoclaw-launcher/internal/ui/index.html` (sidebar section)

```html
<!-- Add after Auth section -->
<div class="sidebar-group" data-group="teams">
    <div class="sidebar-group-title" onclick="toggleGroup(this)">
        <span data-i18n="sidebar.teams">Teams</span> 
        <span class="arrow">&#9662;</span>
        <span class="badge-count" id="teamCount" style="display:none">0</span>
    </div>
    <div class="sidebar-items">
        <div class="sidebar-item" data-panel="panelTeamsDashboard" data-i18n="sidebar.teamsDashboard">
            Dashboard
        </div>
        <!-- Dynamic team items will be inserted here -->
        <div id="dynamicTeamItems"></div>
        <div class="sidebar-divider"></div>
        <div class="sidebar-item" data-panel="panelTeamsCreate" data-i18n="sidebar.createTeam">
            + Create Team
        </div>
    </div>
</div>
```

### 2. Teams Dashboard Panel

```html
<div class="content-panel" id="panelTeamsDashboard">
    <div class="panel-title" data-i18n="teams.dashboardTitle">Teams Dashboard</div>
    <div class="panel-desc" data-i18n="teams.dashboardDesc">
        Manage multi-agent teams and monitor connected workers.
    </div>
    
    <!-- Stats Cards -->
    <div class="stats-grid" id="teamsStatsGrid">
        <div class="stat-card">
            <div class="stat-value" id="statTotalTeams">0</div>
            <div class="stat-label" data-i18n="teams.totalTeams">Active Teams</div>
        </div>
        <div class="stat-card">
            <div class="stat-value" id="statTotalAgents">0</div>
            <div class="stat-label" data-i18n="teams.totalAgents">Connected Agents</div>
        </div>
        <div class="stat-card">
            <div class="stat-value" id="statOnlineAgents">0</div>
            <div class="stat-label" data-i18n="teams.onlineAgents">Online</div>
        </div>
        <div class="stat-card">
            <div class="stat-value" id="statActiveTasks">0</div>
            <div class="stat-label" data-i18n="teams.activeTasks">Active Tasks</div>
        </div>
    </div>
    
    <!-- Teams List -->
    <div class="section-title" data-i18n="teams.yourTeams">Your Teams</div>
    <div class="teams-grid" id="teamsGrid">
        <!-- Team cards will be rendered here -->
    </div>
</div>
```

### 3. Team Detail Panel

```html
<div class="content-panel" id="panelTeamDetail">
    <div class="panel-header-with-actions">
        <div>
            <div class="panel-title" id="teamDetailName">Team Name</div>
            <div class="panel-desc" id="teamDetailDesc">Team description</div>
        </div>
        <div class="panel-actions">
            <button class="btn btn-sm" onclick="editCurrentTeam()" data-i18n="edit">Edit</button>
            <button class="btn btn-sm btn-danger" onclick="deleteCurrentTeam()" data-i18n="delete">Delete</button>
        </div>
    </div>
    
    <!-- Team Key Section -->
    <div class="team-key-section">
        <div class="section-title" data-i18n="teams.teamKey">Team Key</div>
        <div class="key-display">
            <code id="teamKeyDisplay" class="key-masked">pk_team_••••••••</code>
            <button class="btn btn-sm" onclick="showTeamKey()" data-i18n="teams.showKey">Show</button>
            <button class="btn btn-sm" onclick="copyTeamKey()" data-i18n="teams.copyKey">Copy</button>
            <button class="btn btn-sm btn-warning" onclick="rotateTeamKey()" data-i18n="teams.rotateKey">Rotate</button>
        </div>
        <div class="key-warning" data-i18n="teams.keyWarning">
            ⚠️ Keep this key secure. Anyone with this key can join your team.
        </div>
    </div>
    
    <!-- Agents List -->
    <div class="section-title" data-i18n="teams.connectedAgents">Connected Agents</div>
    <div class="agents-table-container">
        <table class="agents-table" id="agentsTable">
            <thead>
                <tr>
                    <th data-i18n="teams.agentId">Agent ID</th>
                    <th data-i18n="teams.role">Role</th>
                    <th data-i18n="teams.status">Status</th>
                    <th data-i18n="teams.tasks">Tasks</th>
                    <th data-i18n="teams.lastSeen">Last Seen</th>
                    <th data-i18n="teams.actions">Actions</th>
                </tr>
            </thead>
            <tbody id="agentsTableBody">
                <!-- Agent rows rendered here -->
            </tbody>
        </table>
    </div>
    
    <!-- Role Definitions -->
    <div class="section-title" data-i18n="teams.roles">Role Definitions</div>
    <div class="roles-list" id="rolesList">
        <!-- Role cards rendered here -->
    </div>
</div>
```

### 4. Create Team Panel

```html
<div class="content-panel" id="panelTeamsCreate">
    <div class="panel-title" data-i18n="teams.createTitle">Create New Team</div>
    <div class="panel-desc" data-i18n="teams.createDesc">
        Create a team to organize and manage multiple agents.
    </div>
    
    <form class="team-form" id="createTeamForm" onsubmit="submitCreateTeam(event)">
        <div class="form-group">
            <label class="form-label" data-i18n="teams.teamName">Team Name</label>
            <input type="text" class="form-input" id="newTeamName" required
                   placeholder="e.g., Development Team Alpha">
        </div>
        
        <div class="form-group">
            <label class="form-label" data-i18n="teams.teamDescription">Description</label>
            <input type="text" class="form-input" id="newTeamDesc"
                   placeholder="What does this team do?">
        </div>
        
        <div class="form-group">
            <label class="form-label" data-i18n="teams.maxAgents">Max Agents</label>
            <input type="number" class="form-input form-input-number" id="newTeamMaxAgents"
                   value="10" min="1" max="100">
        </div>
        
        <div class="form-group">
            <label class="toggle-row">
                <div class="toggle" id="toggleAutoAccept" onclick="toggleThis(this)"></div>
                <span class="toggle-label" data-i18n="teams.autoAccept">
                    Auto-accept new agents
                </span>
            </label>
            <div class="form-hint" data-i18n="teams.autoAcceptHint">
                If disabled, you must manually approve each agent joining the team.
            </div>
        </div>
        
        <div class="form-actions">
            <button type="button" class="btn" onclick="showPanel('panelTeamsDashboard')"
                    data-i18n="cancel">Cancel</button>
            <button type="submit" class="btn btn-primary" data-i18n="teams.createTeamBtn">
                Create Team
            </button>
        </div>
    </form>
</div>
```

## CSS Styles to Add

```css
/* Teams-specific styles */
.stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 16px;
    margin-bottom: 24px;
}

.stat-card {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 20px;
    text-align: center;
}

.stat-value {
    font-size: 32px;
    font-weight: 700;
    color: var(--accent);
    margin-bottom: 4px;
}

.stat-label {
    font-size: 12px;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
}

.teams-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 16px;
}

.team-card {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 20px;
    cursor: pointer;
    transition: all var(--transition);
}

.team-card:hover {
    border-color: var(--text-muted);
    transform: translateY(-2px);
}

.team-card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
}

.team-name {
    font-size: 16px;
    font-weight: 600;
}

.team-status {
    font-size: 11px;
    padding: 4px 10px;
    border-radius: 20px;
    text-transform: uppercase;
    font-weight: 600;
}

.status-active {
    background: var(--success-bg);
    color: var(--success);
    border: 1px solid rgba(34, 197, 94, 0.2);
}

.team-meta {
    font-size: 12px;
    color: var(--text-muted);
    margin-bottom: 12px;
}

.team-agents {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
}

.agent-avatars {
    display: flex;
}

.agent-avatar {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: var(--accent);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 12px;
    font-weight: 600;
    margin-left: -8px;
    border: 2px solid var(--bg-secondary);
}

.agent-avatar:first-child {
    margin-left: 0;
}

/* Key display */
.team-key-section {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 20px;
    margin-bottom: 24px;
}

.key-display {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
    margin-bottom: 12px;
}

.key-display code {
    font-family: 'JetBrains Mono', monospace;
    font-size: 14px;
    background: var(--bg-editor);
    padding: 10px 16px;
    border-radius: 8px;
    border: 1px solid var(--border);
    flex: 1;
    min-width: 200px;
}

.key-warning {
    font-size: 12px;
    color: var(--warning);
    background: var(--warning-bg);
    padding: 10px 14px;
    border-radius: 8px;
    border: 1px solid rgba(245, 158, 11, 0.2);
}

/* Agents table */
.agents-table-container {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    overflow: hidden;
    margin-bottom: 24px;
}

.agents-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
}

.agents-table th {
    text-align: left;
    padding: 12px 16px;
    font-weight: 600;
    color: var(--text-secondary);
    border-bottom: 1px solid var(--border);
    background: var(--bg-elevated);
}

.agents-table td {
    padding: 12px 16px;
    border-bottom: 1px solid var(--border);
    color: var(--text-primary);
}

.agents-table tr:last-child td {
    border-bottom: none;
}

.agent-status {
    display: inline-flex;
    align-items: center;
    gap: 6px;
}

.status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
}

.status-dot.online { background: var(--success); }
.status-dot.busy { background: var(--warning); }
.status-dot.offline { background: var(--text-muted); }

/* Role cards */
.roles-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: 16px;
}

.role-card {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 16px;
}

.role-name {
    font-size: 14px;
    font-weight: 600;
    margin-bottom: 8px;
}

.role-capabilities {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
}

.capability-tag {
    font-size: 11px;
    padding: 3px 8px;
    background: var(--bg-elevated);
    border-radius: 4px;
    color: var(--text-secondary);
}

/* Badge count in sidebar */
.badge-count {
    font-size: 10px;
    padding: 2px 6px;
    background: var(--accent);
    color: white;
    border-radius: 10px;
    margin-left: auto;
}
```

## JavaScript Functions to Add

```javascript
// ── Teams State ─────────────────────────────────────
let teamsData = [];
let currentTeamId = null;

// ── Teams API Integration ───────────────────────────

async function loadTeams() {
    try {
        const response = await fetch('/api/teams');
        teamsData = await response.json();
        renderTeamsInSidebar();
        renderTeamsDashboard();
    } catch (err) {
        showToast('Failed to load teams: ' + err.message, 'error');
    }
}

function renderTeamsInSidebar() {
    const container = document.getElementById('dynamicTeamItems');
    const countBadge = document.getElementById('teamCount');
    
    if (teamsData.length === 0) {
        container.innerHTML = '<div class="sidebar-item empty">No teams yet</div>';
        countBadge.style.display = 'none';
        return;
    }
    
    countBadge.textContent = teamsData.length;
    countBadge.style.display = 'inline';
    
    container.innerHTML = teamsData.map(team => `
        <div class="sidebar-item" data-panel="panelTeamDetail" 
             data-team-id="${team.id}" onclick="selectTeam('${team.id}')">
            ${escapeHtml(team.name)}
        </div>
    `).join('');
}

function renderTeamsDashboard() {
    // Update stats
    const totalAgents = teamsData.reduce((sum, t) => sum + (t.agents?.length || 0), 0);
    const onlineAgents = teamsData.reduce((sum, t) => 
        sum + (t.agents?.filter(a => a.status === 'online').length || 0), 0);
    
    document.getElementById('statTotalTeams').textContent = teamsData.length;
    document.getElementById('statTotalAgents').textContent = totalAgents;
    document.getElementById('statOnlineAgents').textContent = onlineAgents;
    
    // Render team cards
    const grid = document.getElementById('teamsGrid');
    if (teamsData.length === 0) {
        grid.innerHTML = `
            <div class="empty-state">
                <p data-i18n="teams.noTeams">No teams created yet.</p>
                <button class="btn btn-primary" onclick="showPanel('panelTeamsCreate')">
                    ${t('teams.createFirst')}
                </button>
            </div>
        `;
        return;
    }
    
    grid.innerHTML = teamsData.map(team => `
        <div class="team-card" onclick="selectTeam('${team.id}')">
            <div class="team-card-header">
                <span class="team-name">${escapeHtml(team.name)}</span>
                <span class="team-status status-${team.status}">${team.status}</span>
            </div>
            <div class="team-meta">${escapeHtml(team.description || '')}</div>
            <div class="team-agents">
                <div class="agent-avatars">
                    ${(team.agents || []).slice(0, 3).map(a => `
                        <div class="agent-avatar" title="${a.id}">${a.id.charAt(0).toUpperCase()}</div>
                    `).join('')}
                </div>
                <span>${team.agents?.length || 0} agents</span>
            </div>
        </div>
    `).join('');
}

async function selectTeam(teamId) {
    currentTeamId = teamId;
    const team = teamsData.find(t => t.id === teamId);
    if (!team) return;
    
    // Update UI
    document.getElementById('teamDetailName').textContent = team.name;
    document.getElementById('teamDetailDesc').textContent = team.description || '';
    document.getElementById('teamKeyDisplay').textContent = maskKey(team.team_key);
    
    // Render agents table
    renderAgentsTable(team.agents || []);
    
    // Render roles
    renderRoles(team.roles || {});
    
    showPanel('panelTeamDetail');
}

function renderAgentsTable(agents) {
    const tbody = document.getElementById('agentsTableBody');
    if (agents.length === 0) {
        tbody.innerHTML = `<tr><td colspan="6" class="empty-cell">${t('teams.noAgents')}</td></tr>`;
        return;
    }
    
    tbody.innerHTML = agents.map(agent => `
        <tr>
            <td>${escapeHtml(agent.id)}</td>
            <td>${escapeHtml(agent.role)}</td>
            <td>
                <span class="agent-status">
                    <span class="status-dot ${agent.status}"></span>
                    ${agent.status}
                </span>
            </td>
            <td>${agent.active_tasks || 0}/${agent.max_tasks || '-'}</td>
            <td>${formatTime(agent.last_seen)}</td>
            <td>
                <button class="btn btn-sm btn-danger" onclick="evictAgent('${agent.id}')">
                    ${t('teams.evict')}
                </button>
            </td>
        </tr>
    `).join('');
}

async function submitCreateTeam(event) {
    event.preventDefault();
    
    const teamData = {
        name: document.getElementById('newTeamName').value,
        description: document.getElementById('newTeamDesc').value,
        max_agents: parseInt(document.getElementById('newTeamMaxAgents').value),
        auto_accept: document.getElementById('toggleAutoAccept').classList.contains('on')
    };
    
    try {
        const response = await fetch('/api/teams', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(teamData)
        });
        
        if (!response.ok) throw new Error('Failed to create team');
        
        const result = await response.json();
        showToast(t('teams.created'), 'success');
        
        // Reload teams and show the new team
        await loadTeams();
        selectTeam(result.id);
    } catch (err) {
        showToast(err.message, 'error');
    }
}

async function rotateTeamKey() {
    if (!currentTeamId) return;
    
    if (!confirm(t('teams.rotateConfirm'))) return;
    
    try {
        const response = await fetch(`/api/teams/${currentTeamId}/rotate-key`, {
            method: 'POST'
        });
        
        if (!response.ok) throw new Error('Failed to rotate key');
        
        const result = await response.json();
        document.getElementById('teamKeyDisplay').textContent = maskKey(result.team_key);
        showToast(t('teams.keyRotated'), 'success');
    } catch (err) {
        showToast(err.message, 'error');
    }
}

async function evictAgent(agentId) {
    if (!currentTeamId) return;
    
    if (!confirm(t('teams.evictConfirm', { agent: agentId }))) return;
    
    try {
        await fetch(`/api/teams/${currentTeamId}/agents/${agentId}/evict`, {
            method: 'POST'
        });
        
        showToast(t('teams.agentEvicted'), 'success');
        loadTeams(); // Refresh
    } catch (err) {
        showToast(err.message, 'error');
    }
}

// Helper functions
function maskKey(key) {
    if (!key) return '';
    if (key.length < 20) return '•'.repeat(key.length);
    return key.substring(0, 10) + '•'.repeat(key.length - 15) + key.slice(-5);
}

function showTeamKey() {
    const team = teamsData.find(t => t.id === currentTeamId);
    if (team) {
        document.getElementById('teamKeyDisplay').textContent = team.team_key;
    }
}

function copyTeamKey() {
    const team = teamsData.find(t => t.id === currentTeamId);
    if (team && team.team_key) {
        navigator.clipboard.writeText(team.team_key);
        showToast(t('teams.keyCopied'), 'success');
    }
}

function formatTime(timestamp) {
    if (!timestamp) return '-';
    const date = new Date(timestamp);
    const now = new Date();
    const diff = Math.floor((now - date) / 1000);
    
    if (diff < 60) return 'just now';
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
    if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
    return date.toLocaleDateString();
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
```

## API Endpoints to Add

```go
// File: cmd/picoclaw-launcher/internal/server/teams_api.go

package server

import (
    "encoding/json"
    "net/http"
    "github.com/sipeed/picoclaw/pkg/teams"
)

func RegisterTeamsAPI(mux *http.ServeMux, teamService *teams.Service) {
    // GET /api/teams - List all teams
    mux.HandleFunc("GET /api/teams", func(w http.ResponseWriter, r *http.Request) {
        teams, err := teamService.ListTeams()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        json.NewEncoder(w).Encode(teams)
    })
    
    // POST /api/teams - Create new team
    mux.HandleFunc("POST /api/teams", func(w http.ResponseWriter, r *http.Request) {
        var req teams.CreateTeamRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
        team, err := teamService.CreateTeam(req)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(team)
    })
    
    // GET /api/teams/{id} - Get team details
    mux.HandleFunc("GET /api/teams/{id}", func(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        team, err := teamService.GetTeam(id)
        if err != nil {
            http.Error(w, err.Error(), http.StatusNotFound)
            return
        }
        json.NewEncoder(w).Encode(team)
    })
    
    // POST /api/teams/{id}/rotate-key - Rotate team key
    mux.HandleFunc("POST /api/teams/{id}/rotate-key", func(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        newKey, err := teamService.RotateKey(id)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        json.NewEncoder(w).Encode(map[string]string{"team_key": newKey})
    })
    
    // POST /api/teams/{id}/agents/{agentId}/evict - Evict agent
    mux.HandleFunc("POST /api/teams/{id}/agents/{agentId}/evict", func(w http.ResponseWriter, r *http.Request) {
        teamID := r.PathValue("id")
        agentID := r.PathValue("agentId")
        
        if err := teamService.EvictAgent(teamID, agentID); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    })
}
```

## i18n Strings to Add

```javascript
const i18nData = {
    en: {
        // ... existing strings ...
        
        // Teams
        'sidebar.teams': 'Teams',
        'sidebar.teamsDashboard': 'Dashboard',
        'sidebar.createTeam': '+ Create Team',
        'teams.dashboardTitle': 'Teams Dashboard',
        'teams.dashboardDesc': 'Manage multi-agent teams and monitor connected workers.',
        'teams.totalTeams': 'Active Teams',
        'teams.totalAgents': 'Connected Agents',
        'teams.onlineAgents': 'Online',
        'teams.activeTasks': 'Active Tasks',
        'teams.yourTeams': 'Your Teams',
        'teams.noTeams': 'No teams created yet.',
        'teams.createFirst': 'Create First Team',
        'teams.teamKey': 'Team Key',
        'teams.showKey': 'Show',
        'teams.copyKey': 'Copy',
        'teams.rotateKey': 'Rotate',
        'teams.keyWarning': 'Keep this key secure. Anyone with this key can join your team.',
        'teams.rotateConfirm': 'Rotate team key? All agents will need to reconnect with the new key.',
        'teams.keyRotated': 'Team key rotated successfully',
        'teams.keyCopied': 'Key copied to clipboard',
        'teams.connectedAgents': 'Connected Agents',
        'teams.agentId': 'Agent ID',
        'teams.role': 'Role',
        'teams.status': 'Status',
        'teams.tasks': 'Tasks',
        'teams.lastSeen': 'Last Seen',
        'teams.actions': 'Actions',
        'teams.evict': 'Evict',
        'teams.evictConfirm': 'Evict agent "{agent}" from the team?',
        'teams.agentEvicted': 'Agent evicted successfully',
        'teams.noAgents': 'No agents connected to this team yet.',
        'teams.roles': 'Role Definitions',
        'teams.createTitle': 'Create New Team',
        'teams.createDesc': 'Create a team to organize and manage multiple agents.',
        'teams.teamName': 'Team Name',
        'teams.teamDescription': 'Description',
        'teams.maxAgents': 'Max Agents',
        'teams.autoAccept': 'Auto-accept new agents',
        'teams.autoAcceptHint': 'If disabled, you must manually approve each agent joining the team.',
        'teams.createTeamBtn': 'Create Team',
        'teams.created': 'Team created successfully',
    },
    zh: {
        // ... existing Chinese strings ...
        
        // Teams (Chinese)
        'sidebar.teams': '团队',
        'sidebar.teamsDashboard': '概览',
        'sidebar.createTeam': '+ 创建团队',
        'teams.dashboardTitle': '团队概览',
        'teams.dashboardDesc': '管理多智能体团队并监控已连接的工人。',
        'teams.totalTeams': '活跃团队',
        'teams.totalAgents': '已连接智能体',
        'teams.onlineAgents': '在线',
        'teams.activeTasks': '活跃任务',
        // ... more Chinese translations ...
    }
};
```

## Summary

This extension leverages the existing PicoLauncher UI by:

1. **Adding a "Teams" section** to the sidebar
2. **Creating 3 new panels**: Dashboard, Team Detail, Create Team
3. **Reusing existing components**: Cards, tables, forms, toggles
4. **Following existing patterns**: API calls, i18n, theming
5. **Maintaining consistency**: Same styling, same UX patterns

Benefits:
- ✅ No separate UI to build
- ✅ Consistent user experience
- ✅ Reuses existing auth/session
- ✅ Single codebase to maintain
- ✅ Familiar interface for users

