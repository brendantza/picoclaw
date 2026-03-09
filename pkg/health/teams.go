// PicoClaw - Team API registration for gateway
// License: MIT

package health

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sipeed/picoclaw/pkg/teams"
)

// RegisterTeamsAPI registers team management API endpoints on the given mux.
// This allows the gateway to handle team join/heartbeat requests from agents.
func RegisterTeamsAPI(mux *http.ServeMux, teamService *teams.Service) {
	// Agent-facing endpoints (for joining and heartbeats)
	mux.HandleFunc("POST /api/teams/join", func(w http.ResponseWriter, r *http.Request) {
		handleJoinTeam(w, r, teamService)
	})

	mux.HandleFunc("POST /api/teams/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		handleAgentHeartbeat(w, r, teamService)
	})

	// Admin endpoints (for team management)
	mux.HandleFunc("GET /api/teams", func(w http.ResponseWriter, r *http.Request) {
		handleListTeams(w, r, teamService)
	})

	mux.HandleFunc("POST /api/teams", func(w http.ResponseWriter, r *http.Request) {
		handleCreateTeam(w, r, teamService)
	})

	mux.HandleFunc("GET /api/teams/{id}", func(w http.ResponseWriter, r *http.Request) {
		handleGetTeam(w, r, teamService)
	})

	mux.HandleFunc("PUT /api/teams/{id}", func(w http.ResponseWriter, r *http.Request) {
		handleUpdateTeam(w, r, teamService)
	})

	mux.HandleFunc("DELETE /api/teams/{id}", func(w http.ResponseWriter, r *http.Request) {
		handleDeleteTeam(w, r, teamService)
	})

	mux.HandleFunc("POST /api/teams/{id}/rotate-key", func(w http.ResponseWriter, r *http.Request) {
		handleRotateKey(w, r, teamService)
	})

	mux.HandleFunc("GET /api/teams/{id}/agents", func(w http.ResponseWriter, r *http.Request) {
		handleListAgents(w, r, teamService)
	})

	mux.HandleFunc("POST /api/teams/{id}/agents/{agentId}/evict", func(w http.ResponseWriter, r *http.Request) {
		handleEvictAgent(w, r, teamService)
	})

	mux.HandleFunc("POST /api/teams/{id}/agents/{agentId}/wipe", func(w http.ResponseWriter, r *http.Request) {
		handleWipeAgent(w, r, teamService)
	})

	// Task management endpoints
	mux.HandleFunc("POST /api/teams/{id}/tasks", func(w http.ResponseWriter, r *http.Request) {
		handleCreateTask(w, r, teamService)
	})

	mux.HandleFunc("GET /api/teams/{id}/tasks", func(w http.ResponseWriter, r *http.Request) {
		handleListTasks(w, r, teamService)
	})

	mux.HandleFunc("GET /api/teams/{id}/tasks/{taskId}", func(w http.ResponseWriter, r *http.Request) {
		handleGetTask(w, r, teamService)
	})

	// Agent task polling endpoint
	mux.HandleFunc("GET /api/teams/{id}/agents/{agentId}/tasks", func(w http.ResponseWriter, r *http.Request) {
		handlePollTasks(w, r, teamService)
	})

	// Agent task result submission
	mux.HandleFunc("POST /api/teams/{id}/agents/{agentId}/tasks/{taskId}/result", func(w http.ResponseWriter, r *http.Request) {
		handleSubmitTaskResult(w, r, teamService)
	})
}

