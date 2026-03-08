// PicoClaw - Team management service
// License: MIT

package teams

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sipeed/picoclaw/pkg/fileutil"
	"github.com/sipeed/picoclaw/pkg/logger"
)

// Service manages teams and their agents
type Service struct {
	storagePath string
	teams       map[string]*Team
	agents      map[string]*Agent // agent_id -> agent (across all teams)
	sessions    map[string]*AgentSession
	mu          sync.RWMutex
}

// AgentSession represents an authenticated agent session
type AgentSession struct {
	SessionID   string
	AgentID     string
	TeamID      string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	LastActivity time.Time
}

// NewService creates a new team service
func NewService(storagePath string) (*Service, error) {
	s := &Service{
		storagePath: storagePath,
		teams:       make(map[string]*Team),
		agents:      make(map[string]*Agent),
		sessions:    make(map[string]*AgentSession),
	}

	// Load existing teams
	if err := s.loadTeams(); err != nil {
		return nil, fmt.Errorf("failed to load teams: %w", err)
	}

	// Start cleanup goroutine
	go s.cleanupLoop()

	return s, nil
}

// CreateTeam creates a new team
func (s *Service) CreateTeam(req CreateTeamRequest) (*Team, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate unique ID
	id := generateTeamID(req.Name)

	// Generate team key
	teamKey, err := generateTeamKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate team key: %w", err)
	}

	// Use default settings if not provided
	settings := req.Settings
	if settings.MaxAgents == 0 {
		settings = DefaultTeamSettings()
	}

	// Use default roles if not provided
	roles := req.Roles
	if len(roles) == 0 {
		roles = roleMapToSlice(DefaultRoles())
	}
	roleMap := make(map[string]*Role)
	for _, r := range roles {
		roleMap[r.ID] = r
	}

	team := &Team{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		TeamKey:     hashTeamKey(teamKey), // Store hashed version
		Status:      TeamStatusActive,
		CreatedAt:   time.Now().UTC(),
		CreatedBy:   req.CreatedBy,
		Settings:    settings,
		Roles:       roleMap,
		Agents:      make([]*Agent, 0),
		Metadata: map[string]string{
			"raw_key": teamKey, // Store raw key temporarily for display
		},
	}

	s.teams[id] = team

	// Save to storage
	if err := s.saveTeam(team); err != nil {
		delete(s.teams, id)
		return nil, fmt.Errorf("failed to save team: %w", err)
	}

	logger.InfoCF("teams", "Created new team",
		map[string]any{
			"team_id":   id,
			"team_name": req.Name,
		})

	// Return team with raw key for initial display
	return team, nil
}

// GetTeam retrieves a team by ID
func (s *Service) GetTeam(id string) (*Team, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	team, exists := s.teams[id]
	if !exists {
		return nil, fmt.Errorf("team not found: %s", id)
	}

	return team, nil
}

// GetTeamByKey retrieves a team by its raw team key
func (s *Service) GetTeamByKey(rawKey string) (*Team, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hashedKey := hashTeamKey(rawKey)
	for _, team := range s.teams {
		if team.TeamKey == hashedKey {
			return team, nil
		}
	}

	return nil, fmt.Errorf("invalid team key")
}

// ListTeams returns all teams
func (s *Service) ListTeams() ([]*TeamSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summaries := make([]*TeamSummary, 0, len(s.teams))
	for _, team := range s.teams {
		onlineCount := 0
		for _, agent := range team.Agents {
			if agent.Status == AgentStatusOnline || agent.Status == AgentStatusBusy {
				onlineCount++
			}
		}

		summaries = append(summaries, &TeamSummary{
			ID:            team.ID,
			Name:          team.Name,
			Description:   team.Description,
			Status:        team.Status,
			AgentCount:    len(team.Agents),
			OnlineCount:   onlineCount,
			MaxAgents:     team.Settings.MaxAgents,
			CreatedAt:     team.CreatedAt,
		})
	}

	return summaries, nil
}

// UpdateTeam updates a team's information
func (s *Service) UpdateTeam(id string, req UpdateTeamRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	team, exists := s.teams[id]
	if !exists {
		return fmt.Errorf("team not found: %s", id)
	}

	if req.Name != "" {
		team.Name = req.Name
	}
	if req.Description != "" {
		team.Description = req.Description
	}
	if req.Status != "" {
		team.Status = req.Status
	}
	if req.Settings.MaxAgents != 0 {
		team.Settings = req.Settings
	}

	if err := s.saveTeam(team); err != nil {
		return fmt.Errorf("failed to save team: %w", err)
	}

	logger.InfoCF("teams", "Updated team",
		map[string]any{"team_id": id})

	return nil
}

