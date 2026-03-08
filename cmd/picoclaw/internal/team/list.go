// PicoClaw - Team list command
// License: MIT

package team

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newListCommand(teamsDirFn func() (string, error)) *cobra.Command {
	var showKeys bool

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List known teams",
		Example: `picoclaw team list`,
		RunE: func(_ *cobra.Command, _ []string) error {
			teamsDir, err := teamsDirFn()
			if err != nil {
				return err
			}

			return listTeams(teamsDir, showKeys)
		},
	}

	cmd.Flags().BoolVarP(&showKeys, "show-keys", "k", false, "Show team keys (careful: sensitive information)")

	return cmd
}

func listTeams(teamsDir string, showKeys bool) error {
	entries, err := os.ReadDir(teamsDir)
	if err != nil {
		return fmt.Errorf("failed to read teams directory: %w", err)
	}

	var teams []teamStore
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(teamsDir, entry.Name()))
		if err != nil {
			continue
		}

		var team teamStore
		if err := json.Unmarshal(data, &team); err != nil {
			continue
		}

		teams = append(teams, team)
	}

	if len(teams) == 0 {
		fmt.Println("No teams configured. Use 'picoclaw team join' to join a team.")
		return nil
	}

	// Print header
	headerStyle := color.New(color.Bold, color.FgBlue)
	headerStyle.Println("Teams:")
	fmt.Println()

	// Print teams
	for _, team := range teams {
		printTeam(team, showKeys)
		fmt.Println()
	}

	return nil
}

func printTeam(team teamStore, showKeys bool) {
	// Status color
	statusColor := color.FgYellow
	switch team.Status {
	case "active":
		statusColor = color.FgGreen
	case "inactive":
		statusColor = color.FgYellow
	case "evicted":
		statusColor = color.FgRed
	}

	// Team ID (bold)
	color.New(color.Bold).Printf("  %s", team.TeamID)

	// Status
	statusStr := fmt.Sprintf("[%s]", team.Status)
	color.New(statusColor).Printf(" %s", statusStr)

	// Active indicator
	if team.Status == "active" {
		color.New(color.FgGreen).Print(" ●")
	}
	fmt.Println()

	// Details
	fmt.Printf("    Role:     %s\n", team.AgentRole)
	fmt.Printf("    Agent ID: %s\n", team.AgentID)

	if team.GatewayAddr != "" {
		fmt.Printf("    Gateway:  %s\n", team.GatewayAddr)
	}

	if showKeys && team.TeamKey != "" {
		fmt.Printf("    Team Key: %s\n", team.TeamKey)
	}

	if team.JoinedAt != "" {
		if t, err := time.Parse(time.RFC3339, team.JoinedAt); err == nil {
			fmt.Printf("    Joined:   %s\n", t.Format("2006-01-02 15:04"))
		}
	}
}
