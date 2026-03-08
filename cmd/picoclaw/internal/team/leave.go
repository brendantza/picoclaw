// PicoClaw - Team leave command
// License: MIT

package team

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newLeaveCommand(teamsDirFn func() (string, error)) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "leave <team-id>",
		Short: "Leave a team",
		Long: `Leave a team and remove local team data.

This will remove the team key from this agent but does not notify
the team gateway. Use 'evict' if you want to notify the gateway.`,
		Example: `picoclaw team leave my-team`,
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			teamsDir, err := teamsDirFn()
			if err != nil {
				return err
			}

			teamID := args[0]
			return leaveTeam(teamsDir, teamID, force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force leave without confirmation")

	return cmd
}

func leaveTeam(teamsDir, teamID string, force bool) error {
	teamFile := getTeamFilePath(teamsDir, teamID)

	// Check if joined
	_, err := os.Stat(teamFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("not a member of team %s", teamID)
	}

	// Load team info for display
	var team teamStore
	data, err := os.ReadFile(teamFile)
	if err == nil {
		// Best effort to load team info for confirmation message
		_ = json.Unmarshal(data, &team)
	}

	// Confirm unless --force
	if !force {
		fmt.Printf("Leave team %s? This will remove local team data. [y/N] ", teamID)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Remove team file
	if err := os.Remove(teamFile); err != nil {
		return fmt.Errorf("failed to remove team data: %w", err)
	}

	color.Green("✓ Left team %s", teamID)

	// Note about notifying gateway
	if team.GatewayAddr != "" {
		fmt.Println()
		fmt.Println("Note: This only removes local team data.")
		fmt.Printf("To notify the gateway, an admin should run: picoclaw team evict %s %s\n", teamID, team.AgentID)
	}

	return nil
}
