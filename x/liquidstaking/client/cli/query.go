package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
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
		GetCmdQueryStates(),
		GetCmdQueryVotingPower(),
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
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Params(
				cmd.Context(),
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
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			if err != nil {
				return err
			}

			res, err := queryClient.LiquidValidators(
				cmd.Context(),
				&types.QueryLiquidValidatorsRequest{},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryStates implements the query states command.
func GetCmdQueryStates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "states",
		Args:  cobra.NoArgs,
		Short: "Query states",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries states about net amount, mint rate.

Example:
$ %s query %s states
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.States(
				cmd.Context(),
				&types.QueryStatesRequest{},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryVotingPower implements the query voting power command.
func GetCmdQueryVotingPower() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "voting-power [voter]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the voter's voting power",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the voter's staking and liquid staking voting power.

Example:
$ %s query %s voting-power %s1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v
`,
				version.AppName, types.ModuleName, sdk.Bech32MainPrefix,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			voter, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.VotingPower(
				cmd.Context(),
				&types.QueryVotingPowerRequest{Voter: voter.String()},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
