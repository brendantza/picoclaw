// PicoClaw - Team activate command
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

func newActivateCommand(teamsDirFn func() (string, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate <team-id>",
		Short: "Activate a team",
		Long: `Activate a team membership.

This marks the team as the active team for operations.
Only one team can be active at a time.`,
		Example: `picoclaw team activate my-team`,
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			teamsDir, err := teamsDirFn()
			if err != nil {
				return err
			}

			teamID := args[0]
			return activateTeam(teamsDir, teamID)
		},
	}

	return cmd
}

func activateTeam(teamsDir, teamID string) error {
	teamFile := getTeamFilePath(teamsDir, teamID)

	// Check if team exists
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

	// Deactivate all other teams
	entries, err := os.ReadDir(teamsDir)
	if err != nil {
		return fmt.Errorf("failed to read teams directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		otherID := entry.Name()[:len(entry.Name())-5]
		if otherID == teamID {
			continue
		}

		otherFile := getTeamFilePath(teamsDir, otherID)
		otherData, err := os.ReadFile(otherFile)
		if err != nil {
			continue
		}

		var otherTeam teamStore
		if err := json.Unmarshal(otherData, &otherTeam); err != nil {
			continue
		}

		if otherTeam.Status == "active" {
			otherTeam.Status = "inactive"
			otherData, _ := json.MarshalIndent(otherTeam, "", "  ")
			_ = os.WriteFile(otherFile, otherData, 0600)
		}
	}

	// Activate this team
	if team.Status != "active" {
		team.Status = "active"
		data, _ := json.MarshalIndent(team, "", "  ")
		if err := os.WriteFile(teamFile, data, 0600); err != nil {
			return fmt.Errorf("failed to update team status: %w", err)
		}
	}

	color.Green("✓ Activated team %s", teamID)

	return nil
}