func handleJoinTeam(w http.ResponseWriter, r *http.Request, teamService *teams.Service) {
	var req struct {
		TeamID       string   `json:"team_id"`
		AgentID      string   `json:"agent_id"`
		Role         string   `json:"role"`
		TeamKey      string   `json:"team_key"`
		Capabilities []string `json:"capabilities,omitempty"`
		Version      string   `json:"version,omitempty"`
		Address      string   `json:"address,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	joinReq := teams.JoinTeamRequest{
		AgentID:      req.AgentID,
		Role:         req.Role,
		Capabilities: req.Capabilities,
		Version:      req.Version,
		Address:      req.Address,
	}

	resp, err := teamService.JoinTeam(req.TeamID, joinReq, req.TeamKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if resp.Success {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
	json.NewEncoder(w).Encode(resp)
}

func handleAgentHeartbeat(w http.ResponseWriter, r *http.Request, teamService *teams.Service) {
	var req struct {
		SessionID string `json:"session_id"`
		TeamID    string `json:"team_id"`
		AgentID   string `json:"agent_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate session
	_, err := teamService.GetAgentSession(req.SessionID)
	if err != nil {
		http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
		return
	}

	// Update heartbeat
	if err := teamService.UpdateAgentHeartbeat(req.TeamID, req.AgentID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Touch session to extend its lifetime
	_ = teamService.TouchSession(req.SessionID) // Ignore error - heartbeat succeeded

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleListTeams(w http.ResponseWriter, r *http.Request, teamService *teams.Service) {
	teamList, err := teamService.ListTeams()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teamList)
}

func handleCreateTeam(w http.ResponseWriter, r *http.Request, teamService *teams.Service) {
	var req teams.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}

	team, err := teamService.CreateTeam(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(team)
}

func handleGetTeam(w http.ResponseWriter, r *http.Request, teamService *teams.Service) {
	id := r.PathValue("id")
	team, err := teamService.GetTeam(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

func handleUpdateTeam(w http.ResponseWriter, r *http.Request, teamService *teams.Service) {
	id := r.PathValue("id")
	var req teams.UpdateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := teamService.UpdateTeam(id, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleDeleteTeam(w http.ResponseWriter, r *http.Request, teamService *teams.Service) {
	id := r.PathValue("id")
	if err := teamService.DeleteTeam(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleRotateKey(w http.ResponseWriter, r *http.Request, teamService *teams.Service) {
	id := r.PathValue("id")
	newKey, err := teamService.RotateTeamKey(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"team_key": newKey,
		"message":  "Team key rotated. Agents must reconnect with new key.",
	})
}

func handleListAgents(w http.ResponseWriter, r *http.Request, teamService *teams.Service) {
	id := r.PathValue("id")
	team, err := teamService.GetTeam(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Convert to AgentInfo for public API
	agentInfos := make([]*teams.AgentInfo, len(team.Agents))
	for i, agent := range team.Agents {
		agentInfos[i] = &teams.AgentInfo{
			ID:           agent.ID,
			Role:         agent.Role,
			Status:       agent.Status,
			Capabilities: agent.Capabilities,
			ActiveTasks:  agent.ActiveTasks,
			MaxTasks:     agent.MaxTasks,
			LastSeen:     agent.LastSeen,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agentInfos)
}

func handleEvictAgent(w http.ResponseWriter, r *http.Request, teamService *teams.Service) {
	teamID := r.PathValue("id")
	agentID := r.PathValue("agentId")

	// First, send a leave task to the agent so it can clean up
	_, err := teamService.CreateTask(teamID, teams.CreateTaskRequest{
		AgentID:     agentID,
		Type:        "system",
		Title:       "Leave Team",
		Description: "Controller requested agent to leave team",
		Priority:    10, // High priority
		Payload: map[string]any{
			"action": "leave",
			"team_id": teamID,
		},
		CreatedBy: "controller",
	})
	// Continue with eviction even if task creation fails (agent may already be offline)
	_ = err

	// Wait a moment for agent to receive and process the task
	time.Sleep(500 * time.Millisecond)

	if err := teamService.EvictAgent(teamID, agentID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Agent evicted successfully",
	})
}

func handleWipeAgent(w http.ResponseWriter, r *http.Request, teamService *teams.Service) {
	teamID := r.PathValue("id")
	agentID := r.PathValue("agentId")

	if err := teamService.WipeAgent(teamID, agentID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Agent wiped. Team key has been removed from agent.",
	})
}

// Task handlers

func handleCreateTask(w http.ResponseWriter, r *http.Request, teamService *teams.Service) {
	teamID := r.PathValue("id")

	var req teams.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	task, err := teamService.CreateTask(teamID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func handleListTasks(w http.ResponseWriter, r *http.Request, teamService *teams.Service) {
	teamID := r.PathValue("id")

	tasks := teamService.ListTasksForTeam(teamID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func handleGetTask(w http.ResponseWriter, r *http.Request, teamService *teams.Service) {
	taskID := r.PathValue("taskId")

	task, err := teamService.GetTask(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func handlePollTasks(w http.ResponseWriter, r *http.Request, teamService *teams.Service) {
	teamID := r.PathValue("id")
	agentID := r.PathValue("agentId")

	// Validate session from header
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		http.Error(w, "Session required", http.StatusUnauthorized)
		return
	}

	_, err := teamService.GetAgentSession(sessionID)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	tasks, err := teamService.PollTasksForAgent(teamID, agentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func handleSubmitTaskResult(w http.ResponseWriter, r *http.Request, teamService *teams.Service) {
	teamID := r.PathValue("id")
	agentID := r.PathValue("agentId")

	// Validate session from header
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		http.Error(w, "Session required", http.StatusUnauthorized)
		return
	}

	_, err := teamService.GetAgentSession(sessionID)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	var result teams.TaskResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := teamService.SubmitTaskResult(teamID, agentID, result); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
