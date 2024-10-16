package command

import (
	"fmt"
	"github.com/spf13/cobra"
)

var QueryCmd = &cobra.Command{
	Use:                "query",
	Short:              "querying subcommands for station ",
	Run:                queryCommands,
	DisableFlagParsing: true,
}

func queryCommands(cmd *cobra.Command, args []string) {
	const message = `Query-related commands for station. These queries allow users to fetch and display various data related to stations on the Tracks, such as current schema configurations or engagement details. Useful for verifying network status, monitoring station activity, and gathering information necessary for participation in Tracks-based systems.`

	fmt.Println(message)

	// show help for the command
	if err := cmd.Help(); err != nil {
		fmt.Println("Failed to print command help:", err)
	}
}