// DeleteTeam deletes a team
func (s *Service) DeleteTeam(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	team, exists := s.teams[id]
	if !exists {
		return fmt.Errorf("team not found: %s", id)
	}

	// Disconnect all agents
	for _, agent := range team.Agents {
		delete(s.agents, agent.ID)
	}

	delete(s.teams, id)

	// Remove from storage
	teamFile := filepath.Join(s.storagePath, id+".json")
	os.Remove(teamFile)

	logger.InfoCF("teams", "Deleted team",
		map[string]any{"team_id": id})

	return nil
}

// RotateTeamKey generates a new team key
func (s *Service) RotateTeamKey(id string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	team, exists := s.teams[id]
	if !exists {
		return "", fmt.Errorf("team not found: %s", id)
	}

	newKey, err := generateTeamKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}

	team.TeamKey = hashTeamKey(newKey)
	team.Metadata["raw_key"] = newKey
	team.Metadata["last_rotation"] = time.Now().UTC().Format(time.RFC3339)

	// Invalidate all existing sessions
	for sessionID, session := range s.sessions {
		if session.TeamID == id {
			delete(s.sessions, sessionID)
		}
	}

	// Mark all agents as needing reconnection
	for _, agent := range team.Agents {
		agent.Status = AgentStatusOffline
	}

	if err := s.saveTeam(team); err != nil {
		return "", fmt.Errorf("failed to save team: %w", err)
	}

	logger.InfoCF("teams", "Rotated team key",
		map[string]any{"team_id": id})

	return newKey, nil
}

// JoinTeam handles an agent joining a team
func (s *Service) JoinTeam(teamID string, req JoinTeamRequest, rawTeamKey string) (*JoinTeamResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	team, exists := s.teams[teamID]
	if !exists {
		return &JoinTeamResponse{Success: false, Message: "Team not found"}, nil
	}

	if team.Status != TeamStatusActive {
		return &JoinTeamResponse{Success: false, Message: "Team is not active"}, nil
	}

	// Verify team key
	if hashTeamKey(rawTeamKey) != team.TeamKey {
		return &JoinTeamResponse{Success: false, Message: "Invalid team key"}, nil
	}

	// Check if agent already exists
	existingAgent := s.findAgentInTeam(team, req.AgentID)
	if existingAgent != nil {
		// Update existing agent
		existingAgent.Status = AgentStatusOnline
		existingAgent.LastSeen = time.Now().UTC()
		existingAgent.Address = req.Address
		existingAgent.Version = req.Version

		sessionID := generateSessionID()
		s.sessions[sessionID] = &AgentSession{
			SessionID:    sessionID,
			AgentID:      req.AgentID,
			TeamID:       teamID,
			CreatedAt:    time.Now().UTC(),
			ExpiresAt:    time.Now().UTC().Add(24 * time.Hour),
			LastActivity: time.Now().UTC(),
		}

		return &JoinTeamResponse{
			Success:   true,
			SessionID: sessionID,
		}, nil
	}

	// Check max agents
	if len(team.Agents) >= team.Settings.MaxAgents {
		return &JoinTeamResponse{Success: false, Message: "Team is full"}, nil
	}

	// Validate role
	role, roleExists := team.Roles[req.Role]
	if !roleExists {
		return &JoinTeamResponse{Success: false, Message: "Invalid role: " + req.Role}, nil
	}

	// Check if approval is required
	if team.Settings.RequireApproval && !team.Settings.AutoAcceptAgents {
		return &JoinTeamResponse{
			Success:          false,
			Message:          "Waiting for approval",
			RequiresApproval: true,
		}, nil
	}

	// Create new agent
	agent := &Agent{
		ID:           req.AgentID,
		TeamID:       teamID,
		Role:         req.Role,
		Status:       AgentStatusOnline,
		Capabilities: req.Capabilities,
		ActiveTasks:  0,
		MaxTasks:     role.MaxTasks,
		ConnectedAt:  time.Now().UTC(),
		LastSeen:     time.Now().UTC(),
		Address:      req.Address,
		Version:      req.Version,
	}

	team.Agents = append(team.Agents, agent)
	s.agents[req.AgentID] = agent

	sessionID := generateSessionID()
	s.sessions[sessionID] = &AgentSession{
		SessionID:    sessionID,
		AgentID:      req.AgentID,
		TeamID:       teamID,
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(24 * time.Hour),
		LastActivity: time.Now().UTC(),
	}

	if err := s.saveTeam(team); err != nil {
		return nil, fmt.Errorf("failed to save team: %w", err)
	}

	logger.InfoCF("teams", "Agent joined team",
		map[string]any{
			"team_id":  teamID,
			"agent_id": req.AgentID,
			"role":     req.Role,
		})

	return &JoinTeamResponse{
		Success:   true,
		SessionID: sessionID,
	}, nil
}

