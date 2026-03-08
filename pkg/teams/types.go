// PicoClaw - Team management types
// License: MIT

package teams

import "time"

// Team represents a team of agents working together
type Team struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	TeamKey     string            `json:"team_key"` // Hashed key
	Status      string            `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	CreatedBy   string            `json:"created_by,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at,omitempty"`
	Settings    TeamSettings      `json:"settings"`
	Roles       map[string]*Role  `json:"roles"`
	Agents      []*Agent          `json:"agents"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Agent represents an agent instance within a team
type Agent struct {
	ID           string   `json:"id"`
	TeamID       string   `json:"team_id"`
	Role         string   `json:"role"`
	Status       string   `json:"status"`
	Capabilities []string `json:"capabilities"`
	ActiveTasks  int      `json:"active_tasks"`
	MaxTasks     int      `json:"max_tasks"`
	ConnectedAt  time.Time `json:"connected_at"`
	LastSeen     time.Time `json:"last_seen"`
	Address      string   `json:"address,omitempty"`
	Version      string   `json:"version,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Role represents a role configuration within a team
type Role struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Capabilities []string `json:"capabilities"`
	MaxTasks     int      `json:"max_tasks"`
	Tools        []string `json:"tools"`
	AllowedChannels []string `json:"allowed_channels,omitempty"`
	SystemPrompt string   `json:"system_prompt,omitempty"`
}

// TeamSettings contains configuration for the team
type TeamSettings struct {
	MaxAgents          int      `json:"max_agents"`
	AgentTimeout       int      `json:"agent_timeout"` // seconds
	RequireApproval    bool     `json:"require_approval"`
	AllowedCapabilities []string `json:"allowed_capabilities"`
	AutoAcceptAgents   bool     `json:"auto_accept_agents"`
	SharedMemoryEnabled bool    `json:"shared_memory_enabled"`
	MessageEncryption  string   `json:"message_encryption,omitempty"` // e.g., "aes-256-gcm"
}

// Team status constants
const (
	TeamStatusActive   = "active"
	TeamStatusPaused   = "paused"
	TeamStatusArchived = "archived"
)

// Agent status constants
const (
	AgentStatusOnline   = "online"
	AgentStatusOffline  = "offline"
	AgentStatusBusy     = "busy"
	AgentStatusEvicted  = "evicted"
	AgentStatusPending  = "pending"
)

// CreateTeamRequest is the request to create a new team
type CreateTeamRequest struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	CreatedBy   string       `json:"created_by,omitempty"`
	Settings    TeamSettings `json:"settings,omitempty"`
	Roles       []*Role      `json:"roles,omitempty"`
}

// UpdateTeamRequest is the request to update a team
type UpdateTeamRequest struct {
	Name        string       `json:"name,omitempty"`
	Description string       `json:"description,omitempty"`
	Status      string       `json:"status,omitempty"`
	Settings    TeamSettings `json:"settings,omitempty"`
}

// TeamSummary is a summary view of a team for listing
type TeamSummary struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	AgentCount  int       `json:"agent_count"`
	OnlineCount int       `json:"online_count"`
	MaxAgents   int       `json:"max_agents"`
	CreatedAt   time.Time `json:"created_at"`
}

// JoinTeamRequest is sent by an agent wanting to join a team
type JoinTeamRequest struct {
	AgentID      string   `json:"agent_id"`
	Role         string   `json:"role"`
	Capabilities []string `json:"capabilities,omitempty"`
	Version      string   `json:"version,omitempty"`
	Address      string   `json:"address,omitempty"`
}

// JoinTeamResponse is the response to a join request
type JoinTeamResponse struct {
	Success          bool   `json:"success"`
	Message          string `json:"message,omitempty"`
	SessionID        string `json:"session_id,omitempty"`
	TeamID           string `json:"team_id,omitempty"`
	AssignedRole     string `json:"assigned_role,omitempty"`
	RequiresApproval bool   `json:"requires_approval,omitempty"`
}

// DefaultTeamSettings returns the default team settings
func DefaultTeamSettings() TeamSettings {
	return TeamSettings{
		MaxAgents:           10,
		AgentTimeout:        60,
		RequireApproval:     false,
		AllowedCapabilities: []string{},
		AutoAcceptAgents:    true,
		SharedMemoryEnabled: true,
		MessageEncryption:   "aes-256-gcm",
	}
}

// DefaultRoles returns the default set of roles
func DefaultRoles() map[string]*Role {
	return map[string]*Role{
		"frontend": {
			ID:           "frontend",
			Name:         "Frontend Developer",
			Description:  "Specializes in UI/UX and client-side development",
			Capabilities: []string{"javascript", "typescript", "react", "vue", "html", "css"},
			MaxTasks:     3,
			Tools:        []string{"read_file", "write_file", "edit_file", "list_dir", "web_search"},
		},
		"backend": {
			ID:           "backend",
			Name:         "Backend Developer",
			Description:  "Specializes in server-side APIs and databases",
			Capabilities: []string{"go", "python", "rust", "nodejs", "sql", "redis", "grpc"},
			MaxTasks:     3,
			Tools:        []string{"read_file", "write_file", "edit_file", "list_dir", "exec", "web_search"},
		},
		"fullstack": {
			ID:           "fullstack",
			Name:         "Full Stack Developer",
			Description:  "Handles both frontend and backend tasks",
			Capabilities: []string{"javascript", "typescript", "react", "go", "python", "sql"},
			MaxTasks:     4,
			Tools:        []string{"read_file", "write_file", "edit_file", "list_dir", "exec", "web_search"},
		},
		"qa": {
			ID:           "qa",
			Name:         "QA Engineer",
			Description:  "Specializes in testing and quality assurance",
			Capabilities: []string{"testing", "cypress", "jest", "pytest", "go_test"},
			MaxTasks:     2,
			Tools:        []string{"read_file", "write_file", "edit_file", "list_dir", "exec"},
		},
		"devops": {
			ID:           "devops",
			Name:         "DevOps Engineer",
			Description:  "Specializes in infrastructure and deployment",
			Capabilities: []string{"kubernetes", "docker", "github_actions", "terraform", "helm"},
			MaxTasks:     2,
			Tools:        []string{"read_file", "write_file", "edit_file", "list_dir", "exec", "web_search"},
		},
	}
}

// GetAllowedCapabilities returns the union of all allowed capabilities
func (s *TeamSettings) GetAllowedCapabilities() map[string]bool {
	allowed := make(map[string]bool)
	for _, cap := range s.AllowedCapabilities {
		allowed[cap] = true
	}
	return allowed
}

// AgentInfo is the public information about an agent (for API responses)
type AgentInfo struct {
	ID           string    `json:"id"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	Capabilities []string  `json:"capabilities"`
	ActiveTasks  int       `json:"active_tasks"`
	MaxTasks     int       `json:"max_tasks"`
	LastSeen     time.Time `json:"last_seen"`
}
