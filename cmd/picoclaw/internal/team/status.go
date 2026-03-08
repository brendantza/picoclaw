// PicoClaw - Team status command
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

func newStatusCommand(teamsDirFn func() (string, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [team-id]",
		Short: "Show team status",
		Long: `Show detailed status of a team membership.

If no team ID is provided, shows status of the active team.`,
		Example: `picoclaw team status
picoclaw team status my-team`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			teamsDir, err := teamsDirFn()
			if err != nil {
				return err
			}

			var teamID string
			if len(args) > 0 {
				teamID = args[0]
			} else {
				// Find active team
				teamID = getActiveTeam(teamsDir)
				if teamID == "" {
					return fmt.Errorf("no active team (provide team-id or join a team first)")
				}
			}

			return showTeamStatus(teamsDir, teamID)
		},
	}

	return cmd
}

func showTeamStatus(teamsDir, teamID string) error {
	teamFile := getTeamFilePath(teamsDir, teamID)

	data, err := os.ReadFile(teamFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("not a member of team %s", teamID)
		}
		return fmt.Errorf("failed to read team data: %w", err)
	}

	var team teamStore
	if err := json.Unmarshal(data, &team); err != nil {
		return fmt.Errorf("failed to parse team data: %w", err)
	}

	// Print status
	color.New(color.Bold, color.FgBlue).Printf("Team: %s\n", team.TeamID)
	fmt.Println()

	// Status with color
	statusColor := color.FgYellow
	switch team.Status {
	case "active":
		statusColor = color.FgGreen
	case "inactive":
		statusColor = color.FgYellow
	case "evicted":
		statusColor = color.FgRed
	}

	fmt.Printf("  Status:       ")
	color.New(statusColor).Printf("%s\n", team.Status)

	fmt.Printf("  Agent ID:     %s\n", team.AgentID)
	fmt.Printf("  Role:         %s\n", team.AgentRole)

	if team.GatewayAddr != "" {
		fmt.Printf("  Gateway:      %s\n", team.GatewayAddr)
	}

	if team.SessionID != "" {
		fmt.Printf("  Session ID:   %s\n", team.SessionID)
	}

	if team.JoinedAt != "" {
		fmt.Printf("  Joined:       %s\n", team.JoinedAt)
	}

	fmt.Println()

	// Connection hint
	if team.Status == "active" {
		color.Green("✓ You are an active member of this team")
	} else if team.Status == "evicted" {
		color.Red("✗ You have been evicted from this team")
		fmt.Println("  Run 'picoclaw team leave' to remove this team from local storage.")
	}

	return nil
}

func getActiveTeam(teamsDir string) string {
	entries, err := os.ReadDir(teamsDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		teamID := entry.Name()[:len(entry.Name())-5]
		data, err := os.ReadFile(getTeamFilePath(teamsDir, teamID))
		if err != nil {
			continue
		}

		var team teamStore
		if err := json.Unmarshal(data, &team); err != nil {
			continue
		}

		if team.Status == "active" {
			return team.TeamID
		}
	}

	return ""
}
