// PicoClaw - Team join command
// License: MIT

package team

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func newJoinCommand(teamsDirFn func() (string, error)) *cobra.Command {
	var (
		teamKey     string
		gatewayAddr string
		role        string
		agentID     string
	)

	cmd := &cobra.Command{
		Use:   "join <team-id>",
		Short: "Join a team",
		Long: `Join a team using a team key.

The team key should be provided by the team administrator.
You can also set the team key via PICOCLAW_TEAM_KEY environment variable.`,
		Example: `picoclaw team join my-team --key pk_team_xxx --gateway http://localhost:18800`,
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			teamsDir, err := teamsDirFn()
			if err != nil {
				return err
			}

			teamID := args[0]

			// Get team key from flag or environment
			if teamKey == "" {
				teamKey = os.Getenv("PICOCLAW_TEAM_KEY")
			}
			if teamKey == "" {
				return fmt.Errorf("team key is required (use --key flag or PICOCLAW_TEAM_KEY environment variable)")
			}

			// Get gateway from flag or environment
			if gatewayAddr == "" {
				gatewayAddr = os.Getenv("PICOCLAW_GATEWAY_ADDRESS")
			}
			if gatewayAddr == "" {
				gatewayAddr = "http://localhost:18800" // default
			}

			// Generate agent ID if not provided
			if agentID == "" {
				agentID = generateAgentID()
			}

			// Default role
			if role == "" {
				role = os.Getenv("PICOCLAW_AGENT_ROLE")
			}
			if role == "" {
				role = "worker" // default role
			}

			return joinTeam(teamsDir, teamID, teamKey, gatewayAddr, role, agentID)
		},
	}

	cmd.Flags().StringVarP(&teamKey, "key", "k", "", "Team key (or PICOCLAW_TEAM_KEY env var)")
	cmd.Flags().StringVarP(&gatewayAddr, "gateway", "g", "", "Gateway address (or PICOCLAW_GATEWAY_ADDRESS env var)")
	cmd.Flags().StringVarP(&role, "role", "r", "", "Agent role (or PICOCLAW_AGENT_ROLE env var)")
	cmd.Flags().StringVarP(&agentID, "agent-id", "a", "", "Agent ID (auto-generated if not provided)")

	return cmd
}

func joinTeam(teamsDir, teamID, teamKey, gatewayAddr, role, agentID string) error {
	// Check if already joined
	teamFile := getTeamFilePath(teamsDir, teamID)
	if _, err := os.Stat(teamFile); err == nil {
		return fmt.Errorf("already joined team %s (use 'leave' first if you want to rejoin)", teamID)
	}

	// Prepare join request
	joinReq := map[string]any{
		"team_id":       teamID,
		"agent_id":      agentID,
		"role":          role,
		"team_key":      teamKey,
		"capabilities":  []string{}, // TODO: detect capabilities
		"version":       getVersion(),
		"address":       getLocalAddress(),
	}

	reqBody, err := json.Marshal(joinReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send join request
	joinURL := gatewayAddr + "/api/teams/join"
	resp, err := http.Post(joinURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}
	defer resp.Body.Close()

	var joinResp struct {
		Success          bool   `json:"success"`
		Message          string `json:"message"`
		SessionID        string `json:"session_id"`
		RequiresApproval bool   `json:"requires_approval"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&joinResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !joinResp.Success {
		if joinResp.RequiresApproval {
			color.Yellow("Join request sent. Waiting for team administrator approval.")
			return nil
		}
		return fmt.Errorf("failed to join team: %s", joinResp.Message)
	}

	// Store team info locally
	team := teamStore{
		TeamID:      teamID,
		TeamKey:     teamKey,
		GatewayAddr: gatewayAddr,
		AgentRole:   role,
		AgentID:     agentID,
		SessionID:   joinResp.SessionID,
		Status:      "active",
		JoinedAt:    time.Now().UTC().Format(time.RFC3339),
	}

	if err := saveTeam(teamsDir, team); err != nil {
		return fmt.Errorf("failed to save team info: %w", err)
	}

	color.Green("✓ Successfully joined team %s as %s", teamID, role)
	fmt.Printf("  Session ID: %s\n", joinResp.SessionID)

	return nil
}

func generateAgentID() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "agent"
	}
	return fmt.Sprintf("%s-%s", hostname, uuid.New().String()[:8])
}

func getVersion() string {
	return "0.1.0" // TODO: get from build info
}

func getLocalAddress() string {
	// TODO: get actual local address
	return ""
}

func saveTeam(teamsDir string, team teamStore) error {
	teamFile := getTeamFilePath(teamsDir, team.TeamID)
	data, err := json.MarshalIndent(team, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(teamFile, data, 0600)
}
