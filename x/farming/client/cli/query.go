package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/tendermint/farming/x/farming/types"
)

// GetQueryCmd returns a root CLI command handler for all x/farming query commands.
func GetQueryCmd() *cobra.Command {
	farmingQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the farming module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	farmingQueryCmd.AddCommand(
	// TODO: add query commands
	// GetCmdQueryPlan(),
	)

	return farmingQueryCmd
}
