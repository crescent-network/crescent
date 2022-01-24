package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

// GetQueryCmd returns a root CLI command handler for all x/liquidstaking query commands.
func GetQueryCmd() *cobra.Command {
	liquidValidatorQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the liquidstaking module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	liquidValidatorQueryCmd.AddCommand(
		GetCmdQueryParams(),
		GetCmdQueryLiquidValidators(),
	)

	return liquidValidatorQueryCmd
}

// GetCmdQueryParams implements the params query command.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the values set as liquidstaking parameters",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query values set as liquidstaking parameters.

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
			res, err := queryClient.Params(
				context.Background(),
				&types.QueryParamsRequest{},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryLiquidValidators implements the query liquidValidators command.
func GetCmdQueryLiquidValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquid-validators",
		Args:  cobra.NoArgs,
		Short: "Query all liquid validators",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries all liquid validators.

Example:
$ %s query %s liquid-validators
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			//name, _ := cmd.Flags().GetString(FlagName)
			//sourceAddr, _ := cmd.Flags().GetString(FlagSourceAddress)
			//destinationAddr, _ := cmd.Flags().GetString(FlagDestinationAddress)

			queryClient := types.NewQueryClient(clientCtx)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.LiquidValidators(
				context.Background(),
				&types.QueryLiquidValidatorsRequest{
					// TODO: add status
					//Name:               name,
					//SourceAddress:      sourceAddr,
					//DestinationAddress: destinationAddr,
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	//cmd.Flags().AddFlagSet(flagSetLiquidValidators())
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
