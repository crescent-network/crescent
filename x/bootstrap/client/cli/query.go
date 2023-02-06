package cli

// DONTCOVER
// client is excluded from test coverage in MVP version

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/crescent-network/crescent/v4/x/bootstrap/types"
)

// GetQueryCmd returns a root CLI command handler for all x/bootstrap query commands.
func GetQueryCmd() *cobra.Command {
	mmQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the bootstrap module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	mmQueryCmd.AddCommand(
		GetCmdQueryParams(),
		GetQueryBootstrapsCmd(),
		GetCmdQueryIncentive(),
	)
	return mmQueryCmd
}

// GetCmdQueryParams implements the query params command.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current market maker parameters information",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query values set as market maker parameters.

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

			resp, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&resp.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryBootstrapsCmd implements the market maker query command.
func GetQueryBootstrapsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap",
		Args:  cobra.MaximumNArgs(0),
		Short: "Query details of the market makers",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details of the market makers

Example:
$ %s query %s bootstrap --pair-id=1
$ %s query %s bootstrap --address=...
$ %s query %s bootstrap --eligible=true...
`,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			//pairIdStr, _ := cmd.Flags().GetString(FlagPairId)
			//mmAddr, _ := cmd.Flags().GetString(FlagAddress)
			//eligibleStr, _ := cmd.Flags().GetString(FlagEligible)
			//
			//queryClient := types.NewQueryClient(clientCtx)
			//pageReq, err := client.ReadPageRequest(cmd.Flags())
			//if err != nil {
			//	return err
			//}
			//
			//req := &types.QueryBootstrapsRequest{
			//	Pagination: pageReq,
			//}
			//
			//switch {
			//case pairIdStr != "":
			//	pairId, err := strconv.ParseUint(pairIdStr, 10, 64)
			//	if err != nil {
			//		return fmt.Errorf("parse pair id: %w", err)
			//	}
			//	req.PairId = pairId
			//case mmAddr != "":
			//	req.Address = mmAddr
			//case eligibleStr != "":
			//	if _, err := strconv.ParseBool(eligibleStr); err != nil {
			//		return fmt.Errorf("parse eligible flag: %w", err)
			//	}
			//	req.Eligible = eligibleStr
			//}
			//
			//res, err := queryClient.Bootstraps(cmd.Context(), req)
			//if err != nil {
			//	return err
			//}

			return clientCtx.PrintProto(nil)
		},
	}

	cmd.Flags().AddFlagSet(flagSetBootstraps())
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "bootstraps")

	return cmd
}

// GetCmdQueryIncentive implements the query market maker claimable incentive command.
func GetCmdQueryIncentive() *cobra.Command {
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()

	cmd := &cobra.Command{
		Use:   "incentive [mm-address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query claimable incentive of a market maker",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query claimable incentive of a market maker.

Example:
$ %s query %s incentive %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName, types.ModuleName, bech32PrefixAccAddr,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			//queryClient := types.NewQueryClient(clientCtx)
			//
			//mmAddr, err := sdk.AccAddressFromBech32(args[0])
			//if err != nil {
			//	return err
			//}
			//
			//resp, err := queryClient.Incentive(cmd.Context(), &types.QueryIncentiveRequest{
			//	Address: mmAddr.String(),
			//})
			//if err != nil {
			//	return err
			//}

			return clientCtx.PrintProto(nil)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}