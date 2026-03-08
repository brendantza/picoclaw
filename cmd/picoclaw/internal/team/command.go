// PicoClaw - Team management commands for agents
// License: MIT

package team

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/sipeed/picoclaw/cmd/picoclaw/internal"
)

const teamsDirName = "agent_teams"

type deps struct {
	teamsDir string
}

func NewTeamCommand() *cobra.Command {
	var d deps

	cmd := &cobra.Command{
		Use:   "team",
		Short: "Manage team membership",
		Long: `Manage team membership and interact with team services.

Teams allow multiple agents to collaborate on tasks together.
Use these commands to join, leave, and manage team membership.`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			home := internal.GetPicoclawHome()
			d.teamsDir = filepath.Join(home, teamsDirName)

			// Ensure teams directory exists
			if err := os.MkdirAll(d.teamsDir, 0700); err != nil {
				return fmt.Errorf("failed to create teams directory: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	teamsDirFn := func() (string, error) {
		if d.teamsDir == "" {
			return "", fmt.Errorf("teams directory is not initialized")
		}
		return d.teamsDir, nil
	}

	cmd.AddCommand(
		newListCommand(teamsDirFn),
		newJoinCommand(teamsDirFn),
		newLeaveCommand(teamsDirFn),
		newStatusCommand(teamsDirFn),
		newActivateCommand(teamsDirFn),
		newEvictCommand(teamsDirFn),
		newWipeCommand(teamsDirFn),
		newDiscoverCommand(),
	)

	return cmd
}

// getTeamFilePath returns the path to a team's stored data file
func getTeamFilePath(teamsDir, teamID string) string {
	return filepath.Join(teamsDir, teamID+".json")
}

// teamStore represents locally stored team information
type teamStore struct {
	TeamID      string            `json:"team_id"`
	TeamKey     string            `json:"team_key"`
	GatewayAddr string            `json:"gateway_addr"`
	AgentRole   string            `json:"agent_role"`
	AgentID     string            `json:"agent_id"`
	SessionID   string            `json:"session_id,omitempty"`
	Status      string            `json:"status"` // active, inactive, evicted
	JoinedAt    string            `json:"joined_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}
