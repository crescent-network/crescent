package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		QueryParams(),
		// QueryPools(),
		// QueryPool(),
		// QueryPoolsByPair(),
		// QueryPoolsByReserveAcc(),
		// QueryPoolsByPoolCoinDenom(),
		// QueryPairs(),
		// QueryPair(),
		// QueryDepositRequests(),
		// QueryDepositRequest(),
		// QueryWithdrawRequests(),
		// QueryWithdrawRequest(),
		// QuerySwapRequests(),
		// QuerySwapRequest(),
	)

	return cmd
}

func QueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current liquidity parameters information",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query values set as liquidity parameters.
Example:
$ %s query %s params
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			resp, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&resp.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
