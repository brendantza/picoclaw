// PicoClaw - Team management API handlers for launcher
// License: MIT

package server

import (
	"encoding/json"
	"net/http"

	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/teams"
)

// RegisterTeamsAPI registers team management API endpoints
func RegisterTeamsAPI(mux *http.ServeMux, teamService *teams.Service) {
	// GET /api/teams - List all teams
	mux.HandleFunc("GET /api/teams", func(w http.ResponseWriter, r *http.Request) {
		// Reload to pick up changes from gateway
		_ = teamService.Reload()

		teamList, err := teamService.ListTeams()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(teamList)
	})

	// POST /api/teams - Create a new team
	mux.HandleFunc("POST /api/teams", func(w http.ResponseWriter, r *http.Request) {
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
	})

	// GET /api/teams/{id} - Get team details
	mux.HandleFunc("GET /api/teams/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Reload to pick up changes from gateway
		_ = teamService.Reload()

		id := r.PathValue("id")
		team, err := teamService.GetTeam(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(team)
	})

	// PUT /api/teams/{id} - Update team
	mux.HandleFunc("PUT /api/teams/{id}", func(w http.ResponseWriter, r *http.Request) {
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
	})

	// DELETE /api/teams/{id} - Delete team
	mux.HandleFunc("DELETE /api/teams/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := teamService.DeleteTeam(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// POST /api/teams/{id}/rotate-key - Rotate team key
	mux.HandleFunc("POST /api/teams/{id}/rotate-key", func(w http.ResponseWriter, r *http.Request) {
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
	})

	// POST /api/teams/{id}/agents/{agentId}/evict - Evict agent
	mux.HandleFunc("POST /api/teams/{id}/agents/{agentId}/evict", func(w http.ResponseWriter, r *http.Request) {
		teamID := r.PathValue("id")
		agentID := r.PathValue("agentId")

		if err := teamService.EvictAgent(teamID, agentID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.InfoCF("teams-api", "Agent evicted via API",
			map[string]any{
				"team_id":  teamID,
				"agent_id": agentID,
			})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "Agent evicted successfully",
		})
	})

	// POST /api/teams/{id}/agents/{agentId}/wipe - Wipe agent
	mux.HandleFunc("POST /api/teams/{id}/agents/{agentId}/wipe", func(w http.ResponseWriter, r *http.Request) {
		teamID := r.PathValue("id")
		agentID := r.PathValue("agentId")

		if err := teamService.WipeAgent(teamID, agentID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.InfoCF("teams-api", "Agent wiped via API",
			map[string]any{
				"team_id":  teamID,
				"agent_id": agentID,
			})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "Agent wiped. Team key has been removed from agent.",
		})
	})

	// GET /api/teams/{id}/agents - List team agents
	mux.HandleFunc("GET /api/teams/{id}/agents", func(w http.ResponseWriter, r *http.Request) {
		// Reload to pick up changes from gateway
		_ = teamService.Reload()

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
	})

	// POST /api/teams/join - Agent join endpoint (called by agents)
	mux.HandleFunc("POST /api/teams/join", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			TeamID      string              `json:"team_id"`
			AgentID     string              `json:"agent_id"`
			Role        string              `json:"role"`
			TeamKey     string              `json:"team_key"`
			Capabilities []string           `json:"capabilities,omitempty"`
			Version     string              `json:"version,omitempty"`
			Address     string              `json:"address,omitempty"`
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
	})

	// POST /api/teams/heartbeat - Agent heartbeat endpoint
	mux.HandleFunc("POST /api/teams/heartbeat", func(w http.ResponseWriter, r *http.Request) {
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
		_ = teamService.TouchSession(req.SessionID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
}
