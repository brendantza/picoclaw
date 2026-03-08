// PicoClaw - Team discover command
// License: MIT

package team

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/sipeed/picoclaw/pkg/discovery"
)

func newDiscoverCommand() *cobra.Command {
	var timeout int

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover teams on the local network",
		Long: `Discover PicoClaw team gateways on the local network using mDNS.

This command searches for team gateways advertising themselves via mDNS.
The team hash (not the key) is advertised for security.`,
		Example: `picoclaw team discover
picoclaw team discover --timeout 5`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return discoverTeams(timeout)
		},
	}

	cmd.Flags().IntVarP(&timeout, "timeout", "t", 3, "Discovery timeout in seconds")

	return cmd
}

func discoverTeams(timeout int) error {
	fmt.Println("🔍 Searching for teams on the local network...")
	fmt.Println()

	client := discovery.NewClient()
	teams, err := client.Discover(timeout)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}

	if len(teams) == 0 {
		fmt.Println("No teams found on the local network.")
		fmt.Println()
		fmt.Println("Possible reasons:")
		fmt.Println("  • No team gateways are running on the network")
		fmt.Println("  • Team gateways are not advertising via mDNS")
		fmt.Println("  • Network/firewall is blocking mDNS traffic (port 5353)")
		return nil
	}

	fmt.Printf("Found %d team(s):\n", len(teams))
	fmt.Println()

	for i, team := range teams {
		color.New(color.Bold).Printf("  %d. Team Hash: ", i+1)
		color.Cyan(team.TeamHash)
		fmt.Println()
		fmt.Printf("     Gateway: %s\n", team.GetBestAddress())
		if team.Host != "" {
			fmt.Printf("     Host: %s\n", team.Host)
		}
		fmt.Println()
	}

	fmt.Println("To join a team, use:")
	fmt.Println("  picoclaw team join <team-id> --key <team-key> --gateway <gateway-url>")
	fmt.Println()
	fmt.Println("Note: You'll need the full team key (not just the hash) to join.")

	return nil
}
