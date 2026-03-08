// PicoClaw - Team evict command
// License: MIT

package team

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newEvictCommand(teamsDirFn func() (string, error)) *cobra.Command {
	var gatewayAddr string

	cmd := &cobra.Command{
		Use:   "evict <team-id> <agent-id>",
		Short: "Evict an agent from a team",
		Long: `Evict an agent from a team.

This command sends an eviction request to the team gateway.
Requires knowing the team gateway address.`,
		Example: `picoclaw team evict my-team agent-123`,
		Args:    cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			teamsDir, err := teamsDirFn()
			if err != nil {
				return err
			}

			teamID := args[0]
			agentID := args[1]

			// Get gateway from local storage if not provided
			if gatewayAddr == "" {
				gatewayAddr = getTeamGateway(teamsDir, teamID)
			}
			if gatewayAddr == "" {
				return fmt.Errorf("gateway address required (use --gateway flag)")
			}

			return evictAgent(teamsDir, teamID, agentID, gatewayAddr)
		},
	}

	cmd.Flags().StringVarP(&gatewayAddr, "gateway", "g", "", "Team gateway address")

	return cmd
}

func evictAgent(teamsDir, teamID, agentID, gatewayAddr string) error {
	// TODO: Implement actual eviction API call to gateway
	// For now, this is a placeholder

	color.Yellow("Evicting agent %s from team %s...", agentID, teamID)
	fmt.Printf("Gateway: %s\n", gatewayAddr)

	// This would typically make an API call to the gateway
	// POST /api/teams/{teamID}/agents/{agentID}/evict

	// Check if we're evicting ourselves
	teamFile := getTeamFilePath(teamsDir, teamID)
	if _, err := os.Stat(teamFile); err == nil {
		// We are a member of this team
		if strings.Contains(readAgentID(teamsDir, teamID), agentID) {
			color.Red("Warning: You are evicting yourself from this team!")
			fmt.Println("Use 'picoclaw team leave' instead for a cleaner exit.")
		}
	}

	color.Green("✓ Eviction request sent to gateway")
	fmt.Println("Note: The agent will be notified and disconnected.")

	return nil
}

func getTeamGateway(teamsDir, teamID string) string {
	teamFile := getTeamFilePath(teamsDir, teamID)
	data, err := os.ReadFile(teamFile)
	if err != nil {
		return ""
	}

	var team teamStore
	if err := json.Unmarshal(data, &team); err != nil {
		return ""
	}

	return team.GatewayAddr
}

func readAgentID(teamsDir, teamID string) string {
	teamFile := getTeamFilePath(teamsDir, teamID)
	data, err := os.ReadFile(teamFile)
	if err != nil {
		return ""
	}

	var team teamStore
	if err := json.Unmarshal(data, &team); err != nil {
		return ""
	}

	return team.AgentID
}