// EvictAgent removes an agent from a team
func (s *Service) EvictAgent(teamID, agentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	team, exists := s.teams[teamID]
	if !exists {
		return fmt.Errorf("team not found: %s", teamID)
	}

	// Find and remove agent
	for i, agent := range team.Agents {
		if agent.ID == agentID {
			// Remove from team
			team.Agents = append(team.Agents[:i], team.Agents[i+1:]...)
			
			// Update agent status
			agent.Status = AgentStatusEvicted
			
			// Remove from global agents map
			delete(s.agents, agentID)
			
			// Invalidate sessions
			for sessionID, session := range s.sessions {
				if session.AgentID == agentID {
					delete(s.sessions, sessionID)
				}
			}

			if err := s.saveTeam(team); err != nil {
				return fmt.Errorf("failed to save team: %w", err)
			}

			logger.InfoCF("teams", "Evicted agent from team",
				map[string]any{
					"team_id":  teamID,
					"agent_id": agentID,
				})

			return nil
		}
	}

	return fmt.Errorf("agent not found: %s", agentID)
}

// WipeAgent removes team key from an agent and evicts it
func (s *Service) WipeAgent(teamID, agentID string) error {
	// First evict
	if err := s.EvictAgent(teamID, agentID); err != nil {
		return err
	}

	// TODO: Send wipe command to agent if connected

	logger.InfoCF("teams", "Wiped agent",
		map[string]any{
			"team_id":  teamID,
			"agent_id": agentID,
		})

	return nil
}

// UpdateAgentHeartbeat updates an agent's last seen time
func (s *Service) UpdateAgentHeartbeat(teamID, agentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	team, exists := s.teams[teamID]
	if !exists {
		return fmt.Errorf("team not found: %s", teamID)
	}

	agent := s.findAgentInTeam(team, agentID)
	if agent == nil {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	agent.LastSeen = time.Now().UTC()
	if agent.Status == AgentStatusOffline {
		agent.Status = AgentStatusOnline
	}

	return nil
}

// GetAgentSession validates a session ID and returns the session
func (s *Service) GetAgentSession(sessionID string) (*AgentSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// Helper methods

func (s *Service) findAgentInTeam(team *Team, agentID string) *Agent {
	for _, agent := range team.Agents {
		if agent.ID == agentID {
			return agent
		}
	}
	return nil
}

func (s *Service) loadTeams() error {
	if _, err := os.Stat(s.storagePath); os.IsNotExist(err) {
		os.MkdirAll(s.storagePath, 0755)
		return nil
	}

	files, err := os.ReadDir(s.storagePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		path := filepath.Join(s.storagePath, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			logger.WarnCF("teams", "Failed to read team file",
				map[string]any{"file": file.Name(), "error": err})
			continue
		}

		var team Team
		if err := json.Unmarshal(data, &team); err != nil {
			logger.WarnCF("teams", "Failed to parse team file",
				map[string]any{"file": file.Name(), "error": err})
			continue
		}

		s.teams[team.ID] = &team
		for _, agent := range team.Agents {
			s.agents[agent.ID] = agent
		}
	}

	logger.InfoCF("teams", "Loaded teams",
		map[string]any{"count": len(s.teams)})

	return nil
}

func (s *Service) saveTeam(team *Team) error {
	if err := os.MkdirAll(s.storagePath, 0755); err != nil {
		return err
	}

	path := filepath.Join(s.storagePath, team.ID+".json")
	data, err := json.MarshalIndent(team, "", "  ")
	if err != nil {
		return err
	}

	return fileutil.WriteFileAtomic(path, data, 0600)
}

func (s *Service) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanup()
	}
}

func (s *Service) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Clean up expired sessions
	for id, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, id)
		}
	}

	// Mark offline agents
	for _, team := range s.teams {
		for _, agent := range team.Agents {
			if agent.Status == AgentStatusOnline || agent.Status == AgentStatusBusy {
				if now.Sub(agent.LastSeen) > time.Duration(team.Settings.AgentTimeout)*time.Second {
					agent.Status = AgentStatusOffline
				}
			}
		}
	}
}

// Utility functions

func generateTeamID(name string) string {
	// Create a URL-friendly ID from the name
	prefix := ""
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			prefix += string(c)
		} else if c == ' ' || c == '-' {
			prefix += "-"
		}
	}
	if len(prefix) > 20 {
		prefix = prefix[:20]
	}
	if prefix == "" {
		prefix = "team"
	}
	return fmt.Sprintf("%s-%s", prefix, uuid.New().String()[:8])
}

func generateTeamKey() (string, error) {
	// Generate a secure random 256-bit key
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return "pk_team_" + base64.URLEncoding.EncodeToString(b), nil
}

func hashTeamKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return base64.URLEncoding.EncodeToString(hash[:])
}

func generateSessionID() string {
	return "sess_" + uuid.New().String()
}

func roleMapToSlice(roles map[string]*Role) []*Role {
	result := make([]*Role, 0, len(roles))
	for _, r := range roles {
		result = append(result, r)
	}
	return result
}
