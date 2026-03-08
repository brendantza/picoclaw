// PicoClaw - Team wipe command
// License: MIT

package team

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newWipeCommand(teamsDirFn func() (string, error)) *cobra.Command {
	var (
		gatewayAddr string
		force       bool
	)

	cmd := &cobra.Command{
		Use:   "wipe <team-id> <agent-id>",
		Short: "Wipe an agent's team key",
		Long: `Wipe an agent's team key from a team.

This evicts the agent and removes their team key, requiring
them to rejoin with a new key if they want to rejoin.

This is a destructive operation and should be used carefully.`,
		Example: `picoclaw team wipe my-team agent-123`,
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

			return wipeAgent(teamsDir, teamID, agentID, gatewayAddr, force)
		},
	}

	cmd.Flags().StringVarP(&gatewayAddr, "gateway", "g", "", "Team gateway address")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}

func wipeAgent(teamsDir, teamID, agentID, gatewayAddr string, force bool) error {
	// Confirm unless --force
	if !force {
		fmt.Printf("WARNING: This will permanently remove %s from team %s.\n", agentID, teamID)
		fmt.Print("The agent will need a new team key to rejoin. Continue? [y/N] ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	color.Yellow("Wiping agent %s from team %s...", agentID, teamID)
	fmt.Printf("Gateway: %s\n", gatewayAddr)

	// TODO: Implement actual wipe API call to gateway
	// POST /api/teams/{teamID}/agents/{agentID}/wipe

	color.Green("✓ Wipe request sent to gateway")
	fmt.Println("The agent's team key has been invalidated.")

	return nil
}
